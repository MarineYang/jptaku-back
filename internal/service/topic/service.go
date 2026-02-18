package topic

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/jptaku/server/internal/model"
	"github.com/redis/go-redis/v9"
)

const redisKeyPrefix = "topics:"

// Service 토픽 로더 서비스
type Service struct {
	redis *redis.Client
}

// NewService 생성자
func NewService(redisClient *redis.Client) *Service {
	return &Service{redis: redisClient}
}

// GetTopicData 도메인별 전체 토픽 데이터 조회
func (s *Service) GetTopicData(ctx context.Context, domain string) (*model.TopicData, error) {
	key := redisKeyPrefix + domain

	data, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("토픽 데이터 없음: %s", domain)
	}
	if err != nil {
		return nil, fmt.Errorf("Redis 조회 실패: %w", err)
	}

	var topicData model.TopicData
	if err := json.Unmarshal([]byte(data), &topicData); err != nil {
		return nil, fmt.Errorf("토픽 데이터 파싱 실패: %w", err)
	}

	return &topicData, nil
}

// GetRandomContent 도메인에서 랜덤 작품 1개 선택
func (s *Service) GetRandomContent(ctx context.Context, domain string) (*model.TopicContent, error) {
	topicData, err := s.GetTopicData(ctx, domain)
	if err != nil {
		return nil, err
	}

	if len(topicData.Contents) == 0 {
		return nil, fmt.Errorf("해당 도메인에 콘텐츠 없음: %s", domain)
	}

	idx := rand.Intn(len(topicData.Contents))
	return &topicData.Contents[idx], nil
}
