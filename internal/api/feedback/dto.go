package feedback

type StatsResponse struct {
	TotalStudyDays     int64 `json:"total_study_days"`
	TotalSessions      int64 `json:"total_sessions"`
	TotalSentencesUsed int64 `json:"total_sentences_used"`
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

