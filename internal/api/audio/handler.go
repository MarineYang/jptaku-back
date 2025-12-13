package audio

import (
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/jptaku/server/internal/pkg"
)

type Handler struct {
	s3Client *s3.Client
	s3Bucket string
}

func NewHandler(s3Client *s3.Client, s3Bucket string) *Handler {
	return &Handler{
		s3Client: s3Client,
		s3Bucket: s3Bucket,
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	audio := r.Group("/audio")
	{
		audio.GET("/:filename", h.GetAudio)
	}
}

// GetAudio godoc
// @Summary 오디오 파일 프록시
// @Description Object Storage에서 오디오 파일을 가져와 반환
// @Tags Audio
// @Produce audio/wav
// @Param filename path string true "파일명 (예: sentence_1.wav)"
// @Success 200 {file} binary
// @Failure 404 {object} pkg.Response
// @Router /api/audio/{filename} [get]
func (h *Handler) GetAudio(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		pkg.BadRequestResponse(c, "filename is required")
		return
	}

	// Object Storage에서 파일 가져오기
	result, err := h.s3Client.GetObject(c.Request.Context(), &s3.GetObjectInput{
		Bucket: aws.String(h.s3Bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		pkg.NotFoundResponse(c, "audio file not found")
		return
	}
	defer result.Body.Close()

	// 파일 데이터 읽기
	data, err := io.ReadAll(result.Body)
	if err != nil {
		pkg.InternalServerErrorResponse(c, "failed to read audio file")
		return
	}

	// Content-Type 설정 및 응답
	contentType := "audio/wav"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400") // 24시간 캐시
	c.Status(http.StatusOK)
	c.Writer.Write(data)
}

// GetAudioStream godoc (스트리밍 버전 - 대용량 파일용)
func (h *Handler) GetAudioStream(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		pkg.BadRequestResponse(c, "filename is required")
		return
	}

	result, err := h.s3Client.GetObject(c.Request.Context(), &s3.GetObjectInput{
		Bucket: aws.String(h.s3Bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		pkg.NotFoundResponse(c, "audio file not found")
		return
	}
	defer result.Body.Close()

	contentType := "audio/wav"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400")
	c.Status(http.StatusOK)

	// 스트리밍
	buf := make([]byte, 32*1024) // 32KB 버퍼
	for {
		n, err := result.Body.Read(buf)
		if n > 0 {
			c.Writer.Write(buf[:n])
			c.Writer.Flush()
		}
		if err != nil {
			break
		}
	}
}
