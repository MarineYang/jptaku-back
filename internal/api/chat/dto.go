package chat

// CreateSessionRequest 세션 생성 요청
type CreateSessionRequest struct {
	Topic       string `json:"topic" binding:"required"` // 대화 주제
	TopicDetail string `json:"topic_detail"`             // 세부 주제 (선택)
}

// SendMessageRequest 메시지 전송 요청
type SendMessageRequest struct {
	Message string `json:"message" binding:"required"` // 유저 메시지
}

// SessionsQuery 세션 목록 조회 쿼리
type SessionsQuery struct {
	Page    int `form:"page" binding:"min=1"`
	PerPage int `form:"per_page" binding:"min=1,max=50"`
}

// SessionDetailResponse 세션 상세 조회 응답
type SessionDetailResponse struct {
	ID              uint                    `json:"id"`
	UserID          uint                    `json:"user_id"`
	Topic           string                  `json:"topic"`
	TopicDetail     string                  `json:"topic_detail"`
	CurrentTurn     int                     `json:"current_turn"`
	MaxTurn         int                     `json:"max_turn"`
	Status          string                  `json:"status"`
	StartedAt       string                  `json:"started_at"`
	EndedAt         *string                 `json:"ended_at,omitempty"`
	DurationSeconds int                     `json:"duration_seconds"`
	Messages        []MessageWithTranslation `json:"messages"`
	Suggestions     []SuggestionResponse    `json:"suggestions,omitempty"`
}

// MessageWithTranslation 번역 포함 메시지
type MessageWithTranslation struct {
	ID        uint   `json:"id"`
	SessionID uint   `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`    // 일본어
	ContentKr string `json:"content_kr"` // 한국어 번역
	CreatedAt string `json:"created_at"`
}

// SuggestionResponse 제안 응답
type SuggestionResponse struct {
	Text            string `json:"text"`
	TextKr          string `json:"text_kr"`
	IsTodaySentence bool   `json:"is_today_sentence"`
}
