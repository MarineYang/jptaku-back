package service

import (
	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/pkg"
	"github.com/jptaku/server/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db         *repository.DBManager
	userRepo   *repository.UserRepository
	jwtManager *pkg.JWTManager
}

func NewAuthService(db *repository.DBManager, userRepo *repository.UserRepository, jwtManager *pkg.JWTManager) *AuthService {
	return &AuthService{
		db:         db,
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         *model.User `json:"user"`
}

func (s *AuthService) Register(input *RegisterInput) (*TokenResponse, error) {
	// 이메일 중복 확인
	existingUser, _ := s.userRepo.FindByEmail(input.Email)
	if existingUser != nil {
		return nil, pkg.ErrDuplicateEmail
	}

	// 비밀번호 해시
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user *model.User

	// 트랜잭션으로 유저와 설정을 함께 생성
	err = s.db.Transaction(func(tx *gorm.DB) error {
		user = &model.User{
			Email:    input.Email,
			Password: string(hashedPassword),
			Name:     input.Name,
			Provider: "local",
		}

		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// 기본 설정 생성
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

func (s *AuthService) Login(input *LoginInput) (*TokenResponse, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, pkg.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, pkg.ErrInvalidCredentials
	}

	return s.generateTokens(user)
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
