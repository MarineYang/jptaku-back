package auth

import "github.com/jptaku/server/internal/model"

// TokenResponse 토큰 응답
type TokenResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token,omitempty"` // 비회원은 refresh token 없음
	User         *model.User `json:"user,omitempty"`
	IsGuest      bool        `json:"is_guest,omitempty"` // 비회원 토큰 여부
}
