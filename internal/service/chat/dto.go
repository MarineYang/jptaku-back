package chat

import "github.com/jptaku/server/internal/model"

// CreateSessionInput 세션 생성 입력
type CreateSessionInput struct {
	Topic       string `json:"topic"`        // 대화 주제 (예: "anime", "game")
	TopicDetail string `json:"topic_detail"` // 세부 주제 (예: "원피스에 대해서")
}

// CreateSessionResponse 세션 생성 응답
type CreateSessionResponse struct {
	Session     *model.ChatSession       `json:"session"`
	Greeting    string                   `json:"greeting,omitempty"`    // AI 첫 인사 (일본어) - 새 세션인 경우만
	GreetingKr  string                   `json:"greeting_kr,omitempty"` // AI 첫 인사 (한국어 번역) - 새 세션인 경우만
	Suggestions []Suggestion             `json:"suggestions,omitempty"` // 유저 답변 제안 (3개)
	Audio       string                   `json:"audio,omitempty"`       // TTS 오디오 (Base64)
	Messages    []MessageWithTranslation `json:"messages,omitempty"`    // 기존 세션의 메시지들 (번역 포함) - 기존 세션인 경우만
	IsResumed   bool                     `json:"is_resumed"`            // 기존 세션을 재개한 경우 true
}

// Suggestion 답변 제안
type Suggestion struct {
	Text            string `json:"text"`              // 일본어 제안
	TextKr          string `json:"text_kr"`           // 한국어 번역
	IsTodaySentence bool   `json:"is_today_sentence"` // 오늘의 문장 여부
}
