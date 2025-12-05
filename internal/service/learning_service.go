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

// SubmitQuizInput 퀴즈 제출 입력
type SubmitQuizInput struct {
	SentenceID      uint
	DailySetID      uint
	FillBlankAnswer string
	OrderingAnswer  []int
}

// SubmitQuizResult 퀴즈 제출 결과
type SubmitQuizResult struct {
	SentenceID       uint
	FillBlankCorrect bool
	OrderingCorrect  bool
	AllCorrect       bool
	Memorized        bool
}

// SubmitQuiz 퀴즈 제출 및 정답 검증
func (s *LearningService) SubmitQuiz(userID uint, input *SubmitQuizInput) (*SubmitQuizResult, error) {
	// 1. 문장 상세 정보 조회 (정답 확인용)
	detail, err := s.sentenceRepo.GetDetail(input.SentenceID)
	if err != nil {
		return nil, err
	}

	if detail == nil || detail.Quiz == nil {
		return nil, err
	}

	// 2. 빈칸 채우기 정답 확인
	fillBlankCorrect := false
	if detail.Quiz.FillBlank != nil {
		fillBlankCorrect = detail.Quiz.FillBlank.Answer == input.FillBlankAnswer
	}

	// 3. 문장 배열 정답 확인
	orderingCorrect := false
	if detail.Quiz.Ordering != nil {
		orderingCorrect = compareIntSlices(detail.Quiz.Ordering.CorrectOrder, input.OrderingAnswer)
	}

	// 4. 모두 정답인지 확인
	allCorrect := fillBlankCorrect && orderingCorrect

	// 5. 학습 진행 상황 업데이트
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

	// 확인하기(퀴즈) 완료 표시
	progress.Confirm = allCorrect

	// 모두 정답이면 암기 완료로 표시
	if allCorrect {
		progress.Memorized = true
		now := time.Now()
		progress.CompletedAt = &now
	}

	// 저장
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
