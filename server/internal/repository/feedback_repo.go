package repository

import (
	"github.com/jptaku/server/internal/model"
	"gorm.io/gorm"
)

type FeedbackRepository struct {
	db *gorm.DB
}

func NewFeedbackRepository(db *gorm.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

func (r *FeedbackRepository) Create(feedback *model.Feedback) error {
	return r.db.Create(feedback).Error
}

func (r *FeedbackRepository) FindByID(id uint) (*model.Feedback, error) {
	var feedback model.Feedback
	err := r.db.First(&feedback, id).Error
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (r *FeedbackRepository) FindBySessionID(sessionID uint) (*model.Feedback, error) {
	var feedback model.Feedback
	err := r.db.Where("session_id = ?", sessionID).First(&feedback).Error
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (r *FeedbackRepository) Update(feedback *model.Feedback) error {
	return r.db.Save(feedback).Error
}

func (r *FeedbackRepository) GetUserFeedbacks(userID uint, page, perPage int) ([]model.Feedback, int64, error) {
	var feedbacks []model.Feedback
	var total int64

	query := r.db.Model(&model.Feedback{}).
		Joins("JOIN chat_sessions ON feedbacks.session_id = chat_sessions.id").
		Where("chat_sessions.user_id = ?", userID)

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Preload("Session").
		Order("feedbacks.created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&feedbacks).Error
	if err != nil {
		return nil, 0, err
	}

	return feedbacks, total, nil
}
