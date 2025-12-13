package main

import (
	"flag"
	"log"

	"github.com/jptaku/server/internal/config"
	"github.com/jptaku/server/internal/model"
	"gorm.io/gorm"
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

func migrateUp(db *gorm.DB) {
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

func migrateDown(db *gorm.DB) {
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

func seed(db *gorm.DB) {
	log.Println("Seeding database...")

	// Sample sentences for testing (새 스키마에 맞게 수정)
	sampleSentences := []model.Sentence{
		{
			SentenceKey: "102_1",
			JP:          "おはようございます",
			KR:          "안녕하세요 (아침 인사)",
			Romaji:      "ohayou gozaimasu",
			Level:       1,
			SubCategory: 102, // 일상/러브코미·감성
		},
		{
			SentenceKey: "102_1",
			JP:          "いただきます",
			KR:          "잘 먹겠습니다",
			Romaji:      "itadakimasu",
			Level:       1,
			SubCategory: 102,
		},
		{
			SentenceKey: "101_2",
			JP:          "このアニメは面白いですね",
			KR:          "이 애니메이션 재미있네요",
			Romaji:      "kono anime wa omoshiroi desu ne",
			Level:       2,
			SubCategory: 101, // 배틀/판타지·SF
		},
		{
			SentenceKey: "201_3",
			JP:          "ガチャで推しが出た！",
			KR:          "가챠에서 최애가 나왔어!",
			Romaji:      "gacha de oshi ga deta!",
			Level:       3,
			SubCategory: 201, // RPG/가챠
		},
		{
			SentenceKey: "401_3",
			JP:          "聖地巡礼に行きたいです",
			KR:          "성지순례 가고 싶어요",
			Romaji:      "seichi junrei ni ikitai desu",
			Level:       3,
			SubCategory: 401, // 성지순례/여행
		},
		{
			SentenceKey: "403_3",
			JP:          "コミケに参加したことありますか？",
			KR:          "코미케에 참가한 적 있어요?",
			Romaji:      "komike ni sanka shita koto arimasu ka?",
			Level:       3,
			SubCategory: 403, // 코미케/동인
		},
		{
			SentenceKey: "302_3",
			JP:          "推しの生誕祭を祝いたい",
			KR:          "최애 생일 축하해주고 싶어",
			Romaji:      "oshi no seitansai wo iwaitai",
			Level:       3,
			SubCategory: 302, // 아이돌
		},
		{
			SentenceKey: "402_3",
			JP:          "秋葉原でフィギュアを買いました",
			KR:          "아키하바라에서 피규어를 샀어요",
			Romaji:      "akihabara de figua wo kaimashita",
			Level:       3,
			SubCategory: 402, // 굿즈/수집
		},
		{
			SentenceKey: "201_3",
			JP:          "新作ゲームの予約開始だ",
			KR:          "신작 게임 예약 시작이다",
			Romaji:      "shinsaku geemu no yoyaku kaishi da",
			Level:       3,
			SubCategory: 201, // RPG/가챠
		},
		{
			SentenceKey: "502_3",
			JP:          "声優のサイン会に当選した",
			KR:          "성우 사인회에 당첨됐어",
			Romaji:      "seiyuu no sainkai ni tousen shita",
			Level:       3,
			SubCategory: 502, // 현장/라이브
		},
	}

	for _, sentence := range sampleSentences {
		if err := db.Create(&sentence).Error; err != nil {
			log.Printf("Warning: Failed to seed sentence: %v", err)
		}
	}

	log.Println("Seeding completed!")
}
