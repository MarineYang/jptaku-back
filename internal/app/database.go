package app

import (
	"log"

	"github.com/jptaku/server/internal/config"
	"github.com/jptaku/server/internal/model"
	"gorm.io/gorm"
)

// NewDatabase 데이터베이스 연결 초기화
func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	db, err := config.NewDatabase(&cfg.Database)
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

// RunMigrations 데이터베이스 마이그레이션 실행
func RunMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.UserSettings{},
		&model.UserOnboarding{},
		&model.Sentence{},
		&model.SentenceDetail{},
		&model.DailySentenceSet{},
		&model.LearningProgress{},
		&model.ChatSession{},
		&model.ChatMessage{},
		&model.Feedback{},
	); err != nil {
		return err
	}

	log.Println("Database migration completed")
	return nil
}
