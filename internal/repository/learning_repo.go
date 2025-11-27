package repository

import (
	"github.com/jptaku/server/internal/model"
	"gorm.io/gorm"
)

type LearningRepository struct {
	db *gorm.DB
}

func NewLearningRepository(db *gorm.DB) *LearningRepository {
	return &LearningRepository{db: db}
}

func (r *LearningRepository) Create(progress *model.LearningProgress) error {
	return r.db.Create(progress).Error
}

func (r *LearningRepository) FindByID(id uint) (*model.LearningProgress, error) {
	var progress model.LearningProgress
	err := r.db.First(&progress, id).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *LearningRepository) FindByUserAndSentence(userID, sentenceID uint) (*model.LearningProgress, error) {
	var progress model.LearningProgress
	err := r.db.Where("user_id = ? AND sentence_id = ?", userID, sentenceID).
		Order("created_at DESC").
		First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *LearningRepository) FindByDailySet(dailySetID uint) ([]model.LearningProgress, error) {
	var progresses []model.LearningProgress
	err := r.db.Where("daily_set_id = ?", dailySetID).
		Preload("Sentence").
		Find(&progresses).Error
	if err != nil {
		return nil, err
	}
	return progresses, nil
}

func (r *LearningRepository) Update(progress *model.LearningProgress) error {
	return r.db.Save(progress).Error
}

func (r *LearningRepository) GetUserProgress(userID uint, page, perPage int) ([]model.LearningProgress, int64, error) {
	var progresses []model.LearningProgress
	var total int64

	query := r.db.Model(&model.LearningProgress{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Preload("Sentence").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&progresses).Error
	if err != nil {
		return nil, 0, err
	}

	return progresses, total, nil
}

func (r *LearningRepository) GetTodayProgress(userID uint, dailySetID uint) ([]model.LearningProgress, error) {
	var progresses []model.LearningProgress
	err := r.db.Where("user_id = ? AND daily_set_id = ?", userID, dailySetID).
		Preload("Sentence").
		Find(&progresses).Error
	if err != nil {
		return nil, err
	}
	return progresses, nil
}

func (r *LearningRepository) CountCompletedToday(userID uint, dailySetID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LearningProgress{}).
		Where("user_id = ? AND daily_set_id = ? AND memorized = ?", userID, dailySetID, true).
		Count(&count).Error
	return count, err
}
