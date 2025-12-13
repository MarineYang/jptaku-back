package service

import (
	"fmt"
	"time"

	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/repository"
)

type SentenceService struct {
	sentenceRepo *repository.SentenceRepository
	userRepo     *repository.UserRepository
	learningRepo *repository.LearningRepository
}

func NewSentenceService(sentenceRepo *repository.SentenceRepository, userRepo *repository.UserRepository) *SentenceService {
	return &SentenceService{
		sentenceRepo: sentenceRepo,
		userRepo:     userRepo,
	}
}

func (s *SentenceService) SetLearningRepo(learningRepo *repository.LearningRepository) {
	s.learningRepo = learningRepo
}

// SentenceWithDetail 문장 + 상세 정보
type SentenceWithDetail struct {
	model.Sentence
	Words     []model.Word `json:"words"`
	Grammar   []string     `json:"grammar"`
	Examples  []string     `json:"examples"`
	Quiz      *model.Quiz  `json:"quiz"`
	Memorized bool         `json:"memorized"` // 암기 완료 여부
}

// DailySentencesResponse 오늘의 5문장 응답
type DailySentencesResponse struct {
	Date      string               `json:"date"`
	Sentences []SentenceWithDetail `json:"sentences"`
}

// GetTodaySentences 오늘의 5문장 조회 (없으면 생성)
func (s *SentenceService) GetTodaySentences(userID uint) (*DailySentencesResponse, error) {
	today := time.Now().Truncate(24 * time.Hour)
	return s.getSentencesByDate(userID, today)
}

// HistoryItem 지난 학습 기록 아이템
type HistoryItem struct {
	Date      string               `json:"date"`
	Sentences []SentenceWithDetail `json:"sentences"`
}

// HistorySentencesResponse 지난 학습 문장 응답
type HistorySentencesResponse struct {
	History    []HistoryItem `json:"history"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	Total      int64         `json:"total"`
	TotalPages int           `json:"total_pages"`
}

// GetHistorySentences 지난 학습 문장 조회 (오늘 제외)
func (s *SentenceService) GetHistorySentences(userID uint, page, perPage int) (*HistorySentencesResponse, error) {
	dailySets, total, err := s.sentenceRepo.GetPastDailySets(userID, page, perPage)
	if err != nil {
		return nil, err
	}

	history := make([]HistoryItem, 0, len(dailySets))
	for _, dailySet := range dailySets {
		sentences, err := s.sentenceRepo.FindByIDs(dailySet.SentenceIDs)
		if err != nil {
			continue
		}

		// 상세 정보 + 학습 상태 조회
		sentencesWithDetail := make([]SentenceWithDetail, 0, len(sentences))
		for _, sentence := range sentences {
			detail, _ := s.sentenceRepo.GetDetail(sentence.ID)
			swd := SentenceWithDetail{
				Sentence: sentence,
			}
			if detail != nil {
				swd.Words = detail.Words
				swd.Grammar = detail.Grammar
				swd.Examples = detail.Examples
				swd.Quiz = detail.Quiz
			}

			// 학습 상태 조회
			if s.learningRepo != nil {
				progress, _ := s.learningRepo.FindByUserAndSentence(userID, sentence.ID)
				if progress != nil {
					swd.Memorized = progress.Memorized
				}
			}

			sentencesWithDetail = append(sentencesWithDetail, swd)
		}

		history = append(history, HistoryItem{
			Date:      dailySet.Date.Format("2006-01-02"),
			Sentences: sentencesWithDetail,
		})
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &HistorySentencesResponse{
		History:    history,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *SentenceService) getSentencesByDate(userID uint, date time.Time) (*DailySentencesResponse, error) {
	// 해당 날짜의 세트가 있는지 확인
	dailySet, err := s.sentenceRepo.GetDailySet(userID, date)
	if err == nil && dailySet != nil {
		sentences, err := s.sentenceRepo.FindByIDs(dailySet.SentenceIDs)
		if err != nil {
			return nil, err
		}

		// 상세 정보 + 학습 상태 조회
		sentencesWithDetail := make([]SentenceWithDetail, 0, len(sentences))
		for _, sentence := range sentences {
			detail, _ := s.sentenceRepo.GetDetail(sentence.ID)
			swd := SentenceWithDetail{
				Sentence: sentence,
			}
			if detail != nil {
				swd.Words = detail.Words
				swd.Grammar = detail.Grammar
				swd.Examples = detail.Examples
				swd.Quiz = detail.Quiz
			}

			// 학습 상태 조회
			if s.learningRepo != nil {
				progress, _ := s.learningRepo.FindByUserAndSentence(userID, sentence.ID)
				if progress != nil {
					swd.Memorized = progress.Memorized
				}
			}

			sentencesWithDetail = append(sentencesWithDetail, swd)
		}

		return &DailySentencesResponse{
			Date:      date.Format("2006-01-02"),
			Sentences: sentencesWithDetail,
		}, nil
	}

	// 오늘이 아니면 생성하지 않음
	today := time.Now().Truncate(24 * time.Hour)
	if !date.Equal(today) {
		return nil, fmt.Errorf("해당 날짜의 문장이 없습니다")
	}

	// 오늘 세트가 없으면 새로 생성
	return s.createDailySet(userID, date)
}

func (s *SentenceService) createDailySet(userID uint, date time.Time) (*DailySentencesResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	level := 1
	var interests []int
	if user.Onboarding != nil {
		level = user.Onboarding.Level
		interests = user.Onboarding.Interests
	}

	// 이미 학습한 문장 ID 조회
	learnedIDs, err := s.sentenceRepo.GetUserLearnedSentenceIDs(userID)
	if err != nil {
		learnedIDs = []uint{} // 에러 시 빈 배열로 진행
	}

	// 미리 생성된 문장 pool에서 조건에 맞는 5개 가져오기
	// - level 이하의 문장
	// - interests(SubCategory)에 해당하는 문장
	// - 이미 학습한 문장 제외
	sentences, err := s.sentenceRepo.FindRandom(level, interests, 5, learnedIDs)
	if err != nil {
		return nil, fmt.Errorf("문장 조회 실패: %w", err)
	}

	if len(sentences) == 0 {
		return nil, fmt.Errorf("조건에 맞는 문장이 없습니다. 문장 pool이 비어있거나 모든 문장을 학습했습니다")
	}

	// 상세 정보 조회
	sentencesWithDetail := make([]SentenceWithDetail, 0, len(sentences))
	sentenceIDs := make([]uint, 0, len(sentences))

	for _, sentence := range sentences {
		detail, _ := s.sentenceRepo.GetDetail(sentence.ID)
		swd := SentenceWithDetail{
			Sentence: sentence,
		}
		if detail != nil {
			swd.Words = detail.Words
			swd.Grammar = detail.Grammar
			swd.Examples = detail.Examples
			swd.Quiz = detail.Quiz
		}
		sentencesWithDetail = append(sentencesWithDetail, swd)
		sentenceIDs = append(sentenceIDs, sentence.ID)
	}

	// DailySentenceSet 저장
	dailySet := &model.DailySentenceSet{
		UserID:      userID,
		Date:        date,
		SentenceIDs: sentenceIDs,
	}

	if err := s.sentenceRepo.CreateDailySet(dailySet); err != nil {
		return nil, err
	}

	return &DailySentencesResponse{
		Date:      date.Format("2006-01-02"),
		Sentences: sentencesWithDetail,
	}, nil
}

