package feedback

import "github.com/jptaku/server/internal/model"

// FeedbackRepository 피드백 저장소 인터페이스
type FeedbackRepository interface {
	FindBySessionID(sessionID uint) (*model.Feedback, error)
	Create(feedback *model.Feedback) error
}

// ChatRepository 채팅 저장소 인터페이스
type ChatRepository interface {
	// 필요한 메서드가 있으면 추가
}

// Provider 서비스 인터페이스 (외부에서 사용)
type Provider interface {
	GetFeedback(sessionID uint) (*model.Feedback, error)
	CreateFeedback(sessionID uint, feedback *model.Feedback) (*model.Feedback, error)
	GetTodayStats(userID uint) (*StatsResponse, error)
	GetCategoryProgress(userID uint) ([]CategoryProgress, error)
	GetWeeklyStats(userID uint) ([]WeeklyStats, error)
}
