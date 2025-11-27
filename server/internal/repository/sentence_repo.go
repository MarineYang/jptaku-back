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

func (r *SentenceRepository) FindByLevel(level int, limit int) ([]model.Sentence, error) {
	var sentences []model.Sentence
	err := r.db.Where("level = ?", level).Limit(limit).Find(&sentences).Error
	if err != nil {
		return nil, err
	}
	return sentences, nil
}

func (r *SentenceRepository) FindByTags(tags []string, limit int) ([]model.Sentence, error) {
	var sentences []model.Sentence
	err := r.db.Where("tags && ?", tags).Limit(limit).Find(&sentences).Error
	if err != nil {
		return nil, err
	}
	return sentences, nil
}

func (r *SentenceRepository) FindRandom(level int, tags []string, limit int, excludeIDs []uint) ([]model.Sentence, error) {
	var sentences []model.Sentence
	query := r.db.Model(&model.Sentence{})

	if level > 0 {
		query = query.Where("level <= ?", level)
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

func (r *SentenceRepository) GetHistory(userID uint, page, perPage int) ([]model.Sentence, int64, error) {
	var sentences []model.Sentence
	var total int64

	// DailySentenceSet에서 유저가 학습한 모든 문장 조회
	subQuery := r.db.Model(&model.DailySentenceSet{}).
		Select("UNNEST(sentence_ids)").
		Where("user_id = ?", userID)

	query := r.db.Model(&model.Sentence{}).
		Where("id IN (?)", subQuery)

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

func (r *SentenceRepository) GetUserLearnedSentenceIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&model.DailySentenceSet{}).
		Select("UNNEST(sentence_ids)").
		Where("user_id = ?", userID).
		Find(&ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}
