package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"go_integration/internal/models"

	"cloud.google.com/go/pubsub"
)

// Client wraps Google Cloud Pub/Sub client
type Client struct {
	client    *pubsub.Client
	projectID string
}

// NewClient creates a new Pub/Sub client
func NewClient(ctx context.Context, projectID string) (*Client, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	return &Client{
		client:    client,
		projectID: projectID,
	}, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.client.Close()
}

// EnsureTopic creates a topic if it doesn't exist
func (c *Client) EnsureTopic(ctx context.Context, topicID string) (*pubsub.Topic, error) {
	topic := c.client.Topic(topicID)

	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if topic exists: %w", err)
	}

	if !exists {
		topic, err = c.client.CreateTopic(ctx, topicID)
		if err != nil {
			return nil, fmt.Errorf("failed to create topic: %w", err)
		}
		log.Printf("Created topic: %s", topicID)
	}

	return topic, nil
}

// EnsureSubscription creates a subscription if it doesn't exist
func (c *Client) EnsureSubscription(ctx context.Context, subID string, topic *pubsub.Topic) (*pubsub.Subscription, error) {
	sub := c.client.Subscription(subID)

	exists, err := sub.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if subscription exists: %w", err)
	}

	if !exists {
		sub, err = c.client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
			Topic: topic,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create subscription: %w", err)
		}
		log.Printf("Created subscription: %s", subID)
	}

	return sub, nil
}

// Receive wraps the subscription Receive method with a handler function
func (c *Client) Receive(ctx context.Context, sub *pubsub.Subscription, handler func(context.Context, *models.EmailPayload) error) error {
	return sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var payload models.EmailPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			msg.Nack()
			return
		}

		if err := handler(ctx, &payload); err != nil {
			log.Printf("Failed to handle message: %v", err)
			msg.Nack()
			return
		}

		msg.Ack()
	})
}
