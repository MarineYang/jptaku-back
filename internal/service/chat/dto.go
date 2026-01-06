package chat

// CreateSessionInput 세션 생성 입력
type CreateSessionInput struct {
	Topic       string `json:"topic"`        // 대화 주제 (예: "anime", "game")
	TopicDetail string `json:"topic_detail"` // 세부 주제 (예: "원피스에 대해서")
}
