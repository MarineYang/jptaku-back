package service

import (
	"time"

	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/repository"
	"gorm.io/gorm"
)

type LearningService struct {
	learningRepo *repository.LearningRepository
	sentenceRepo *repository.SentenceRepository
}

func NewLearningService(learningRepo *repository.LearningRepository, sentenceRepo *repository.SentenceRepository) *LearningService {
	return &LearningService{
		learningRepo: learningRepo,
		sentenceRepo: sentenceRepo,
	}
}

type UpdateProgressInput struct {
	SentenceID uint  `json:"sentence_id" binding:"required"`
	DailySetID uint  `json:"daily_set_id"`
	Understand *bool `json:"understand,omitempty"`
	Speak      *bool `json:"speak,omitempty"`
	Confirm    *bool `json:"confirm,omitempty"`
	Memorized  *bool `json:"memorized,omitempty"`
}

type TodayProgressResponse struct {
	DailySetID     uint                     `json:"daily_set_id"`
	TotalSentences int                      `json:"total_sentences"`
	Completed      int                      `json:"completed"`
	Progress       []model.LearningProgress `json:"progress"`
}

func (s *LearningService) UpdateProgress(userID uint, input *UpdateProgressInput) (*model.LearningProgress, error) {
	// 기존 진행 상황 확인
	progress, err := s.learningRepo.FindByUserAndSentence(userID, input.SentenceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 새로 생성
			progress = &model.LearningProgress{
				UserID:     userID,
				SentenceID: input.SentenceID,
				DailySetID: input.DailySetID,
			}
		} else {
			return nil, err
		}
	}

	// 진행 상황 업데이트
	if input.Understand != nil {
		progress.Understand = *input.Understand
	}
	if input.Speak != nil {
		progress.Speak = *input.Speak
	}
	if input.Confirm != nil {
		progress.Confirm = *input.Confirm
	}
	if input.Memorized != nil {
		progress.Memorized = *input.Memorized
		if *input.Memorized && progress.CompletedAt == nil {
			now := time.Now()
			progress.CompletedAt = &now
		}
	}

	if progress.ID == 0 {
		if err := s.learningRepo.Create(progress); err != nil {
			return nil, err
		}
	} else {
		if err := s.learningRepo.Update(progress); err != nil {
			return nil, err
		}
	}

	return progress, nil
}

func (s *LearningService) GetTodayProgress(userID uint, dailySetID uint) (*TodayProgressResponse, error) {
	progresses, err := s.learningRepo.GetTodayProgress(userID, dailySetID)
	if err != nil {
		return nil, err
	}

	completed := 0
	for _, p := range progresses {
		if p.Memorized {
			completed++
		}
	}

	return &TodayProgressResponse{
		DailySetID:     dailySetID,
		TotalSentences: 5,
		Completed:      completed,
		Progress:       progresses,
	}, nil
}

func (s *LearningService) GetProgress(userID uint, page, perPage int) ([]model.LearningProgress, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	return s.learningRepo.GetUserProgress(userID, page, perPage)
}
