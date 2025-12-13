package user

// UpdateProfileInput 프로필 업데이트 입력
type UpdateProfileInput struct {
	Name string `json:"name"`
}

// OnboardingInput 온보딩 입력
type OnboardingInput struct {
	Level     int   `json:"level" binding:"min=0,max=5"`
	Interests []int `json:"interests"` // pkg.SubCategory 값들
	Purposes  []int `json:"purposes"`  // pkg.Purpose 값들
}

// UpdateSettingsInput 설정 업데이트 입력
type UpdateSettingsInput struct {
	NotificationEnabled *bool    `json:"notification_enabled,omitempty"`
	DailyReminderTime   *string  `json:"daily_reminder_time,omitempty"`
	PreferredVoiceSpeed *float64 `json:"preferred_voice_speed,omitempty"`
	ShowRomaji          *bool    `json:"show_romaji,omitempty"`
	ShowTranslation     *bool    `json:"show_translation,omitempty"`
}
