package flash

import (
	"context"
	"log"
	"time"

	"github.com/jptaku/server/internal/model"
	"gorm.io/gorm"
)

// SRS 간격 설정
const (
	ReviewIntervalBad  = 10 * time.Minute // bad: 10분 후
	ReviewIntervalMid  = 1 * time.Hour    // mid: 1시간 후
	ReviewIntervalGood = 24 * time.Hour   // good: 오늘은 끝 (내일)
)

// Service Flash 서비스 구현체
type Service struct {
	sentenceRepo SentenceRepository
	learningRepo LearningRepository
}

// 컴파일 타임에 인터페이스 구현 확인
var _ Provider = (*Service)(nil)

// NewService Flash 서비스 생성자
func NewService(
	sentenceRepo SentenceRepository,
	learningRepo LearningRepository,
) *Service {
	return &Service{
		sentenceRepo: sentenceRepo,
		learningRepo: learningRepo,
	}
}

// GetTodayFlash 오늘의 Flash 문장 조회 (복습 대상만 반환)
func (s *Service) GetTodayFlash(ctx context.Context, userID uint) (*TodayFlashResponse, error) {
	today := time.Now().Truncate(24 * time.Hour)
	now := time.Now()

	// 오늘의 DailySet 조회
	dailySet, err := s.sentenceRepo.GetDailySet(userID, today)
	if err != nil {
		return nil, err
	}

	// 문장 조회
	sentences, err := s.sentenceRepo.FindByIDs(dailySet.SentenceIDs)
	if err != nil {
		return nil, err
	}

	// Flash 문장 목록 생성 (복습 대상만 필터링)
	flashSentences := make([]FlashSentence, 0, len(sentences))

	for _, sentence := range sentences {
		flashSentence, shouldShow := s.buildFlashSentenceWithFilter(userID, sentence, now)
		if shouldShow {
			flashSentences = append(flashSentences, flashSentence)
		}
	}

	return &TodayFlashResponse{
		Date:      today.Format("2006-01-02"),
		Sentences: flashSentences,
	}, nil
}

// buildFlashSentenceWithFilter Flash용 문장 빌드 + 복습 대상 여부 판단
func (s *Service) buildFlashSentenceWithFilter(userID uint, sentence model.Sentence, now time.Time) (FlashSentence, bool) {
	flashSentence := FlashSentence{
		Sentence: sentence,
	}

	// SentenceDetail 조회 (phrase/tip/alt는 cron job에서 미리 생성됨)
	detail, err := s.sentenceRepo.GetDetail(sentence.ID)
	if err != nil {
		log.Printf("Failed to get detail for sentence %d: %v", sentence.ID, err)
	}

	if detail != nil {
		flashSentence.Phrase = detail.Phrase
		flashSentence.Tip = detail.Tip
		flashSentence.Alt = detail.Alt
	}

	// LearningProgress에서 Flash 상태 조회
	progress, err := s.learningRepo.FindByUserAndSentence(userID, sentence.ID)
	if err != nil {
		// Progress가 없으면 아직 학습 안 한 문장 → 복습 대상
		return flashSentence, true
	}

	if progress.FlashGrade != nil {
		flashSentence.FlashGrade = *progress.FlashGrade
	}
	flashSentence.FlashCount = progress.FlashCount

	if progress.NextReviewAt != nil {
		flashSentence.NextReviewAt = progress.NextReviewAt
	}

	// 복습 대상 판단: NextReviewAt이 없거나 현재 시간 이전이면 복습 대상
	shouldShow := progress.NextReviewAt == nil || now.After(*progress.NextReviewAt) || now.Equal(*progress.NextReviewAt)

	return flashSentence, shouldShow
}

// UpdateFlashProgress Flash 진행 상황 업데이트 (SRS 포함)
func (s *Service) UpdateFlashProgress(ctx context.Context, userID uint, input *UpdateFlashInput) (*FlashProgressResult, error) {
	now := time.Now()
	nextReview := calculateNextReview(now, input.Grade)

	// 기존 Progress 조회 또는 생성
	progress, err := s.learningRepo.FindByUserAndSentence(userID, input.SentenceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 새로 생성
			progress = &model.LearningProgress{
				UserID:         userID,
				SentenceID:     input.SentenceID,
				FlashGrade:     &input.Grade,
				FlashUpdatedAt: &now,
				FlashCount:     1,
				NextReviewAt:   &nextReview,
			}
			if err := s.learningRepo.Create(progress); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// 기존 Progress 업데이트
		progress.FlashGrade = &input.Grade
		progress.FlashUpdatedAt = &now
		progress.FlashCount++
		progress.NextReviewAt = &nextReview

		if err := s.learningRepo.Update(progress); err != nil {
			return nil, err
		}
	}

	return &FlashProgressResult{
		SentenceID:   input.SentenceID,
		Grade:        input.Grade,
		FlashCount:   progress.FlashCount,
		NextReviewAt: &nextReview,
	}, nil
}

// calculateNextReview grade에 따른 다음 복습 시간 계산
func calculateNextReview(now time.Time, grade string) time.Time {
	switch grade {
	case "bad":
		return now.Add(ReviewIntervalBad) // 10분 후
	case "mid":
		return now.Add(ReviewIntervalMid) // 1시간 후
	case "good":
		return now.Add(ReviewIntervalGood) // 내일
	default:
		return now.Add(ReviewIntervalMid)
	}
}
