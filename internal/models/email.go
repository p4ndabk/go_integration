package models

import "encoding/json"

// EmailPayload represents the structure of an email message
type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Validate validates the email payload
func (e *EmailPayload) Validate() error {
	if e.To == "" {
		return ErrMissingRecipient
	}
	if e.Subject == "" {
		return ErrMissingSubject
	}
	if e.Body == "" {
		return ErrMissingBody
	}
	return nil
}

// ToJSON converts the payload to JSON bytes
func (e *EmailPayload) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON creates an EmailPayload from JSON bytes
func FromJSON(data []byte) (*EmailPayload, error) {
	var payload EmailPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
