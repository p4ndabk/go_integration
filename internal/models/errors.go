package models

import (
	"errors"
	"fmt"
)

var (
	// ErrMissingRecipient is returned when the "to" field is empty
	ErrMissingRecipient = errors.New("recipient email is required")

	// ErrMissingSubject is returned when the "subject" field is empty
	ErrMissingSubject = errors.New("email subject is required")

	// ErrMissingBody is returned when the "body" field is empty
	ErrMissingBody = errors.New("email body is required")
)

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string
	Message string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", v.Field, v.Message)
}
