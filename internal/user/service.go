package user

import (
	"context"
	"fmt"
	"log"

	"go_integration/internal/models"

	"cloud.google.com/go/pubsub"
)

// Service handles user-related operations
type Service struct {
	userTopic *pubsub.Topic
}

// NewService creates a new user service
func NewService(userTopic *pubsub.Topic) *Service {
	return &Service{
		userTopic: userTopic,
	}
}

// CreateUser publishes a user creation message to the topic
func (s *Service) CreateUser(ctx context.Context, payload *models.UserPayload) (string, error) {
	if err := payload.Validate(); err != nil {
		return "", fmt.Errorf("invalid payload: %w", err)
	}

	data, err := payload.ToJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	result := s.userTopic.Publish(ctx, &pubsub.Message{Data: data})
	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published user creation message with ID: %s", id)
	return id, nil
}

// MessageHandler defines the function signature for processing user messages
type MessageHandler func(ctx context.Context, payload *models.UserPayload) error

// ProcessMessage processes a received Pub/Sub message for user creation
func ProcessMessage(ctx context.Context, msg *pubsub.Message, handler MessageHandler) {
	payload, err := models.UserFromJSON(msg.Data)
	if err != nil {
		log.Printf("Failed to unmarshal user message: %v", err)
		msg.Nack()
		return
	}

	log.Printf("Processing user creation: ID=%s, Email=%s, Name=%s", payload.ID, payload.Email, payload.Name)

	if err := handler(ctx, payload); err != nil {
		log.Printf("Failed to process user message: %v", err)
		msg.Nack()
		return
	}

	msg.Ack()
}
