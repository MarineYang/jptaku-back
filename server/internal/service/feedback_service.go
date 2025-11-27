package service

import (
	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/repository"
)

type FeedbackService struct {
	feedbackRepo *repository.FeedbackRepository
	chatRepo     *repository.ChatRepository
}

func NewFeedbackService(feedbackRepo *repository.FeedbackRepository, chatRepo *repository.ChatRepository) *FeedbackService {
	return &FeedbackService{
		feedbackRepo: feedbackRepo,
		chatRepo:     chatRepo,
	}
}

func (s *FeedbackService) GetFeedback(sessionID uint) (*model.Feedback, error) {
	return s.feedbackRepo.FindBySessionID(sessionID)
}

func (s *FeedbackService) CreateFeedback(sessionID uint, feedback *model.Feedback) (*model.Feedback, error) {
	feedback.SessionID = sessionID

	if err := s.feedbackRepo.Create(feedback); err != nil {
		return nil, err
	}

	return feedback, nil
}

type StatsResponse struct {
	TotalSessions        int64   `json:"total_sessions"`
	TotalLearningMinutes int64   `json:"total_learning_minutes"`
	TotalSentencesUsed   int64   `json:"total_sentences_used"`
	AverageScore         float64 `json:"average_score"`
	CurrentStreak        int     `json:"current_streak"`
}

type CategoryProgress struct {
	Category string  `json:"category"`
	Progress float64 `json:"progress"` // 0 ~ 100
	Count    int     `json:"count"`
}

type WeeklyStats struct {
	Date             string `json:"date"`
	SessionCount     int    `json:"session_count"`
	SentencesLearned int    `json:"sentences_learned"`
	MinutesSpent     int    `json:"minutes_spent"`
}

func (s *FeedbackService) GetTodayStats(userID uint) (*StatsResponse, error) {
	// TODO: 실제 통계 계산 구현
	// 현재는 목업 데이터 반환
	return &StatsResponse{
		TotalSessions:        0,
		TotalLearningMinutes: 0,
		TotalSentencesUsed:   0,
		AverageScore:         0,
		CurrentStreak:        0,
	}, nil
}

func (s *FeedbackService) GetCategoryProgress(userID uint) ([]CategoryProgress, error) {
	// TODO: 실제 카테고리별 진행도 계산 구현
	// 현재는 목업 데이터 반환
	return []CategoryProgress{
		{Category: "애니", Progress: 0, Count: 0},
		{Category: "게임", Progress: 0, Count: 0},
		{Category: "성지순례", Progress: 0, Count: 0},
		{Category: "이벤트", Progress: 0, Count: 0},
	}, nil
}

func (s *FeedbackService) GetWeeklyStats(userID uint) ([]WeeklyStats, error) {
	// TODO: 실제 주간 통계 계산 구현
	// 현재는 목업 데이터 반환
	return []WeeklyStats{}, nil
}
