package sentences

import (
	"github.com/gin-gonic/gin"
	"github.com/jptaku/server/internal/middleware"
	"github.com/jptaku/server/internal/pkg"
	"github.com/jptaku/server/internal/service"
)

type Handler struct {
	sentenceService *service.SentenceService
}

func NewHandler(sentenceService *service.SentenceService) *Handler {
	return &Handler{sentenceService: sentenceService}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	sentences := r.Group("/sentences")
	sentences.Use(authMiddleware)
	{
		sentences.GET("/today", h.GetTodaySentences)
		sentences.GET("/yesterday", h.GetYesterdaySentences)
	}
}

// GetTodaySentences godoc
// @Summary 오늘의 5문장 조회
// @Description 유저의 레벨/관심사 기반 오늘의 5문장 반환 (첫 호출 시 생성)
// @Tags Sentences
// @Security BearerAuth
// @Produce json
// @Success 200 {object} DailySentencesResponse
// @Router /api/sentences/today [get]
func (h *Handler) GetTodaySentences(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	result, err := h.sentenceService.GetTodaySentences(userID)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "오늘의 문장을 불러오는데 실패했습니다")
		return
	}

	response := convertToResponse(result)
	pkg.SuccessResponse(c, response)
}

// GetYesterdaySentences godoc
// @Summary 어제의 5문장 조회
// @Description 어제 학습한 5문장 반환
// @Tags Sentences
// @Security BearerAuth
// @Produce json
// @Success 200 {object} DailySentencesResponse
// @Router /api/sentences/yesterday [get]
func (h *Handler) GetYesterdaySentences(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	result, err := h.sentenceService.GetYesterdaySentences(userID)
	if err != nil {
		pkg.NotFoundResponse(c, "어제의 문장이 없습니다")
		return
	}

	response := convertToResponse(result)
	pkg.SuccessResponse(c, response)
}

func convertToResponse(result *service.DailySentencesResponse) *DailySentencesResponse {
	sentences := make([]SentenceResponse, len(result.Sentences))
	for i, s := range result.Sentences {
		sentences[i] = SentenceResponse{
			ID:         s.ID,
			JP:         s.JP,
			KR:         s.KR,
			Romaji:     s.Romaji,
			Level:      s.Level,
			Categories: s.Categories,
		}
	}

	return &DailySentencesResponse{
		Date:      result.Date,
		Sentences: sentences,
	}
}
