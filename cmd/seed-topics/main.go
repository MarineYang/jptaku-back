package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/jptaku/server/internal/cache"
	"github.com/jptaku/server/internal/config"
)

const redisKeyPrefix = "topics:"

func main() {
	// .env 로드
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")

	cfg := config.Load()

	redisClient, err := cache.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Redis 연결 실패: %v", err)
	}
	defer redisClient.Close()

	ctx := context.Background()

	// 시딩할 파일 목록 (domain → file path)
	files := map[string]string{
		"anime": "docs/anime_20260210.json",
		"drama": "docs/drama_20260210.json",
		"game":  "docs/game_20260210.json",
		"movie": "docs/movie_20260210.json",
		"music": "docs/music_20260210.json",
	}

	for domain, filePath := range files {
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("파일 읽기 실패 (%s): %v", filePath, err)
			continue
		}

		key := redisKeyPrefix + domain
		err = redisClient.Set(ctx, key, string(data), 0).Err()
		if err != nil {
			log.Printf("Redis 저장 실패 (%s): %v", key, err)
			continue
		}

		fmt.Printf("✓ %s → Redis key '%s' (%d bytes)\n", filePath, key, len(data))
	}

	fmt.Println("\n시딩 완료!")
}
