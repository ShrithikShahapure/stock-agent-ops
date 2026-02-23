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

	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	httpserver "github.com/shrithkshahapure/stock-agent-ops/internal/http"
	"github.com/shrithkshahapure/stock-agent-ops/internal/metrics"
	redisclient "github.com/shrithkshahapure/stock-agent-ops/internal/services/redis"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Stock Agent Ops API Server...")

	// Load configuration
	cfg := config.Load()

	// Initialize metrics
	registry := prometheus.NewRegistry()
	m := metrics.New(registry)

	// Connect to Redis (with retry)
	redis, err := redisclient.New(cfg, m)
	if err != nil {
		log.Printf("Warning: Redis unavailable: %v", err)
		log.Println("Server will start but caching and rate limiting will be disabled")
	}

	// Create HTTP server
	server := httpserver.NewServer(cfg, redis, m)

	// Create HTTP server with timeouts
	addr := fmt.Sprintf(":%s", cfg.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Server listening on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.GracefulTimeout)*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close Redis connection if available
	if redis != nil {
		if err := redis.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}

	log.Println("Server stopped")
}
