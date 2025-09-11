package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	ProjectID           string
	Host                string
	TopicID             string
	SubID               string
	VerificationTopicID string
	VerificationSubID   string
}

// Load loads configuration from environment variables and .env file
func Load() *Config {
	// Try to load .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		ProjectID:           getEnv("PUBSUB_PROJECT_ID", "test-project"),
		Host:                getEnv("HOST", "8080"),
		TopicID:             getEnv("TOPIC_ID", "send-email"),
		SubID:               getEnv("SUBSCRIPTION_ID", "send-email-sub"),
		VerificationTopicID: getEnv("VERIFICATION_TOPIC_ID", "email-verification"),
		VerificationSubID:   getEnv("VERIFICATION_SUB_ID", "email-verification-sub"),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
