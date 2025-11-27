package model

import (
	"time"
)

type LearningProgress struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	SentenceID  uint       `gorm:"index;not null" json:"sentence_id"`
	DailySetID  uint       `gorm:"index" json:"daily_set_id"`
	Understand  bool       `gorm:"default:false" json:"understand"` // 이해하기 완료
	Speak       bool       `gorm:"default:false" json:"speak"`      // 말하기 완료
	Confirm     bool       `gorm:"default:false" json:"confirm"`    // 확인하기 완료
	Memorized   bool       `gorm:"default:false" json:"memorized"`  // 암기 완료
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relations
	User     *User     `gorm:"foreignKey:UserID" json:"-"`
	Sentence *Sentence `gorm:"foreignKey:SentenceID" json:"sentence,omitempty"`
}

func (LearningProgress) TableName() string {
	return "learning_progress"
}
