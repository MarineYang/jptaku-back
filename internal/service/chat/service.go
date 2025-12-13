package chat

import (
	"time"

	"github.com/jptaku/server/internal/model"
)

// Service 채팅 서비스
type Service struct {
	chatRepo     ChatRepository
	sentenceRepo SentenceRepository
}

// 컴파일 타임 인터페이스 검증
var _ Provider = (*Service)(nil)

// NewService 서비스 생성자
func NewService(chatRepo ChatRepository, sentenceRepo SentenceRepository) *Service {
	return &Service{
		chatRepo:     chatRepo,
		sentenceRepo: sentenceRepo,
	}
}

// CreateSession 세션 생성
func (s *Service) CreateSession(userID uint, input *CreateSessionInput) (*model.ChatSession, error) {
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

// GetSession 세션 조회
func (s *Service) GetSession(sessionID uint) (*model.ChatSession, error) {
	return s.chatRepo.FindSessionByID(sessionID)
}

// EndSession 세션 종료
func (s *Service) EndSession(sessionID uint, input *EndSessionInput) (*model.ChatSession, error) {
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

// GetSessions 세션 목록 조회
func (s *Service) GetSessions(userID uint, page, perPage int) ([]model.ChatSession, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	return s.chatRepo.GetUserSessions(userID, page, perPage)
}

// GetRecentSessions 최근 세션 조회
func (s *Service) GetRecentSessions(userID uint, limit int) ([]model.ChatSession, error) {
	if limit < 1 || limit > 10 {
		limit = 5
	}
	return s.chatRepo.GetRecentSessions(userID, limit)
}

// AddMessage 메시지 추가
func (s *Service) AddMessage(sessionID uint, speaker, jpText, krText string, usedSentenceID *uint) (*model.ChatMessage, error) {
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
