package chat

// CreateSessionRequest 세션 생성 요청
type CreateSessionRequest struct {
	Topic       string `json:"topic" binding:"required"`        // 대화 주제
	TopicDetail string `json:"topic_detail" binding:"required"` // 세부 주제
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
