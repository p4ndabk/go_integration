package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go_integration/internal/email"
	"go_integration/internal/models"
)

// EmailHandler handles HTTP requests for sending emails
type EmailHandler struct {
	emailService *email.Service
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(emailService *email.Service) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

// SendEmail handles POST /send-email requests
func (h *EmailHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload models.EmailPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	id, err := h.emailService.SendEmail(context.Background(), &payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send email: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": fmt.Sprintf("Mensagem publicada com ID: %s", id),
		"id":      id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
