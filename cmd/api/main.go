package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go_integration/internal/config"
	"go_integration/internal/email"
	"go_integration/internal/handlers"
	"go_integration/internal/pubsub"
	"go_integration/internal/user"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()

	// Create context with signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize Pub/Sub client
	client, err := pubsub.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to create pub/sub client: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			slog.Error("Failed to close pub/sub client", "error", closeErr)
		}
	}()

	// Ensure topics exist
	topic, err := client.EnsureTopic(ctx, cfg.EmailTopic)
	if err != nil {
		return fmt.Errorf("failed to ensure email topic (%s): %w", cfg.EmailTopic, err)
	}

	verificationTopic, err := client.EnsureTopic(ctx, cfg.VerificationTopic)
	if err != nil {
		return fmt.Errorf("failed to ensure verification topic (%s): %w", cfg.VerificationTopic, err)
	}

	userTopic, err := client.EnsureTopic(ctx, cfg.UserTopic)
	if err != nil {
		return fmt.Errorf("failed to ensure user topic (%s): %w", cfg.UserTopic, err)
	}

	// Initialize services
	emailService := email.NewServiceWithVerification(topic, verificationTopic)
	emailHandler := handlers.NewEmailHandler(emailService)

	userService := user.NewService(userTopic)
	userHandler := handlers.NewUserHandler(userService)

	// Setup HTTP router
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
	
	mux.HandleFunc("POST /send-email", emailHandler.SendEmail)
	mux.HandleFunc("POST /send-verification-email", handlers.SendVerificationEmail(emailService))
	mux.HandleFunc("POST /create-user", userHandler.CreateUser)

	// Configure HTTP server with proper timeouts
	server := &http.Server{
		Addr:         ":" + cfg.Host,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("Starting HTTP server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		slog.Info("Shutdown signal received")
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	slog.Info("Server shutdown completed")
	return nil
}
