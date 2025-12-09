package service

import (
	"context"

	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/pkg"
	"github.com/jptaku/server/internal/repository"
	"gorm.io/gorm"
)

type AuthService struct {
	db          *repository.DBManager
	userRepo    *repository.UserRepository
	jwtManager  *pkg.JWTManager
	googleOAuth *pkg.GoogleOAuthManager
}

func NewAuthService(db *repository.DBManager, userRepo *repository.UserRepository, jwtManager *pkg.JWTManager) *AuthService {
	return &AuthService{
		db:         db,
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) SetGoogleOAuth(googleOAuth *pkg.GoogleOAuthManager) {
	s.googleOAuth = googleOAuth
}

type TokenResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         *model.User `json:"user"`
}

func (s *AuthService) RefreshToken(refreshToken string) (*TokenResponse, error) {
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

func (s *AuthService) generateTokens(user *model.User) (*TokenResponse, error) {
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

func (s *AuthService) GetGoogleAuthURL(state string) string {
	if s.googleOAuth == nil {
		return ""
	}
	return s.googleOAuth.GetAuthURL(state)
}

func (s *AuthService) GoogleCallback(ctx context.Context, code string) (*TokenResponse, error) {
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

func (s *AuthService) createGoogleUser(userInfo *pkg.GoogleUserInfo) (*TokenResponse, error) {
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
