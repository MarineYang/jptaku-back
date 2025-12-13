package auth

import "github.com/jptaku/server/internal/model"

// TokenResponse 토큰 응답
type TokenResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         *model.User `json:"user"`
}
