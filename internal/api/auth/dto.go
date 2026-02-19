package auth

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	User         any    `json:"user,omitempty"`
	IsGuest      bool   `json:"is_guest,omitempty"`
}

// GoogleIDTokenRequest 네이티브 앱에서 Google SDK로 로그인 후 받은 ID Token
type GoogleIDTokenRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}
