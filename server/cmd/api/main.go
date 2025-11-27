package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jptaku/server/internal/api/auth"
	"github.com/jptaku/server/internal/api/chat"
	"github.com/jptaku/server/internal/api/feedback"
	"github.com/jptaku/server/internal/api/learning"
	"github.com/jptaku/server/internal/api/sentences"
	"github.com/jptaku/server/internal/api/user"
	"github.com/jptaku/server/internal/config"
	"github.com/jptaku/server/internal/middleware"
	"github.com/jptaku/server/internal/model"
	"github.com/jptaku/server/internal/pkg"
	"github.com/jptaku/server/internal/repository"
	"github.com/jptaku/server/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize database
	db, err := config.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate models
	if err := db.AutoMigrate(
		&model.User{},
		&model.UserSettings{},
		&model.UserOnboarding{},
		&model.Sentence{},
		&model.SentenceDetail{},
		&model.DailySentenceSet{},
		&model.LearningProgress{},
		&model.ChatSession{},
		&model.ChatMessage{},
		&model.Feedback{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")

	// Initialize JWT Manager
	jwtManager := pkg.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpirationHours)

	// Initialize DB Manager
	dbManager := repository.NewDBManager(db)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	sentenceRepo := repository.NewSentenceRepository(db)
	learningRepo := repository.NewLearningRepository(db)
	chatRepo := repository.NewChatRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(db)

	// Initialize async service (워커 4개, 큐 사이즈 100)
	asyncService := service.NewAsyncService(4, 100)

	// Initialize services
	authService := service.NewAuthService(dbManager, userRepo, jwtManager)
	userService := service.NewUserService(userRepo)
	sentenceService := service.NewSentenceService(dbManager, sentenceRepo, userRepo)
	learningService := service.NewLearningService(learningRepo, sentenceRepo)
	chatService := service.NewChatService(chatRepo, sentenceRepo)
	feedbackService := service.NewFeedbackService(feedbackRepo, chatRepo)

	// asyncService는 나중에 다른 서비스에서 사용할 수 있도록 보관
	_ = asyncService

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService, jwtManager)
	sentencesHandler := sentences.NewHandler(sentenceService)
	learningHandler := learning.NewHandler(learningService)
	chatHandler := chat.NewHandler(chatService)
	feedbackHandler := feedback.NewHandler(feedbackService)

	// Initialize Gin router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "jptaku API server is running",
		})
	})

	// API routes
	api := r.Group("/api")
	{
		// Auth routes (no auth middleware)
		authHandler.RegisterRoutes(api)

		// Auth middleware for protected routes
		authMiddleware := middleware.AuthMiddleware(jwtManager)

		// Protected routes
		userHandler.RegisterRoutes(api, authMiddleware)
		sentencesHandler.RegisterRoutes(api, authMiddleware)
		learningHandler.RegisterRoutes(api, authMiddleware)
		chatHandler.RegisterRoutes(api, authMiddleware)
		feedbackHandler.RegisterRoutes(api, authMiddleware)
	}

	// Create HTTP server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown 처리
	go func() {
		log.Printf("Server is starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 종료 시그널 대기 (Ctrl+C, kill 등)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown (최대 30초 대기)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 비동기 서비스 종료
	asyncService.Stop()

	// DB 연결 종료
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	// HTTP 서버 종료
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
