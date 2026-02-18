package model

import (
	"time"

	"gorm.io/gorm"
)

type Sentence struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UUID          string         `gorm:"size:50;uniqueIndex" json:"uuid"`               // 외부 생성 UUID
	Domain        string         `gorm:"size:20;index;not null" json:"domain"`           // anime, game, music, movie, drama
	Topic         string         `gorm:"size:255" json:"topic"`                          // 작품명 (예: "BLEACH")
	Level         int            `gorm:"index;not null" json:"level"`                    // 5=N5(초급), 4=N4(중급), 3=N3(고급)
	FunctionMacro string         `gorm:"size:100" json:"function_macro,omitempty"`       // 상호작용, 표현 등
	FunctionMicro string         `gorm:"size:100" json:"function_micro,omitempty"`       // 관계형성, 의견교환 등
	JP            string         `gorm:"type:text;not null" json:"jp"`                   // 일본어 문장
	KR            string         `gorm:"type:text;not null" json:"kr"`                   // 한국어 번역
	Words         string         `gorm:"type:text" json:"words,omitempty"`               // 파이프 구분 단어
	Keywords      string         `gorm:"type:text" json:"keywords,omitempty"`            // 파이프 구분 키워드
	AudioURL      string         `gorm:"size:500" json:"audio_url,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type SentenceDetail struct {
	ID         uint     `gorm:"primaryKey" json:"id"`
	SentenceID uint     `gorm:"index;not null" json:"sentence_id"`
	Words      []Word   `gorm:"type:jsonb;serializer:json" json:"words"`    // 단어 풀이
	Grammar    []string `gorm:"type:jsonb;serializer:json" json:"grammar"`  // 핵심 문법
	Examples   []string `gorm:"type:jsonb;serializer:json" json:"examples"` // 예문
	Quiz       *Quiz    `gorm:"type:jsonb;serializer:json" json:"quiz"`     // 퀴즈
	// Flash 카드용 필드
	Phrase string `gorm:"type:text" json:"phrase,omitempty"` // 핵심 표현 1개
	Tip    string `gorm:"type:text" json:"tip,omitempty"`    // 한 줄 설명
	Alt    string `gorm:"type:text" json:"alt,omitempty"`    // 같은 구조 예문 1개
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
