package sentence

import (
	"fmt"
	"time"

	"github.com/jptaku/server/internal/model"
)

// GetTodaySentences 오늘의 5문장 조회 (없으면 생성)
func (s *Service) GetTodaySentences(userID uint) (*DailySentencesResponse, error) {
	today := time.Now().Truncate(24 * time.Hour)
	return s.getSentencesByDate(userID, today)
}

// GetHistorySentences 지난 학습 문장 조회 (오늘 제외)
func (s *Service) GetHistorySentences(userID uint, page, perPage int) (*HistorySentencesResponse, error) {
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

		sentencesWithDetail := s.buildSentencesWithDetail(userID, sentences)
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

// getSentencesByDate 특정 날짜의 문장 조회
func (s *Service) getSentencesByDate(userID uint, date time.Time) (*DailySentencesResponse, error) {
	// 해당 날짜의 세트가 있는지 확인
	dailySet, err := s.sentenceRepo.GetDailySet(userID, date)
	if err == nil && dailySet != nil {
		sentences, err := s.sentenceRepo.FindByIDs(dailySet.SentenceIDs)
		if err != nil {
			return nil, err
		}

		sentencesWithDetail := s.buildSentencesWithDetail(userID, sentences)
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

// createDailySet 오늘의 문장 세트 생성
func (s *Service) createDailySet(userID uint, date time.Time) (*DailySentencesResponse, error) {
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
		learnedIDs = []uint{}
	}

	// 미리 생성된 문장 pool에서 조건에 맞는 5개 가져오기
	sentences, err := s.sentenceRepo.FindRandom(level, interests, 5, learnedIDs)
	if err != nil {
		return nil, fmt.Errorf("문장 조회 실패: %w", err)
	}

	if len(sentences) == 0 {
		return nil, fmt.Errorf("조건에 맞는 문장이 없습니다. 문장 pool이 비어있거나 모든 문장을 학습했습니다")
	}

	// 상세 정보 조회 및 ID 수집
	sentencesWithDetail := make([]SentenceWithDetail, 0, len(sentences))
	sentenceIDs := make([]uint, 0, len(sentences))

	for _, sentence := range sentences {
		swd := s.buildSentenceWithDetail(userID, sentence)
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

// buildSentencesWithDetail 문장 목록에 상세 정보 추가
func (s *Service) buildSentencesWithDetail(userID uint, sentences []model.Sentence) []SentenceWithDetail {
	result := make([]SentenceWithDetail, 0, len(sentences))
	for _, sentence := range sentences {
		swd := s.buildSentenceWithDetail(userID, sentence)
		result = append(result, swd)
	}
	return result
}

// buildSentenceWithDetail 단일 문장에 상세 정보 추가
func (s *Service) buildSentenceWithDetail(userID uint, sentence model.Sentence) SentenceWithDetail {
	swd := SentenceWithDetail{
		Sentence: sentence,
	}

	// 상세 정보 조회
	if detail, _ := s.sentenceRepo.GetDetail(sentence.ID); detail != nil {
		swd.Words = detail.Words
		swd.Grammar = detail.Grammar
		swd.Examples = detail.Examples
		swd.Quiz = detail.Quiz
	}

	// 학습 상태 조회
	if s.learningRepo != nil {
		if progress, _ := s.learningRepo.FindByUserAndSentence(userID, sentence.ID); progress != nil {
			swd.Memorized = progress.Memorized
		}
	}

	return swd
}
