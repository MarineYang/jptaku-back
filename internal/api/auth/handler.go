package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jptaku/server/internal/pkg"
	"github.com/jptaku/server/internal/service"
)

type Handler struct {
	authService *service.AuthService
}

func NewHandler(authService *service.AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)
	}
}

// Register godoc
// @Summary 회원가입
// @Description 이메일과 비밀번호로 회원가입
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "회원가입 정보"
// @Success 201 {object} TokenResponse
// @Router /api/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	input := &service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	result, err := h.authService.Register(input)
	if err != nil {
		if err == pkg.ErrDuplicateEmail {
			pkg.BadRequestResponse(c, "이미 등록된 이메일입니다")
			return
		}
		pkg.InternalServerErrorResponse(c, "회원가입 실패")
		return
	}

	pkg.CreatedResponse(c, result)
}

// Login godoc
// @Summary 로그인
// @Description 이메일과 비밀번호로 로그인
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "로그인 정보"
// @Success 200 {object} TokenResponse
// @Router /api/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	input := &service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := h.authService.Login(input)
	if err != nil {
		if err == pkg.ErrInvalidCredentials {
			pkg.UnauthorizedResponse(c, "이메일 또는 비밀번호가 올바르지 않습니다")
			return
		}
		pkg.InternalServerErrorResponse(c, "로그인 실패")
		return
	}

	pkg.SuccessResponse(c, result)
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
	// 실제 로그아웃 처리는 클라이언트에서 토큰을 삭제하면 됨
	// 서버에서는 블랙리스트 처리 등을 할 수 있음 (옵션)
	pkg.SuccessMessageResponse(c, "로그아웃 되었습니다")
}
