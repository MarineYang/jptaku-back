package service

import (
	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

type UpdateProfileInput struct {
	Name string `json:"name"`
}

type OnboardingInput struct {
	Level     int      `json:"level" binding:"min=0,max=5"`
	Interests []string `json:"interests"`
	Purposes  []string `json:"purposes"`
}

type UpdateSettingsInput struct {
	NotificationEnabled *bool    `json:"notification_enabled,omitempty"`
	DailyReminderTime   *string  `json:"daily_reminder_time,omitempty"`
	PreferredVoiceSpeed *float64 `json:"preferred_voice_speed,omitempty"`
	ShowRomaji          *bool    `json:"show_romaji,omitempty"`
	ShowTranslation     *bool    `json:"show_translation,omitempty"`
}

func (s *UserService) GetMe(userID uint) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *UserService) UpdateProfile(userID uint, input *UpdateProfileInput) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		user.Name = input.Name
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) SaveOnboarding(userID uint, input *OnboardingInput) (*model.UserOnboarding, error) {
	// 기존 온보딩 정보 확인
	onboarding, err := s.userRepo.GetOnboarding(userID)
	if err != nil {
		// 새로 생성
		onboarding = &model.UserOnboarding{
			UserID:    userID,
			Level:     input.Level,
			Interests: input.Interests,
			Purposes:  input.Purposes,
		}
		if err := s.userRepo.CreateOnboarding(onboarding); err != nil {
			return nil, err
		}
	} else {
		// 업데이트
		onboarding.Level = input.Level
		onboarding.Interests = input.Interests
		onboarding.Purposes = input.Purposes
		if err := s.userRepo.UpdateOnboarding(onboarding); err != nil {
			return nil, err
		}
	}

	return onboarding, nil
}

func (s *UserService) GetSettings(userID uint) (*model.UserSettings, error) {
	return s.userRepo.GetSettings(userID)
}

func (s *UserService) UpdateSettings(userID uint, input *UpdateSettingsInput) (*model.UserSettings, error) {
	settings, err := s.userRepo.GetSettings(userID)
	if err != nil {
		// 설정이 없으면 생성
		settings = &model.UserSettings{UserID: userID}
		if err := s.userRepo.CreateSettings(settings); err != nil {
			return nil, err
		}
	}

	if input.NotificationEnabled != nil {
		settings.NotificationEnabled = *input.NotificationEnabled
	}
	if input.DailyReminderTime != nil {
		settings.DailyReminderTime = *input.DailyReminderTime
	}
	if input.PreferredVoiceSpeed != nil {
		settings.PreferredVoiceSpeed = *input.PreferredVoiceSpeed
	}
	if input.ShowRomaji != nil {
		settings.ShowRomaji = *input.ShowRomaji
	}
	if input.ShowTranslation != nil {
		settings.ShowTranslation = *input.ShowTranslation
	}

	if err := s.userRepo.UpdateSettings(settings); err != nil {
		return nil, err
	}

	return settings, nil
}
