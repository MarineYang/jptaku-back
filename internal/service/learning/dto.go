package learning

import "github.com/jptaku/server/internal/model"

// UpdateProgressInput 진행 상황 업데이트 입력
type UpdateProgressInput struct {
	SentenceID uint  `json:"sentence_id" binding:"required"`
	DailySetID uint  `json:"daily_set_id"`
	Understand *bool `json:"understand,omitempty"`
	Speak      *bool `json:"speak,omitempty"`
	Confirm    *bool `json:"confirm,omitempty"`
	Memorized  *bool `json:"memorized,omitempty"`
}

// TodayProgressResponse 오늘의 진행 상황 응답
type TodayProgressResponse struct {
	DailySetID     uint                     `json:"daily_set_id"`
	TotalSentences int                      `json:"total_sentences"`
	Completed      int                      `json:"completed"`
	Progress       []model.LearningProgress `json:"progress"`
}

// SubmitQuizInput 퀴즈 제출 입력
type SubmitQuizInput struct {
	SentenceID      uint
	DailySetID      uint
	FillBlankAnswer string
	OrderingAnswer  []int
}

// SubmitQuizResult 퀴즈 제출 결과
type SubmitQuizResult struct {
	SentenceID       uint
	FillBlankCorrect bool
	OrderingCorrect  bool
	AllCorrect       bool
	Memorized        bool
}
