package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"go_integration/internal/email"
	"go_integration/internal/models"
)

// SendVerificationEmail handles sending verification emails
func SendVerificationEmail(emailService *email.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload models.VerificationEmailPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := payload.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Publish verification email to pub/sub
		if err := emailService.PublishVerificationEmail(r.Context(), &payload); err != nil {
			log.Printf("Failed to publish verification email: %v", err)
			http.Error(w, "Failed to send verification email", http.StatusInternalServerError)
			return
		}

		log.Printf("Verification email published successfully to: %s", payload.To)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Verification email sent successfully",
		})
	}
}
