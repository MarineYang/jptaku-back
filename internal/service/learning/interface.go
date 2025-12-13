package learning

import "github.com/jptaku/server/internal/model"

// LearningRepository 학습 저장소 인터페이스
type LearningRepository interface {
	FindByUserAndSentence(userID, sentenceID uint) (*model.LearningProgress, error)
	Create(progress *model.LearningProgress) error
	Update(progress *model.LearningProgress) error
	GetTodayProgress(userID, dailySetID uint) ([]model.LearningProgress, error)
	GetUserProgress(userID uint, page, perPage int) ([]model.LearningProgress, int64, error)
}

// SentenceRepository 문장 저장소 인터페이스
type SentenceRepository interface {
	GetDetail(sentenceID uint) (*model.SentenceDetail, error)
}

// Provider 서비스 인터페이스 (외부에서 사용)
type Provider interface {
	UpdateProgress(userID uint, input *UpdateProgressInput) (*model.LearningProgress, error)
	GetTodayProgress(userID, dailySetID uint) (*TodayProgressResponse, error)
	GetProgress(userID uint, page, perPage int) ([]model.LearningProgress, int64, error)
	SubmitQuiz(userID uint, input *SubmitQuizInput) (*SubmitQuizResult, error)
}
