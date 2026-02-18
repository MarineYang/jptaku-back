package model

import (
	"time"
)

const (
	MaxTurnsPerSession = 8 // 세션당 최대 턴 수 (유저 메시지 기준)
)

type ChatSession struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UserID          uint       `gorm:"index;not null" json:"user_id"`
	Topic           string     `gorm:"size:255" json:"topic"`              // 대화 주제 (예: "애니메이션", "게임")
	TopicDetail     string     `gorm:"size:500" json:"topic_detail"`       // 세부 주제 (예: "원피스에 대해서")
	CurrentTurn     int        `gorm:"default:0" json:"current_turn"`      // 현재 턴 수
	MaxTurn         int        `gorm:"default:8" json:"max_turn"`          // 최대 턴 수
	PersonaName     string     `gorm:"size:50" json:"persona_name"`              // 캐릭터 이름
	PersonaGender   string     `gorm:"size:10" json:"persona_gender"`            // "male" or "female"
	Domain          string     `gorm:"size:50" json:"domain"`                    // "anime", "drama" 등
	ContentID       string     `gorm:"size:100" json:"content_id"`               // TopicContent.ID
	ContentTitle    string     `gorm:"size:255" json:"content_title"`            // 작품 제목 (일본어)
	Status          string     `gorm:"size:20;default:'active'" json:"status"` // active, completed, ended
	StartedAt       time.Time  `gorm:"not null" json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationSeconds int        `gorm:"default:0" json:"duration_seconds"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	User     *User         `gorm:"foreignKey:UserID" json:"-"`
	Messages []ChatMessage `gorm:"foreignKey:SessionID" json:"messages,omitempty"`
}

type ChatMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID uint      `gorm:"index;not null" json:"session_id"`
	Role      string    `gorm:"size:20;not null" json:"role"` // "user" or "assistant"
	Content   string    `gorm:"type:text" json:"content"`     // 일본어 메시지
	CreatedAt time.Time `json:"created_at"`
}

func (ChatSession) TableName() string {
	return "chat_sessions"
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
