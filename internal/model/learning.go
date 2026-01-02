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
	// Flash 카드용 필드
	FlashGrade     *string    `gorm:"size:10" json:"flash_grade,omitempty"`      // bad/mid/good
	FlashUpdatedAt *time.Time `json:"flash_updated_at,omitempty"`                // Flash 평가 시간
	FlashCount     int        `gorm:"default:0" json:"flash_count"`              // Flash 평가 횟수
	NextReviewAt   *time.Time `json:"next_review_at,omitempty"`                  // 다음 복습 시간 (SRS)

	// Relations
	User     *User     `gorm:"foreignKey:UserID" json:"-"`
	Sentence *Sentence `gorm:"foreignKey:SentenceID" json:"sentence,omitempty"`
}

func (LearningProgress) TableName() string {
	return "learning_progress"
}
