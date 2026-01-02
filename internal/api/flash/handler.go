package flash

import (
	"github.com/gin-gonic/gin"
	"github.com/jptaku/server/internal/middleware"
	"github.com/jptaku/server/internal/pkg"
	flashSvc "github.com/jptaku/server/internal/service/flash"
)

type Handler struct {
	flashService flashSvc.Provider
}

func NewHandler(flashService flashSvc.Provider) *Handler {
	return &Handler{flashService: flashService}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	flash := r.Group("/flash")
	flash.Use(authMiddleware)
	{
		flash.GET("/today", h.GetTodayFlash)
		flash.POST("/progress", h.UpdateFlashProgress)
	}
}

// GetTodayFlash godoc
// @Summary 오늘의 Flash 문장 조회
// @Description 오늘의 5문장을 Flash 카드 형식으로 조회. 상세 정보(phrase/tip/alt)가 없으면 자동 생성
// @Tags Flash
// @Security BearerAuth
// @Produce json
// @Success 200 {object} flashSvc.TodayFlashResponse
// @Router /api/flash/today [get]
func (h *Handler) GetTodayFlash(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	result, err := h.flashService.GetTodayFlash(c.Request.Context(), userID)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "Flash 문장을 불러오는데 실패했습니다: "+err.Error())
		return
	}

	pkg.SuccessResponse(c, result)
}

// UpdateFlashProgress godoc
// @Summary Flash 진행 상황 업데이트
// @Description Flash 카드 학습 후 평가(bad/mid/good) 저장
// @Tags Flash
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateFlashProgressRequest true "Flash 평가"
// @Success 200 {object} flashSvc.FlashProgressResult
// @Router /api/flash/progress [post]
func (h *Handler) UpdateFlashProgress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	var req UpdateFlashProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	input := &flashSvc.UpdateFlashInput{
		SentenceID: req.SentenceID,
		Grade:      req.Grade,
	}

	result, err := h.flashService.UpdateFlashProgress(c.Request.Context(), userID, input)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "Flash 진행 상황 업데이트 실패: "+err.Error())
		return
	}

	pkg.SuccessResponse(c, result)
}
