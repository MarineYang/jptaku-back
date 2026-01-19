package chat

import (
	"context"
	"encoding/json"
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
	sentenceRepo   SentenceRepository
	openaiClient   *openai.Client
	model          string
	voicevoxClient *pkg.VoiceVoxClient
	ttsEnabled     bool
}

// 컴파일 타임 인터페이스 검증
var _ Provider = (*Service)(nil)

// NewService 서비스 생성자
func NewService(chatRepo ChatRepository, sentenceRepo SentenceRepository, openaiAPIKey, openaiModel string) *Service {
	client := openai.NewClient(openaiAPIKey)
	return &Service{
		chatRepo:     chatRepo,
		sentenceRepo: sentenceRepo,
		openaiClient: client,
		model:        openaiModel,
		ttsEnabled:   false,
	}
}

// getTodaySentenceTexts 오늘의 문장 일본어 텍스트 목록 조회
func (s *Service) getTodaySentenceTexts(userID uint) []string {
	if s.sentenceRepo == nil {
		return nil
	}

	today := time.Now().Truncate(24 * time.Hour)
	dailySet, err := s.sentenceRepo.GetDailySet(userID, today)
	if err != nil {
		log.Printf("Failed to get daily set for today's sentences: %v", err)
		return nil
	}

	sentences, err := s.sentenceRepo.FindByIDs(dailySet.SentenceIDs)
	if err != nil {
		log.Printf("Failed to get sentences by IDs: %v", err)
		return nil
	}

	texts := make([]string, 0, len(sentences))
	for _, sent := range sentences {
		texts = append(texts, sent.JP)
	}
	return texts
}

// markTodaySentences 제안 목록에서 오늘의 문장 표시
func (s *Service) markTodaySentences(suggestions []Suggestion, todayTexts []string) []Suggestion {
	if len(todayTexts) == 0 {
		return suggestions
	}

	// 오늘의 문장 텍스트를 map으로 변환 (빠른 조회)
	todayMap := make(map[string]bool)
	for _, text := range todayTexts {
		todayMap[text] = true
	}

	// 제안 중 오늘의 문장이 있으면 표시
	for i := range suggestions {
		if todayMap[suggestions[i].Text] {
			suggestions[i].IsTodaySentence = true
		}
	}

	return suggestions
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

// GreetingResponse 인사말 통합 응답 구조체
type GreetingResponse struct {
	Greeting    string       `json:"greeting"`
	GreetingKr  string       `json:"greeting_kr"`
	Suggestions []Suggestion `json:"suggestions"`
}

// CreateSession 세션 생성 (오늘 이미 활성 세션이 있으면 해당 세션 반환)
func (s *Service) CreateSession(userID uint, input *CreateSessionInput) (*CreateSessionResponse, error) {
	// 오늘 이미 활성 세션이 있는지 확인
	existingSession, err := s.chatRepo.GetTodayActiveSession(userID)
	if err == nil && existingSession != nil {
		log.Printf("CreateSession: found existing today's session, sessionID=%d", existingSession.ID)
		// 기존 세션 반환 (상세 정보 포함)
		return s.buildResumedSessionResponse(existingSession)
	}

	// 새 세션 생성
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

	// AI 첫 인사 + 번역 + 제안을 한 번에 생성
	ctx := context.Background()
	greetingResp, err := s.generateGreetingWithSuggestions(ctx, session)
	if err != nil {
		log.Printf("Failed to generate greeting: %v", err)
		// 실패 시 기본 인사
		greetingResp = &GreetingResponse{
			Greeting:   fmt.Sprintf("こんにちは！「%s」について話しましょう。日本語で話してみてください！", session.Topic),
			GreetingKr: fmt.Sprintf("안녕하세요! \"%s\"에 대해 이야기해요. 일본어로 말해보세요!", session.Topic),
			Suggestions: []Suggestion{
				{Text: "はい、よろしくお願いします！", TextKr: "네, 잘 부탁드려요!"},
				{Text: "何から話しましょうか？", TextKr: "뭐부터 이야기할까요?"},
				{Text: "楽しみです！", TextKr: "기대돼요!"},
			},
		}
	}

	// AI 인사 메시지 저장
	aiMsg := &model.ChatMessage{
		SessionID: session.ID,
		Role:      "assistant",
		Content:   greetingResp.Greeting,
	}
	if err := s.chatRepo.CreateMessage(aiMsg); err != nil {
		log.Printf("Failed to save greeting message: %v", err)
	}

	// 세션에 메시지 포함
	session.Messages = []model.ChatMessage{*aiMsg}

	// 오늘의 문장과 제안 비교하여 표시
	todayTexts := s.getTodaySentenceTexts(userID)
	suggestions := s.markTodaySentences(greetingResp.Suggestions, todayTexts)

	// 응답 구성
	response := &CreateSessionResponse{
		Session:     session,
		Greeting:    greetingResp.Greeting,
		GreetingKr:  greetingResp.GreetingKr,
		Suggestions: suggestions,
	}

	// TTS 생성 (VoiceVox가 활성화된 경우)
	log.Printf("CreateSession: TTS enabled=%v, voicevoxClient=%v", s.ttsEnabled, s.voicevoxClient != nil)
	if s.ttsEnabled && s.voicevoxClient != nil && greetingResp.Greeting != "" {
		log.Printf("CreateSession: generating TTS for greeting...")
		audioBase64, err := s.voicevoxClient.TextToSpeech(greetingResp.Greeting)
		if err != nil {
			log.Printf("CreateSession: TTS generation failed: %v", err)
		} else if audioBase64 != "" {
			log.Printf("CreateSession: TTS generated, audio length=%d", len(audioBase64))
			response.Audio = audioBase64
		}
	}

	return response, nil
}

// buildResumedSessionResponse 기존 세션 재개 응답 구성
func (s *Service) buildResumedSessionResponse(session *model.ChatSession) (*CreateSessionResponse, error) {
	ctx := context.Background()

	// 메시지에 번역 추가
	messagesWithTranslation := make([]MessageWithTranslation, 0, len(session.Messages))
	var lastAIMessage string

	for _, msg := range session.Messages {
		msgWithTrans := MessageWithTranslation{
			ID:        msg.ID,
			SessionID: msg.SessionID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		// AI 메시지인 경우 번역 생성
		if msg.Role == "assistant" && msg.Content != "" {
			translation, err := s.translateToKorean(ctx, msg.Content)
			if err != nil {
				log.Printf("buildResumedSessionResponse: translation failed for message %d: %v", msg.ID, err)
			} else {
				msgWithTrans.ContentKr = translation
			}
			lastAIMessage = msg.Content
		}

		messagesWithTranslation = append(messagesWithTranslation, msgWithTrans)
	}

	// 응답 구성
	response := &CreateSessionResponse{
		Session:   session,
		Messages:  messagesWithTranslation,
		IsResumed: true,
	}

	// 세션이 활성 상태이고 마지막 턴이 아닌 경우 제안 생성
	if session.Status == "active" && session.CurrentTurn < session.MaxTurn && lastAIMessage != "" {
		suggestions, err := s.generateSuggestionsOnly(ctx, session, lastAIMessage)
		if err != nil {
			log.Printf("buildResumedSessionResponse: suggestions generation failed: %v", err)
		} else {
			// 오늘의 문장과 비교
			todayTexts := s.getTodaySentenceTexts(session.UserID)
			response.Suggestions = s.markTodaySentences(suggestions, todayTexts)
		}
	}

	log.Printf("buildResumedSessionResponse: session %d resumed with %d messages, %d suggestions",
		session.ID, len(messagesWithTranslation), len(response.Suggestions))

	return response, nil
}

// generateGreetingWithSuggestions AI 첫 인사 + 번역 + 제안을 한 번에 생성
func (s *Service) generateGreetingWithSuggestions(ctx context.Context, session *model.ChatSession) (*GreetingResponse, error) {
	prompt := fmt.Sprintf(`あなたは日本語会話の練習パートナーです。
ユーザーと「%s」というトピックで会話を始めます。

以下の3つを一度に生成してください：
1. 最初の挨拶（日本語、1-2文）
2. その挨拶の韓国語訳
3. ユーザーが返答できる日本語の提案3つ（各提案に韓国語訳付き）

ルール:
- 友達のように親しみやすく話してください
- トピックに関連した質問や話題を投げかけてください
- 簡単な日本語を使ってください
- 提案は10〜20文字程度で、異なるニュアンスにしてください

必ず以下のJSON形式で出力してください（他の文字は出力しないでください）:
{"greeting":"日本語の挨拶","greeting_kr":"한국어 번역","suggestions":[{"text":"日本語の提案1","text_kr":"한국어 번역1"},{"text":"日本語の提案2","text_kr":"한국어 번역2"},{"text":"日本語の提案3","text_kr":"한국어 번역3"}]}`, session.Topic)

	resp, err := s.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   400,
		Temperature: 0.8,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no greeting generated")
	}

	content := resp.Choices[0].Message.Content
	var greetingResp GreetingResponse
	if err := json.Unmarshal([]byte(content), &greetingResp); err != nil {
		log.Printf("Failed to parse greeting JSON: %v, content: %s", err, content)
		return nil, err
	}

	return &greetingResp, nil
}

// GetSession 세션 조회
func (s *Service) GetSession(sessionID uint) (*model.ChatSession, error) {
	return s.chatRepo.FindSessionByID(sessionID)
}

// GetSessionDetail 세션 상세 조회 (번역 + 제안 포함)
func (s *Service) GetSessionDetail(ctx context.Context, sessionID uint) (*SessionDetailResponse, error) {
	session, err := s.chatRepo.FindSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	// 메시지 조회
	messages, err := s.chatRepo.GetSessionMessages(sessionID)
	if err != nil {
		return nil, err
	}

	// 메시지에 번역 추가
	messagesWithTranslation := make([]MessageWithTranslation, 0, len(messages))
	var lastAIMessage string

	for _, msg := range messages {
		msgWithTrans := MessageWithTranslation{
			ID:        msg.ID,
			SessionID: msg.SessionID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		// AI 메시지인 경우 번역 생성
		if msg.Role == "assistant" && msg.Content != "" {
			translation, err := s.translateToKorean(ctx, msg.Content)
			if err != nil {
				log.Printf("GetSessionDetail: translation failed for message %d: %v", msg.ID, err)
			} else {
				msgWithTrans.ContentKr = translation
			}
			lastAIMessage = msg.Content
		}

		messagesWithTranslation = append(messagesWithTranslation, msgWithTrans)
	}

	// 응답 구성
	response := &SessionDetailResponse{
		Session:  session,
		Messages: messagesWithTranslation,
	}

	// 세션이 활성 상태이고 마지막 턴이 아닌 경우 제안 생성
	if session.Status == "active" && session.CurrentTurn < session.MaxTurn && lastAIMessage != "" {
		suggestions, err := s.generateSuggestionsOnly(ctx, session, lastAIMessage)
		if err != nil {
			log.Printf("GetSessionDetail: suggestions generation failed: %v", err)
		} else {
			// 오늘의 문장과 비교
			todayTexts := s.getTodaySentenceTexts(session.UserID)
			response.Suggestions = s.markTodaySentences(suggestions, todayTexts)
		}
	}

	return response, nil
}

// translateToKorean 일본어를 한국어로 번역
func (s *Service) translateToKorean(ctx context.Context, japaneseText string) (string, error) {
	prompt := fmt.Sprintf(`다음 일본어를 한국어로 자연스럽게 번역해주세요. 번역만 출력하세요.

일본어: %s`, japaneseText)

	resp, err := s.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   200,
		Temperature: 0.3,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no translation generated")
	}

	return resp.Choices[0].Message.Content, nil
}

// generateSuggestionsOnly 제안만 생성 (번역 없이)
func (s *Service) generateSuggestionsOnly(ctx context.Context, session *model.ChatSession, aiResponse string) ([]Suggestion, error) {
	prompt := fmt.Sprintf(`다음 AI 발언에 대해 유저 답변 제안 3개를 생성해주세요.

회화 토픽: %s
AI 발언: %s

규칙:
1. 초급~중급 일본어 학습자가 사용할 수 있는 간단한 표현 (10~20자)
2. 각 제안은 다른 뉘앙스로
3. 각 제안에 한국어 번역 포함

반드시 아래 JSON 형식으로만 출력:
[{"text":"日本語の提案1","text_kr":"한국어1"},{"text":"日本語の提案2","text_kr":"한국어2"},{"text":"日本語の提案3","text_kr":"한국어3"}]`, session.Topic, aiResponse)

	resp, err := s.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   300,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no suggestions generated")
	}

	var suggestions []Suggestion
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &suggestions); err != nil {
		log.Printf("Failed to parse suggestions JSON: %v", err)
		return nil, err
	}

	return suggestions, nil
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

	log.Printf("SendMessageStream: sessionID=%d, message=%s", sessionID, userMessage)

	// 세션 조회
	session, err := s.chatRepo.FindSessionByID(sessionID)
	if err != nil {
		log.Printf("SendMessageStream: session not found: %v", err)
		streamChan <- StreamChunk{Type: "error", Error: "세션을 찾을 수 없습니다"}
		return err
	}

	log.Printf("SendMessageStream: session status=%s, currentTurn=%d, maxTurn=%d", session.Status, session.CurrentTurn, session.MaxTurn)

	// 세션 상태 확인
	if session.Status != "active" {
		log.Printf("SendMessageStream: session is not active")
		streamChan <- StreamChunk{Type: "error", Error: "이미 종료된 세션입니다"}
		return errors.New("session is not active")
	}

	// 턴 수 확인
	if session.CurrentTurn >= session.MaxTurn {
		log.Printf("SendMessageStream: max turns reached")
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
	remainingTurns := session.MaxTurn - session.CurrentTurn

	if isLastTurn {
		messages = append(messages, openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleSystem,
			Content: `これが最後の会話です。以下の形式で自然に締めくくってください：

1. まず相手の発言に対して自然に返答してください
2. その後、「今日はたくさん話せて楽しかったね！」のように会話を締めくくってください
3. 最後に、会話の中で良かった表現や、次回試してみてほしい表現を1-2個、励ましの言葉と一緒に伝えてください

例：「そうだね！[返答]。今日は楽しかったよ！○○という表現、とても自然だったよ。次は△△も使ってみてね。また話そうね！」

自然な日本語で、友達との会話を終えるように話してください。`,
		})
	} else if remainingTurns <= 2 {
		// 残り2ターン以下なら、会話を締めくくる準備を促す
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("残り%dターンです。会話を自然に続けながら、そろそろ締めくくりに向かう準備をしてください。", remainingTurns),
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

	log.Printf("SendMessageStream: AI response completed, length=%d", len(fullResponse))

	// AI 응답 저장
	aiMsg := &model.ChatMessage{
		SessionID: sessionID,
		Role:      "assistant",
		Content:   fullResponse,
	}
	if err := s.chatRepo.CreateMessage(aiMsg); err != nil {
		log.Printf("SendMessageStream: failed to save AI message: %v", err)
		return err
	}

	// AI 응답 번역 + 제안을 한 번에 생성 (마지막 턴이 아닌 경우에만 제안 포함)
	if fullResponse != "" {
		log.Printf("SendMessageStream: generating translation and suggestions...")
		postResp, err := s.generatePostResponse(ctx, session, fullResponse, !isLastTurn)
		if err != nil {
			log.Printf("SendMessageStream: Post response generation failed: %v", err)
		} else {
			// 번역 전송
			if postResp.TranslationKr != "" {
				log.Printf("SendMessageStream: sending translation, length=%d", len(postResp.TranslationKr))
				streamChan <- StreamChunk{
					Type:      "translation",
					ContentKr: postResp.TranslationKr,
				}
			}
			// 제안 전송 (마지막 턴이 아닌 경우)
			if !isLastTurn && len(postResp.Suggestions) > 0 {
				// 오늘의 문장과 비교하여 표시
				todayTexts := s.getTodaySentenceTexts(session.UserID)
				suggestions := s.markTodaySentences(postResp.Suggestions, todayTexts)

				log.Printf("SendMessageStream: sending %d suggestions", len(suggestions))
				streamChan <- StreamChunk{
					Type:        "suggestions",
					Suggestions: suggestions,
				}
			}
		}
	}

	// TTS 처리 (VoiceVox가 활성화된 경우)
	log.Printf("SendMessageStream: TTS enabled=%v, voicevoxClient=%v", s.ttsEnabled, s.voicevoxClient != nil)
	if s.ttsEnabled && s.voicevoxClient != nil && fullResponse != "" {
		log.Printf("SendMessageStream: generating TTS...")
		audioBase64, err := s.voicevoxClient.TextToSpeech(fullResponse)
		if err != nil {
			log.Printf("SendMessageStream: TTS generation failed: %v", err)
			// TTS 실패해도 대화는 계속 진행
		} else if audioBase64 != "" {
			log.Printf("SendMessageStream: sending audio, length=%d", len(audioBase64))
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
		log.Printf("SendMessageStream: failed to update session: %v", err)
		return err
	}

	log.Printf("SendMessageStream: completed, currentTurn=%d, remainingTurns=%d", session.CurrentTurn, session.MaxTurn-session.CurrentTurn)

	// 완료 신호 전송 (남은 턴 수 포함)
	streamChan <- StreamChunk{
		Type:    "done",
		Content: fmt.Sprintf(`{"current_turn":%d,"max_turn":%d,"remaining_turns":%d,"is_completed":%v}`, session.CurrentTurn, session.MaxTurn, session.MaxTurn-session.CurrentTurn, isLastTurn),
	}

	return nil
}

// PostResponse 대화 후 응답 구조체 (번역 + 제안)
type PostResponse struct {
	TranslationKr string       `json:"translation_kr"`
	Suggestions   []Suggestion `json:"suggestions,omitempty"`
}

// generatePostResponse AI 응답에 대한 번역 + 제안을 한 번에 생성
func (s *Service) generatePostResponse(ctx context.Context, session *model.ChatSession, aiResponse string, includeSuggestions bool) (*PostResponse, error) {
	var prompt string

	if includeSuggestions {
		prompt = fmt.Sprintf(`다음 일본어 발언에 대해 한국어 번역과 유저 답변 제안 3개를 생성해주세요.

회화 토픽: %s
AI 발언: %s

규칙:
1. 번역은 자연스러운 한국어로
2. 제안은 초급~중급 일본어 학습자가 사용할 수 있는 간단한 표현 (10~20자)
3. 각 제안은 다른 뉘앙스로

반드시 아래 JSON 형식으로만 출력 (다른 텍스트 없이):
{"translation_kr":"한국어 번역","suggestions":[{"text":"日本語の提案1","text_kr":"한국어1"},{"text":"日本語の提案2","text_kr":"한국어2"},{"text":"日本語の提案3","text_kr":"한국어3"}]}`, session.Topic, aiResponse)
	} else {
		// 마지막 턴: 번역만
		prompt = fmt.Sprintf(`다음 일본어를 한국어로 자연스럽게 번역해주세요.

일본어: %s

반드시 아래 JSON 형식으로만 출력 (다른 텍스트 없이):
{"translation_kr":"한국어 번역"}`, aiResponse)
	}

	resp, err := s.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   400,
		Temperature: 0.5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate post response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no post response generated")
	}

	content := resp.Choices[0].Message.Content
	var postResp PostResponse
	if err := json.Unmarshal([]byte(content), &postResp); err != nil {
		log.Printf("Failed to parse post response JSON: %v, content: %s", err, content)
		return nil, err
	}

	return &postResp, nil
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
