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
	learningRepo *repository.LearningRepository
	openaiClient *openai.Client
	openaiModel  string
}

func NewSentenceService(sentenceRepo *repository.SentenceRepository, userRepo *repository.UserRepository) *SentenceService {
	return &SentenceService{
		sentenceRepo: sentenceRepo,
		userRepo:     userRepo,
	}
}

func (s *SentenceService) SetLearningRepo(learningRepo *repository.LearningRepository) {
	s.learningRepo = learningRepo
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

// SentenceWithDetail 문장 + 상세 정보
type SentenceWithDetail struct {
	model.Sentence
	Words     []model.Word `json:"words"`
	Grammar   []string     `json:"grammar"`
	Examples  []string     `json:"examples"`
	Quiz      *model.Quiz  `json:"quiz"`
	Memorized bool         `json:"memorized"` // 암기 완료 여부
}

// DailySentencesResponse 오늘의 5문장 응답
type DailySentencesResponse struct {
	Date      string               `json:"date"`
	Sentences []SentenceWithDetail `json:"sentences"`
}

// GetTodaySentences 오늘의 5문장 조회 (없으면 생성)
func (s *SentenceService) GetTodaySentences(userID uint) (*DailySentencesResponse, error) {
	today := time.Now().Truncate(24 * time.Hour)
	return s.getSentencesByDate(userID, today)
}

// HistoryItem 지난 학습 기록 아이템
type HistoryItem struct {
	Date      string               `json:"date"`
	Sentences []SentenceWithDetail `json:"sentences"`
}

// HistorySentencesResponse 지난 학습 문장 응답
type HistorySentencesResponse struct {
	History    []HistoryItem `json:"history"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	Total      int64         `json:"total"`
	TotalPages int           `json:"total_pages"`
}

// GetHistorySentences 지난 학습 문장 조회 (오늘 제외)
func (s *SentenceService) GetHistorySentences(userID uint, page, perPage int) (*HistorySentencesResponse, error) {
	dailySets, total, err := s.sentenceRepo.GetPastDailySets(userID, page, perPage)
	if err != nil {
		return nil, err
	}

	history := make([]HistoryItem, 0, len(dailySets))
	for _, dailySet := range dailySets {
		sentences, err := s.sentenceRepo.FindByIDs(dailySet.SentenceIDs)
		if err != nil {
			continue
		}

		// 상세 정보 + 학습 상태 조회
		sentencesWithDetail := make([]SentenceWithDetail, 0, len(sentences))
		for _, sentence := range sentences {
			detail, _ := s.sentenceRepo.GetDetail(sentence.ID)
			swd := SentenceWithDetail{
				Sentence: sentence,
			}
			if detail != nil {
				swd.Words = detail.Words
				swd.Grammar = detail.Grammar
				swd.Examples = detail.Examples
				swd.Quiz = detail.Quiz
			}

			// 학습 상태 조회
			if s.learningRepo != nil {
				progress, _ := s.learningRepo.FindByUserAndSentence(userID, sentence.ID)
				if progress != nil {
					swd.Memorized = progress.Memorized
				}
			}

			sentencesWithDetail = append(sentencesWithDetail, swd)
		}

		history = append(history, HistoryItem{
			Date:      dailySet.Date.Format("2006-01-02"),
			Sentences: sentencesWithDetail,
		})
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &HistorySentencesResponse{
		History:    history,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *SentenceService) getSentencesByDate(userID uint, date time.Time) (*DailySentencesResponse, error) {
	// 해당 날짜의 세트가 있는지 확인
	dailySet, err := s.sentenceRepo.GetDailySet(userID, date)
	if err == nil && dailySet != nil {
		sentences, err := s.sentenceRepo.FindByIDs(dailySet.SentenceIDs)
		if err != nil {
			return nil, err
		}

		// 상세 정보 + 학습 상태 조회
		sentencesWithDetail := make([]SentenceWithDetail, 0, len(sentences))
		for _, sentence := range sentences {
			detail, _ := s.sentenceRepo.GetDetail(sentence.ID)
			swd := SentenceWithDetail{
				Sentence: sentence,
			}
			if detail != nil {
				swd.Words = detail.Words
				swd.Grammar = detail.Grammar
				swd.Examples = detail.Examples
				swd.Quiz = detail.Quiz
			}

			// 학습 상태 조회
			if s.learningRepo != nil {
				progress, _ := s.learningRepo.FindByUserAndSentence(userID, sentence.ID)
				if progress != nil {
					swd.Memorized = progress.Memorized
				}
			}

			sentencesWithDetail = append(sentencesWithDetail, swd)
		}

		return &DailySentencesResponse{
			Date:      date.Format("2006-01-02"),
			Sentences: sentencesWithDetail,
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
	sentencesWithDetail, err := s.generateSentences(ctx, level, interests, purposes, 5)
	if err != nil {
		return nil, fmt.Errorf("문장 생성 실패: %w", err)
	}

	// 문장 ID 추출
	sentenceIDs := make([]uint, len(sentencesWithDetail))
	for i, swd := range sentencesWithDetail {
		sentenceIDs[i] = swd.ID
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
		Sentences: sentencesWithDetail,
	}, nil
}

// GeneratedSentence OpenAI 응답 파싱용 구조체
type GeneratedSentence struct {
	JP         string       `json:"jp"`
	KR         string       `json:"kr"`
	Romaji     string       `json:"romaji"`
	Categories []int        `json:"categories"`
	Words      []model.Word `json:"words"`
	Grammar    []string     `json:"grammar"`
	Examples   []string     `json:"examples"`
	Quiz       *model.Quiz  `json:"quiz"`
}

func (s *SentenceService) generateSentences(ctx context.Context, level int, interests, purposes []int, count int) ([]SentenceWithDetail, error) {
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
		MaxTokens:   8000,
	})
	if err != nil {
		return nil, fmt.Errorf("OpenAI API 호출 실패: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI 응답이 비어있습니다")
	}

	content := resp.Choices[0].Message.Content
	var generated []GeneratedSentence
	if err := json.Unmarshal([]byte(content), &generated); err != nil {
		content = extractJSON(content)
		if err := json.Unmarshal([]byte(content), &generated); err != nil {
			log.Printf("JSON 파싱 실패: %s", content)
			return nil, fmt.Errorf("생성된 문장 파싱 실패: %w", err)
		}
	}

	// DB에 저장
	sentencesWithDetail := make([]SentenceWithDetail, 0, len(generated))
	for _, g := range generated {
		// Sentence 저장
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

		// SentenceDetail 저장
		detail := model.SentenceDetail{
			SentenceID: sentence.ID,
			Words:      g.Words,
			Grammar:    g.Grammar,
			Examples:   g.Examples,
			Quiz:       g.Quiz,
		}

		if err := s.sentenceRepo.CreateDetail(&detail); err != nil {
			log.Printf("문장 상세 저장 실패: %v", err)
		}

		sentencesWithDetail = append(sentencesWithDetail, SentenceWithDetail{
			Sentence: sentence,
			Words:    g.Words,
			Grammar:  g.Grammar,
			Examples: g.Examples,
			Quiz:     g.Quiz,
		})
	}

	log.Printf("문장 %d개 생성 및 저장 완료", len(sentencesWithDetail))
	return sentencesWithDetail, nil
}

func (s *SentenceService) getSystemPrompt() string {
	return `당신은 일본어 학습용 문장 + 문장 분석 + 퀴즈를 생성하는 전문가입니다.
사용자의 관심사와 레벨에 맞는 실용적인 일본어 문장을 생성하고, 각 문장에 대한 단어 풀이 / 문법 / 예문 / 퀴즈까지 한 번에 만들어주세요.

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
    "categories": number[],    // 관련 서브카테고리 코드 배열 (예: [101, 201])

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
- "categories"에는 아래 SubCategory 코드 중에서 1개 이상을 선택해 넣으세요.
- 가능한 한 문장과 가장 관련 있는 코드만 고르세요.

- Anime: 101(이세계/판타지), 102(러브코미디), 103(일상물), 104(배틀/액션), 105(스포츠물), 106(SF/로봇), 107(음악/아이돌물), 108(미스터리/추리)
- Game: 201(JRPG), 202(모바일가챠), 203(리듬게임), 204(FPS), 205(닌텐도), 206(격투게임)
- Music: 301(Jpop), 302(Vocaloid), 303(애니송), 304(아이돌), 305(버튜버)
- Lifestyle: 401(성지순례), 402(굿즈구매), 403(피규어/프라모델), 404(코미케/행사), 405(애니카페), 406(게임센터)
- Situation: 501(굿즈예약), 502(행사인사), 503(애니얘기), 504(일본사이트주문), 505(오타쿠여행), 506(콘서트/라이브)

[레벨/어휘/문법 가이드라인]
- 사용자의 일본어 레벨에 맞추어 난이도를 조절하세요.
- Lv0~Lv1: 짧고 쉬운 표현, 기본 인사, N5 수준
- Lv2: 일상 회화, N4 수준, て형/ない형 등 기본 활용
- Lv3: 자신의 생각 표현, N3 수준, 조금 길어도 됨
- Lv4: 보다 복잡한 문장, 관용구 일부 포함, N2 수준
- Lv5: 자연스러운 원어민 표현, N1 수준

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

func (s *SentenceService) buildPrompt(level int, interests, purposes []int, count int) string {
	levelDesc := s.getLevelDescription(level)

	return fmt.Sprintf(`다음 조건에 맞는 일본어 학습 문장과 상세 정보를 JSON으로 생성해 주세요.

[사용자 정보]
- 일본어 레벨: %s
- 관심사 SubCategory 코드: %v
- 학습 목적 코드: %v

[생성할 문장 개수]
- 총 %d개의 서로 다른 문장을 생성해 주세요.

[요구 사항]
- 각 문장은 다음 요소를 모두 포함해야 합니다:
  - jp / kr / romaji / categories
  - words (단어 풀이 배열)
  - grammar (핵심 문법 배열)
  - examples (추가 예문 배열)
  - quiz.fill_blank / quiz.ordering

[출력 형식]
- 반드시 시스템 메시지에서 정의한 JSON 스키마를 그대로 따르세요.
- JSON 배열 이외의 설명 텍스트는 절대 포함하지 마세요.`, levelDesc, interests, purposes, count)
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
