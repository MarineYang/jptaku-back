package sentences

type SentenceResponse struct {
	ID         uint     `json:"id"`
	JP         string   `json:"jp"`
	KR         string   `json:"kr"`
	Romaji     string   `json:"romaji"`
	Level      int      `json:"level"`
	Categories []int    `json:"categories"`
}

type DailySentencesResponse struct {
	Date      string             `json:"date"`
	Sentences []SentenceResponse `json:"sentences"`
}
