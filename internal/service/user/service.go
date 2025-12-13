package user

import "github.com/jptaku/server/internal/model"

// Service 사용자 서비스
type Service struct {
	userRepo        UserRepository
	sentenceService SentenceProvider
}

// 컴파일 타임 인터페이스 검증
var _ Provider = (*Service)(nil)

// NewService 서비스 생성자
func NewService(userRepo UserRepository, sentenceService SentenceProvider) *Service {
	return &Service{
		userRepo:        userRepo,
		sentenceService: sentenceService,
	}
}

// GetMe 사용자 정보 조회
func (s *Service) GetMe(userID uint) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

// UpdateProfile 프로필 업데이트
func (s *Service) UpdateProfile(userID uint, input *UpdateProfileInput) (*model.User, error) {
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

// SaveOnboarding 온보딩 저장
func (s *Service) SaveOnboarding(userID uint, input *OnboardingInput) (*model.UserOnboarding, error) {
	onboarding, err := s.userRepo.GetOnboarding(userID)
	isNewOnboarding := err != nil

	if isNewOnboarding {
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
		onboarding.Level = input.Level
		onboarding.Interests = input.Interests
		onboarding.Purposes = input.Purposes
		if err := s.userRepo.UpdateOnboarding(onboarding); err != nil {
			return nil, err
		}
	}

	return onboarding, nil
}

// GetSettings 설정 조회
func (s *Service) GetSettings(userID uint) (*model.UserSettings, error) {
	return s.userRepo.GetSettings(userID)
}

// UpdateSettings 설정 업데이트
func (s *Service) UpdateSettings(userID uint, input *UpdateSettingsInput) (*model.UserSettings, error) {
	settings, err := s.userRepo.GetSettings(userID)
	if err != nil {
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
