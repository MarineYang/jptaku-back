package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/repository"
	openai "github.com/sashabaranov/go-openai"
)

type SentenceService struct {
	sentenceRepo *repository.SentenceRepository
	userRepo     *repository.UserRepository
	openaiClient *openai.Client
	openaiModel  string
}

func NewSentenceService(sentenceRepo *repository.SentenceRepository, userRepo *repository.UserRepository) *SentenceService {
	return &SentenceService{
		sentenceRepo: sentenceRepo,
		userRepo:     userRepo,
	}
}

func (s *SentenceService) SetOpenAIClient(apiKey, model string) {
	if apiKey == "" {
		return
	}
	s.openaiClient = openai.NewClient(apiKey)
	if model == "" {
		model = openai.GPT4oMini
	}
	s.openaiModel = model
	log.Println("SentenceService: OpenAI client initialized")
}

func (s *SentenceService) IsGeneratorEnabled() bool {
	return s.openaiClient != nil
}

type DailySentencesResponse struct {
	Date      string           `json:"date"`
	Sentences []model.Sentence `json:"sentences"`
}

// GetTodaySentences 오늘의 5문장 조회 (없으면 생성)
func (s *SentenceService) GetTodaySentences(userID uint) (*DailySentencesResponse, error) {
	today := time.Now().Truncate(24 * time.Hour)
	return s.getSentencesByDate(userID, today)
}

// GetYesterdaySentences 어제의 5문장 조회
func (s *SentenceService) GetYesterdaySentences(userID uint) (*DailySentencesResponse, error) {
	yesterday := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	return s.getSentencesByDate(userID, yesterday)
}

func (s *SentenceService) getSentencesByDate(userID uint, date time.Time) (*DailySentencesResponse, error) {
	// 해당 날짜의 세트가 있는지 확인
	dailySet, err := s.sentenceRepo.GetDailySet(userID, date)
	if err == nil && dailySet != nil {
		sentences, err := s.sentenceRepo.FindByIDs(dailySet.SentenceIDs)
		if err != nil {
			return nil, err
		}
		return &DailySentencesResponse{
			Date:      date.Format("2006-01-02"),
			Sentences: sentences,
		}, nil
	}

	// 오늘이 아니면 생성하지 않음
	today := time.Now().Truncate(24 * time.Hour)
	if !date.Equal(today) {
		return nil, fmt.Errorf("해당 날짜의 문장이 없습니다")
	}

	// 오늘 세트가 없으면 새로 생성
	return s.createDailySet(userID, date)
}

func (s *SentenceService) createDailySet(userID uint, date time.Time) (*DailySentencesResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	level := 1
	var interests, purposes []int
	if user.Onboarding != nil {
		level = user.Onboarding.Level
		interests = user.Onboarding.Interests
		purposes = user.Onboarding.Purposes
	}

	if !s.IsGeneratorEnabled() {
		return nil, fmt.Errorf("문장 생성 기능이 비활성화되어 있습니다")
	}

	ctx := context.Background()
	sentences, err := s.generateSentences(ctx, level, interests, purposes, 5)
	if err != nil {
		return nil, fmt.Errorf("문장 생성 실패: %w", err)
	}

	// 문장 ID 추출
	sentenceIDs := make([]uint, len(sentences))
	for i, sentence := range sentences {
		sentenceIDs[i] = sentence.ID
	}

	// DailySentenceSet 저장
	dailySet := &model.DailySentenceSet{
		UserID:      userID,
		Date:        date,
		SentenceIDs: sentenceIDs,
	}

	if err := s.sentenceRepo.CreateDailySet(dailySet); err != nil {
		return nil, err
	}

	return &DailySentencesResponse{
		Date:      date.Format("2006-01-02"),
		Sentences: sentences,
	}, nil
}

func (s *SentenceService) generateSentences(ctx context.Context, level int, interests, purposes []int, count int) ([]model.Sentence, error) {
	if s.openaiClient == nil {
		return nil, fmt.Errorf("OpenAI client not initialized")
	}

	prompt := s.buildPrompt(level, interests, purposes, count)

	resp, err := s.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.openaiModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: s.getSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.8,
		MaxTokens:   4000,
	})
	if err != nil {
		return nil, fmt.Errorf("OpenAI API 호출 실패: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI 응답이 비어있습니다")
	}

	content := resp.Choices[0].Message.Content
	var generated []model.Sentence
	if err := json.Unmarshal([]byte(content), &generated); err != nil {
		content = extractJSON(content)
		if err := json.Unmarshal([]byte(content), &generated); err != nil {
			log.Printf("JSON 파싱 실패: %s", content)
			return nil, fmt.Errorf("생성된 문장 파싱 실패: %w", err)
		}
	}

	// DB에 저장
	sentences := make([]model.Sentence, 0, len(generated))
	for _, g := range generated {
		sentence := model.Sentence{
			JP:         g.JP,
			KR:         g.KR,
			Romaji:     g.Romaji,
			Level:      level,
			Categories: g.Categories,
		}

		if err := s.sentenceRepo.Create(&sentence); err != nil {
			log.Printf("문장 저장 실패: %v", err)
			continue
		}

		sentences = append(sentences, sentence)
	}

	log.Printf("문장 %d개 생성 및 저장 완료", len(sentences))
	return sentences, nil
}

func (s *SentenceService) getSystemPrompt() string {
	return `당신은 일본어 학습용 문장을 생성하는 전문가입니다.
사용자의 관심사와 레벨에 맞는 실용적인 일본어 문장을 생성해주세요.

규칙:
1. 반드시 JSON 배열 형식으로만 응답하세요. 다른 텍스트는 포함하지 마세요.
2. 각 문장은 {"jp": "일본어", "kr": "한국어번역", "romaji": "로마자", "categories": [101, 102]} 형식입니다.
3. 오타쿠 문화에서 실제로 사용되는 자연스러운 표현을 사용하세요.
4. 레벨에 맞는 문법과 어휘를 사용하세요.
5. categories는 아래 카테고리 코드 중 관련된 것을 선택하세요:
   - Anime: 101(이세계/판타지), 102(러브코미디), 103(일상물), 104(배틀/액션), 105(스포츠물), 106(SF/로봇), 107(음악/아이돌물), 108(미스터리/추리)
   - Game: 201(JRPG), 202(모바일가챠), 203(리듬게임), 204(FPS), 205(닌텐도), 206(격투게임)
   - Music: 301(Jpop), 302(Vocaloid), 303(애니송), 304(아이돌), 305(버튜버)
   - Lifestyle: 401(성지순례), 402(굿즈구매), 403(피규어/프라모델), 404(코미케/행사), 405(애니카페), 406(게임센터)
   - Situation: 501(굿즈예약), 502(행사인사), 503(애니얘기), 504(일본사이트주문), 505(오타쿠여행), 506(콘서트/라이브)`
}

func (s *SentenceService) buildPrompt(level int, interests, purposes []int, count int) string {
	levelDesc := s.getLevelDescription(level)

	return fmt.Sprintf(`다음 조건에 맞는 일본어 학습 문장 %d개를 생성해주세요.

## 사용자 정보
- 일본어 레벨: %s
- 관심사: %v
- 학습 목적: %v

## 레벨별 기준
%s

## 출력 형식
JSON 배열로만 응답하세요:
[
  {"jp": "日本語文章", "kr": "한국어 번역", "romaji": "romaji", "categories": [101, 102]}
]`,
		count,
		levelDesc,
		interests,
		purposes,
		s.getLevelGuideline(level),
	)
}

func (s *SentenceService) getLevelDescription(level int) string {
	descriptions := map[int]string{
		0: "Lv0 - 완전 초입문 (히라가나/가타카나 학습 중)",
		1: "Lv1 - 기본 인사 가능 (N5 수준)",
		2: "Lv2 - 일상 회화 조금 가능 (N4 수준)",
		3: "Lv3 - 생각 표현 가능 (N3 수준)",
		4: "Lv4 - 능숙 (N2 수준)",
		5: "Lv5 - 거의 원어민 수준 (N1 수준)",
	}
	if desc, ok := descriptions[level]; ok {
		return desc
	}
	return descriptions[1]
}

func (s *SentenceService) getLevelGuideline(level int) string {
	guidelines := map[int]string{
		0: `- 히라가나/가타카나로만 구성
- 5글자 이하의 매우 짧은 단어/문장
- 예: すき, かわいい, ありがとう`,
		1: `- 기초 한자 포함 가능 (N5 한자)
- です/ます 기본 문형
- 10단어 이하의 짧은 문장
- 예: これは何ですか, 好きです`,
		2: `- N4 수준 한자와 문법
- て형, ない형 사용 가능
- 15단어 이하 문장
- 예: 一緒に見ませんか, 〜したいです`,
		3: `- N3 수준 한자와 문법
- 경어, 가정형 사용 가능
- 자연스러운 회화체
- 예: 〜と思います, 〜かもしれない`,
		4: `- N2 수준 한자와 문법
- 복잡한 문장 구조 가능
- 관용표현 사용
- 예: 〜わけではない, 〜ことになっている`,
		5: `- N1 수준 한자와 문법
- 뉴스, 비즈니스 표현
- 고급 관용구/속담
- 원어민이 사용하는 자연스러운 표현`,
	}
	if guideline, ok := guidelines[level]; ok {
		return guideline
	}
	return guidelines[1]
}

func extractJSON(content string) string {
	start := 0
	end := len(content)

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
