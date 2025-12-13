package auth

import (
	"context"

	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/pkg"
	"github.com/jptaku/server/internal/repository"
	"gorm.io/gorm"
)

// Service 인증 서비스
type Service struct {
	db          *repository.DBManager
	userRepo    UserRepository
	jwtManager  *pkg.JWTManager
	googleOAuth *pkg.GoogleOAuthManager
}

// 컴파일 타임 인터페이스 검증
var _ Provider = (*Service)(nil)

// NewService 서비스 생성자
func NewService(db *repository.DBManager, userRepo UserRepository, jwtManager *pkg.JWTManager) *Service {
	return &Service{
		db:         db,
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// SetGoogleOAuth Google OAuth 설정
func (s *Service) SetGoogleOAuth(googleOAuth *pkg.GoogleOAuthManager) {
	s.googleOAuth = googleOAuth
}

// RefreshToken 토큰 갱신
func (s *Service) RefreshToken(refreshToken string) (*TokenResponse, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, pkg.ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, pkg.ErrNotFound
	}

	return s.generateTokens(user)
}

// generateTokens 토큰 생성
func (s *Service) generateTokens(user *model.User) (*TokenResponse, error) {
	accessToken, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// GetGoogleAuthURL Google 로그인 URL 조회
func (s *Service) GetGoogleAuthURL(state string) string {
	if s.googleOAuth == nil {
		return ""
	}
	return s.googleOAuth.GetAuthURL(state)
}

// GoogleCallback Google 로그인 콜백 처리
func (s *Service) GoogleCallback(ctx context.Context, code string) (*TokenResponse, error) {
	if s.googleOAuth == nil {
		return nil, pkg.ErrInvalidCredentials
	}

	token, err := s.googleOAuth.Exchange(ctx, code)
	if err != nil {
		return nil, pkg.ErrInvalidCredentials
	}

	userInfo, err := s.googleOAuth.GetUserInfo(ctx, token)
	if err != nil {
		return nil, pkg.ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByProviderID("google", userInfo.ID)
	if err != nil {
		if repository.IsNotFound(err) {
			return s.createGoogleUser(userInfo)
		}
		return nil, err
	}

	return s.generateTokens(user)
}

// createGoogleUser Google 사용자 생성
func (s *Service) createGoogleUser(userInfo *pkg.GoogleUserInfo) (*TokenResponse, error) {
	var user *model.User

	err := s.db.Transaction(func(tx *gorm.DB) error {
		user = &model.User{
			Email:      userInfo.Email,
			Name:       userInfo.Name,
			Provider:   "google",
			ProviderID: userInfo.ID,
		}

		if err := tx.Create(user).Error; err != nil {
			return err
		}

		settings := &model.UserSettings{
			UserID: user.ID,
		}
		return tx.Create(settings).Error
	})

	if err != nil {
		return nil, err
	}

	return s.generateTokens(user)
}
