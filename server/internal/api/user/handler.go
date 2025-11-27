package user

import (
	"github.com/gin-gonic/gin"
	"github.com/jptaku/server/internal/middleware"
	"github.com/jptaku/server/internal/pkg"
	"github.com/jptaku/server/internal/service"
)

type Handler struct {
	userService *service.UserService
	jwtManager  *pkg.JWTManager
}

func NewHandler(userService *service.UserService, jwtManager *pkg.JWTManager) *Handler {
	return &Handler{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	user := r.Group("/user")
	user.Use(authMiddleware)
	{
		user.GET("/me", h.GetMe)
		user.PUT("/profile", h.UpdateProfile)
		user.POST("/onboarding", h.SaveOnboarding)
		user.GET("/settings", h.GetSettings)
		user.PUT("/settings", h.UpdateSettings)
	}
}

// GetMe godoc
// @Summary 내 정보 조회
// @Description 현재 로그인한 유저의 정보 조회
// @Tags User
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.User
// @Router /api/user/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	user, err := h.userService.GetMe(userID)
	if err != nil {
		pkg.NotFoundResponse(c, "유저를 찾을 수 없습니다")
		return
	}

	pkg.SuccessResponse(c, user)
}

// UpdateProfile godoc
// @Summary 프로필 수정
// @Description 유저 프로필 정보 수정
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "프로필 정보"
// @Success 200 {object} model.User
// @Router /api/user/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	input := &service.UpdateProfileInput{
		Name: req.Name,
	}

	user, err := h.userService.UpdateProfile(userID, input)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "프로필 수정 실패")
		return
	}

	pkg.SuccessResponse(c, user)
}

// SaveOnboarding godoc
// @Summary 온보딩 정보 저장
// @Description 유저 온보딩 정보 저장 (레벨, 관심사, 목적)
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body OnboardingRequest true "온보딩 정보"
// @Success 200 {object} model.UserOnboarding
// @Router /api/user/onboarding [post]
func (h *Handler) SaveOnboarding(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	var req OnboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	input := &service.OnboardingInput{
		Level:     req.Level,
		Interests: req.Interests,
		Purposes:  req.Purposes,
	}

	onboarding, err := h.userService.SaveOnboarding(userID, input)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "온보딩 정보 저장 실패")
		return
	}

	pkg.SuccessResponse(c, onboarding)
}

// GetSettings godoc
// @Summary 설정 조회
// @Description 유저 설정 정보 조회
// @Tags User
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.UserSettings
// @Router /api/user/settings [get]
func (h *Handler) GetSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	settings, err := h.userService.GetSettings(userID)
	if err != nil {
		pkg.NotFoundResponse(c, "설정을 찾을 수 없습니다")
		return
	}

	pkg.SuccessResponse(c, settings)
}

// UpdateSettings godoc
// @Summary 설정 수정
// @Description 유저 설정 정보 수정
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateSettingsRequest true "설정 정보"
// @Success 200 {object} model.UserSettings
// @Router /api/user/settings [put]
func (h *Handler) UpdateSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	input := &service.UpdateSettingsInput{
		NotificationEnabled: req.NotificationEnabled,
		DailyReminderTime:   req.DailyReminderTime,
		PreferredVoiceSpeed: req.PreferredVoiceSpeed,
		ShowRomaji:          req.ShowRomaji,
		ShowTranslation:     req.ShowTranslation,
	}

	settings, err := h.userService.UpdateSettings(userID, input)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "설정 수정 실패")
		return
	}

	pkg.SuccessResponse(c, settings)
}
