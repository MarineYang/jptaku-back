package model

import "encoding/json"

// TopicData Redis에 저장되는 도메인별 토픽 데이터
type TopicData struct {
	Domain     string         `json:"domain"`
	FetchedAt  string         `json:"fetched_at"`
	UpdatedAt  string         `json:"updated_at"`
	TotalCount int            `json:"total_count"`
	Contents   []TopicContent `json:"contents"`
}

// TopicContent 개별 작품/콘텐츠 정보
type TopicContent struct {
	ID          string          `json:"id"`
	Domain      string          `json:"domain"`
	Title       string          `json:"title"`
	TitleNative string          `json:"title_native"`
	Intro       string          `json:"intro"`
	CoverImage  string          `json:"cover_image"`
	SourceURL   string          `json:"source_url"`
	FetchedAt   string          `json:"fetched_at"`
	UpdatedAt   string          `json:"updated_at"`
	Metadata    json.RawMessage `json:"metadata"`
}
