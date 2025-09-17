package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go_integration/internal/config"
	"go_integration/internal/models"
	"go_integration/internal/pubsub"
)

func main() {
	cfg := config.Load()

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

// handleEmailMessage simulates processing an email message
func handleEmailMessage(ctx context.Context, payload *models.EmailPayload) error {
	fmt.Println()
	fmt.Printf("Email processado com sucesso!\n")
	fmt.Printf("Destinatário: %s\n", payload.To)
	fmt.Printf("Assunto: %s\n", payload.Subject)
	fmt.Printf("Mensagem: %s\n", payload.Body)
	fmt.Printf("Status: Enviado via sistema Pub/Sub\n")
	fmt.Printf("Tipo: Email Regular\n")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	// Here you would integrate with actual email service
	// like SendGrid, AWS SES, etc.

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
