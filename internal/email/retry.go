package email

import (
	"context"
	"log/slog"
	"strings"
	"time"
)

// IsWelcomeSubject checks if an email subject indicates a welcome email
func IsWelcomeSubject(subject string) bool {
	subjectLower := strings.ToLower(subject)
	return strings.Contains(subjectLower, "bem-vindo") || strings.Contains(subjectLower, "welcome")
}

// RetryConfig defines retry parameters
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
}

// DefaultRetryConfig returns standard retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Delay:       2 * time.Second,
	}
}

// ExecuteWithRetry executes a function with retry logic
func ExecuteWithRetry(ctx context.Context, config RetryConfig, fn func() error, logger *slog.Logger) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		attemptLogger := logger.With("attempt", attempt, "max_attempts", config.MaxAttempts)
		attemptLogger.Info("Executing attempt")

		err := fn()
		if err == nil {
			attemptLogger.Info("Operation successful")
			return nil
		}

		lastErr = err
		attemptLogger.Error("Operation failed", "error", err)

		// If this is not the last attempt, wait before retrying
		if attempt < config.MaxAttempts {
			attemptLogger.Info("Waiting before retry", "delay", config.Delay)
			time.Sleep(config.Delay)
		}
	}

	logger.Error("All retry attempts failed", "max_attempts", config.MaxAttempts, "last_error", lastErr)
	// Return nil to acknowledge the message and remove it from queue
	// Even though sending failed, we don't want to keep retrying indefinitely
	return nil
}
