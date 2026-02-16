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

// CreateSession 새 세션 생성
func (r *ChatRepository) CreateSession(session *model.ChatSession) error {
	return r.db.Create(session).Error
}

// FindSessionByID ID로 세션 조회 (메시지 포함)
func (r *ChatRepository) FindSessionByID(id uint) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSession 세션 업데이트
func (r *ChatRepository) UpdateSession(session *model.ChatSession) error {
	return r.db.Save(session).Error
}

// GetUserSessions 유저의 세션 목록 조회 (페이지네이션)
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

// GetActiveSession 활성 세션 조회
func (r *ChatRepository) GetActiveSession(userID uint) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.Where("user_id = ? AND status = ?", userID, "active").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetTodayActiveSession 오늘 생성된 활성 세션 조회
func (r *ChatRepository) GetTodayActiveSession(userID uint) (*model.ChatSession, error) {
	var session model.ChatSession
	// 오늘 00:00:00 기준
	today := "DATE(started_at) = DATE(NOW())"

	err := r.db.Where("user_id = ? AND status = ? AND "+today, userID, "active").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// CreateMessage 메시지 생성
func (r *ChatRepository) CreateMessage(message *model.ChatMessage) error {
	return r.db.Create(message).Error
}

// GetSessionMessages 세션의 모든 메시지 조회
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

// CountUserSessions 유저의 총 채팅 세션 수
func (r *ChatRepository) CountUserSessions(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ChatSession{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
