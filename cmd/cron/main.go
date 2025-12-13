package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"github.com/jptaku/server/internal/config"
	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/pkg"
	"github.com/robfig/cron/v3"
	openai "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GeneratedSentence OpenAI 응답 파싱용
type GeneratedSentence struct {
	JP          string        `json:"jp"`
	KR          string        `json:"kr"`
	Romaji      string        `json:"romaji"`
	SubCategory int           `json:"sub_category"`
	Words       []model.Word  `json:"words"`
	Grammar     StringOrArray `json:"grammar"`
	Examples    StringOrArray `json:"examples"`
	Quiz        *model.Quiz   `json:"quiz"`
}

// StringOrArray GPT가 문자열 또는 배열로 반환할 수 있는 필드 처리
type StringOrArray []string

func (s *StringOrArray) UnmarshalJSON(data []byte) error {
	// 배열로 시도
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*s = arr
		return nil
	}

	// 문자열로 시도
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str != "" {
			*s = []string{str}
		} else {
			*s = []string{}
		}
		return nil
	}

	// 둘 다 실패하면 빈 배열
	*s = []string{}
	return nil
}

const TargetCount = 15 // 조합당 목표 문장 수

func main() {
	log.Println("Starting sentence pre-generation cron job...")

	// Load .env file
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")

	// 기존 config 패키지 사용
	cfg := config.Load()

	if cfg.OpenAI.APIKey == "" {
		log.Fatal("OPEN_AI_API_KEY is required")
	}

	// Initialize database (기존 config 패키지 사용)
	db, err := config.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 로그 레벨 설정
	db.Logger = logger.Default.LogMode(logger.Error)

	// Auto migrate (기존 model 패키지 사용)
	if err := db.AutoMigrate(&model.Sentence{}, &model.SentenceDetail{}); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	// Initialize OpenAI client
	openaiClient := openai.NewClient(cfg.OpenAI.APIKey)

	// S3 호환 클라이언트 생성 (config에서 설정 가져오기)
	s3Client := s3.New(s3.Options{
		Region:       "kr-standard",
		BaseEndpoint: aws.String(cfg.NCP_Storage.Endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.NCP_Storage.AccessKey, cfg.NCP_Storage.SecretKey, ""),
	})

	// Create generator
	gen := &Generator{
		db:          db,
		openai:      openaiClient,
		model:       cfg.OpenAI.Model,
		targetCount: TargetCount,
		voicevoxURL: cfg.VoiceVox.VoiceVoxURL,
		s3Client:    s3Client,
		s3Bucket:    cfg.NCP_Storage.BucketName,
	}

	log.Printf("VoiceVox URL: %s", cfg.VoiceVox.VoiceVoxURL)
	log.Printf("NCP Endpoint: %s", cfg.NCP_Storage.Endpoint)
	log.Printf("NCP Bucket: %s", cfg.NCP_Storage.BucketName)

	// Create cron scheduler
	c := cron.New(cron.WithSeconds())

	// Run every 5 minutes
	_, err = c.AddFunc("0 */1 * * * *", func() {
		log.Println("=== Starting generation job ===")
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
		defer cancel()

		if err := gen.Run(ctx); err != nil {
			log.Printf("Generation failed: %v", err)
		}
		log.Println("=== Generation job completed ===")
	})
	if err != nil {
		log.Fatalf("Failed to add cron job: %v", err)
	}

	c.Start()
	log.Println("Cron scheduler started (every 1 minutes)")

	// Run once immediately
	log.Println("Running initial generation...")
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	if err := gen.Run(ctx); err != nil {
		log.Printf("Initial generation failed: %v", err)
	}
	cancel()

	// Print status
	gen.PrintStatus()

	// Wait for shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	c.Stop()

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	log.Println("Cron job exited gracefully")
}

// ========== Generator ==========

type Generator struct {
	db          *gorm.DB
	openai      *openai.Client
	model       string
	targetCount int
	voicevoxURL string
	s3Client    *s3.Client
	s3Bucket    string
}

// ========== VoiceVox TTS ==========

func (g *Generator) generateTTS(ctx context.Context, sentenceID uint, sentenceKey string, text string) (string, error) {
	// 1. audio_query 생성
	queryURL := fmt.Sprintf("%s/audio_query?text=%s&speaker=1", g.voicevoxURL, url.QueryEscape(text))
	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, nil)
	if err != nil {
		return "", fmt.Errorf("create query request failed: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("audio_query failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("audio_query error: %s", string(body))
	}

	queryBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read query response failed: %w", err)
	}

	// 2. synthesis로 WAV 생성
	synthesisURL := fmt.Sprintf("%s/synthesis?speaker=1", g.voicevoxURL)
	req, err = http.NewRequestWithContext(ctx, "POST", synthesisURL, bytes.NewReader(queryBody))
	if err != nil {
		return "", fmt.Errorf("create synthesis request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("synthesis failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("synthesis error: %s", string(body))
	}

	// 3. WAV 데이터 읽기
	wavData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read wav data failed: %w", err)
	}

	// 4. Object Storage에 업로드
	// 파일명: {sentenceKey}_{날짜}_{sentenceID}.wav (예: 101_0_20241214_123.wav)
	today := time.Now().Format("20060102")
	fileName := fmt.Sprintf("%s_%s_%d.wav", sentenceKey, today, sentenceID)
	_, err = g.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(g.s3Bucket),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(wavData),
		ContentType: aws.String("audio/wav"),
	})
	if err != nil {
		return "", fmt.Errorf("upload to object storage failed: %w", err)
	}

	// 5. 객체 키 반환 (프록시를 통해 접근하므로 키만 저장)
	return fileName, nil
}

func (g *Generator) Run(ctx context.Context) error {
	// 부족한 조합 찾기
	deficient := g.findDeficientKeys()

	if len(deficient) == 0 {
		log.Println("All combinations have enough sentences!")
		return nil
	}

	log.Printf("Found %d deficient combinations", len(deficient))

	// 한 번에 하나의 조합만 처리
	key := deficient[0]
	sentenceKey := string(key)

	count, _ := g.countByKey(sentenceKey)
	needed := g.targetCount - int(count)
	if needed <= 0 {
		return nil
	}

	// 최대 5개씩 생성
	batchSize := 5
	if needed < batchSize {
		batchSize = needed
	}

	subCat := key.SubCategory()
	level := key.Level()

	log.Printf("Generating %d sentences for %s (%s, %s) - current: %d",
		batchSize, sentenceKey, subCat.Name(), level.Name(), count)

	return g.generate(ctx, int(subCat), int(level), subCat.Name(), batchSize)
}

func (g *Generator) findDeficientKeys() []pkg.SentenceKey {
	var deficient []pkg.SentenceKey

	for _, key := range pkg.AllSentenceKeys {
		count, _ := g.countByKey(string(key))
		if int(count) < g.targetCount {
			deficient = append(deficient, key)
		}
	}

	return deficient
}

func (g *Generator) countByKey(sentenceKey string) (int64, error) {
	var count int64
	err := g.db.Model(&model.Sentence{}).Where("sentence_key = ?", sentenceKey).Count(&count).Error
	return count, err
}

func (g *Generator) generate(ctx context.Context, subCategory, level int, categoryName string, count int) error {
	prompt := g.buildPrompt(subCategory, level, categoryName, count)

	resp, err := g.openai.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: g.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: g.getSystemPrompt()},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: 0.8,
		MaxTokens:   8000,
	})
	if err != nil {
		return fmt.Errorf("OpenAI API failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("empty response")
	}

	content := resp.Choices[0].Message.Content
	var generated []GeneratedSentence
	if err := json.Unmarshal([]byte(content), &generated); err != nil {
		content = extractJSON(content)
		if err := json.Unmarshal([]byte(content), &generated); err != nil {
			log.Printf("JSON parse failed: %s", content)
			return fmt.Errorf("parse failed: %w", err)
		}
	}

	// 저장
	sentenceKey := fmt.Sprintf("%d_%d", subCategory, level)
	savedCount := 0

	for _, gen := range generated {
		sentence := model.Sentence{
			SentenceKey: sentenceKey,
			JP:          gen.JP,
			KR:          gen.KR,
			Romaji:      gen.Romaji,
			Level:       level,
			SubCategory: subCategory,
		}

		if err := g.db.Create(&sentence).Error; err != nil {
			log.Printf("Failed to save sentence: %v", err)
			continue
		}

		detail := model.SentenceDetail{
			SentenceID: sentence.ID,
			Words:      gen.Words,
			Grammar:    []string(gen.Grammar),
			Examples:   []string(gen.Examples),
			Quiz:       gen.Quiz,
		}

		if err := g.db.Create(&detail).Error; err != nil {
			log.Printf("Failed to save detail: %v", err)
		}

		// Step 2: VoiceVox TTS 생성
		audioURL, err := g.generateTTS(ctx, sentence.ID, sentenceKey, gen.JP)
		if err != nil {
			log.Printf("Failed to generate TTS for sentence %d: %v", sentence.ID, err)
			// TTS 실패해도 문장은 저장됨, 나중에 재시도 가능
		} else {
			// audio_url DB 업데이트
			if err := g.db.Model(&sentence).Update("audio_url", audioURL).Error; err != nil {
				log.Printf("Failed to update audio_url for sentence %d: %v", sentence.ID, err)
			} else {
				log.Printf("Generated TTS for sentence %d: %s", sentence.ID, audioURL)
			}
		}

		savedCount++
	}

	log.Printf("Saved %d sentences for %s", savedCount, sentenceKey)
	return nil
}

func (g *Generator) PrintStatus() {
	log.Println("=== Current Pool Status ===")

	total := 0
	deficient := 0

	for _, key := range pkg.AllSentenceKeys {
		sentenceKey := string(key)
		count, _ := g.countByKey(sentenceKey)
		total += int(count)
		if int(count) < g.targetCount {
			deficient++
			subCat := key.SubCategory()
			level := key.Level()
			log.Printf("  [NEED] %s (%s, %s): %d/%d", sentenceKey, subCat.Name(), level.Name(), count, g.targetCount)
		}
	}

	log.Printf("Total: %d sentences, %d combinations need more", total, deficient)
}

func (g *Generator) buildPrompt(subCategory, level int, categoryName string, count int) string {
	levelDesc := getLevelDescription(level)

	return fmt.Sprintf(`다음 조건에 맞는 일본어 학습 문장을 JSON으로 생성해 주세요.

[조건]
- 카테고리: %s (코드: %d)
- 레벨: %s
- 생성 개수: %d개

[요구 사항]
- 각 문장은 jp, kr, romaji, sub_category(%d 고정), words, grammar, examples, quiz 포함
- 오타쿠 문화에서 실제로 쓰일 법한 자연스러운 표현
- JSON 배열만 출력`, categoryName, subCategory, levelDesc, count, subCategory)
}

func (g *Generator) getSystemPrompt() string {
	return `당신은 일본어 학습용 문장 + 문장 분석 + 퀴즈를 생성하는 전문가입니다.
오타쿠(애니, 게임, 버튜버, 굿즈 등)에 관심 있는 한국인 학습자를 위한 실용적인 일본어 문장을 생성하고, 각 문장에 대한 단어 풀이 / 문법 / 예문 / 퀴즈까지 한 번에 만들어주세요.

[아주 중요]
- 반드시 **JSON 배열** 형식으로만 응답하세요.
- JSON 이외의 텍스트(설명, 코멘트, 마크다운, 코드블록 등)는 절대 포함하지 마세요.
- 모든 키는 아래에서 정의하는 영문 소문자 snake_case만 사용하세요.

[생성해야 하는 데이터 구조]

응답 전체는 다음과 같은 객체의 배열입니다.

[
  {
    "jp": string,              // 일본어 문장 (최종 학습 문장)
    "kr": string,              // 한국어 번역
    "romaji": string,          // 로마자 표기
    "sub_category": number,    // 단일 SubCategory 코드 (예: 101)

    "words": [                 // 단어 풀이
      {
        "japanese": string,    // 단어 원형 (일본어)
        "reading": string,     // 읽는 법(히라가나/가타카나)
        "meaning": string,     // 한국어 뜻
        "part_of": string      // 품사 (명사, 동사, 형용사, 부사, 조사 등)
      }
    ],

    "grammar": [               // 핵심 문법
      string                   // 예: "~ている: 진행/상태를 나타내는 표현"
    ],

    "examples": [              // 예문
      string                   // 일본어 예문 (가능하면 같은 문법/단어를 활용)
    ],

    "quiz": {                  // 확인하기 퀴즈
      "fill_blank": {          // 빈칸 채우기
        "question_jp": string, // 빈칸(____)이 포함된 일본어 문장
        "options": [string],   // 보기 3~4개
        "answer": string       // 정답 (options 중 하나와 동일한 문자열)
      },
      "ordering": {            // 문장 배열하기
        "fragments": [string], // 문장을 3~5조각으로 나눈 배열 (순서 섞어서 제공)
        "correct_order": [number] // 정답 순서를 나타내는 인덱스 배열 (0부터 시작)
      }
    }
  }
]

[카테고리 규칙]
- "sub_category"에는 아래 SubCategory 코드 중에서 **1개만** 선택해 넣으세요.
- 문장과 가장 관련 있는 코드를 고르세요.

- Anime: 101(배틀/판타지·SF), 102(일상/러브코미·감성), 103(서사/추리)
- Game: 201(RPG/가챠), 202(리듬게임), 203(액션/대전·슈터)
- Music: 301(J-POP), 302(아이돌), 303(애니송)
- VTuber: 351(버튜버)
- Lifestyle: 401(성지순례/여행), 402(굿즈/수집), 403(코미케/동인)
- Situation: 501(쇼핑/주문), 502(현장/라이브), 503(오타쿠 대화), 504(콜라보카페/게임센터)

[레벨/어휘/문법 가이드라인]
- 사용자의 일본어 레벨에 맞추어 난이도를 조절하세요.
- Lv0: 가장 짧고 쉬운 표현, 기본 단어만
- Lv1: 짧고 쉬운 표현, 기본 인사, N5 수준
- Lv2: 일상 회화, N4 수준, て형/ない형 등 기본 활용
- Lv3: 자신의 생각 표현, N3 수준, 조금 길어도 됨

[퀴즈 생성 규칙]

1) 빈칸 채우기 (fill_blank)
- 원문 "jp" 문장을 살짝 변형해서, 핵심 단어나 표현 한 곳을 "____" 로 비우세요.
- "options"에는 정답 포함 3~4개 정도 단어/구를 넣으세요.
- "answer"는 options 중 정답과 **완전히 같은 문자열**이어야 합니다.

2) 문장 배열 (ordering)
- "jp" 문장을 3~5개의 의미 있는 조각으로 나누세요.
- "fragments" 에는 섞인 상태의 조각들을 넣습니다.
- "correct_order"에는 올바른 순서를 인덱스로 표현합니다. (0부터 시작)
  예: fragments가 ["ゲームをしました。", "昨日", "友達と"] 이고,
      올바른 문장이 "昨日 友達と ゲームをしました。" 라면,
      correct_order는 [1, 2, 0] 입니다.

[스타일 규칙]
- 오타쿠 문화(애니, 게임, 버튜버, 굿즈 등)에서 실제로 쓰일 법한 자연스러운 표현을 사용하세요.
- 학습자가 실제로 쓸 수 있는 실용적인 문장으로 만드세요.
- 문장은 너무 길지 않게, 레벨에 맞는 길이와 단어를 사용하세요.
- 로마자(romaji)는 일반적인 표기법으로 적어주세요.

[출력 시 주의사항]
- JSON 배열 이외에는 어떤 텍스트도 절대 출력하지 마세요.
- 들여쓰기는 해도 되고 안 해도 되지만, JSON이 파싱 가능해야 합니다.`
}

func getLevelDescription(level int) string {
	switch level {
	case 0:
		return "Lv0 - 완전 초입문"
	case 1:
		return "Lv1 - N5 수준"
	case 2:
		return "Lv2 - N4 수준"
	case 3:
		return "Lv3 - N3 수준"
	default:
		return "Lv1 - N5 수준"
	}
}

func extractJSON(content string) string {
	start, end := 0, len(content)
	for i := 0; i < len(content); i++ {
		if content[i] == '[' {
			start = i
			break
		}
	}
	for i := len(content) - 1; i >= 0; i-- {
		if content[i] == ']' {
			end = i + 1
			break
		}
	}
	if start < end {
		return content[start:end]
	}
	return content
}
