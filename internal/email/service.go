package email

import (
	"context"
	"fmt"
	"log"

	"go_integration/internal/models"

	"cloud.google.com/go/pubsub"
)

const (
	// DefaultTopicID is the default topic name for email messages
	DefaultTopicID = "send-email"

	// DefaultSubscriptionID is the default subscription name
	DefaultSubscriptionID = "send-email-sub"
)

// Service handles email-related operations
type Service struct {
	emailTopic        *pubsub.Topic
	verificationTopic *pubsub.Topic
}

// NewService creates a new email service
func NewService(emailTopic *pubsub.Topic) *Service {
	return &Service{
		emailTopic: emailTopic,
	}
}

// NewServiceWithVerification creates a new email service with verification support
func NewServiceWithVerification(emailTopic, verificationTopic *pubsub.Topic) *Service {
	return &Service{
		emailTopic:        emailTopic,
		verificationTopic: verificationTopic,
	}
}

// SendEmail publishes an email message to the topic
func (s *Service) SendEmail(ctx context.Context, payload *models.EmailPayload) (string, error) {
	if err := payload.Validate(); err != nil {
		return "", fmt.Errorf("invalid payload: %w", err)
	}

	data, err := payload.ToJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	result := s.emailTopic.Publish(ctx, &pubsub.Message{Data: data})
	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published email message with ID: %s", id)
	return id, nil
}

// PublishVerificationEmail publishes a verification email message to the verification topic
func (s *Service) PublishVerificationEmail(ctx context.Context, payload *models.VerificationEmailPayload) error {
	if s.verificationTopic == nil {
		return fmt.Errorf("verification topic not configured")
	}

	if err := payload.Validate(); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	data, err := payload.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	result := s.verificationTopic.Publish(ctx, &pubsub.Message{Data: data})
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish verification message: %w", err)
	}

	log.Printf("Published verification email message with ID: %s", id)
	return nil
}

// MessageHandler defines the function signature for processing messages
type MessageHandler func(ctx context.Context, payload *models.EmailPayload) error

// ProcessMessage processes a received Pub/Sub message
func ProcessMessage(ctx context.Context, msg *pubsub.Message, handler MessageHandler) {
	payload, err := models.FromJSON(msg.Data)
	if err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		msg.Nack()
		return
	}

	log.Printf("Processing email: To=%s, Subject=%s", payload.To, payload.Subject)

	if err := handler(ctx, payload); err != nil {
		log.Printf("Failed to process message: %v", err)
		msg.Nack()
		return
	}

	msg.Ack()
}
