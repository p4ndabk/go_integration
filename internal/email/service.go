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
	topic *pubsub.Topic
}

// NewService creates a new email service
func NewService(topic *pubsub.Topic) *Service {
	return &Service{
		topic: topic,
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

	result := s.topic.Publish(ctx, &pubsub.Message{Data: data})
	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published email message with ID: %s", id)
	return id, nil
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
