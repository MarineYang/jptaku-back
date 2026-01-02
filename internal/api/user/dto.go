package user

type UpdateProfileRequest struct {
	Name string `json:"name"`
}

type OnboardingRequest struct {
	Categories []int `json:"categories" binding:"required,min=1"` // pkg.OnboardingCategory 값들 (애니/게임/음악/영화/드라마)
	Level      int   `json:"level" binding:"required,oneof=3 4 5"` // pkg.Level 값들 (N3=3, N4=4, N5=5)
}

type UpdateSettingsRequest struct {
	NotificationEnabled *bool    `json:"notification_enabled,omitempty"`
	DailyReminderTime   *string  `json:"daily_reminder_time,omitempty"`
	PreferredVoiceSpeed *float64 `json:"preferred_voice_speed,omitempty"`
	ShowRomaji          *bool    `json:"show_romaji,omitempty"`
	ShowTranslation     *bool    `json:"show_translation,omitempty"`
}
