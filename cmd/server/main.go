package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thediligencedev/betteridn/internal/config"
	"github.com/thediligencedev/betteridn/internal/db"
	"github.com/thediligencedev/betteridn/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	pool, err := db.Init(cfg.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Create and start server
	srv := server.New(pool, cfg)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited properly")

	// TODO: implement cleanup for other goroutine functions if needed
}
