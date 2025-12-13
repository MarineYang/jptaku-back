package learning

import (
	"time"

	"github.com/jptaku/server/internal/model"
	"gorm.io/gorm"
)

// Service 학습 서비스
type Service struct {
	learningRepo LearningRepository
	sentenceRepo SentenceRepository
}

// 컴파일 타임 인터페이스 검증
var _ Provider = (*Service)(nil)

// NewService 서비스 생성자
func NewService(learningRepo LearningRepository, sentenceRepo SentenceRepository) *Service {
	return &Service{
		learningRepo: learningRepo,
		sentenceRepo: sentenceRepo,
	}
}

// UpdateProgress 진행 상황 업데이트
func (s *Service) UpdateProgress(userID uint, input *UpdateProgressInput) (*model.LearningProgress, error) {
	progress, err := s.learningRepo.FindByUserAndSentence(userID, input.SentenceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			progress = &model.LearningProgress{
				UserID:     userID,
				SentenceID: input.SentenceID,
				DailySetID: input.DailySetID,
			}
		} else {
			return nil, err
		}
	}

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

// GetTodayProgress 오늘의 진행 상황 조회
func (s *Service) GetTodayProgress(userID, dailySetID uint) (*TodayProgressResponse, error) {
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

// GetProgress 진행 상황 조회
func (s *Service) GetProgress(userID uint, page, perPage int) ([]model.LearningProgress, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	return s.learningRepo.GetUserProgress(userID, page, perPage)
}

// SubmitQuiz 퀴즈 제출 및 정답 검증
func (s *Service) SubmitQuiz(userID uint, input *SubmitQuizInput) (*SubmitQuizResult, error) {
	detail, err := s.sentenceRepo.GetDetail(input.SentenceID)
	if err != nil {
		return nil, err
	}

	if detail == nil || detail.Quiz == nil {
		return nil, err
	}

	fillBlankCorrect := false
	if detail.Quiz.FillBlank != nil {
		fillBlankCorrect = detail.Quiz.FillBlank.Answer == input.FillBlankAnswer
	}

	orderingCorrect := false
	if detail.Quiz.Ordering != nil {
		orderingCorrect = compareIntSlices(detail.Quiz.Ordering.CorrectOrder, input.OrderingAnswer)
	}

	allCorrect := fillBlankCorrect && orderingCorrect

	progress, err := s.learningRepo.FindByUserAndSentence(userID, input.SentenceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			progress = &model.LearningProgress{
				UserID:     userID,
				SentenceID: input.SentenceID,
				DailySetID: input.DailySetID,
			}
		} else {
			return nil, err
		}
	}

	progress.Confirm = allCorrect

	if allCorrect {
		progress.Memorized = true
		now := time.Now()
		progress.CompletedAt = &now
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

	return &SubmitQuizResult{
		SentenceID:       input.SentenceID,
		FillBlankCorrect: fillBlankCorrect,
		OrderingCorrect:  orderingCorrect,
		AllCorrect:       allCorrect,
		Memorized:        progress.Memorized,
	}, nil
}

// compareIntSlices 두 int 슬라이스가 같은지 비교
func compareIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
