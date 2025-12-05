package model

import (
	"time"

	"gorm.io/gorm"
)

type Sentence struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	JP         string         `gorm:"type:text;not null" json:"jp"`                 // 일본어 문장
	KR         string         `gorm:"type:text;not null" json:"kr"`                 // 한국어 번역
	Romaji     string         `gorm:"type:text" json:"romaji,omitempty"`            // 로마지
	Level      int            `gorm:"default:1" json:"level"`                       // 난이도 1~5
	Categories []int          `gorm:"type:jsonb;serializer:json" json:"categories"` // pkg.SubCategory 값들
	AudioURL   string         `gorm:"size:500" json:"audio_url,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type SentenceDetail struct {
	ID         uint     `gorm:"primaryKey" json:"id"`
	SentenceID uint     `gorm:"index;not null" json:"sentence_id"`
	Words      []Word   `gorm:"type:jsonb;serializer:json" json:"words"`    // 단어 풀이
	Grammar    []string `gorm:"type:jsonb;serializer:json" json:"grammar"`  // 핵심 문법
	Examples   []string `gorm:"type:jsonb;serializer:json" json:"examples"` // 예문
	Quiz       *Quiz    `gorm:"type:jsonb;serializer:json" json:"quiz"`     // 퀴즈
}

type Word struct {
	Japanese string `json:"japanese"`
	Reading  string `json:"reading"`
	Meaning  string `json:"meaning"`
	PartOf   string `json:"part_of"` // 품사
}

// Quiz 확인하기 퀴즈
type Quiz struct {
	FillBlank *QuizFillBlank `json:"fill_blank,omitempty"`
	Ordering  *QuizOrdering  `json:"ordering,omitempty"`
}

// QuizFillBlank 빈칸 채우기
type QuizFillBlank struct {
	QuestionJP string   `json:"question_jp"`
	Options    []string `json:"options"`
	Answer     string   `json:"answer"`
}

// QuizOrdering 문장 배열하기
type QuizOrdering struct {
	Fragments    []string `json:"fragments"`
	CorrectOrder []int    `json:"correct_order"`
}

func (Sentence) TableName() string {
	return "sentences"
}

func (SentenceDetail) TableName() string {
	return "sentence_details"
}
