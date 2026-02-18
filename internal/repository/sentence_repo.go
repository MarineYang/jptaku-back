package repository

import (
	"time"

	"github.com/jptaku/server/internal/model"
	"gorm.io/gorm"
)

type SentenceRepository struct {
	db *gorm.DB
}

func NewSentenceRepository(db *gorm.DB) *SentenceRepository {
	return &SentenceRepository{db: db}
}

func (r *SentenceRepository) FindByID(id uint) (*model.Sentence, error) {
	var sentence model.Sentence
	err := r.db.First(&sentence, id).Error
	if err != nil {
		return nil, err
	}
	return &sentence, nil
}

func (r *SentenceRepository) FindByIDs(ids []uint) ([]model.Sentence, error) {
	var sentences []model.Sentence
	err := r.db.Where("id IN ?", ids).Find(&sentences).Error
	if err != nil {
		return nil, err
	}
	return sentences, nil
}

// FindRandom 조건에 맞는 랜덤 문장 조회
func (r *SentenceRepository) FindRandom(levels []int, domains []string, limit int, excludeIDs []uint) ([]model.Sentence, error) {
	var sentences []model.Sentence
	query := r.db.Model(&model.Sentence{})

	if len(levels) > 0 {
		query = query.Where("level IN ?", levels)
	}

	if len(domains) > 0 {
		query = query.Where("domain IN ?", domains)
	}

	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}

	err := query.Order("RANDOM()").Limit(limit).Find(&sentences).Error
	if err != nil {
		return nil, err
	}
	return sentences, nil
}

// FindSequential 레벨 오름차순으로 문장 조회 (N5→N4→N3)
func (r *SentenceRepository) FindSequential(levels []int, domains []string, limit int, excludeIDs []uint) ([]model.Sentence, error) {
	var sentences []model.Sentence
	query := r.db.Model(&model.Sentence{})

	if len(levels) > 0 {
		query = query.Where("level IN ?", levels)
	}

	if len(domains) > 0 {
		query = query.Where("domain IN ?", domains)
	}

	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}

	err := query.Order("level DESC, id ASC").Limit(limit).Find(&sentences).Error
	if err != nil {
		return nil, err
	}
	return sentences, nil
}

// GetUserMemorizedSentenceIDs Memorized=true인 문장 ID만 반환
func (r *SentenceRepository) GetUserMemorizedSentenceIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&model.LearningProgress{}).
		Select("sentence_id").
		Where("user_id = ? AND memorized = true", userID).
		Find(&ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// FindByUUID UUID로 문장 조회
func (r *SentenceRepository) FindByUUID(uuid string) (*model.Sentence, error) {
	var sentence model.Sentence
	err := r.db.Where("uuid = ?", uuid).First(&sentence).Error
	if err != nil {
		return nil, err
	}
	return &sentence, nil
}

func (r *SentenceRepository) GetDetail(sentenceID uint) (*model.SentenceDetail, error) {
	var detail model.SentenceDetail
	err := r.db.Where("sentence_id = ?", sentenceID).First(&detail).Error
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (r *SentenceRepository) Create(sentence *model.Sentence) error {
	return r.db.Create(sentence).Error
}

func (r *SentenceRepository) CreateDetail(detail *model.SentenceDetail) error {
	return r.db.Create(detail).Error
}

func (r *SentenceRepository) GetHistory(userID uint, page, perPage int) ([]model.Sentence, int64, error) {
	var sentences []model.Sentence
	var total int64

	subQuery := r.db.Model(&model.DailySentenceSet{}).Select("jsonb_array_elements_text(sentence_ids)::int").Where("user_id = ?", userID)
	query := r.db.Model(&model.Sentence{}).Where("id IN (?)", subQuery)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Offset(offset).Limit(perPage).Find(&sentences).Error
	if err != nil {
		return nil, 0, err
	}

	return sentences, total, nil
}

// DailySentenceSet methods
func (r *SentenceRepository) GetDailySet(userID uint, date time.Time) (*model.DailySentenceSet, error) {
	var dailySet model.DailySentenceSet
	err := r.db.Where("user_id = ? AND date = ?", userID, date.Format("2006-01-02")).First(&dailySet).Error
	if err != nil {
		return nil, err
	}
	return &dailySet, nil
}

func (r *SentenceRepository) CreateDailySet(dailySet *model.DailySentenceSet) error {
	return r.db.Create(dailySet).Error
}

// DeleteDailySet 특정 날짜의 DailySentenceSet 삭제
func (r *SentenceRepository) DeleteDailySet(userID uint, date time.Time) error {
	return r.db.Where("user_id = ? AND date = ?", userID, date.Format("2006-01-02")).Delete(&model.DailySentenceSet{}).Error
}

// DeleteAllDailySets 유저의 모든 DailySentenceSet 삭제 (난이도 변경 시 완전 초기화)
func (r *SentenceRepository) DeleteAllDailySets(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.DailySentenceSet{}).Error
}

func (r *SentenceRepository) GetUserLearnedSentenceIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&model.DailySentenceSet{}).Select("jsonb_array_elements_text(sentence_ids)::int").Where("user_id = ?", userID).Find(&ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// CountUserDailySets 유저의 총 학습일 수 (DailySentenceSet 개수)
func (r *SentenceRepository) CountUserDailySets(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.DailySentenceSet{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *SentenceRepository) GetPastDailySets(userID uint, page, perPage int) ([]model.DailySentenceSet, int64, error) {
	var dailySets []model.DailySentenceSet
	var total int64

	today := time.Now().Truncate(24 * time.Hour)
	query := r.db.Model(&model.DailySentenceSet{}).Where("user_id = ? AND date < ?", userID, today.Format("2006-01-02"))
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("date DESC").Offset(offset).Limit(perPage).Find(&dailySets).Error
	if err != nil {
		return nil, 0, err
	}

	return dailySets, total, nil
}
