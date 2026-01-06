package pkg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// VoiceVoxClient VoiceVox API 클라이언트
type VoiceVoxClient struct {
	baseURL    string
	httpClient *http.Client
	speakerID  int // 기본 화자 ID (0: 四国めたん, 1: ずんだもん 등)
}

// NewVoiceVoxClient VoiceVox 클라이언트 생성
func NewVoiceVoxClient(baseURL string) *VoiceVoxClient {
	return &VoiceVoxClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		speakerID: 1, // ずんだもん (자연스러운 여성 목소리)
	}
}

// SetSpeaker 화자 변경
func (c *VoiceVoxClient) SetSpeaker(speakerID int) {
	c.speakerID = speakerID
}

// AudioQuery 음성 합성을 위한 쿼리 생성
type AudioQuery struct {
	AccentPhrases      []AccentPhrase `json:"accent_phrases"`
	SpeedScale         float64        `json:"speedScale"`
	PitchScale         float64        `json:"pitchScale"`
	IntonationScale    float64        `json:"intonationScale"`
	VolumeScale        float64        `json:"volumeScale"`
	PrePhonemeLength   float64        `json:"prePhonemeLength"`
	PostPhonemeLength  float64        `json:"postPhonemeLength"`
	OutputSamplingRate int            `json:"outputSamplingRate"`
	OutputStereo       bool           `json:"outputStereo"`
	Kana               string         `json:"kana,omitempty"`
}

type AccentPhrase struct {
	Moras           []Mora `json:"moras"`
	Accent          int    `json:"accent"`
	PauseMora       *Mora  `json:"pause_mora,omitempty"`
	IsInterrogative bool   `json:"is_interrogative"`
}

type Mora struct {
	Text            string   `json:"text"`
	Consonant       *string  `json:"consonant,omitempty"`
	ConsonantLength *float64 `json:"consonant_length,omitempty"`
	Vowel           string   `json:"vowel"`
	VowelLength     float64  `json:"vowel_length"`
	Pitch           float64  `json:"pitch"`
}

// TextToSpeech 텍스트를 음성으로 변환 (Base64 WAV 반환)
func (c *VoiceVoxClient) TextToSpeech(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	// 1. audio_query API 호출
	audioQuery, err := c.createAudioQuery(text)
	if err != nil {
		return "", fmt.Errorf("audio_query failed: %w", err)
	}

	// 2. synthesis API 호출
	wavData, err := c.synthesis(audioQuery)
	if err != nil {
		return "", fmt.Errorf("synthesis failed: %w", err)
	}

	// 3. Base64 인코딩
	base64Audio := base64.StdEncoding.EncodeToString(wavData)

	return base64Audio, nil
}

// createAudioQuery 텍스트에서 AudioQuery 생성
func (c *VoiceVoxClient) createAudioQuery(text string) (*AudioQuery, error) {
	// URL 인코딩된 텍스트로 쿼리 파라미터 생성
	apiURL := fmt.Sprintf("%s/audio_query?text=%s&speaker=%d",
		c.baseURL,
		url.QueryEscape(text),
		c.speakerID,
	)

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("audio_query returned status %d: %s", resp.StatusCode, string(body))
	}

	var query AudioQuery
	if err := json.NewDecoder(resp.Body).Decode(&query); err != nil {
		return nil, err
	}

	return &query, nil
}

// synthesis AudioQuery에서 WAV 음성 생성
func (c *VoiceVoxClient) synthesis(query *AudioQuery) ([]byte, error) {
	// AudioQuery를 JSON으로 변환
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s/synthesis?speaker=%d", c.baseURL, c.speakerID)

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(queryJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("synthesis returned status %d: %s", resp.StatusCode, string(body))
	}

	// WAV 데이터 읽기
	wavData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return wavData, nil
}

// HealthCheck VoiceVox 서버 상태 확인
func (c *VoiceVoxClient) HealthCheck() error {
	resp, err := c.httpClient.Get(c.baseURL + "/version")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("VoiceVox health check failed with status: %d", resp.StatusCode)
	}

	return nil
}
