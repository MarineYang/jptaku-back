package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/jptaku/server/internal/app"
	"github.com/jptaku/server/internal/config"
)

func main() {
	// Load .env file
	_ = godotenv.Load(".env")       // Docker / production
	_ = godotenv.Load("../../.env") // Local development

	// Load configuration
	cfg := config.Load()

	// Initialize application
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Start server in goroutine
	go func() {
		if err := application.Run(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
}
