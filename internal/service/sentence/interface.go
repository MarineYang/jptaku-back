package sentence

import (
	"time"

	"github.com/jptaku/server/internal/model"
)

// SentenceRepository 문장 저장소 인터페이스
type SentenceRepository interface {
	GetDailySet(userID uint, date time.Time) (*model.DailySentenceSet, error)
	GetPastDailySets(userID uint, page, perPage int) ([]model.DailySentenceSet, int64, error)
	FindByIDs(ids []uint) ([]model.Sentence, error)
	FindRandom(level int, interests []int, count int, excludeIDs []uint) ([]model.Sentence, error)
	GetDetail(sentenceID uint) (*model.SentenceDetail, error)
	GetUserLearnedSentenceIDs(userID uint) ([]uint, error)
	CreateDailySet(dailySet *model.DailySentenceSet) error
}

// UserRepository 사용자 저장소 인터페이스
type UserRepository interface {
	FindByID(id uint) (*model.User, error)
}

// LearningRepository 학습 저장소 인터페이스
type LearningRepository interface {
	FindByUserAndSentence(userID, sentenceID uint) (*model.LearningProgress, error)
}

// Provider 서비스 인터페이스 (외부에서 사용)
type Provider interface {
	GetTodaySentences(userID uint) (*DailySentencesResponse, error)
	GetHistorySentences(userID uint, page, perPage int) (*HistorySentencesResponse, error)
	SetLearningRepo(learningRepo LearningRepository)
}
