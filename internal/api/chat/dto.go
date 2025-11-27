package chat

type CreateSessionRequest struct {
	DailySetID uint `json:"daily_set_id"`
}

type EndSessionRequest struct {
	DurationSeconds int `json:"duration_seconds"`
}

type SessionsQuery struct {
	Page    int `form:"page" binding:"min=1"`
	PerPage int `form:"per_page" binding:"min=1,max=50"`
}

