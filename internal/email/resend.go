package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// ResendService handles email sending via Resend API
type ResendService struct {
	apiKey    string
	fromEmail string
}

// NewResendService creates a new Resend email service
func NewResendService() *ResendService {
	return &ResendService{
		apiKey:    os.Getenv("RESEND_API_KEY"),
		fromEmail: os.Getenv("RESEND_FROM_EMAIL"),
	}
}

// EmailRequest represents the Resend API request structure
type EmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

// EmailResponse represents the Resend API response
type EmailResponse struct {
	ID string `json:"id"`
}

// SendEmail sends an email using the Resend API
func (r *ResendService) SendEmail(to, subject, body string) error {
	// Add delay to avoid rate limit (max 2 requests per second)
	time.Sleep(600 * time.Millisecond)

	if r.apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY not configured")
	}

	if r.fromEmail == "" {
		return fmt.Errorf("RESEND_FROM_EMAIL not configured")
	}

	// Prepare request payload
	emailReq := EmailRequest{
		From:    r.fromEmail,
		To:      []string{to},
		Subject: subject,
		Text:    body,
	}

	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		// Read response body for error details
		var respBody bytes.Buffer
		respBody.ReadFrom(resp.Body)

		return fmt.Errorf("resend API returned status %d - Response: %s", resp.StatusCode, respBody.String())
	}

	var emailResp EmailResponse
	if err := json.NewDecoder(resp.Body).Decode(&emailResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Email enviado com sucesso! ID: %s\n", emailResp.ID)
	return nil
}

// SendEmailWithHTML sends an email with HTML content using the Resend API
func (r *ResendService) SendEmailWithHTML(to, subject, htmlBody string) error {
	// Add delay to avoid rate limit (max 2 requests per second)
	time.Sleep(600 * time.Millisecond)

	if r.apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY not configured")
	}

	if r.fromEmail == "" {
		return fmt.Errorf("RESEND_FROM_EMAIL not configured")
	}

	// Prepare request payload with HTML
	emailReq := EmailRequest{
		From:    r.fromEmail,
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
	}

	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		// Read the error response body for more details
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		return fmt.Errorf("resend API returned status %d: %s", resp.StatusCode, errorBody.String())
	}

	var emailResp EmailResponse
	if err := json.NewDecoder(resp.Body).Decode(&emailResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Email HTML enviado com sucesso! ID: %s\n", emailResp.ID)
	return nil
}
