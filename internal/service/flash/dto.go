package flash

import (
	"time"

	"github.com/jptaku/server/internal/model"
)

// FlashSentence Flash용 문장 정보
type FlashSentence struct {
	model.Sentence
	// Detail 정보
	Phrase string `json:"phrase"`
	Tip    string `json:"tip"`
	Alt    string `json:"alt"`
	// Progress 정보
	FlashGrade   string     `json:"flash_grade,omitempty"`    // bad/mid/good
	FlashCount   int        `json:"flash_count"`              // 복습 횟수
	NextReviewAt *time.Time `json:"next_review_at,omitempty"` // 다음 복습 시간
}

// TodayFlashResponse 오늘의 Flash 응답
type TodayFlashResponse struct {
	Date      string          `json:"date"`
	Sentences []FlashSentence `json:"sentences"`
}

// UpdateFlashInput Flash 진행 상황 업데이트 입력
type UpdateFlashInput struct {
	SentenceID uint   `json:"sentence_id" binding:"required"`
	Grade      string `json:"grade" binding:"required,oneof=bad mid good"`
}

// FlashProgressResult Flash 진행 상황 업데이트 결과
type FlashProgressResult struct {
	SentenceID   uint       `json:"sentence_id"`
	Grade        string     `json:"grade"`
	FlashCount   int        `json:"flash_count"`
	NextReviewAt *time.Time `json:"next_review_at,omitempty"` // 다음 복습 시간
}
