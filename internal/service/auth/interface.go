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

// Provider 서비스 인터페이스 (외부에서 사용)
type Provider interface {
	RefreshToken(refreshToken string) (*TokenResponse, error)
	SetGoogleOAuth(googleOAuth *pkg.GoogleOAuthManager)
	GuestLogin() (*TokenResponse, error)

	// 네이티브 앱용 Google ID Token 검증
	GoogleIDTokenLogin(ctx context.Context, idToken string) (*TokenResponse, error)
}
