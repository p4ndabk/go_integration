package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go_integration/internal/email"
	"go_integration/internal/models"
)

// EmailQueueHandler handles email queue message processing
type EmailQueueHandler struct {
	emailService *email.ResendService
}

// NewEmailQueueHandler creates a new email queue handler
func NewEmailQueueHandler(emailService *email.ResendService) *EmailQueueHandler {
	return &EmailQueueHandler{
		emailService: emailService,
	}
}

// retry executes a function with retry logic using structured logging
func (h *EmailQueueHandler) retry(ctx context.Context, maxRetries int, delay time.Duration, fn func() error, logger *slog.Logger, operation string) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		attemptLogger := logger.With(
			"attempt", attempt,
			"max_retries", maxRetries,
			"operation", operation,
		)

		attemptLogger.Info("Starting attempt")

		err := fn()
		if err == nil {
			attemptLogger.Info("Operation completed successfully")
			return nil
		}

		lastErr = err
		attemptLogger.Error("Operation failed", "error", err)

		// If this is not the last attempt, wait before retrying
		if attempt < maxRetries {
			attemptLogger.Info("Waiting before retry", "delay", delay)
			time.Sleep(delay)
		}
	}

	logger.Error("All retry attempts failed",
		"operation", operation,
		"max_retries", maxRetries,
		"last_error", lastErr,
	)

	// Return nil to acknowledge the message and remove it from queue
	// Even though sending failed, we don't want to keep retrying indefinitely
	return nil
}

// HandleEmailMessage processes and sends a regular email message with retry logic
func (h *EmailQueueHandler) HandleEmailMessage(ctx context.Context, payload *models.EmailPayload) error {
	logger := slog.With(
		"recipient", payload.To,
		"subject", payload.Subject,
		"type", "regular_email",
	)

	logger.Info("Processing regular email message")

	return h.retry(ctx, 3, 2*time.Second, func() error {
		htmlContent := email.GetDefaultEmailHTML(payload.Subject, payload.Body, "NorthFi")
		return h.emailService.SendEmailWithHTML(payload.To, payload.Subject, htmlContent)
	}, logger, "send_regular_email")
}

// HandleWelcomeMessage processes and sends a welcome email with retry logic
func (h *EmailQueueHandler) HandleWelcomeMessage(ctx context.Context, payload *models.EmailPayload, userName string) error {
	logger := slog.With(
		"recipient", payload.To,
		"subject", payload.Subject,
		"user_name", userName,
		"type", "welcome_email",
	)

	logger.Info("Processing welcome email message")

	return h.retry(ctx, 3, 2*time.Second, func() error {
		htmlContent := email.GetWelcomeEmailHTML(userName, "NorthFi")
		return h.emailService.SendEmailWithHTML(payload.To, payload.Subject, htmlContent)
	}, logger, "send_welcome_email")
}

// HandleVerificationMessage processes and sends a verification email message with retry logic
func (h *EmailQueueHandler) HandleVerificationMessage(ctx context.Context, payload *models.VerificationEmailPayload) error {
	logger := slog.With(
		"recipient", payload.To,
		"username", payload.Username,
		"token", payload.Token,
		"type", "verification_email",
	)

	logger.Info("Processing verification email message")

	return h.retry(ctx, 3, 2*time.Second, func() error {
		htmlContent := email.GetVerificationEmailHTML(payload.Username, "NorthFi", payload.VerifyURL)
		return h.emailService.SendEmailWithHTML(payload.To, payload.GenerateSubject(), htmlContent)
	}, logger, "send_verification_email")
}

// HandleUserMessage processes a user creation message and sends a welcome email
func (h *EmailQueueHandler) HandleUserMessage(ctx context.Context, payload *models.UserPayload) error {
	logger := slog.With(
		"user_id", payload.ID,
		"user_email", payload.Email,
		"user_name", payload.Name,
		"type", "user_creation",
	)

	logger.Info("Processing user creation message")

	// Create welcome email payload
	welcomeEmail := &models.EmailPayload{
		To:      payload.Email,
		Subject: "Bem-vindo(a) à NorthFi!",
		Body:    fmt.Sprintf("Olá %s, seja bem-vindo(a) à NorthFi! Sua conta foi criada com sucesso.", payload.Name),
	}

	logger.Info("Sending welcome email for new user", "recipient", payload.Email)

	// Send welcome email using the welcome email handler
	err := h.HandleWelcomeMessage(ctx, welcomeEmail, payload.Name)
	if err != nil {
		logger.Error("Failed to send welcome email", "error", err)
		return fmt.Errorf("failed to send welcome email for user %s: %w", payload.ID, err)
	}

	logger.Info("User creation processed successfully")
	return nil
}
