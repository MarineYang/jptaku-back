package sentence

import (
	"fmt"
	"time"

	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/pkg"
)

// GetTodaySentences 오늘의 5문장 조회 (없으면 생성)
func (s *Service) GetTodaySentences(userID uint) (*DailySentencesResponse, error) {
	today := time.Now().Truncate(24 * time.Hour)
	return s.getSentencesByDate(userID, today)
}

// ResetTodaySentences 모든 문장 세트 삭제 (난이도/카테고리 변경 시 완전 초기화)
func (s *Service) ResetTodaySentences(userID uint) error {
	return s.sentenceRepo.DeleteAllDailySets(userID)
}

// GetGuestSentences 비회원용 N5 랜덤 5문장 (DB 저장 없음)
func (s *Service) GetGuestSentences() (*DailySentencesResponse, error) {
	sentences, err := s.sentenceRepo.FindRandom([]int{5}, nil, 5, nil)
	if err != nil {
		return nil, err
	}

	sentencesWithDetail := s.buildSentencesWithDetail(0, sentences)
	return &DailySentencesResponse{
		Date:      time.Now().Format("2006-01-02"),
		Sentences: sentencesWithDetail,
	}, nil
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

// createDailySet 오늘의 문장 세트 생성 (미완료 이월 + 순차 제공)
func (s *Service) createDailySet(userID uint, date time.Time) (*DailySentencesResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	onboardingLevel := 5 // 기본값 N5
	var categories []int
	if user.Onboarding != nil {
		onboardingLevel = user.Onboarding.Level
		categories = user.Onboarding.Categories
	}

	// 온보딩 값을 DB 쿼리용으로 변환
	levels := pkg.LevelsForUser(onboardingLevel)
	domains := pkg.DomainsFromCategories(categories)

	const dailyCount = 5

	// 1. 과거 미완료(Memorized=false) 이월 문장 조회
	var carryOverIDs []uint
	if s.learningRepo != nil {
		carryOverIDs, err = s.learningRepo.GetUnmemorizedSentenceIDs(userID, date, levels)
		if err != nil {
			carryOverIDs = []uint{}
		}
	}

	// 2. 새로 뽑을 문장 수 결정
	newCount := dailyCount - len(carryOverIDs)
	if newCount < 0 {
		newCount = 0
	}

	// 3. Memorized=true인 문장 제외 대상 조회
	memorizedIDs, err := s.sentenceRepo.GetUserMemorizedSentenceIDs(userID)
	if err != nil {
		memorizedIDs = []uint{}
	}

	// excludeIDs = memorizedIDs + carryOverIDs (중복 방지)
	excludeSet := make(map[uint]bool)
	for _, id := range memorizedIDs {
		excludeSet[id] = true
	}
	for _, id := range carryOverIDs {
		excludeSet[id] = true
	}
	excludeIDs := make([]uint, 0, len(excludeSet))
	for id := range excludeSet {
		excludeIDs = append(excludeIDs, id)
	}

	// 4. 새 문장 순차 선택 (N5→N4→N3)
	var newSentences []model.Sentence
	if newCount > 0 {
		newSentences, err = s.sentenceRepo.FindSequential(levels, domains, newCount, excludeIDs)
		if err != nil {
			return nil, fmt.Errorf("문장 조회 실패: %w", err)
		}
	}

	// 5. 이월 문장 조회
	var carryOverSentences []model.Sentence
	if len(carryOverIDs) > 0 {
		carryOverSentences, err = s.sentenceRepo.FindByIDs(carryOverIDs)
		if err != nil {
			carryOverSentences = []model.Sentence{}
		}
	}

	// 6. 최종 문장 = 새 문장 + 이월 문장
	allSentences := append(newSentences, carryOverSentences...)
	if len(allSentences) == 0 {
		return nil, fmt.Errorf("조건에 맞는 문장이 없습니다. 문장 pool이 비어있거나 모든 문장을 학습했습니다")
	}

	// 상세 정보 조회 및 ID 수집
	sentencesWithDetail := make([]SentenceWithDetail, 0, len(allSentences))
	sentenceIDs := make([]uint, 0, len(allSentences))

	for _, sentence := range allSentences {
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
