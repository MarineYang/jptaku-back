package main

import (
	"flag"
	"log"

	"github.com/jptaku/server/internal/config"
	"github.com/jptaku/server/internal/model"
)

func main() {
	// Parse command line flags
	action := flag.String("action", "up", "Migration action: up, down, seed")
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := config.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	switch *action {
	case "up":
		migrateUp(db)
	case "down":
		migrateDown(db)
	case "seed":
		seed(db)
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

func migrateUp(db interface {
	AutoMigrate(dst ...interface{}) error
}) {
	log.Println("Running migrations...")

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
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}

func migrateDown(db interface {
	Migrator() interface {
		DropTable(dst ...interface{}) error
	}
}) {
	log.Println("Rolling back migrations...")

	migrator := db.Migrator()

	// Drop tables in reverse order due to foreign key constraints
	tables := []interface{}{
		&model.Feedback{},
		&model.ChatMessage{},
		&model.ChatSession{},
		&model.LearningProgress{},
		&model.DailySentenceSet{},
		&model.SentenceDetail{},
		&model.Sentence{},
		&model.UserOnboarding{},
		&model.UserSettings{},
		&model.User{},
	}

	for _, table := range tables {
		if err := migrator.DropTable(table); err != nil {
			log.Printf("Warning: Failed to drop table: %v", err)
		}
	}

	log.Println("Rollback completed!")
}

func seed(db interface {
	Create(value interface{}) interface{ Error() error }
}) {
	log.Println("Seeding database...")

	// Sample sentences for testing
	sampleSentences := []model.Sentence{
		{
			JP:     "おはようございます",
			KR:     "안녕하세요 (아침 인사)",
			Romaji: "ohayou gozaimasu",
			Level:  1,
			Tags:   []string{"기본", "인사"},
		},
		{
			JP:     "いただきます",
			KR:     "잘 먹겠습니다",
			Romaji: "itadakimasu",
			Level:  1,
			Tags:   []string{"기본", "식사"},
		},
		{
			JP:     "このアニメは面白いですね",
			KR:     "이 애니메이션 재미있네요",
			Romaji: "kono anime wa omoshiroi desu ne",
			Level:  2,
			Tags:   []string{"애니"},
		},
		{
			JP:     "ガチャで推しが出た！",
			KR:     "가챠에서 최애가 나왔어!",
			Romaji: "gacha de oshi ga deta!",
			Level:  3,
			Tags:   []string{"게임", "가챠"},
		},
		{
			JP:     "聖地巡礼に行きたいです",
			KR:     "성지순례 가고 싶어요",
			Romaji: "seichi junrei ni ikitai desu",
			Level:  3,
			Tags:   []string{"성지순례"},
		},
		{
			JP:     "コミケに参加したことありますか？",
			KR:     "코미케에 참가한 적 있어요?",
			Romaji: "komike ni sanka shita koto arimasu ka?",
			Level:  3,
			Tags:   []string{"이벤트"},
		},
		{
			JP:     "推しの生誕祭を祝いたい",
			KR:     "최애 생일 축하해주고 싶어",
			Romaji: "oshi no seitansai wo iwaitai",
			Level:  4,
			Tags:   []string{"애니", "이벤트"},
		},
		{
			JP:     "秋葉原でフィギュアを買いました",
			KR:     "아키하바라에서 피규어를 샀어요",
			Romaji: "akihabara de figua wo kaimashita",
			Level:  3,
			Tags:   []string{"성지순례", "쇼핑"},
		},
		{
			JP:     "新作ゲームの予約開始だ",
			KR:     "신작 게임 예약 시작이다",
			Romaji: "shinsaku geemu no yoyaku kaishi da",
			Level:  3,
			Tags:   []string{"게임"},
		},
		{
			JP:     "声優のサイン会に当選した",
			KR:     "성우 사인회에 당첨됐어",
			Romaji: "seiyuu no sainkai ni tousen shita",
			Level:  4,
			Tags:   []string{"이벤트", "성우"},
		},
	}

	for _, sentence := range sampleSentences {
		if err := db.Create(&sentence).Error(); err != nil {
			log.Printf("Warning: Failed to seed sentence: %v", err)
		}
	}

	log.Println("Seeding completed!")
}
