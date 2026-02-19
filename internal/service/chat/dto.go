package chat

import "github.com/jptaku/server/internal/model"

// CreateSessionInput 세션 생성 입력
type CreateSessionInput struct {
	Domain string `json:"domain"` // 도메인 (예: "anime", "drama", "game", "movie", "music")
}

// CreateSessionResponse 세션 생성 응답
type CreateSessionResponse struct {
	Session        *model.ChatSession       `json:"session"`
	Greeting       string                   `json:"greeting,omitempty"`         // AI 첫 인사 (일본어) - 새 세션인 경우만
	GreetingKr     string                   `json:"greeting_kr,omitempty"`      // AI 첫 인사 (한국어 번역) - 새 세션인 경우만
	Suggestions    []Suggestion             `json:"suggestions,omitempty"`      // 유저 답변 제안 (3개)
	Audio          string                   `json:"audio,omitempty"`            // TTS 오디오 (Base64)
	Messages       []MessageWithTranslation `json:"messages,omitempty"`         // 기존 세션의 메시지들 (번역 포함) - 기존 세션인 경우만
	IsResumed      bool                     `json:"is_resumed"`                 // 기존 세션을 재개한 경우 true
	ScenarioTextKr string                   `json:"scenario_text_kr,omitempty"` // 대화 상황 한국어 설명 (화면 상단 표시용)
}

// Suggestion 답변 제안
type Suggestion struct {
	Text            string `json:"text"`              // 일본어 제안
	TextKr          string `json:"text_kr"`           // 한국어 번역
	IsTodaySentence bool   `json:"is_today_sentence"` // 오늘의 문장 여부
}

// GuestStartInput 비회원 대화 시작 입력
type GuestStartInput struct {
	Domain string `json:"domain"` // 도메인 ("anime", "drama", "game", "movie", "music")
}

// GuestSessionResponse 비회원 세션 시작 응답
type GuestSessionResponse struct {
	Domain         string       `json:"domain"`
	ContentTitle   string       `json:"content_title"`
	PersonaName    string       `json:"persona_name"`
	PersonaGender  string       `json:"persona_gender"`
	Greeting       string       `json:"greeting"`
	GreetingKr     string       `json:"greeting_kr"`
	Suggestions    []Suggestion `json:"suggestions"`
	ScenarioTextKr string       `json:"scenario_text_kr,omitempty"`
	MaxTurn        int          `json:"max_turn"`
}

// GuestMessageItem 비회원 대화 메시지 항목
type GuestMessageItem struct {
	Role    string `json:"role"`    // "user" 또는 "assistant"
	Content string `json:"content"` // 일본어 내용
}

// GuestMessageInput 비회원 메시지 전송 입력 (클라이언트가 히스토리를 직접 전달)
type GuestMessageInput struct {
	Domain        string             `json:"domain"`
	ContentTitle  string             `json:"content_title"`
	PersonaName   string             `json:"persona_name"`
	PersonaGender string             `json:"persona_gender"`
	Messages      []GuestMessageItem `json:"messages"`     // 지금까지의 대화 히스토리
	Message       string             `json:"message"`      // 새 유저 메시지
	CurrentTurn   int                `json:"current_turn"` // 현재 턴
	MaxTurn       int                `json:"max_turn"`     // 최대 턴
}
