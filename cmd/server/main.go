package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ghstouch/one-go/internal/config"
	"github.com/ghstouch/one-go/internal/database"
	"github.com/ghstouch/one-go/internal/router"
	"github.com/ghstouch/one-go/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting One Go...")
	logger.Infof("Server configuration: %s:%s", cfg.Server.Host, cfg.Server.Port)

	// Initialize database
	if err := database.Init(&cfg.Database); err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.AutoMigrate(); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed initial data (password hashing is handled inside Seed)
	if err := database.Seed(cfg); err != nil {
		logger.Fatalf("Failed to seed database: %v", err)
	}

	// Setup router
	r := router.Setup(cfg)

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Infof("Server listening on http://%s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited gracefully")
}
