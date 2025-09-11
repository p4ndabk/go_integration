package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"go_integration/internal/config"
	"go_integration/internal/email"
	"go_integration/internal/handlers"
	"go_integration/internal/pubsub"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

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

	// Ensure verification topic exists
	verificationTopic, err := client.EnsureTopic(ctx, cfg.VerificationTopicID)
	if err != nil {
		log.Fatalf("Failed to ensure verification topic: %v", err)
	}

	// Initialize services
	emailService := email.NewServiceWithVerification(topic, verificationTopic)
	emailHandler := handlers.NewEmailHandler(emailService)

	// Setup routes
	http.HandleFunc("/send-email", emailHandler.SendEmail)
	http.HandleFunc("/send-verification-email", handlers.SendVerificationEmail(emailService))

	// Start server
	addr := ":" + cfg.Host
	fmt.Printf("API rodando na porta %s\n", cfg.Host)
	log.Fatal(http.ListenAndServe(addr, nil))
}
