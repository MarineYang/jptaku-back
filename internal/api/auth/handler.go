package auth

import (
	"fmt"

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

		// Google OAuth
		auth.GET("/google", h.GoogleAuth)
		auth.GET("/google/callback", h.GoogleCallback)
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

// GoogleAuth godoc
// @Summary Google OAuth 로그인 URL 생성
// @Description Google 로그인을 위한 인증 URL을 반환합니다
// @Tags Auth
// @Produce json
// @Success 200 {object} GoogleAuthURLResponse
// @Router /api/auth/google [get]
func (h *Handler) GoogleAuth(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		state = "default"
	}

	url := h.authService.GetGoogleAuthURL(state)
	if url == "" {
		pkg.InternalServerErrorResponse(c, "Google OAuth가 설정되지 않았습니다")
		return
	}

	pkg.SuccessResponse(c, GoogleAuthURLResponse{URL: url})
}

// GoogleCallback godoc
// @Summary Google OAuth 콜백
// @Description Google 로그인 후 콜백을 처리합니다. 모바일 앱인 경우 딥링크로 리다이렉트합니다.
// @Tags Auth
// @Produce json
// @Param code query string true "Authorization code from Google"
// @Param state query string false "State parameter (mobile: 모바일 앱으로 리다이렉트)"
// @Success 200 {object} TokenResponse
// @Router /api/auth/google/callback [get]
func (h *Handler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		pkg.BadRequestResponse(c, "인증 코드가 필요합니다")
		return
	}

	state := c.Query("state")

	result, err := h.authService.GoogleCallback(c.Request.Context(), code)
	if err != nil {
		fmt.Println("Google callback error:", err)
		// 모바일 앱인 경우 에러도 딥링크로 전달
		if state == "mobile" {
			c.Redirect(302, "jptaku://auth/callback?error=login_failed")
			return
		}
		pkg.UnauthorizedResponse(c, "Google 로그인에 실패했습니다")
		return
	}

	// 모바일 앱인 경우 딥링크로 리다이렉트
	if state == "mobile" {
		redirectURL := fmt.Sprintf("jptaku://auth/callback?access_token=%s&refresh_token=%s",
			result.AccessToken,
			result.RefreshToken,
		)
		c.Redirect(302, redirectURL)
		return
	}

	// 웹인 경우 JSON 응답
	pkg.SuccessResponse(c, result)
}
