package app

import (
	"encoding/json"
	"log"
	"os"

	"github.com/jptaku/server/internal/model"
	"gorm.io/gorm"
)

// MockData JSON 파일 구조
type MockData struct {
	Sentences []MockSentence `json:"sentences"`
}

// MockSentence 문장 mock 데이터 구조
type MockSentence struct {
	SentenceKey string           `json:"sentence_key"`
	JP          string           `json:"jp"`
	KR          string           `json:"kr"`
	Romaji      string           `json:"romaji"`
	Level       int              `json:"level"`
	Category    int              `json:"category"`
	Detail      MockSentenceDetail `json:"detail"`
}

// MockSentenceDetail 문장 상세 mock 데이터 구조
type MockSentenceDetail struct {
	Words    []model.Word `json:"words"`
	Grammar  []string     `json:"grammar"`
	Examples []string     `json:"examples"`
	Quiz     *model.Quiz  `json:"quiz"`
	Phrase   string       `json:"phrase"`
	Tip      string       `json:"tip"`
	Alt      string       `json:"alt"`
}

// SeedMockData mock 데이터 시딩 (데이터가 없을 때만)
func SeedMockData(db *gorm.DB) error {
	// 이미 문장 데이터가 있으면 스킵
	var count int64
	if err := db.Model(&model.Sentence{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		log.Printf("Sentences already exist (%d), skipping seed", count)
		return nil
	}

	// Mock 데이터 파일 경로들 (Docker/Local 모두 지원)
	paths := []string{
		"data/mock_sentences.json",      // Docker
		"../../data/mock_sentences.json", // Local development
	}

	var data []byte
	var err error
	var usedPath string

	for _, path := range paths {
		data, err = os.ReadFile(path)
		if err == nil {
			usedPath = path
			break
		}
	}

	if err != nil {
		log.Printf("Mock data file not found, skipping seed")
		return nil
	}

	log.Printf("Loading mock data from: %s", usedPath)

	// JSON 파싱
	var mockData MockData
	if err := json.Unmarshal(data, &mockData); err != nil {
		return err
	}

	// 트랜잭션으로 데이터 삽입
	return db.Transaction(func(tx *gorm.DB) error {
		for _, ms := range mockData.Sentences {
			// Sentence 생성
			sentence := &model.Sentence{
				SentenceKey: ms.SentenceKey,
				JP:          ms.JP,
				KR:          ms.KR,
				Romaji:      ms.Romaji,
				Level:       ms.Level,
				Category:    ms.Category,
			}

			if err := tx.Create(sentence).Error; err != nil {
				return err
			}

			// SentenceDetail 생성
			detail := &model.SentenceDetail{
				SentenceID: sentence.ID,
				Words:      ms.Detail.Words,
				Grammar:    ms.Detail.Grammar,
				Examples:   ms.Detail.Examples,
				Quiz:       ms.Detail.Quiz,
				Phrase:     ms.Detail.Phrase,
				Tip:        ms.Detail.Tip,
				Alt:        ms.Detail.Alt,
			}

			if err := tx.Create(detail).Error; err != nil {
				return err
			}
		}

		log.Printf("Seeded %d mock sentences", len(mockData.Sentences))
		return nil
	})
}
