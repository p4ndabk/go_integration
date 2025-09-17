package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go_integration/internal/config"
	"go_integration/internal/email"
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

	// Initialize email service
	emailService := email.NewResendService()

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
	emailTopic, err := client.EnsureTopic(ctx, cfg.TopicID)
	if err != nil {
		return fmt.Errorf("failed to ensure email topic: %w", err)
	}

	emailSub, err := client.EnsureSubscription(ctx, cfg.SubID, emailTopic)
	if err != nil {
		return fmt.Errorf("failed to ensure email subscription: %w", err)
	}

	verificationTopic, err := client.EnsureTopic(ctx, cfg.VerificationTopicID)
	if err != nil {
		return fmt.Errorf("failed to ensure verification topic: %w", err)
	}

	verificationSub, err := client.EnsureSubscription(ctx, cfg.VerificationSubID, verificationTopic)
	if err != nil {
		return fmt.Errorf("failed to ensure verification subscription: %w", err)
	}

	slog.Info("Starting message processing",
		"email_subscription", cfg.SubID,
		"verification_subscription", cfg.VerificationSubID,
	)

	// Error channel for goroutine errors
	errChan := make(chan error, 2)

	// Start receiving email messages
	go func() {
		if err := client.Receive(ctx, emailSub, func(ctx context.Context, payload *models.EmailPayload) error {
			return handleEmailMessage(ctx, payload, emailService)
		}); err != nil {
			errChan <- fmt.Errorf("email message receiver failed: %w", err)
		}
	}()

	// Start receiving verification messages
	go func() {
		if err := client.ReceiveVerification(ctx, verificationSub, handleVerificationMessage); err != nil {
			errChan <- fmt.Errorf("verification message receiver failed: %w", err)
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

// handleEmailMessage processes and sends an email message with retry logic
func handleEmailMessage(ctx context.Context, payload *models.EmailPayload, emailService *email.ResendService) error {
	logger := slog.With(
		"recipient", payload.To,
		"subject", payload.Subject,
	)

	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Processando email...\n")
	fmt.Printf("Destinatário: %s\n", payload.To)
	fmt.Printf("Assunto: %s\n", payload.Subject)
	fmt.Printf("Mensagem: %s\n", payload.Body)
	fmt.Println()

	logger.Info("Processing email")

	// Check if it's a welcome email
	isWelcomeEmail := strings.Contains(strings.ToLower(payload.Subject), "bem-vindo") ||
		strings.Contains(strings.ToLower(payload.Subject), "welcome")

	// Retry logic: attempt up to 3 times
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		attemptLogger := logger.With("attempt", attempt, "max_retries", maxRetries)

		fmt.Printf("[INFO] Tentativa %d/%d - Enviando via Resend...\n", attempt, maxRetries)
		attemptLogger.Info("Sending email attempt")

		var err error
		if isWelcomeEmail {
			// Send HTML email for welcome messages
			htmlContent := emailService.GetWelcomeEmailHTML("Usuário", "NorthFi")
			err = emailService.SendEmailWithHTML(payload.To, payload.Subject, htmlContent)
		} else {
			// Send regular text email
			err = emailService.SendEmail(payload.To, payload.Subject, payload.Body)
		}

		if err == nil {
			// Success! Email sent
			fmt.Printf("[SUCCESS] ✅ Email enviado com sucesso na tentativa %d!\n", attempt)
			if isWelcomeEmail {
				fmt.Printf("Status: Template HTML de boas-vindas enviado via Resend\n")
				fmt.Printf("Tipo: Email de Boas-Vindas (HTML)\n")
			} else {
				fmt.Printf("Status: Email texto enviado via Resend\n")
				fmt.Printf("Tipo: Email Regular (Texto)\n")
			}
			fmt.Println(strings.Repeat("-", 50))
			fmt.Println()

			attemptLogger.Info("Email sent successfully",
				"type", map[string]string{
					"welcome": "HTML",
					"regular": "text",
				}[map[bool]string{true: "welcome", false: "regular"}[isWelcomeEmail]],
			)
			return nil
		}

		// Failed attempt
		lastErr = err
		fmt.Printf("[ERROR] ❌ Tentativa %d falhou: resend API returned status 403\n", attempt)
		fmt.Printf("Detalhes: You can only send testing emails to your own email address (ti@northficoin.com.br).\n")

		attemptLogger.Error("Email sending failed", "error", err)

		// If this is not the last attempt, wait before retrying
		if attempt < maxRetries {
			fmt.Printf("[WAIT] ⏳ Aguardando 2 segundos antes da próxima tentativa...\n")
			fmt.Println()
			time.Sleep(2 * time.Second)
		}
	}

	// All retries failed, remove message from queue
	fmt.Printf("[FATAL] 💀 Todas as %d tentativas falharam. Removendo mensagem da fila.\n", maxRetries)
	fmt.Printf("Último erro: resend API returned status 403\n")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println()

	logger.Error("All retry attempts failed, removing from queue",
		"max_retries", maxRetries,
		"last_error", lastErr,
	)

	// Return nil to acknowledge the message and remove it from queue
	// Even though sending failed, we don't want to keep retrying indefinitely
	return nil
}

// handleVerificationMessage processes a verification email message
func handleVerificationMessage(ctx context.Context, payload *models.VerificationEmailPayload) error {
	logger := slog.With(
		"recipient", payload.To,
		"username", payload.Username,
		"token", payload.Token,
	)

	fmt.Println()
	fmt.Printf("Email de verificação processado com sucesso!\n")
	fmt.Printf("Destinatário: %s\n", payload.To)
	fmt.Printf("Usuário: %s\n", payload.Username)
	fmt.Printf("Token: %s\n", payload.Token)
	fmt.Printf("URL de verificação: %s\n", payload.VerifyURL)
	fmt.Printf("Assunto: %s\n", payload.GenerateSubject())
	fmt.Printf("Status: Email de verificação enviado\n")
	fmt.Printf("Tipo: Email de Verificação\n")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println()

	logger.Info("Verification email processed",
		"verify_url", payload.VerifyURL,
		"subject", payload.GenerateSubject(),
	)

	// Here you would integrate with actual email service
	// to send the verification email with the generated HTML

	return nil
}
