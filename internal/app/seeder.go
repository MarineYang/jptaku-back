package app

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"github.com/jptaku/server/internal/model"
	"gorm.io/gorm"
)

// parseLevelToInt CSV의 레벨 문자열을 int로 변환
// 초급/Beginner/N5 → 1, 중급/Intermediate/N4 → 2, 고급/Advanced/N3 → 3
func parseLevelToInt(level string) int {
	l := strings.ToLower(level)
	switch {
	case strings.Contains(l, "초급") || strings.Contains(l, "beginner") || l == "n5":
		return 1
	case strings.Contains(l, "중급") || strings.Contains(l, "intermediate") || l == "n4":
		return 2
	case strings.Contains(l, "고급") || strings.Contains(l, "advanced") || l == "n3":
		return 3
	default:
		return 1
	}
}

// SeedFromCSV CSV 파일로부터 문장 데이터 시딩 (데이터가 없을 때만)
func SeedFromCSV(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.Sentence{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		log.Printf("Sentences already exist (%d), skipping seed", count)
		return nil
	}

	// CSV 파일 경로들
	paths := []string{
		"docs/sentences_v2.csv",      // Local
		"/app/docs/sentences_v2.csv", // Docker
	}

	var file *os.File
	var err error
	var usedPath string

	for _, path := range paths {
		file, err = os.Open(path)
		if err == nil {
			usedPath = path
			break
		}
	}

	if err != nil {
		log.Printf("CSV file not found, skipping seed")
		return nil
	}
	defer file.Close()

	log.Printf("Loading sentence data from: %s", usedPath)

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable field count
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 2 {
		log.Printf("CSV file is empty, skipping seed")
		return nil
	}

	// 헤더: UUID, Domain, Topic, Level, FunctionMacro, FunctionMicro, GeneratedSentence, KoreanTranslation, Words, Keywords
	inserted := 0
	return db.Transaction(func(tx *gorm.DB) error {
		for _, row := range records[1:] {
			if len(row) < 10 {
				continue
			}

			uuid := row[0]
			jp := row[6]
			kr := row[7]

			if jp == "" || kr == "" {
				continue
			}

			sentence := &model.Sentence{
				UUID:          uuid,
				Domain:        row[1],
				Topic:         row[2],
				Level:         parseLevelToInt(row[3]),
				FunctionMacro: row[4],
				FunctionMicro: row[5],
				JP:            jp,
				KR:            kr,
				Words:         row[8],
				Keywords:      row[9],
			}

			if err := tx.Create(sentence).Error; err != nil {
				log.Printf("Failed to insert sentence (UUID=%s): %v", uuid, err)
				continue
			}
			inserted++
		}

		log.Printf("Seeded %d sentences from CSV", inserted)
		return nil
	})
}
