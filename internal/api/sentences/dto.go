package sentences

// WordResponse 단어 풀이
type WordResponse struct {
	Japanese string `json:"japanese"`
	Reading  string `json:"reading"`
	Meaning  string `json:"meaning"`
	PartOf   string `json:"part_of"`
}

// QuizFillBlankResponse 빈칸 채우기 퀴즈
type QuizFillBlankResponse struct {
	QuestionJP string   `json:"question_jp"`
	Options    []string `json:"options"`
	Answer     string   `json:"answer"`
}

// QuizOrderingResponse 문장 배열 퀴즈
type QuizOrderingResponse struct {
	Fragments    []string `json:"fragments"`
	CorrectOrder []int    `json:"correct_order"`
}

// QuizResponse 퀴즈 응답
type QuizResponse struct {
	FillBlank *QuizFillBlankResponse `json:"fill_blank,omitempty"`
	Ordering  *QuizOrderingResponse  `json:"ordering,omitempty"`
}

// SentenceResponse 문장 + 상세 정보 응답
type SentenceResponse struct {
	ID          uint           `json:"id"`
	SentenceKey string         `json:"sentence_key"`
	JP          string         `json:"jp"`
	KR          string         `json:"kr"`
	Romaji      string         `json:"romaji"`
	Level       int            `json:"level"`
	SubCategory int            `json:"sub_category"`
	Words       []WordResponse `json:"words"`
	Grammar     []string       `json:"grammar"`
	Examples    []string       `json:"examples"`
	Quiz        *QuizResponse  `json:"quiz,omitempty"`
	Memorized   bool           `json:"memorized"` // 암기 완료 여부
}

// DailySentencesResponse 오늘의 5문장 응답
type DailySentencesResponse struct {
	Date      string             `json:"date"`
	Sentences []SentenceResponse `json:"sentences"`
}

// HistoryItemResponse 지난 학습 기록 아이템
type HistoryItemResponse struct {
	Date      string             `json:"date"`
	Sentences []SentenceResponse `json:"sentences"`
}

// HistorySentencesResponse 지난 학습 문장 응답
type HistorySentencesResponse struct {
	History    []HistoryItemResponse `json:"history"`
	Page       int                   `json:"page"`
	PerPage    int                   `json:"per_page"`
	Total      int64                 `json:"total"`
	TotalPages int                   `json:"total_pages"`
}
