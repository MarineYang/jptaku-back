package sentence

import "github.com/jptaku/server/internal/model"

// SentenceWithDetail 문장 + 상세 정보
type SentenceWithDetail struct {
	model.Sentence
	Words     []model.Word `json:"words"`
	Grammar   []string     `json:"grammar"`
	Examples  []string     `json:"examples"`
	Quiz      *model.Quiz  `json:"quiz"`
	Memorized bool         `json:"memorized"`
}

// DailySentencesResponse 오늘의 5문장 응답
type DailySentencesResponse struct {
	Date      string               `json:"date"`
	Sentences []SentenceWithDetail `json:"sentences"`
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
