package repository

import (
	"github.com/jptaku/server/internal/model"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// ChatSession methods
func (r *ChatRepository) CreateSession(session *model.ChatSession) error {
	return r.db.Create(session).Error
}

func (r *ChatRepository) FindSessionByID(id uint) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.Preload("Messages").First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ChatRepository) UpdateSession(session *model.ChatSession) error {
	return r.db.Save(session).Error
}

func (r *ChatRepository) GetUserSessions(userID uint, page, perPage int) ([]model.ChatSession, int64, error) {
	var sessions []model.ChatSession
	var total int64

	query := r.db.Model(&model.ChatSession{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&sessions).Error
	if err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

func (r *ChatRepository) GetRecentSessions(userID uint, limit int) ([]model.ChatSession, error) {
	var sessions []model.ChatSession
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// ChatMessage methods
func (r *ChatRepository) CreateMessage(message *model.ChatMessage) error {
	return r.db.Create(message).Error
}

func (r *ChatRepository) GetSessionMessages(sessionID uint) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	err := r.db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}
