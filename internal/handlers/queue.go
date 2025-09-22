// Package handlers provides message handling functionality for queue processing.
// It includes retry logic and structured logging for various email operations.
package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go_integration/internal/email"
	"go_integration/internal/models"
)

// RetryableFunc represents a function that can be retried
type RetryableFunc func() error

// Retry executes a function with exponential backoff retry logic.
// It logs structured information about each attempt and removes the message
// from the queue after all retries are exhausted to prevent infinite retries.
//
// Parameters:
//   - ctx: Context for cancellation and deadline control
//   - maxRetries: Maximum number of retry attempts
//   - delay: Fixed delay between retry attempts
//   - fn: Function to execute with retry logic
//   - logger: Structured logger for operation tracking
//
// Returns nil to acknowledge the message even on final failure.
func Retry(ctx context.Context, maxRetries int, delay time.Duration, fn RetryableFunc, logger *slog.Logger) error {
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		attemptLogger := logger.With("attempt", attempt, "max_retries", maxRetries)
		attemptLogger.Info("Attempting operation")
		
		err := fn()
		if err == nil {
			attemptLogger.Info("Operation succeeded")
			return nil
		}
		
		lastErr = err
		attemptLogger.Error("Operation attempt failed", "error", err)
		
		// If this is not the last attempt, wait before retrying
		if attempt < maxRetries {
			attemptLogger.Info("Waiting before retry", "delay_seconds", delay.Seconds())
			time.Sleep(delay)
		}
	}
	
	logger.Error("All retry attempts failed, removing from queue",
		"max_retries", maxRetries,
		"last_error", lastErr,
	)
	
	// Return nil to acknowledge the message and remove it from queue
	// Even though sending failed, we don't want to keep retrying indefinitely
	return nil
}

// HandleEmailMessage processes and sends a regular email message with HTML template and retry logic.
// It uses structured logging and the generic Retry function to handle failures gracefully.
//
// Parameters:
//   - ctx: Context for cancellation and deadline control
//   - payload: Email payload containing recipient, subject, and body
//   - emailService: Configured ResendService for email delivery
//
// Returns error if processing fails, nil if successful or after exhausting retries.
func HandleEmailMessage(ctx context.Context, payload *models.EmailPayload, emailService *email.ResendService) error {
	logger := slog.With(
		"recipient", payload.To,
		"subject", payload.Subject,
	)
	
	logger.Info("Processing regular email")
	
	return Retry(ctx, 3, 2*time.Second, func() error {
		// Send HTML email using default template
		htmlContent := email.GetDefaultEmailHTML(payload.Subject, payload.Body, "NorthFi")
		err := emailService.SendEmailWithHTML(payload.To, payload.Subject, htmlContent)
		if err != nil {
			return fmt.Errorf("failed to send email via Resend: %w", err)
		}
		
		logger.Info("Regular email sent successfully", "type", "HTML")
		return nil
	}, logger)
}

// HandleWelcomeEmailMessage processes and sends a welcome email with HTML template and retry logic.
// It generates personalized welcome content using the provided username and company template.
//
// Parameters:
//   - ctx: Context for cancellation and deadline control
//   - payload: Email payload containing recipient and subject
//   - emailService: Configured ResendService for email delivery
//   - userName: Name of the user to personalize the welcome email
//
// Returns error if processing fails, nil if successful or after exhausting retries.
func HandleWelcomeEmailMessage(ctx context.Context, payload *models.EmailPayload, emailService *email.ResendService, userName string) error {
	logger := slog.With(
		"recipient", payload.To,
		"subject", payload.Subject,
		"user_name", userName,
	)
	
	logger.Info("Processing welcome email")
	
	return Retry(ctx, 3, 2*time.Second, func() error {
		// Send HTML welcome email
		htmlContent := email.GetWelcomeEmailHTML(userName, "NorthFi")
		err := emailService.SendEmailWithHTML(payload.To, payload.Subject, htmlContent)
		if err != nil {
			return fmt.Errorf("failed to send welcome email via Resend: %w", err)
		}
		
		logger.Info("Welcome email sent successfully", "type", "HTML")
		return nil
	}, logger)
}

// HandleVerificationMessage processes and sends a verification email message.
// It creates a new ResendService instance and generates verification email content
// with the provided verification URL and user details.
//
// Parameters:
//   - ctx: Context for cancellation and deadline control
//   - payload: Verification email payload containing recipient, username, token, and verification URL
//
// Returns error if processing fails, nil if successful or after exhausting retries.
func HandleVerificationMessage(ctx context.Context, payload *models.VerificationEmailPayload) error {
	logger := slog.With(
		"recipient", payload.To,
		"username", payload.Username,
		"token", payload.Token,
	)
	
	logger.Info("Processing verification email")
	
	// Create ResendService instance for verification email
	resendService := email.NewResendService()
	
	return Retry(ctx, 3, 2*time.Second, func() error {
		// Generate HTML content for verification email
		htmlContent := email.GetVerificationEmailHTML(payload.Username, "NorthFi", payload.VerifyURL)
		
		// Send verification email with HTML template
		err := resendService.SendEmailWithHTML(payload.To, payload.GenerateSubject(), htmlContent)
		if err != nil {
			return fmt.Errorf("failed to send verification email via Resend: %w", err)
		}
		
		logger.Info("Verification email sent successfully", "type", "HTML")
		return nil
	}, logger)
}

// HandleUserMessage processes a user creation message and sends a welcome email.
// It constructs a personalized welcome email payload and delegates to HandleWelcomeEmailMessage
// for consistent processing and retry logic.
//
// Parameters:
//   - ctx: Context for cancellation and deadline control
//   - payload: User payload containing ID, email, name, and optional username
//   - emailService: Configured ResendService for email delivery
//
// Returns error if user processing or welcome email sending fails.
func HandleUserMessage(ctx context.Context, payload *models.UserPayload, emailService *email.ResendService) error {
	logger := slog.With(
		"user_id", payload.ID,
		"user_email", payload.Email,
		"user_name", payload.Name,
	)
	
	logger.Info("Processing user creation")
	
	// Create welcome email payload
	welcomeEmail := &models.EmailPayload{
		To:      payload.Email,
		Subject: "Bem-vindo(a) à NorthFi!",
		Body:    fmt.Sprintf("Olá %s,\n\nSeja bem-vindo(a) à NorthFi! Sua conta foi criada com sucesso.\n\nID do usuário: %s\nEmail: %s\n\nObrigado por se juntar a nós!\n\nEquipe NorthFi", payload.Name, payload.ID, payload.Email),
	}
	
	// Send welcome email using the specific welcome email handler
	logger.Info("Sending welcome email", "recipient", payload.Email)
	
	err := HandleWelcomeEmailMessage(ctx, welcomeEmail, emailService, payload.Name)
	if err != nil {
		logger.Error("Failed to send welcome email", "error", err)
		return fmt.Errorf("failed to send welcome email: %w", err)
	}
	
	logger.Info("User creation processed successfully")
	return nil
}