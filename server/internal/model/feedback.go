package model

import (
	"time"
)

type Feedback struct {
	ID                 uint                `gorm:"primaryKey" json:"id"`
	SessionID          uint                `gorm:"uniqueIndex;not null" json:"session_id"`
	TotalScore         float64             `gorm:"default:0" json:"total_score"`                 // 총점
	GrammarScore       float64             `gorm:"default:0" json:"grammar_score"`               // 문법 점수
	PronunciationScore float64             `gorm:"default:0" json:"pronunciation_score"`         // 발음 점수
	NaturalnessScore   float64             `gorm:"default:0" json:"naturalness_score"`           // 자연스러움 점수
	Summary            string              `gorm:"type:text" json:"summary"`                     // 요약 문장
	Highlights         []FeedbackHighlight `gorm:"type:jsonb;serializer:json" json:"highlights"` // 하이라이트
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`

	// Relations
	Session *ChatSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

type FeedbackHighlight struct {
	Title   string `json:"title"`
	JP      string `json:"jp"`
	KR      string `json:"kr"`
	Comment string `json:"comment"`
}

func (Feedback) TableName() string {
	return "feedbacks"
}
