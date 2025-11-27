package sentences

type DailySentencesResponse struct {
	DailySetID uint `json:"daily_set_id"`
	Date       string `json:"date"`
	Sentences  []SentenceResponse `json:"sentences"`
}

type SentenceResponse struct {
	ID       uint     `json:"id"`
	JP       string   `json:"jp"`
	KR       string   `json:"kr"`
	Romaji   string   `json:"romaji,omitempty"`
	Level    int      `json:"level"`
	Tags     []string `json:"tags"`
	AudioURL string   `json:"audio_url,omitempty"`
}

type SentenceDetailResponse struct {
	Sentence SentenceResponse `json:"sentence"`
	Words    []WordResponse   `json:"words,omitempty"`
	Grammar  []string         `json:"grammar,omitempty"`
	Examples []string         `json:"examples,omitempty"`
}

type WordResponse struct {
	Japanese string `json:"japanese"`
	Reading  string `json:"reading"`
	Meaning  string `json:"meaning"`
	PartOf   string `json:"part_of"`
}

type HistoryQuery struct {
	Page    int `form:"page" binding:"min=1"`
	PerPage int `form:"per_page" binding:"min=1,max=50"`
}

