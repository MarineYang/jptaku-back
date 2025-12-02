package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Password  string         `gorm:"size:255" json:"-"`
	Name      string         `gorm:"size:100" json:"name"`
	Provider  string         `gorm:"size:50" json:"provider"` // local, google, kakao, etc.
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Settings   *UserSettings   `gorm:"foreignKey:UserID" json:"settings,omitempty"`
	Onboarding *UserOnboarding `gorm:"foreignKey:UserID" json:"onboarding,omitempty"`
}

type UserSettings struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	UserID              uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	NotificationEnabled bool      `gorm:"default:true" json:"notification_enabled"`
	DailyReminderTime   string    `gorm:"size:10;default:'09:00'" json:"daily_reminder_time"`
	PreferredVoiceSpeed float64   `gorm:"default:1.0" json:"preferred_voice_speed"`
	ShowRomaji          bool      `gorm:"default:true" json:"show_romaji"`
	ShowTranslation     bool      `gorm:"default:true" json:"show_translation"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type UserOnboarding struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	Level     int       `gorm:"default:0" json:"level"`                 // 0 ~ 5 (pkg.Level)
	Interests []int     `gorm:"type:jsonb;serializer:json" json:"interests"` // pkg.SubCategory 값들
	Purposes  []int     `gorm:"type:jsonb;serializer:json" json:"purposes"`  // pkg.Purpose 값들
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

func (UserSettings) TableName() string {
	return "user_settings"
}

func (UserOnboarding) TableName() string {
	return "user_onboardings"
}
