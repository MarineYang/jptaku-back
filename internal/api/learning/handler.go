package learning

import (
	"github.com/gin-gonic/gin"
	"github.com/jptaku/server/internal/middleware"
	"github.com/jptaku/server/internal/pkg"
	"github.com/jptaku/server/internal/service"
)

type Handler struct {
	learningService *service.LearningService
}

func NewHandler(learningService *service.LearningService) *Handler {
	return &Handler{learningService: learningService}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	learning := r.Group("/learning")
	learning.Use(authMiddleware)
	{
		learning.POST("/progress", h.UpdateProgress)
		learning.POST("/quiz", h.SubmitQuiz)
		learning.GET("/today", h.GetTodayProgress)
		learning.GET("/history", h.GetProgressHistory)
	}
}

// UpdateProgress godoc
// @Summary 학습 진행 상황 업데이트
// @Description 문장별 이해/말하기/확인/암기 완료 상태 업데이트
// @Tags Learning
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProgressRequest true "진행 상황"
// @Success 200 {object} model.LearningProgress
// @Router /api/learning/progress [post]
func (h *Handler) UpdateProgress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	var req UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	input := &service.UpdateProgressInput{
		SentenceID: req.SentenceID,
		DailySetID: req.DailySetID,
		Understand: req.Understand,
		Speak:      req.Speak,
		Confirm:    req.Confirm,
		Memorized:  req.Memorized,
	}

	progress, err := h.learningService.UpdateProgress(userID, input)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "진행 상황 업데이트 실패")
		return
	}

	pkg.SuccessResponse(c, progress)
}

// GetTodayProgress godoc
// @Summary 오늘의 학습 진행 상황 조회
// @Description 오늘의 5문장 학습 진행 상황 조회
// @Tags Learning
// @Security BearerAuth
// @Produce json
// @Param daily_set_id query int true "Daily Set ID"
// @Success 200 {object} service.TodayProgressResponse
// @Router /api/learning/today [get]
func (h *Handler) GetTodayProgress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	var query TodayProgressQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		pkg.BadRequestResponse(c, "daily_set_id가 필요합니다")
		return
	}

	progress, err := h.learningService.GetTodayProgress(userID, query.DailySetID)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "진행 상황을 불러오는데 실패했습니다")
		return
	}

	pkg.SuccessResponse(c, progress)
}

// GetProgressHistory godoc
// @Summary 학습 히스토리 조회
// @Description 전체 학습 진행 히스토리 조회
// @Tags Learning
// @Security BearerAuth
// @Produce json
// @Param page query int false "페이지 번호" default(1)
// @Param per_page query int false "페이지당 개수" default(20)
// @Success 200 {object} pkg.PaginatedResponse
// @Router /api/learning/history [get]
func (h *Handler) GetProgressHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	var query ProgressQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		query.Page = 1
		query.PerPage = 20
	}

	progresses, total, err := h.learningService.GetProgress(userID, query.Page, query.PerPage)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "히스토리를 불러오는데 실패했습니다")
		return
	}

	pkg.PaginatedSuccessResponse(c, progresses, query.Page, query.PerPage, total)
}

// SubmitQuiz godoc
// @Summary 퀴즈 제출
// @Description 빈칸 채우기/문장 배열 퀴즈 정답 제출 및 검증. 모두 맞으면 해당 문장 암기 완료로 표시
// @Tags Learning
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body SubmitQuizRequest true "퀴즈 답안"
// @Success 200 {object} SubmitQuizResponse
// @Router /api/learning/quiz [post]
func (h *Handler) SubmitQuiz(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	var req SubmitQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	input := &service.SubmitQuizInput{
		SentenceID:      req.SentenceID,
		DailySetID:      req.DailySetID,
		FillBlankAnswer: req.FillBlankAnswer,
		OrderingAnswer:  req.OrderingAnswer,
	}

	result, err := h.learningService.SubmitQuiz(userID, input)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "퀴즈 제출 실패")
		return
	}

	response := SubmitQuizResponse{
		SentenceID:       result.SentenceID,
		FillBlankCorrect: result.FillBlankCorrect,
		OrderingCorrect:  result.OrderingCorrect,
		AllCorrect:       result.AllCorrect,
		Memorized:        result.Memorized,
	}

	pkg.SuccessResponse(c, response)
}
