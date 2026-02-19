package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jptaku/server/internal/pkg"
	authSvc "github.com/jptaku/server/internal/service/auth"
)

type Handler struct {
	authService authSvc.Provider
}

func NewHandler(authService authSvc.Provider) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)
		auth.POST("/guest", h.GuestLogin)

		// Google OAuth (네이티브 앱용 - ID Token 검증)
		auth.POST("/google/token", h.GoogleIDTokenLogin)
	}
}

// Refresh godoc
// @Summary 토큰 갱신
// @Description Refresh Token으로 Access Token 갱신
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh Token"
// @Success 200 {object} TokenResponse
// @Router /api/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	result, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		pkg.UnauthorizedResponse(c, "유효하지 않은 토큰입니다")
		return
	}

	pkg.SuccessResponse(c, result)
}

// Logout godoc
// @Summary 로그아웃
// @Description 로그아웃 (클라이언트에서 토큰 삭제)
// @Tags Auth
// @Success 200 {object} pkg.Response
// @Router /api/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	pkg.SuccessMessageResponse(c, "로그아웃 되었습니다")
}

// GuestLogin godoc
// @Summary 비회원 로그인
// @Description DB 저장 없이 24시간짜리 비회원 토큰 발급
// @Tags Auth
// @Produce json
// @Success 200 {object} TokenResponse
// @Router /api/auth/guest [post]
func (h *Handler) GuestLogin(c *gin.Context) {
	result, err := h.authService.GuestLogin()
	if err != nil {
		pkg.InternalServerErrorResponse(c, "비회원 토큰 발급 실패")
		return
	}

	pkg.SuccessResponse(c, result)
}

// GoogleIDTokenLogin godoc
// @Summary Google ID Token 로그인 (네이티브 앱용)
// @Description 네이티브 앱에서 Google SDK로 로그인 후 받은 ID Token으로 로그인합니다
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body GoogleIDTokenRequest true "Google ID Token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} pkg.Response "잘못된 요청"
// @Failure 401 {object} pkg.Response "유효하지 않은 토큰"
// @Router /api/auth/google/token [post]
func (h *Handler) GoogleIDTokenLogin(c *gin.Context) {
	var req GoogleIDTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, "ID Token이 필요합니다")
		return
	}

	result, err := h.authService.GoogleIDTokenLogin(c.Request.Context(), req.IDToken)
	if err != nil {
		pkg.UnauthorizedResponse(c, "Google 로그인에 실패했습니다")
		return
	}

	pkg.SuccessResponse(c, result)
}
