package flash

import (
	"context"
	"time"

	"github.com/jptaku/server/internal/model"
)

// SentenceRepository 문장 저장소 인터페이스
type SentenceRepository interface {
	GetDailySet(userID uint, date time.Time) (*model.DailySentenceSet, error)
	FindByIDs(ids []uint) ([]model.Sentence, error)
	GetDetail(sentenceID uint) (*model.SentenceDetail, error)
}

// LearningRepository 학습 저장소 인터페이스
type LearningRepository interface {
	FindByUserAndSentence(userID, sentenceID uint) (*model.LearningProgress, error)
	Create(progress *model.LearningProgress) error
	Update(progress *model.LearningProgress) error
}

// Provider 서비스 인터페이스 (외부에서 사용)
type Provider interface {
	GetTodayFlash(ctx context.Context, userID uint) (*TodayFlashResponse, error)
	UpdateFlashProgress(ctx context.Context, userID uint, input *UpdateFlashInput) (*FlashProgressResult, error)
}
