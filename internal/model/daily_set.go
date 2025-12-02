package model

import (
	"time"
)

type DailySentenceSet struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	Date        time.Time `gorm:"type:date;index;not null" json:"date"`          // 날짜
	SentenceIDs []uint    `gorm:"type:jsonb;serializer:json" json:"sentence_ids"` // 5개 문장 ID
	CreatedAt   time.Time `json:"created_at"`

	// Relations
	User      *User      `gorm:"foreignKey:UserID" json:"-"`
	Sentences []Sentence `gorm:"-" json:"sentences,omitempty"`
}

func (DailySentenceSet) TableName() string {
	return "daily_sentence_sets"
}
