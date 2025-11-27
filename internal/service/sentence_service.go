package service

import (
	"time"

	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/repository"
	"gorm.io/gorm"
)

type SentenceService struct {
	db           *repository.DBManager
	sentenceRepo *repository.SentenceRepository
	userRepo     *repository.UserRepository
}

func NewSentenceService(db *repository.DBManager, sentenceRepo *repository.SentenceRepository, userRepo *repository.UserRepository) *SentenceService {
	return &SentenceService{
		db:           db,
		sentenceRepo: sentenceRepo,
		userRepo:     userRepo,
	}
}

type DailySentencesResponse struct {
	DailySetID uint             `json:"daily_set_id"`
	Date       string           `json:"date"`
	Sentences  []model.Sentence `json:"sentences"`
}

func (s *SentenceService) GetDailySentences(userID uint) (*DailySentencesResponse, error) {
	today := time.Now().Truncate(24 * time.Hour)

	// 오늘의 세트가 이미 있는지 확인
	dailySet, err := s.sentenceRepo.GetDailySet(userID, today)
	if err == nil && dailySet != nil {
		// 기존 세트 반환
		sentences, err := s.sentenceRepo.FindByIDs(dailySet.SentenceIDs)
		if err != nil {
			return nil, err
		}

		return &DailySentencesResponse{
			DailySetID: dailySet.ID,
			Date:       today.Format("2006-01-02"),
			Sentences:  sentences,
		}, nil
	}

	// 새 세트 생성
	return s.createDailySet(userID, today)
}

func (s *SentenceService) createDailySet(userID uint, date time.Time) (*DailySentencesResponse, error) {
	// 유저 정보 및 온보딩 데이터 가져오기
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	level := 1
	var tags []string
	if user.Onboarding != nil {
		level = user.Onboarding.Level
		tags = user.Onboarding.Interests
	}

	// 이미 학습한 문장 제외
	learnedIDs, _ := s.sentenceRepo.GetUserLearnedSentenceIDs(userID)

	// 랜덤으로 5개 문장 선택
	sentences, err := s.sentenceRepo.FindRandom(level, tags, 5, learnedIDs)
	if err != nil {
		return nil, err
	}

	// 문장이 부족한 경우
	if len(sentences) < 5 {
		// 이미 학습한 문장도 포함해서 가져오기
		additionalSentences, err := s.sentenceRepo.FindRandom(level, tags, 5-len(sentences), nil)
		if err == nil {
			sentences = append(sentences, additionalSentences...)
		}
	}

	// 문장 ID 추출
	sentenceIDs := make([]uint, len(sentences))
	for i, sentence := range sentences {
		sentenceIDs[i] = sentence.ID
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
		DailySetID: dailySet.ID,
		Date:       date.Format("2006-01-02"),
		Sentences:  sentences,
	}, nil
}

func (s *SentenceService) GetSentence(id uint) (*model.Sentence, error) {
	return s.sentenceRepo.FindByID(id)
}

func (s *SentenceService) GetSentenceDetail(id uint) (*model.Sentence, *model.SentenceDetail, error) {
	sentence, err := s.sentenceRepo.FindByID(id)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, nil, gorm.ErrRecordNotFound
		}
		return nil, nil, err
	}

	detail, err := s.sentenceRepo.GetDetail(id)
	if err != nil && !repository.IsNotFound(err) {
		return nil, nil, err
	}

	return sentence, detail, nil
}

func (s *SentenceService) GetHistory(userID uint, page, perPage int) ([]model.Sentence, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	return s.sentenceRepo.GetHistory(userID, page, perPage)
}
