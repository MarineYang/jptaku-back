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
		sentences.GET("/today", h.GetTodaySentences)
		sentences.GET("/history", h.GetHistorySentences)
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

// GetHistorySentences godoc
// @Summary 지난 학습 문장 조회
// @Description 오늘을 제외한 과거 학습 문장들을 날짜별로 조회 (페이지네이션)
// @Tags Sentences
// @Security BearerAuth
// @Produce json
// @Param page query int false "페이지 번호" default(1)
// @Param per_page query int false "페이지당 개수" default(10)
// @Success 200 {object} HistorySentencesResponse
// @Router /api/sentences/history [get]
func (h *Handler) GetHistorySentences(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.UnauthorizedResponse(c, "")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 10
	}

	result, err := h.sentenceService.GetHistorySentences(userID, page, perPage)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "지난 학습 문장을 불러오는데 실패했습니다")
		return
	}

	response := convertHistoryToResponse(result)
	pkg.SuccessResponse(c, response)
}

func convertToResponse(result *service.DailySentencesResponse) *DailySentencesResponse {
	sentences := make([]SentenceResponse, len(result.Sentences))
	for i, s := range result.Sentences {
		// Words 변환
		words := make([]WordResponse, len(s.Words))
		for j, w := range s.Words {
			words[j] = WordResponse{
				Japanese: w.Japanese,
				Reading:  w.Reading,
				Meaning:  w.Meaning,
				PartOf:   w.PartOf,
			}
		}

		// Quiz 변환
		var quiz *QuizResponse
		if s.Quiz != nil {
			quiz = &QuizResponse{}
			if s.Quiz.FillBlank != nil {
				quiz.FillBlank = &QuizFillBlankResponse{
					QuestionJP: s.Quiz.FillBlank.QuestionJP,
					Options:    s.Quiz.FillBlank.Options,
					Answer:     s.Quiz.FillBlank.Answer,
				}
			}
			if s.Quiz.Ordering != nil {
				quiz.Ordering = &QuizOrderingResponse{
					Fragments:    s.Quiz.Ordering.Fragments,
					CorrectOrder: s.Quiz.Ordering.CorrectOrder,
				}
			}
		}

		sentences[i] = SentenceResponse{
			ID:          s.ID,
			SentenceKey: s.SentenceKey,
			JP:          s.JP,
			KR:          s.KR,
			Romaji:      s.Romaji,
			Level:       s.Level,
			SubCategory: s.SubCategory,
			Words:       words,
			Grammar:     s.Grammar,
			Examples:    s.Examples,
			Quiz:        quiz,
			Memorized:   s.Memorized,
		}
	}

	return &DailySentencesResponse{
		Date:      result.Date,
		Sentences: sentences,
	}
}

func convertHistoryToResponse(result *service.HistorySentencesResponse) *HistorySentencesResponse {
	history := make([]HistoryItemResponse, len(result.History))

	for i, item := range result.History {
		sentences := make([]SentenceResponse, len(item.Sentences))
		for j, s := range item.Sentences {
			// Words 변환
			words := make([]WordResponse, len(s.Words))
			for k, w := range s.Words {
				words[k] = WordResponse{
					Japanese: w.Japanese,
					Reading:  w.Reading,
					Meaning:  w.Meaning,
					PartOf:   w.PartOf,
				}
			}

			// Quiz 변환
			var quiz *QuizResponse
			if s.Quiz != nil {
				quiz = &QuizResponse{}
				if s.Quiz.FillBlank != nil {
					quiz.FillBlank = &QuizFillBlankResponse{
						QuestionJP: s.Quiz.FillBlank.QuestionJP,
						Options:    s.Quiz.FillBlank.Options,
						Answer:     s.Quiz.FillBlank.Answer,
					}
				}
				if s.Quiz.Ordering != nil {
					quiz.Ordering = &QuizOrderingResponse{
						Fragments:    s.Quiz.Ordering.Fragments,
						CorrectOrder: s.Quiz.Ordering.CorrectOrder,
					}
				}
			}

			sentences[j] = SentenceResponse{
				ID:          s.ID,
				SentenceKey: s.SentenceKey,
				JP:          s.JP,
				KR:          s.KR,
				Romaji:      s.Romaji,
				Level:       s.Level,
				SubCategory: s.SubCategory,
				Words:       words,
				Grammar:     s.Grammar,
				Examples:    s.Examples,
				Quiz:        quiz,
				Memorized:   s.Memorized,
			}
		}

		history[i] = HistoryItemResponse{
			Date:      item.Date,
			Sentences: sentences,
		}
	}

	return &HistorySentencesResponse{
		History:    history,
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	}
}
