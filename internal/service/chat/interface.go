package chat

import (
	"context"

	"github.com/jptaku/server/internal/model"
)

// ChatRepository 채팅 저장소 인터페이스
type ChatRepository interface {
	CreateSession(session *model.ChatSession) error
	FindSessionByID(id uint) (*model.ChatSession, error)
	UpdateSession(session *model.ChatSession) error
	GetUserSessions(userID uint, page, perPage int) ([]model.ChatSession, int64, error)
	GetActiveSession(userID uint) (*model.ChatSession, error)
	CreateMessage(message *model.ChatMessage) error
	GetSessionMessages(sessionID uint) ([]model.ChatMessage, error)
}

// Provider 서비스 인터페이스 (외부에서 사용)
type Provider interface {
	// 세션 관리
	CreateSession(userID uint, input *CreateSessionInput) (*CreateSessionResponse, error)
	GetSession(sessionID uint) (*model.ChatSession, error)
	GetSessions(userID uint, page, perPage int) ([]model.ChatSession, int64, error)
	EndSession(sessionID uint) (*model.ChatSession, error)

	// 메시지 전송 (SSE 스트리밍)
	SendMessageStream(ctx context.Context, sessionID uint, userMessage string, streamChan chan<- StreamChunk) error
}

// StreamChunk SSE 스트리밍 청크
type StreamChunk struct {
	Type        string       `json:"type"`                    // "content", "done", "error", "audio", "suggestions", "translation"
	Content     string       `json:"content"`                 // 텍스트 내용 (일본어)
	ContentKr   string       `json:"content_kr,omitempty"`    // 텍스트 내용 (한국어 번역)
	Audio       string       `json:"audio,omitempty"`         // Base64 인코딩된 WAV 오디오
	Error       string       `json:"error,omitempty"`
	Suggestions []Suggestion `json:"suggestions,omitempty"`   // 유저 답변 제안 (3개)
}
