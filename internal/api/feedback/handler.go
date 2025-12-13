package feedback

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jptaku/server/internal/middleware"
	"github.com/jptaku/server/internal/pkg"
	feedbackSvc "github.com/jptaku/server/internal/service/feedback"
)

type Handler struct {
	feedbackService feedbackSvc.Provider
}

func NewHandler(feedbackService feedbackSvc.Provider) *Handler {
	return &Handler{feedbackService: feedbackService}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	feedback := r.Group("/feedback")
	feedback.Use(authMiddleware)
	{
		feedback.GET("/:sessionId", h.GetFeedback)
	}

	stats := r.Group("/stats")
	stats.Use(authMiddleware)
	{
		stats.GET("/today", h.GetTodayStats)
		stats.GET("/categories", h.GetCategoryProgress)
		stats.GET("/weekly", h.GetWeeklyStats)
	}
}

// GetFeedback godoc
// @Summary 대화 피드백 조회
// @Description 세션 ID로 피드백 조회 (총점, 문법, 발음, 자연스러움 등)
// @Tags Feedback
// @Security BearerAuth
// @Produce json
// @Param sessionId path int true "세션 ID"
// @Success 200 {object} model.Feedback
// @Router /api/feedback/{sessionId} [get]
func (h *Handler) GetFeedback(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	sessionIDParam := c.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDParam, 10, 32)
	if err != nil {
		pkg.BadRequestResponse(c, "유효하지 않은 세션 ID입니다")
		return
	}

	feedback, err := h.feedbackService.GetFeedback(uint(sessionID))
	if err != nil {
		pkg.NotFoundResponse(c, "피드백을 찾을 수 없습니다")
		return
	}

	pkg.SuccessResponse(c, feedback)
}

// GetTodayStats godoc
// @Summary 오늘의 통계 조회
// @Description 오늘의 학습 요약 통계 (사용 문장 수, 학습 시간, streak 등)
// @Tags Stats
// @Security BearerAuth
// @Produce json
// @Success 200 {object} StatsResponse
// @Router /api/stats/today [get]
func (h *Handler) GetTodayStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	stats, err := h.feedbackService.GetTodayStats(userID)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "통계를 불러오는데 실패했습니다")
		return
	}

	pkg.SuccessResponse(c, stats)
}

// GetCategoryProgress godoc
// @Summary 카테고리별 진행도 조회
// @Description 오타쿠 카테고리별 학습 진행도 (애니, 게임, 성지순례, 이벤트)
// @Tags Stats
// @Security BearerAuth
// @Produce json
// @Success 200 {array} CategoryProgress
// @Router /api/stats/categories [get]
func (h *Handler) GetCategoryProgress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	progress, err := h.feedbackService.GetCategoryProgress(userID)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "카테고리 진행도를 불러오는데 실패했습니다")
		return
	}

	pkg.SuccessResponse(c, progress)
}

// GetWeeklyStats godoc
// @Summary 주간 통계 조회
// @Description 최근 7일간의 학습 통계
// @Tags Stats
// @Security BearerAuth
// @Produce json
// @Success 200 {array} WeeklyStats
// @Router /api/stats/weekly [get]
func (h *Handler) GetWeeklyStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	stats, err := h.feedbackService.GetWeeklyStats(userID)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "주간 통계를 불러오는데 실패했습니다")
		return
	}

	pkg.SuccessResponse(c, stats)
}

