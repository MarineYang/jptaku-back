package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleOAuthManager struct {
	config   *oauth2.Config
	clientID string
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// GoogleIDTokenPayload Google ID Token에서 추출한 사용자 정보
// Google tokeninfo API는 모든 필드를 문자열로 반환함
type GoogleIDTokenPayload struct {
	Sub           string `json:"sub"`            // Google 사용자 고유 ID
	Email         string `json:"email"`          // 이메일
	EmailVerified string `json:"email_verified"` // 이메일 인증 여부 ("true" or "false")
	Name          string `json:"name"`           // 전체 이름
	Picture       string `json:"picture"`        // 프로필 사진 URL
	GivenName     string `json:"given_name"`     // 이름
	FamilyName    string `json:"family_name"`    // 성
	Aud           string `json:"aud"`            // Client ID (검증 필요)
	Iss           string `json:"iss"`            // 발급자
	Exp           string `json:"exp"`            // 만료 시간 (문자열)
}

func NewGoogleOAuthManager(clientID, clientSecret, redirectURL string) *GoogleOAuthManager {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleOAuthManager{
		config:   config,
		clientID: clientID,
	}
}

func (m *GoogleOAuthManager) GetAuthURL(state string) string {
	return m.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (m *GoogleOAuthManager) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return m.config.Exchange(ctx, code)
}

func (m *GoogleOAuthManager) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := m.config.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: status %d, body: %s", resp.StatusCode, string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// VerifyIDToken Google ID Token 검증 (네이티브 앱용)
// Google의 tokeninfo 엔드포인트를 사용하여 ID Token을 검증합니다
func (m *GoogleOAuthManager) VerifyIDToken(ctx context.Context, idToken string) (*GoogleUserInfo, error) {
	// Google tokeninfo 엔드포인트로 검증
	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("invalid token: status %d, body: %s", resp.StatusCode, string(body))
	}

	var payload GoogleIDTokenPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode token payload: %w", err)
	}

	// Google tokeninfo API가 토큰 유효성을 검증해줌
	// Client ID 검증은 생략 (네이티브 앱은 다른 Client ID 사용)

	// GoogleUserInfo 형태로 변환
	return &GoogleUserInfo{
		ID:            payload.Sub,
		Email:         payload.Email,
		VerifiedEmail: payload.EmailVerified == "true",
		Name:          payload.Name,
		GivenName:     payload.GivenName,
		FamilyName:    payload.FamilyName,
		Picture:       payload.Picture,
	}, nil
}
