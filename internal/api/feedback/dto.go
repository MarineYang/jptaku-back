package feedback

type StatsResponse struct {
	TotalSessions        int64   `json:"total_sessions"`
	TotalLearningMinutes int64   `json:"total_learning_minutes"`
	TotalSentencesUsed   int64   `json:"total_sentences_used"`
	AverageScore         float64 `json:"average_score"`
	CurrentStreak        int     `json:"current_streak"`
}

type CategoryProgress struct {
	Category string  `json:"category"`
	Progress float64 `json:"progress"`
	Count    int     `json:"count"`
}

type WeeklyStats struct {
	Date             string `json:"date"`
	SessionCount     int    `json:"session_count"`
	SentencesLearned int    `json:"sentences_learned"`
	MinutesSpent     int    `json:"minutes_spent"`
}

