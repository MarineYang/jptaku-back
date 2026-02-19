package auth

import (
	"context"
	"log"

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

// RefreshToken 토큰 갱신 (비회원 토큰은 갱신 불가)
func (s *Service) RefreshToken(refreshToken string) (*TokenResponse, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, pkg.ErrInvalidToken
	}

	// 비회원 토큰은 갱신 불가
	if claims.UserID == 0 {
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

// GuestLogin 비회원 토큰 발급 (DB 저장 없음, 24시간 유효)
func (s *Service) GuestLogin() (*TokenResponse, error) {
	accessToken, err := s.jwtManager.GenerateToken(0, "guest")
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken: accessToken,
		IsGuest:     true,
	}, nil
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

// GoogleIDTokenLogin 네이티브 앱에서 Google SDK로 로그인 후 받은 ID Token으로 로그인
func (s *Service) GoogleIDTokenLogin(ctx context.Context, idToken string) (*TokenResponse, error) {
	if s.googleOAuth == nil {
		log.Println("GoogleIDTokenLogin: googleOAuth is nil")
		return nil, pkg.ErrInvalidCredentials
	}

	// ID Token 검증 및 사용자 정보 추출
	userInfo, err := s.googleOAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		log.Printf("GoogleIDTokenLogin: VerifyIDToken failed: %v", err)
		return nil, pkg.ErrInvalidCredentials
	}

	// 기존 사용자 조회
	user, err := s.userRepo.FindByProviderID("google", userInfo.ID)
	if err != nil {
		if repository.IsNotFound(err) {
			return s.createGoogleUser(userInfo)
		}
		return nil, err
	}

	return s.generateTokens(user)
}
