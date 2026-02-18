package repository

import (
	"time"

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

// GetUnmemorizedSentenceIDs 과거 DailySet에 포함된 문장 중 Memorized=false인 문장 ID 조회 (유저 레벨 필터 포함)
func (r *LearningRepository) GetUnmemorizedSentenceIDs(userID uint, today time.Time, levels []int) ([]uint, error) {
	var ids []uint
	// 과거 DailySet의 sentence_ids 중 memorized=true인 progress가 없고, 현재 레벨에 해당하는 것
	err := r.db.Raw(`
		SELECT DISTINCT s_id::int
		FROM daily_sentence_sets, jsonb_array_elements_text(sentence_ids) AS s_id
		WHERE user_id = ? AND date < ?
		AND s_id::int NOT IN (
			SELECT sentence_id FROM learning_progresses
			WHERE user_id = ? AND memorized = true
		)
		AND s_id::int IN (
			SELECT id FROM sentences WHERE level IN ?
		)
	`, userID, today.Format("2006-01-02"), userID, levels).Scan(&ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}
