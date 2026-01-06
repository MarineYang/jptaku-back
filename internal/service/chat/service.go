package chat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/pkg"
	"github.com/sashabaranov/go-openai"
)

// Service 채팅 서비스
type Service struct {
	chatRepo       ChatRepository
	openaiClient   *openai.Client
	model          string
	voicevoxClient *pkg.VoiceVoxClient
	ttsEnabled     bool
}

// 컴파일 타임 인터페이스 검증
var _ Provider = (*Service)(nil)

// NewService 서비스 생성자
func NewService(chatRepo ChatRepository, openaiAPIKey, openaiModel string) *Service {
	client := openai.NewClient(openaiAPIKey)
	return &Service{
		chatRepo:     chatRepo,
		openaiClient: client,
		model:        openaiModel,
		ttsEnabled:   false,
	}
}

// SetVoiceVox VoiceVox 클라이언트 설정
func (s *Service) SetVoiceVox(voicevoxURL string) {
	s.voicevoxClient = pkg.NewVoiceVoxClient(voicevoxURL)

	// VoiceVox 연결 확인
	if err := s.voicevoxClient.HealthCheck(); err != nil {
		log.Printf("Warning: VoiceVox not available: %v", err)
		s.ttsEnabled = false
	} else {
		s.ttsEnabled = true
		log.Println("VoiceVox TTS enabled")
	}
}

// CreateSession 세션 생성
func (s *Service) CreateSession(userID uint, input *CreateSessionInput) (*model.ChatSession, error) {
	session := &model.ChatSession{
		UserID:      userID,
		Topic:       input.Topic,
		TopicDetail: input.TopicDetail,
		CurrentTurn: 0,
		MaxTurn:     model.MaxTurnsPerSession,
		Status:      "active",
		StartedAt:   time.Now(),
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

// EndSession 세션 종료
func (s *Service) EndSession(sessionID uint) (*model.ChatSession, error) {
	session, err := s.chatRepo.FindSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session.Status = "ended"
	session.EndedAt = &now
	session.DurationSeconds = int(now.Sub(session.StartedAt).Seconds())

	if err := s.chatRepo.UpdateSession(session); err != nil {
		return nil, err
	}

	return session, nil
}

// SendMessageStream SSE 스트리밍으로 메시지 전송
func (s *Service) SendMessageStream(ctx context.Context, sessionID uint, userMessage string, streamChan chan<- StreamChunk) error {
	defer close(streamChan)

	// 세션 조회
	session, err := s.chatRepo.FindSessionByID(sessionID)
	if err != nil {
		streamChan <- StreamChunk{Type: "error", Error: "세션을 찾을 수 없습니다"}
		return err
	}

	// 세션 상태 확인
	if session.Status != "active" {
		streamChan <- StreamChunk{Type: "error", Error: "이미 종료된 세션입니다"}
		return errors.New("session is not active")
	}

	// 턴 수 확인
	if session.CurrentTurn >= session.MaxTurn {
		streamChan <- StreamChunk{Type: "error", Error: "최대 대화 횟수에 도달했습니다"}
		return errors.New("max turns reached")
	}

	// 유저 메시지 저장
	userMsg := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   userMessage,
	}
	if err := s.chatRepo.CreateMessage(userMsg); err != nil {
		streamChan <- StreamChunk{Type: "error", Error: "메시지 저장 실패"}
		return err
	}

	// 턴 증가
	session.CurrentTurn++

	// OpenAI 메시지 구성
	messages := s.buildOpenAIMessages(session, userMessage)

	// 마지막 턴이면 대화 마무리 지시 추가
	isLastTurn := session.CurrentTurn >= session.MaxTurn
	if isLastTurn {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: "これが最後のターンです。会話を自然に締めくくり、ユーザーの良かった点と改善点を簡潔にフィードバックしてください。",
		})
	}

	// OpenAI 스트리밍 요청
	stream, err := s.openaiClient.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    messages,
		MaxTokens:   500,
		Temperature: 0.8,
		Stream:      true,
	})
	if err != nil {
		streamChan <- StreamChunk{Type: "error", Error: "AI 응답 생성 실패"}
		return err
	}
	defer stream.Close()

	// 전체 응답 수집
	var fullResponse string

	// 스트리밍 응답 처리
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			streamChan <- StreamChunk{Type: "error", Error: "스트리밍 오류"}
			return err
		}

		content := response.Choices[0].Delta.Content
		if content != "" {
			fullResponse += content
			streamChan <- StreamChunk{Type: "content", Content: content}
		}
	}

	// AI 응답 저장
	aiMsg := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "assistant",
		Content:   fullResponse,
	}
	if err := s.chatRepo.CreateMessage(aiMsg); err != nil {
		return err
	}

	// TTS 처리 (VoiceVox가 활성화된 경우)
	if s.ttsEnabled && s.voicevoxClient != nil && fullResponse != "" {
		audioBase64, err := s.voicevoxClient.TextToSpeech(fullResponse)
		if err != nil {
			log.Printf("TTS generation failed: %v", err)
			// TTS 실패해도 대화는 계속 진행
		} else if audioBase64 != "" {
			streamChan <- StreamChunk{
				Type:  "audio",
				Audio: audioBase64,
			}
		}
	}

	// 마지막 턴이면 세션 완료 처리
	if isLastTurn {
		now := time.Now()
		session.Status = "completed"
		session.EndedAt = &now
		session.DurationSeconds = int(now.Sub(session.StartedAt).Seconds())
	}

	// 세션 업데이트
	if err := s.chatRepo.UpdateSession(session); err != nil {
		return err
	}

	// 완료 신호 전송
	streamChan <- StreamChunk{
		Type:    "done",
		Content: fmt.Sprintf(`{"current_turn":%d,"max_turn":%d,"is_completed":%v}`, session.CurrentTurn, session.MaxTurn, isLastTurn),
	}

	return nil
}

// buildOpenAIMessages OpenAI 메시지 구성
func (s *Service) buildOpenAIMessages(session *model.ChatSession, newUserMessage string) []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf(`あなたは日本語会話の練習パートナーです。以下のルールに従ってください：

1. 日本語のみで会話してください
2. ユーザーのレベルに合わせて、簡単な日本語を使ってください
3. 友達のように親しみやすく話してください
4. 会話のトピック: %s
5. 詳細: %s
6. 相手の発言に対して自然に反応し、会話を続けてください
7. 1-2文程度で簡潔に返答してください`, session.Topic, session.TopicDetail),
		},
	}

	// 기존 대화 히스토리 추가
	for _, msg := range session.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 새 유저 메시지 추가
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: newUserMessage,
	})

	return messages
}
