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

// SubmitQuizRequest 퀴즈 제출 요청
type SubmitQuizRequest struct {
	SentenceID      uint   `json:"sentence_id" binding:"required"`
	DailySetID      uint   `json:"daily_set_id"`
	FillBlankAnswer string `json:"fill_blank_answer" binding:"required"` // 빈칸 채우기 정답
	OrderingAnswer  []int  `json:"ordering_answer" binding:"required"`   // 문장 배열 정답 (인덱스 배열)
}

// SubmitQuizResponse 퀴즈 제출 응답
type SubmitQuizResponse struct {
	SentenceID       uint `json:"sentence_id"`
	FillBlankCorrect bool `json:"fill_blank_correct"` // 빈칸 채우기 정답 여부
	OrderingCorrect  bool `json:"ordering_correct"`   // 문장 배열 정답 여부
	AllCorrect       bool `json:"all_correct"`        // 모두 정답 여부
	Memorized        bool `json:"memorized"`          // 암기 완료 여부
}
