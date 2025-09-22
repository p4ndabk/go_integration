package models

import (
	"encoding/json"
	"fmt"
)

// UserPayload represents the structure of a user creation message
type UserPayload struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username,omitempty"`
}

// Validate validates the user payload
func (u *UserPayload) Validate() error {
	if u.ID == "" {
		return fmt.Errorf("missing user ID")
	}
	if u.Email == "" {
		return fmt.Errorf("missing user email")
	}
	if u.Name == "" {
		return fmt.Errorf("missing user name")
	}
	return nil
}

// ToJSON converts the payload to JSON bytes
func (u *UserPayload) ToJSON() ([]byte, error) {
	return json.Marshal(u)
}

// UserFromJSON parses JSON bytes into a UserPayload
func UserFromJSON(data []byte) (*UserPayload, error) {
	var payload UserPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user payload: %w", err)
	}
	return &payload, nil
}
