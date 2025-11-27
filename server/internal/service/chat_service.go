package service

import (
	"time"

	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/repository"
)

type ChatService struct {
	chatRepo     *repository.ChatRepository
	sentenceRepo *repository.SentenceRepository
}

func NewChatService(chatRepo *repository.ChatRepository, sentenceRepo *repository.SentenceRepository) *ChatService {
	return &ChatService{
		chatRepo:     chatRepo,
		sentenceRepo: sentenceRepo,
	}
}

type CreateSessionInput struct {
	DailySetID uint `json:"daily_set_id"`
}

type EndSessionInput struct {
	DurationSeconds int `json:"duration_seconds"`
}

func (s *ChatService) CreateSession(userID uint, input *CreateSessionInput) (*model.ChatSession, error) {
	session := &model.ChatSession{
		UserID:     userID,
		DailySetID: input.DailySetID,
		StartedAt:  time.Now(),
	}

	if err := s.chatRepo.CreateSession(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *ChatService) GetSession(sessionID uint) (*model.ChatSession, error) {
	return s.chatRepo.FindSessionByID(sessionID)
}

func (s *ChatService) EndSession(sessionID uint, input *EndSessionInput) (*model.ChatSession, error) {
	session, err := s.chatRepo.FindSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session.EndedAt = &now
	session.DurationSeconds = input.DurationSeconds

	if err := s.chatRepo.UpdateSession(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *ChatService) GetSessions(userID uint, page, perPage int) ([]model.ChatSession, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	return s.chatRepo.GetUserSessions(userID, page, perPage)
}

func (s *ChatService) GetRecentSessions(userID uint, limit int) ([]model.ChatSession, error) {
	if limit < 1 || limit > 10 {
		limit = 5
	}
	return s.chatRepo.GetRecentSessions(userID, limit)
}

func (s *ChatService) AddMessage(sessionID uint, speaker, jpText, krText string, usedSentenceID *uint) (*model.ChatMessage, error) {
	message := &model.ChatMessage{
		SessionID:           sessionID,
		Speaker:             speaker,
		JPText:              jpText,
		KRText:              krText,
		UsedTodaySentenceID: usedSentenceID,
	}

	if err := s.chatRepo.CreateMessage(message); err != nil {
		return nil, err
	}

	// 오늘 문장 사용 카운트 업데이트
	if usedSentenceID != nil {
		session, err := s.chatRepo.FindSessionByID(sessionID)
		if err == nil {
			session.TodaySentenceUsedCount++
			session.TotalMessages++
			s.chatRepo.UpdateSession(session)
		}
	}

	return message, nil
}
