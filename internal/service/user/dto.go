package user

// UpdateProfileInput 프로필 업데이트 입력
type UpdateProfileInput struct {
	Name string `json:"name"`
}

// OnboardingInput 온보딩 입력
type OnboardingInput struct {
	Categories []int `json:"categories"` // pkg.OnboardingCategory 값들 (애니/게임/음악/영화/드라마)
	Level      int   `json:"level"`      // pkg.Level 값들 (N3=3, N4=4, N5=5)
}

// UpdateSettingsInput 설정 업데이트 입력
type UpdateSettingsInput struct {
	NotificationEnabled *bool    `json:"notification_enabled,omitempty"`
	DailyReminderTime   *string  `json:"daily_reminder_time,omitempty"`
	PreferredVoiceSpeed *float64 `json:"preferred_voice_speed,omitempty"`
	ShowRomaji          *bool    `json:"show_romaji,omitempty"`
	ShowTranslation     *bool    `json:"show_translation,omitempty"`
}
