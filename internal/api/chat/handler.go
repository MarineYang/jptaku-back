package chat

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jptaku/server/internal/middleware"
	"github.com/jptaku/server/internal/pkg"
	chatSvc "github.com/jptaku/server/internal/service/chat"
)

type Handler struct {
	chatService chatSvc.Provider
}

func NewHandler(chatService chatSvc.Provider) *Handler {
	return &Handler{chatService: chatService}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	chat := r.Group("/chat")
	chat.Use(authMiddleware)
	{
		chat.POST("/session", h.CreateSession)
		chat.GET("/session/:id", h.GetSession)
		chat.POST("/session/:id/end", h.EndSession)
		chat.GET("/sessions", h.GetSessions)
	}
}

// CreateSession godoc
// @Summary 대화 세션 생성
// @Description 새로운 대화 세션 생성 (오늘의 5문장 세트와 연결)
// @Tags Chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateSessionRequest true "세션 정보"
// @Success 201 {object} model.ChatSession
// @Router /api/chat/session [post]
func (h *Handler) CreateSession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	input := &chatSvc.CreateSessionInput{
		DailySetID: req.DailySetID,
	}

	session, err := h.chatService.CreateSession(userID, input)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "세션 생성 실패")
		return
	}

	pkg.CreatedResponse(c, session)
}

// GetSession godoc
// @Summary 대화 세션 조회
// @Description 대화 세션 상세 정보 조회 (메시지 포함)
// @Tags Chat
// @Security BearerAuth
// @Produce json
// @Param id path int true "세션 ID"
// @Success 200 {object} model.ChatSession
// @Router /api/chat/session/{id} [get]
func (h *Handler) GetSession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		pkg.BadRequestResponse(c, "유효하지 않은 세션 ID입니다")
		return
	}

	session, err := h.chatService.GetSession(uint(id))
	if err != nil {
		pkg.NotFoundResponse(c, "세션을 찾을 수 없습니다")
		return
	}

	// 권한 확인
	if session.UserID != userID {
		pkg.ForbiddenResponse(c, "접근 권한이 없습니다")
		return
	}

	pkg.SuccessResponse(c, session)
}

// EndSession godoc
// @Summary 대화 세션 종료
// @Description 대화 세션을 종료하고 통계 저장
// @Tags Chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "세션 ID"
// @Param request body EndSessionRequest true "종료 정보"
// @Success 200 {object} model.ChatSession
// @Router /api/chat/session/{id}/end [post]
func (h *Handler) EndSession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		pkg.BadRequestResponse(c, "유효하지 않은 세션 ID입니다")
		return
	}

	var req EndSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	// 권한 확인
	existingSession, err := h.chatService.GetSession(uint(id))
	if err != nil {
		pkg.NotFoundResponse(c, "세션을 찾을 수 없습니다")
		return
	}
	if existingSession.UserID != userID {
		pkg.ForbiddenResponse(c, "접근 권한이 없습니다")
		return
	}

	input := &chatSvc.EndSessionInput{
		DurationSeconds: req.DurationSeconds,
	}

	session, err := h.chatService.EndSession(uint(id), input)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "세션 종료 실패")
		return
	}

	pkg.SuccessResponse(c, session)
}

// GetSessions godoc
// @Summary 대화 세션 목록 조회
// @Description 유저의 최근 대화 세션 목록 조회
// @Tags Chat
// @Security BearerAuth
// @Produce json
// @Param page query int false "페이지 번호" default(1)
// @Param per_page query int false "페이지당 개수" default(20)
// @Success 200 {object} pkg.PaginatedResponse
// @Router /api/chat/sessions [get]
func (h *Handler) GetSessions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	var query SessionsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		query.Page = 1
		query.PerPage = 20
	}

	sessions, total, err := h.chatService.GetSessions(userID, query.Page, query.PerPage)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "세션 목록을 불러오는데 실패했습니다")
		return
	}

	pkg.PaginatedSuccessResponse(c, sessions, query.Page, query.PerPage, total)
}

