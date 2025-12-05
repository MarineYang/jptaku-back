package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

// ========== 응답 구조체 정의 ==========

type Word struct {
	Japanese string `json:"japanese"`
	Reading  string `json:"reading"`
	Meaning  string `json:"meaning"`
	PartOf   string `json:"part_of"`
}

type QuizFillBlank struct {
	QuestionJP string   `json:"question_jp"`
	Options    []string `json:"options"`
	Answer     string   `json:"answer"`
}

type QuizOrdering struct {
	Fragments    []string `json:"fragments"`
	CorrectOrder []int    `json:"correct_order"`
}

type Quiz struct {
	FillBlank *QuizFillBlank `json:"fill_blank,omitempty"`
	Ordering  *QuizOrdering  `json:"ordering,omitempty"`
}

type GeneratedSentence struct {
	JP         string   `json:"jp"`
	KR         string   `json:"kr"`
	Romaji     string   `json:"romaji"`
	Categories []int    `json:"categories"`
	Words      []Word   `json:"words"`
	Grammar    []string `json:"grammar"`
	Examples   []string `json:"examples"`
	Quiz       Quiz     `json:"quiz"`
}

// ========== 프롬프트 정의 ==========

func getSystemPrompt() string {
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

func buildUserPrompt(level int, interests, purposes []int, count int) string {
	levelDesc := getLevelDescription(level)

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

func getLevelDescription(level int) string {
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

// JSON에서 배열 부분만 추출
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

func main() {
	// .env 파일 로드
	_ = godotenv.Load(".env")       // Docker / production
	_ = godotenv.Load("../../.env") // Local development (cmd/test)

	apiKey := os.Getenv("OPEN_AI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPEN_AI_API_KEY 환경변수가 설정되지 않았습니다")
	}

	// OpenAI 클라이언트 생성
	client := openai.NewClient(apiKey)

	// 테스트 파라미터 설정
	level := 2                   // N4 수준
	interests := []int{101, 201} // 이세계/판타지, JRPG
	purposes := []int{1, 5}      // 애니 감상, 게임
	count := 3                   // 3문장 생성 (테스트용으로 줄임)

	fmt.Println("========================================")
	fmt.Println("OpenAI API 문장 생성 테스트")
	fmt.Println("========================================")
	fmt.Printf("레벨: %d, 관심사: %v, 목적: %v\n", level, interests, purposes)
	fmt.Printf("생성할 문장 수: %d\n", count)
	fmt.Println("========================================")
	fmt.Println("API 호출 중...")
	fmt.Println()

	ctx := context.Background()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini, // 또는 openai.GPT4o
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: getSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: buildUserPrompt(level, interests, purposes, count),
			},
		},
		Temperature: 0.8,
		MaxTokens:   4000,
	})

	if err != nil {
		log.Fatalf("OpenAI API 호출 실패: %v", err)
	}

	if len(resp.Choices) == 0 {
		log.Fatal("OpenAI 응답이 비어있습니다")
	}

	content := resp.Choices[0].Message.Content

	fmt.Println("========================================")
	fmt.Println("원본 응답:")
	fmt.Println("========================================")
	fmt.Println(content)
	fmt.Println()

	// JSON 파싱 시도
	var sentences []GeneratedSentence
	if err := json.Unmarshal([]byte(content), &sentences); err != nil {
		// JSON 추출 후 재시도
		content = extractJSON(content)
		if err := json.Unmarshal([]byte(content), &sentences); err != nil {
			log.Fatalf("JSON 파싱 실패: %v\n추출된 내용: %s", err, content)
		}
	}

	fmt.Println("========================================")
	fmt.Println("파싱된 결과:")
	fmt.Println("========================================")

	for i, s := range sentences {
		fmt.Printf("\n[문장 %d]\n", i+1)
		fmt.Printf("JP: %s\n", s.JP)
		fmt.Printf("KR: %s\n", s.KR)
		fmt.Printf("Romaji: %s\n", s.Romaji)
		fmt.Printf("Categories: %v\n", s.Categories)

		fmt.Println("\n  📚 단어 풀이:")
		for _, w := range s.Words {
			fmt.Printf("    - %s (%s): %s [%s]\n", w.Japanese, w.Reading, w.Meaning, w.PartOf)
		}

		fmt.Println("\n  📝 핵심 문법:")
		for _, g := range s.Grammar {
			fmt.Printf("    - %s\n", g)
		}

		fmt.Println("\n  💡 예문:")
		for _, e := range s.Examples {
			fmt.Printf("    - %s\n", e)
		}

		fmt.Println("\n  ❓ 퀴즈:")
		if s.Quiz.FillBlank != nil {
			fmt.Println("    [빈칸 채우기]")
			fmt.Printf("    문제: %s\n", s.Quiz.FillBlank.QuestionJP)
			fmt.Printf("    보기: %v\n", s.Quiz.FillBlank.Options)
			fmt.Printf("    정답: %s\n", s.Quiz.FillBlank.Answer)
		}
		if s.Quiz.Ordering != nil {
			fmt.Println("    [문장 배열]")
			fmt.Printf("    조각: %v\n", s.Quiz.Ordering.Fragments)
			fmt.Printf("    정답 순서: %v\n", s.Quiz.Ordering.CorrectOrder)
		}

		fmt.Println("\n  ----------------------------------------")
	}

	// 예쁘게 JSON으로 다시 출력
	fmt.Println("\n========================================")
	fmt.Println("정제된 JSON 출력:")
	fmt.Println("========================================")
	prettyJSON, _ := json.MarshalIndent(sentences, "", "  ")
	fmt.Println(string(prettyJSON))

	fmt.Println("\n✅ 테스트 완료!")
	fmt.Printf("총 %d개 문장 생성됨\n", len(sentences))
}
