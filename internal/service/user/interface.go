package user

import "github.com/jptaku/server/internal/model"

// UserRepository 사용자 저장소 인터페이스
type UserRepository interface {
	FindByID(id uint) (*model.User, error)
	Update(user *model.User) error
	GetOnboarding(userID uint) (*model.UserOnboarding, error)
	CreateOnboarding(onboarding *model.UserOnboarding) error
	UpdateOnboarding(onboarding *model.UserOnboarding) error
	GetSettings(userID uint) (*model.UserSettings, error)
	CreateSettings(settings *model.UserSettings) error
	UpdateSettings(settings *model.UserSettings) error
}

// SentenceProvider 문장 서비스 인터페이스
type SentenceProvider interface {
	// 필요한 메서드가 있으면 추가
}

// Provider 서비스 인터페이스 (외부에서 사용)
type Provider interface {
	GetMe(userID uint) (*model.User, error)
	UpdateProfile(userID uint, input *UpdateProfileInput) (*model.User, error)
	SaveOnboarding(userID uint, input *OnboardingInput) (*model.UserOnboarding, error)
	GetSettings(userID uint) (*model.UserSettings, error)
	UpdateSettings(userID uint, input *UpdateSettingsInput) (*model.UserSettings, error)
}
