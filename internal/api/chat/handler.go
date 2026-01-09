package chat

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
		chat.POST("/session/:id/message", h.SendMessageSSE) // SSE 스트리밍
	}
}

// CreateSession godoc
// @Summary 대화 세션 생성
// @Description 새로운 AI 대화 세션 생성
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
		log.Println("CreateSession: userID is 0, unauthorized")
		pkg.UnauthorizedResponse(c, "")
		return
	}

	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateSession: JSON binding error: %v", err)
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	log.Printf("CreateSession: userID=%d, topic=%s, topicDetail=%s", userID, req.Topic, req.TopicDetail)

	input := &chatSvc.CreateSessionInput{
		Topic:       req.Topic,
		TopicDetail: req.TopicDetail,
	}

	response, err := h.chatService.CreateSession(userID, input)
	if err != nil {
		log.Printf("CreateSession: service error: %v", err)
		pkg.InternalServerErrorResponse(c, "세션 생성 실패")
		return
	}

	log.Printf("CreateSession: session created successfully, sessionID=%d", response.Session.ID)
	pkg.CreatedResponse(c, response)
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
// @Description 대화 세션을 수동으로 종료
// @Tags Chat
// @Security BearerAuth
// @Produce json
// @Param id path int true "세션 ID"
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

	session, err := h.chatService.EndSession(uint(id))
	if err != nil {
		pkg.InternalServerErrorResponse(c, "세션 종료 실패")
		return
	}

	pkg.SuccessResponse(c, session)
}

// GetSessions godoc
// @Summary 대화 세션 목록 조회
// @Description 유저의 대화 세션 목록 조회
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

// SendMessageSSE godoc
// @Summary 메시지 전송 (SSE 스트리밍)
// @Description AI에게 메시지를 보내고 SSE로 스트리밍 응답 받기
// @Tags Chat
// @Security BearerAuth
// @Accept json
// @Produce text/event-stream
// @Param id path int true "세션 ID"
// @Param request body SendMessageRequest true "메시지"
// @Success 200 {string} string "SSE 스트림"
// @Router /api/chat/session/{id}/message [post]
func (h *Handler) SendMessageSSE(c *gin.Context) {
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

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequestResponse(c, err.Error())
		return
	}

	// SSE 헤더 설정
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // nginx 버퍼링 비활성화

	// 스트리밍 채널 생성
	streamChan := make(chan chatSvc.StreamChunk, 100)

	// 백그라운드에서 OpenAI 스트리밍 시작
	go func() {
		h.chatService.SendMessageStream(c.Request.Context(), uint(id), req.Message, streamChan)
	}()

	// 클라이언트로 SSE 전송
	c.Stream(func(w io.Writer) bool {
		select {
		case chunk, ok := <-streamChan:
			if !ok {
				return false // 채널 닫힘
			}

			data, _ := json.Marshal(chunk)
			c.SSEvent("message", string(data))
			c.Writer.Flush()

			// done 또는 error면 종료
			if chunk.Type == "done" || chunk.Type == "error" {
				return false
			}
			return true

		case <-c.Request.Context().Done():
			return false // 클라이언트 연결 끊김
		}
	})

	// 연결 종료 시 마지막 이벤트
	c.SSEvent("close", "connection closed")
	c.Writer.Flush()
}

// sendSSEError SSE 에러 전송 헬퍼
func sendSSEError(c *gin.Context, message string) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	errorData := fmt.Sprintf(`{"type":"error","error":"%s"}`, message)
	c.SSEvent("message", errorData)
	c.Writer.Flush()
}
