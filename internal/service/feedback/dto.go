package feedback

// StatsResponse 통계 응답
type StatsResponse struct {
	TotalSessions        int64   `json:"total_sessions"`
	TotalLearningMinutes int64   `json:"total_learning_minutes"`
	TotalSentencesUsed   int64   `json:"total_sentences_used"`
	AverageScore         float64 `json:"average_score"`
	CurrentStreak        int     `json:"current_streak"`
}

// CategoryProgress 카테고리별 진행도
type CategoryProgress struct {
	Category string  `json:"category"`
	Progress float64 `json:"progress"` // 0 ~ 100
	Count    int     `json:"count"`
}

// WeeklyStats 주간 통계
type WeeklyStats struct {
	Date             string `json:"date"`
	SessionCount     int    `json:"session_count"`
	SentencesLearned int    `json:"sentences_learned"`
	MinutesSpent     int    `json:"minutes_spent"`
}
