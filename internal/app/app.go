package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jptaku/server/internal/config"
	"gorm.io/gorm"
)

// App 애플리케이션 구조체
type App struct {
	cfg    *config.Config
	db     *gorm.DB
	deps   *Dependencies
	router *gin.Engine
	server *http.Server
}

// New 애플리케이션 생성
func New(cfg *config.Config) (*App, error) {
	// Database
	db, err := NewDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Dependencies (repos, services, infra)
	deps := NewDependencies(db, cfg)

	// Router
	router := NewRouter(deps, cfg)

	// HTTP Server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		cfg:    cfg,
		db:     db,
		deps:   deps,
		router: router,
		server: server,
	}, nil
}

// Run 서버 시작
func (a *App) Run() error {
	log.Printf("Server is starting on %s", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Shutdown graceful shutdown
func (a *App) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")

	// Async service 종료
	a.deps.Services.Async.Stop()

	// DB 연결 종료
	if sqlDB, err := a.db.DB(); err == nil && sqlDB != nil {
		sqlDB.Close()
	}

	// HTTP 서버 종료
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited gracefully")
	return nil
}
