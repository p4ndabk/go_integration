package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go_integration/internal/models"
	"go_integration/internal/user"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService *user.Service
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *user.Service) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser handles POST /create-user requests
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload models.UserPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	id, err := h.userService.CreateUser(context.Background(), &payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": fmt.Sprintf("User creation message published with ID: %s", id),
		"id":      id,
		"user":    payload,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
