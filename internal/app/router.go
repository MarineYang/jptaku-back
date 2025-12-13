package app

import (
	"github.com/gin-gonic/gin"
	_ "github.com/jptaku/server/docs" // Swagger docs
	"github.com/jptaku/server/internal/api/audio"
	"github.com/jptaku/server/internal/api/auth"
	"github.com/jptaku/server/internal/api/chat"
	"github.com/jptaku/server/internal/api/feedback"
	"github.com/jptaku/server/internal/api/learning"
	"github.com/jptaku/server/internal/api/sentences"
	"github.com/jptaku/server/internal/api/user"
	"github.com/jptaku/server/internal/config"
	"github.com/jptaku/server/internal/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewRouter Gin 라우터 초기화
func NewRouter(deps *Dependencies, cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", healthCheck)

	// Initialize handlers
	authHandler := auth.NewHandler(deps.Services.Auth)
	userHandler := user.NewHandler(deps.Services.User, deps.Infra.JWTManager)
	sentencesHandler := sentences.NewHandler(deps.Services.Sentence)
	learningHandler := learning.NewHandler(deps.Services.Learning)
	chatHandler := chat.NewHandler(deps.Services.Chat)
	feedbackHandler := feedback.NewHandler(deps.Services.Feedback)
	audioHandler := audio.NewHandler(deps.Infra.S3Client, deps.Infra.BucketName)

	// API routes
	api := r.Group("/api")
	{
		// Public routes
		authHandler.RegisterRoutes(api)
		audioHandler.RegisterRoutes(api)

		// Protected routes
		authMiddleware := middleware.AuthMiddleware(deps.Infra.JWTManager)
		userHandler.RegisterRoutes(api, authMiddleware)
		sentencesHandler.RegisterRoutes(api, authMiddleware)
		learningHandler.RegisterRoutes(api, authMiddleware)
		chatHandler.RegisterRoutes(api, authMiddleware)
		feedbackHandler.RegisterRoutes(api, authMiddleware)
	}

	return r
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "jptaku API server is running",
	})
}
