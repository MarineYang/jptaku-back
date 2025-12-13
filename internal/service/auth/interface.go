package auth

import (
	"context"

	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/pkg"
	"gorm.io/gorm"
)

// UserRepository 사용자 저장소 인터페이스
type UserRepository interface {
	FindByID(id uint) (*model.User, error)
	FindByProviderID(provider, providerID string) (*model.User, error)
}

// DBManager 데이터베이스 매니저 인터페이스
type DBManager interface {
	Transaction(fc func(tx *gorm.DB) error) error
}

// JWTManager JWT 매니저 인터페이스
type JWTManager interface {
	GenerateToken(userID uint, email string) (string, error)
	GenerateRefreshToken(userID uint, email string) (string, error)
	ValidateToken(tokenString string) (*pkg.JWTClaims, error)
}

// GoogleOAuthManager Google OAuth 매니저 인터페이스
type GoogleOAuthManager interface {
	GetAuthURL(state string) string
	Exchange(ctx context.Context, code string) (interface{}, error)
	GetUserInfo(ctx context.Context, token interface{}) (*pkg.GoogleUserInfo, error)
}

// Provider 서비스 인터페이스 (외부에서 사용)
type Provider interface {
	RefreshToken(refreshToken string) (*TokenResponse, error)
	GetGoogleAuthURL(state string) string
	GoogleCallback(ctx context.Context, code string) (*TokenResponse, error)
	SetGoogleOAuth(googleOAuth *pkg.GoogleOAuthManager)
}
