package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	// General application config
	ProjectID string
	Host      string

	// Email processing topic and subscription
	EmailTopic        string
	EmailSubscription string

	// Email verification topic and subscription
	VerificationTopic        string
	VerificationSubscription string

	// User creation topic and subscription
	UserTopic        string
	UserSubscription string
}

// Load loads configuration from environment variables and .env file
func Load() *Config {
	// Try to load .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		ProjectID:                getEnv("PUBSUB_PROJECT_ID", "northfi-integration"),
		Host:                     getEnv("HOST", "8080"),
		EmailTopic:               getEnv("EMAIL_TOPIC", "northfi.email.processing.v1"),
		EmailSubscription:        getEnv("EMAIL_SUBSCRIPTION", "northfi.email.processing.worker.v1"),
		VerificationTopic:        getEnv("VERIFICATION_TOPIC", "northfi.email.verification.v1"),
		VerificationSubscription: getEnv("VERIFICATION_SUBSCRIPTION", "northfi.email.verification.worker.v1"),
		UserTopic:                getEnv("USER_TOPIC", "northfi.user.creation.v1"),
		UserSubscription:         getEnv("USER_SUBSCRIPTION", "northfi.user.creation.worker.v1"),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
