package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"go_integration/internal/config"
	"go_integration/internal/email"
	"go_integration/internal/handlers"
	"go_integration/internal/models"
	"go_integration/internal/pubsub"
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

	// Initialize email service and handlers
	emailService := email.NewResendService()
	emailHandler := handlers.NewEmailQueueHandler(emailService)

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

	// Ensure topics and subscriptions exist
	emailTopic, err := client.EnsureTopic(ctx, cfg.EmailTopic)
	if err != nil {
		return fmt.Errorf("failed to ensure email topic (%s): %w", cfg.EmailTopic, err)
	}

	emailSub, err := client.EnsureSubscription(ctx, cfg.EmailSubscription, emailTopic)
	if err != nil {
		return fmt.Errorf("failed to ensure email subscription (%s): %w", cfg.EmailSubscription, err)
	}

	verificationTopic, err := client.EnsureTopic(ctx, cfg.VerificationTopic)
	if err != nil {
		return fmt.Errorf("failed to ensure verification topic (%s): %w", cfg.VerificationTopic, err)
	}

	verificationSub, err := client.EnsureSubscription(ctx, cfg.VerificationSubscription, verificationTopic)
	if err != nil {
		return fmt.Errorf("failed to ensure verification subscription (%s): %w", cfg.VerificationSubscription, err)
	}

	userTopic, err := client.EnsureTopic(ctx, cfg.UserTopic)
	if err != nil {
		return fmt.Errorf("failed to ensure user topic (%s): %w", cfg.UserTopic, err)
	}

	userSub, err := client.EnsureSubscription(ctx, cfg.UserSubscription, userTopic)
	if err != nil {
		return fmt.Errorf("failed to ensure user subscription (%s): %w", cfg.UserSubscription, err)
	}

	slog.Info("Starting message processing",
		"email_topic", cfg.EmailTopic,
		"email_subscription", cfg.EmailSubscription,
		"verification_topic", cfg.VerificationTopic,
		"verification_subscription", cfg.VerificationSubscription,
		"user_topic", cfg.UserTopic,
		"user_subscription", cfg.UserSubscription,
	)

	// Error channel for goroutine errors
	errChan := make(chan error, 3)

	// Start receiving email messages
	go func() {
		if err := client.Receive(ctx, emailSub, func(ctx context.Context, payload *models.EmailPayload) error {
			return emailHandler.HandleEmailMessage(ctx, payload)
		}); err != nil {
			errChan <- fmt.Errorf("email message receiver failed: %w", err)
		}
	}()

	// Start receiving verification messages
	go func() {
		if err := client.ReceiveVerification(ctx, verificationSub, emailHandler.HandleVerificationMessage); err != nil {
			errChan <- fmt.Errorf("verification message receiver failed: %w", err)
		}
	}()

	// Start receiving user creation messages
	go func() {
		if err := client.ReceiveUser(ctx, userSub, func(ctx context.Context, payload *models.UserPayload) error {
			return emailHandler.HandleUserMessage(ctx, payload)
		}); err != nil {
			errChan <- fmt.Errorf("user message receiver failed: %w", err)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		slog.Info("Shutdown signal received")
	}

	slog.Info("Worker shutdown completed")
	return nil
}
