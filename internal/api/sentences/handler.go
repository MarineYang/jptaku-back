package sentences

import (
	"strconv"

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
		sentences.GET("/daily", h.GetDailySentences)
		sentences.GET("/:id", h.GetSentence)
		sentences.GET("/history", h.GetHistory)
	}
}

// GetDailySentences godoc
// @Summary 오늘의 5문장 조회
// @Description 유저의 레벨/관심사 기반 오늘의 5문장 반환 (첫 호출 시 생성)
// @Tags Sentences
// @Security BearerAuth
// @Produce json
// @Success 200 {object} DailySentencesResponse
// @Router /api/sentences/daily [get]
func (h *Handler) GetDailySentences(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	result, err := h.sentenceService.GetDailySentences(userID)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "오늘의 문장을 불러오는데 실패했습니다")
		return
	}

	pkg.SuccessResponse(c, result)
}

// GetSentence godoc
// @Summary 문장 상세 조회
// @Description 문장 ID로 문장 상세 정보 조회 (단어 풀이, 문법, 예문 포함)
// @Tags Sentences
// @Security BearerAuth
// @Produce json
// @Param id path int true "문장 ID"
// @Success 200 {object} SentenceDetailResponse
// @Router /api/sentences/{id} [get]
func (h *Handler) GetSentence(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		pkg.BadRequestResponse(c, "유효하지 않은 문장 ID입니다")
		return
	}

	sentence, detail, err := h.sentenceService.GetSentenceDetail(uint(id))
	if err != nil {
		pkg.NotFoundResponse(c, "문장을 찾을 수 없습니다")
		return
	}

	response := SentenceDetailResponse{
		Sentence: SentenceResponse{
			ID:       sentence.ID,
			JP:       sentence.JP,
			KR:       sentence.KR,
			Romaji:   sentence.Romaji,
			Level:    sentence.Level,
			Tags:     sentence.Tags,
			AudioURL: sentence.AudioURL,
		},
	}

	if detail != nil {
		words := make([]WordResponse, len(detail.Words))
		for i, w := range detail.Words {
			words[i] = WordResponse{
				Japanese: w.Japanese,
				Reading:  w.Reading,
				Meaning:  w.Meaning,
				PartOf:   w.PartOf,
			}
		}
		response.Words = words
		response.Grammar = detail.Grammar
		response.Examples = detail.Examples
	}

	pkg.SuccessResponse(c, response)
}

// GetHistory godoc
// @Summary 문장 학습 히스토리
// @Description 유저가 학습한 문장 히스토리 조회
// @Tags Sentences
// @Security BearerAuth
// @Produce json
// @Param page query int false "페이지 번호" default(1)
// @Param per_page query int false "페이지당 개수" default(20)
// @Success 200 {object} pkg.PaginatedResponse
// @Router /api/sentences/history [get]
func (h *Handler) GetHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	var query HistoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		query.Page = 1
		query.PerPage = 20
	}

	sentences, total, err := h.sentenceService.GetHistory(userID, query.Page, query.PerPage)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "히스토리를 불러오는데 실패했습니다")
		return
	}

	pkg.PaginatedSuccessResponse(c, sentences, query.Page, query.PerPage, total)
}

