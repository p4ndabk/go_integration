package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
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

	// Ensure topic exists
	topic, err := client.EnsureTopic(ctx, cfg.TopicID)
	if err != nil {
		log.Fatalf("Failed to ensure topic: %v", err)
	}

	// Ensure subscription exists
	sub, err := client.EnsureSubscription(ctx, cfg.SubID, topic)
	if err != nil {
		log.Fatalf("Failed to ensure subscription: %v", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Starting to receive messages from subscription: %s", cfg.SubID)

	// Start receiving messages
	go func() {
		err := client.Receive(ctx, sub, handleEmailMessage)
		if err != nil {
			log.Printf("Error receiving messages: %v", err)
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
	fmt.Printf("📧 Email enviado para: %s\n", payload.To)
	fmt.Printf("   Assunto: %s\n", payload.Subject)
	fmt.Printf("   Mensagem: %s\n", payload.Body)
	fmt.Println()

	// Here you would integrate with actual email service
	// like SendGrid, AWS SES, etc.

	return nil
}
