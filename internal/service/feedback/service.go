package feedback

import "github.com/jptaku/server/internal/model"

// Service 피드백 서비스
type Service struct {
	feedbackRepo FeedbackRepository
	chatRepo     ChatRepository
}

// 컴파일 타임 인터페이스 검증
var _ Provider = (*Service)(nil)

// NewService 서비스 생성자
func NewService(feedbackRepo FeedbackRepository, chatRepo ChatRepository) *Service {
	return &Service{
		feedbackRepo: feedbackRepo,
		chatRepo:     chatRepo,
	}
}

// GetFeedback 피드백 조회
func (s *Service) GetFeedback(sessionID uint) (*model.Feedback, error) {
	return s.feedbackRepo.FindBySessionID(sessionID)
}

// CreateFeedback 피드백 생성
func (s *Service) CreateFeedback(sessionID uint, feedback *model.Feedback) (*model.Feedback, error) {
	feedback.SessionID = sessionID

	if err := s.feedbackRepo.Create(feedback); err != nil {
		return nil, err
	}

	return feedback, nil
}

// GetTodayStats 오늘의 통계 조회
func (s *Service) GetTodayStats(userID uint) (*StatsResponse, error) {
	// TODO: 실제 통계 계산 구현
	return &StatsResponse{
		TotalSessions:        0,
		TotalLearningMinutes: 0,
		TotalSentencesUsed:   0,
		AverageScore:         0,
		CurrentStreak:        0,
	}, nil
}

// GetCategoryProgress 카테고리별 진행도 조회
func (s *Service) GetCategoryProgress(userID uint) ([]CategoryProgress, error) {
	// TODO: 실제 카테고리별 진행도 계산 구현
	return []CategoryProgress{
		{Category: "애니", Progress: 0, Count: 0},
		{Category: "게임", Progress: 0, Count: 0},
		{Category: "성지순례", Progress: 0, Count: 0},
		{Category: "이벤트", Progress: 0, Count: 0},
	}, nil
}

// GetWeeklyStats 주간 통계 조회
func (s *Service) GetWeeklyStats(userID uint) ([]WeeklyStats, error) {
	// TODO: 실제 주간 통계 계산 구현
	return []WeeklyStats{}, nil
}
