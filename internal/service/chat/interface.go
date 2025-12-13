package chat

import "github.com/jptaku/server/internal/model"

// ChatRepository 채팅 저장소 인터페이스
type ChatRepository interface {
	CreateSession(session *model.ChatSession) error
	FindSessionByID(id uint) (*model.ChatSession, error)
	UpdateSession(session *model.ChatSession) error
	GetUserSessions(userID uint, page, perPage int) ([]model.ChatSession, int64, error)
	GetRecentSessions(userID uint, limit int) ([]model.ChatSession, error)
	CreateMessage(message *model.ChatMessage) error
}

// SentenceRepository 문장 저장소 인터페이스
type SentenceRepository interface {
	// 필요한 메서드가 있으면 추가
}

// Provider 서비스 인터페이스 (외부에서 사용)
type Provider interface {
	CreateSession(userID uint, input *CreateSessionInput) (*model.ChatSession, error)
	GetSession(sessionID uint) (*model.ChatSession, error)
	EndSession(sessionID uint, input *EndSessionInput) (*model.ChatSession, error)
	GetSessions(userID uint, page, perPage int) ([]model.ChatSession, int64, error)
	GetRecentSessions(userID uint, limit int) ([]model.ChatSession, error)
	AddMessage(sessionID uint, speaker, jpText, krText string, usedSentenceID *uint) (*model.ChatMessage, error)
}
