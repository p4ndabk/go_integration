package main

import (
	"context"
	"fmt"
	"log"
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

// Global email service
var emailService *email.ResendService

func main() {
	cfg := config.Load()

	// Initialize email service
	emailService = email.NewResendService()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Pub/Sub client
	client, err := pubsub.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create pub/sub client: %v", err)
	}
	defer client.Close()

	// Ensure email topic and subscription exist
	emailTopic, err := client.EnsureTopic(ctx, cfg.TopicID)
	if err != nil {
		log.Fatalf("Failed to ensure email topic: %v", err)
	}

	emailSub, err := client.EnsureSubscription(ctx, cfg.SubID, emailTopic)
	if err != nil {
		log.Fatalf("Failed to ensure email subscription: %v", err)
	}

	// Ensure verification topic and subscription exist
	verificationTopic, err := client.EnsureTopic(ctx, cfg.VerificationTopicID)
	if err != nil {
		log.Fatalf("Failed to ensure verification topic: %v", err)
	}

	verificationSub, err := client.EnsureSubscription(ctx, cfg.VerificationSubID, verificationTopic)
	if err != nil {
		log.Fatalf("Failed to ensure verification subscription: %v", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Starting to receive messages from subscriptions:")
	log.Printf("  - Email: %s", cfg.SubID)
	log.Printf("  - Verification: %s", cfg.VerificationSubID)

	// Start receiving email messages
	go func() {
		err := client.Receive(ctx, emailSub, handleEmailMessage)
		if err != nil {
			log.Printf("Error receiving email messages: %v", err)
			cancel()
		}
	}()

	// Start receiving verification messages
	go func() {
		err := client.ReceiveVerification(ctx, verificationSub, handleVerificationMessage)
		if err != nil {
			log.Printf("Error receiving verification messages: %v", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down gracefully...")
	cancel()
}

// handleEmailMessage processes and sends an email message with retry logic
func handleEmailMessage(ctx context.Context, payload *models.EmailPayload) error {
	fmt.Println()
	fmt.Printf("Processando email...\n")
	fmt.Printf("Destinatário: %s\n", payload.To)
	fmt.Printf("Assunto: %s\n", payload.Subject)
	fmt.Printf("Mensagem: %s\n", payload.Body)

	// Check if it's a welcome email (you can customize this logic)
	isWelcomeEmail := strings.Contains(strings.ToLower(payload.Subject), "bem-vindo") ||
		strings.Contains(strings.ToLower(payload.Subject), "welcome")

	// Retry logic: attempt up to 3 times
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("Tentativa %d/%d - Enviando via Resend...\n", attempt, maxRetries)

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
			fmt.Printf("✅ Email enviado com sucesso na tentativa %d!\n", attempt)
			if isWelcomeEmail {
				fmt.Printf("Status: Template HTML de boas-vindas enviado via Resend\n")
				fmt.Printf("Tipo: Email de Boas-Vindas (HTML)\n")
			} else {
				fmt.Printf("Status: Email texto enviado via Resend\n")
				fmt.Printf("Tipo: Email Regular (Texto)\n")
			}
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println()
			return nil
		}

		// Failed attempt
		lastErr = err
		fmt.Printf("❌ Tentativa %d falhou: %v\n", attempt, err)

		// If this is not the last attempt, wait before retrying
		if attempt < maxRetries {
			fmt.Printf("⏳ Aguardando 2 segundos antes da próxima tentativa...\n")
			time.Sleep(2 * time.Second)
		}
	}

	// All retries failed, remove message from queue
	fmt.Printf("💀 Todas as %d tentativas falharam. Removendo mensagem da fila.\n", maxRetries)
	fmt.Printf("Último erro: %v\n", lastErr)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	// Return nil to acknowledge the message and remove it from queue
	// Even though sending failed, we don't want to keep retrying indefinitely
	return nil
}

// handleVerificationMessage simulates processing a verification email message
func handleVerificationMessage(ctx context.Context, payload *models.VerificationEmailPayload) error {
	fmt.Println()
	fmt.Printf("Email de verificação processado com sucesso!\n")
	fmt.Printf("Destinatário: %s\n", payload.To)
	fmt.Printf("Usuário: %s\n", payload.Username)
	fmt.Printf("Token: %s\n", payload.Token)
	fmt.Printf("URL de verificação: %s\n", payload.VerifyURL)
	fmt.Printf("Assunto: %s\n", payload.GenerateSubject())
	fmt.Printf("Status: Email de verificação enviado\n")
	fmt.Printf("Tipo: Email de Verificação\n")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	// Here you would integrate with actual email service
	// to send the verification email with the generated HTML

	return nil
}
