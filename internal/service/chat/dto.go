package chat

// CreateSessionInput 세션 생성 입력
type CreateSessionInput struct {
	DailySetID uint `json:"daily_set_id"`
}

// EndSessionInput 세션 종료 입력
type EndSessionInput struct {
	DurationSeconds int `json:"duration_seconds"`
}
