package learning

type UpdateProgressRequest struct {
	SentenceID uint  `json:"sentence_id" binding:"required"`
	DailySetID uint  `json:"daily_set_id"`
	Understand *bool `json:"understand,omitempty"`
	Speak      *bool `json:"speak,omitempty"`
	Confirm    *bool `json:"confirm,omitempty"`
	Memorized  *bool `json:"memorized,omitempty"`
}

type TodayProgressQuery struct {
	DailySetID uint `form:"daily_set_id" binding:"required"`
}

type ProgressQuery struct {
	Page    int `form:"page" binding:"min=1"`
	PerPage int `form:"per_page" binding:"min=1,max=50"`
}
