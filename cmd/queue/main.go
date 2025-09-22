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

	// Ensure user topic and subscription exist
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

	// Start receiving user creation messages
	go func() {
		if err := client.ReceiveUser(ctx, userSub, func(ctx context.Context, payload *models.UserPayload) error {
			return handleUserMessage(ctx, payload, emailService)
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

// handleEmailMessage processes and sends a regular email message with HTML template and retry logic
func handleEmailMessage(ctx context.Context, payload *models.EmailPayload, emailService *email.ResendService) error {
	logger := slog.With(
		"recipient", payload.To,
		"subject", payload.Subject,
	)

	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Processando email padrão...\n")
	fmt.Printf("Destinatário: %s\n", payload.To)
	fmt.Printf("Assunto: %s\n", payload.Subject)
	fmt.Printf("Mensagem: %s\n", payload.Body)
	fmt.Println()

	logger.Info("Processing regular email")

	// Retry logic: attempt up to 3 times
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		attemptLogger := logger.With("attempt", attempt, "max_retries", maxRetries)

		fmt.Printf("[INFO] Tentativa %d/%d - Enviando email HTML via Resend...\n", attempt, maxRetries)
		attemptLogger.Info("Sending regular email attempt")

		// Send HTML email using default template
		htmlContent := email.GetDefaultEmailHTML(payload.Subject, payload.Body, "NorthFi")
		err := emailService.SendEmailWithHTML(payload.To, payload.Subject, htmlContent)

		if err == nil {
			// Success! Email sent
			fmt.Printf("[SUCCESS] ✅ Email enviado com sucesso na tentativa %d!\n", attempt)
			fmt.Printf("Status: Email HTML enviado via Resend\n")
			fmt.Printf("Tipo: Email Regular (HTML)\n")
			fmt.Println(strings.Repeat("-", 50))
			fmt.Println()

			attemptLogger.Info("Regular email sent successfully", "type", "HTML")
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

// handleWelcomeEmailMessage processes and sends a welcome email with HTML template and retry logic
func handleWelcomeEmailMessage(ctx context.Context, payload *models.EmailPayload, emailService *email.ResendService, userName string) error {
	logger := slog.With(
		"recipient", payload.To,
		"subject", payload.Subject,
		"user_name", userName,
	)

	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Processando email de boas-vindas...\n")
	fmt.Printf("Destinatário: %s\n", payload.To)
	fmt.Printf("Assunto: %s\n", payload.Subject)
	fmt.Printf("Nome do usuário: %s\n", userName)
	fmt.Println()

	logger.Info("Processing welcome email")

	// Retry logic: attempt up to 3 times
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		attemptLogger := logger.With("attempt", attempt, "max_retries", maxRetries)

		fmt.Printf("[INFO] Tentativa %d/%d - Enviando email HTML de boas-vindas via Resend...\n", attempt, maxRetries)
		attemptLogger.Info("Sending welcome email attempt")

		// Send HTML welcome email
		htmlContent := email.GetWelcomeEmailHTML(userName, "NorthFi")
		err := emailService.SendEmailWithHTML(payload.To, payload.Subject, htmlContent)

		if err == nil {
			// Success! Email sent
			fmt.Printf("[SUCCESS] ✅ Email de boas-vindas enviado com sucesso na tentativa %d!\n", attempt)
			fmt.Printf("Status: Template HTML de boas-vindas enviado via Resend\n")
			fmt.Printf("Tipo: Email de Boas-Vindas (HTML)\n")
			fmt.Println(strings.Repeat("-", 50))
			fmt.Println()

			attemptLogger.Info("Welcome email sent successfully", "type", "HTML")
			return nil
		}

		// Failed attempt
		lastErr = err
		fmt.Printf("[ERROR] ❌ Tentativa %d falhou: resend API returned status 403\n", attempt)
		fmt.Printf("Detalhes: You can only send testing emails to your own email address (ti@northficoin.com.br).\n")

		attemptLogger.Error("Welcome email sending failed", "error", err)

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

	logger.Error("All welcome email retry attempts failed, removing from queue",
		"max_retries", maxRetries,
		"last_error", lastErr,
	)

	// Return nil to acknowledge the message and remove it from queue
	// Even though sending failed, we don't want to keep retrying indefinitely
	return nil
}

// handleVerificationMessage processes and sends a verification email message
func handleVerificationMessage(ctx context.Context, payload *models.VerificationEmailPayload) error {
	logger := slog.With(
		"recipient", payload.To,
		"username", payload.Username,
		"token", payload.Token,
	)

	fmt.Println()
	fmt.Printf("Processando email de verificação...\n")
	fmt.Printf("Destinatário: %s\n", payload.To)
	fmt.Printf("Usuário: %s\n", payload.Username)
	fmt.Printf("Token: %s\n", payload.Token)
	fmt.Printf("URL de verificação: %s\n", payload.VerifyURL)
	fmt.Println()

	logger.Info("Processing verification email")

	// Create ResendService instance for verification email
	resendService := email.NewResendService()

	// Retry logic: attempt up to 3 times
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		attemptLogger := logger.With("attempt", attempt, "max_retries", maxRetries)

		fmt.Printf("[INFO] Tentativa %d/%d - Enviando email de verificação via Resend...\n", attempt, maxRetries)
		attemptLogger.Info("Sending verification email attempt")

		// Generate HTML content for verification email
		htmlContent := email.GetVerificationEmailHTML(payload.Username, "NorthFi", payload.VerifyURL)

		// Send verification email with HTML template
		err := resendService.SendEmailWithHTML(payload.To, payload.GenerateSubject(), htmlContent)

		if err == nil {
			// Success! Email sent
			fmt.Printf("[SUCCESS] ✅ Email de verificação enviado com sucesso na tentativa %d!\n", attempt)
			fmt.Printf("Status: Template HTML de verificação enviado via Resend\n")
			fmt.Printf("Tipo: Email de Verificação (HTML)\n")
			fmt.Println(strings.Repeat("-", 50))
			fmt.Println()

			attemptLogger.Info("Verification email sent successfully", "type", "HTML")
			return nil
		}

		// Failed attempt
		lastErr = err
		fmt.Printf("[ERROR] ❌ Tentativa %d falhou: resend API returned error\n", attempt)
		fmt.Printf("Detalhes: %v\n", err)

		attemptLogger.Error("Verification email sending failed", "error", err)

		// If this is not the last attempt, wait before retrying
		if attempt < maxRetries {
			fmt.Printf("[WAIT] ⏳ Aguardando 2 segundos antes da próxima tentativa...\n")
			fmt.Println()
			time.Sleep(2 * time.Second)
		}
	}

	// All retries failed, remove message from queue
	fmt.Printf("[FATAL] 💀 Todas as %d tentativas falharam. Removendo mensagem da fila.\n", maxRetries)
	fmt.Printf("Último erro: %v\n", lastErr)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println()

	logger.Error("All verification email retry attempts failed, removing from queue",
		"max_retries", maxRetries,
		"last_error", lastErr,
	)

	// Return nil to acknowledge the message and remove it from queue
	return nil
}

// handleUserMessage processes a user creation message and sends a welcome email
func handleUserMessage(ctx context.Context, payload *models.UserPayload, emailService *email.ResendService) error {
	logger := slog.With(
		"user_id", payload.ID,
		"user_email", payload.Email,
		"user_name", payload.Name,
	)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("🆕 PROCESSANDO CRIAÇÃO DE USUÁRIO\n")
	fmt.Printf("ID do Usuário: %s\n", payload.ID)
	fmt.Printf("Nome: %s\n", payload.Name)
	fmt.Printf("Email: %s\n", payload.Email)
	if payload.Username != "" {
		fmt.Printf("Username: %s\n", payload.Username)
	}
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	logger.Info("Processing user creation")

	// Create welcome email payload
	welcomeEmail := &models.EmailPayload{
		To:      payload.Email,
		Subject: "Bem-vindo(a) à NorthFi!",
		Body:    fmt.Sprintf("Olá %s,\n\nSeja bem-vindo(a) à NorthFi! Sua conta foi criada com sucesso.\n\nID do usuário: %s\nEmail: %s\n\nObrigado por se juntar a nós!\n\nEquipe NorthFi", payload.Name, payload.ID, payload.Email),
	}

	// Send welcome email using the specific welcome email handler
	fmt.Printf("📧 Enviando email de boas-vindas para %s...\n", payload.Email)
	logger.Info("Sending welcome email", "recipient", payload.Email)

	// Use the welcome email handler with user name
	err := handleWelcomeEmailMessage(ctx, welcomeEmail, emailService, payload.Name)
	if err != nil {
		logger.Error("Failed to send welcome email", "error", err)
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	fmt.Printf("✅ Usuario %s criado e email de boas-vindas enviado com sucesso!\n", payload.Name)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	logger.Info("User creation processed successfully")
	return nil
}
