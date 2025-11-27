package model

import (
	"time"
)

type ChatSession struct {
	ID                     uint       `gorm:"primaryKey" json:"id"`
	UserID                 uint       `gorm:"index;not null" json:"user_id"`
	DailySetID             uint       `gorm:"index" json:"daily_set_id"`
	StartedAt              time.Time  `gorm:"not null" json:"started_at"`
	EndedAt                *time.Time `json:"ended_at,omitempty"`
	TodaySentenceUsedCount int        `gorm:"default:0" json:"today_sentence_used_count"` // 오늘 5문장 중 사용한 수
	TotalMessages          int        `gorm:"default:0" json:"total_messages"`
	DurationSeconds        int        `gorm:"default:0" json:"duration_seconds"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`

	// Relations
	User     *User         `gorm:"foreignKey:UserID" json:"-"`
	Messages []ChatMessage `gorm:"foreignKey:SessionID" json:"messages,omitempty"`
}

type ChatMessage struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	SessionID           uint      `gorm:"index;not null" json:"session_id"`
	Speaker             string    `gorm:"size:10;not null" json:"speaker"` // "ai" or "user"
	JPText              string    `gorm:"type:text" json:"jp_text"`
	KRText              string    `gorm:"type:text" json:"kr_text"`
	UsedTodaySentenceID *uint     `json:"used_today_sentence_id,omitempty"` // 오늘 5문장 중 사용한 경우
	CreatedAt           time.Time `json:"created_at"`

	// Relations
	UsedSentence *Sentence `gorm:"foreignKey:UsedTodaySentenceID" json:"used_sentence,omitempty"`
}

func (ChatSession) TableName() string {
	return "chat_sessions"
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
