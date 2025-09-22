package models

import (
	"encoding/json"
	"fmt"
)

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

// GenerateSubject returns the email subject (already provided)
func (e *EmailPayload) GenerateSubject() string {
	return e.Subject
}

// GenerateBody generates the HTML email body for regular emails
func (e *EmailPayload) GenerateBody() string {
	return fmt.Sprintf(`
		<html>
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>%s</title>
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
			<div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px; border-radius: 10px 10px 0 0; text-align: center;">
				<h1 style="color: white; margin: 0; font-size: 28px;">📧 Nova Mensagem</h1>
			</div>
			<div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
				<h2 style="color: #333; margin-top: 0; border-bottom: 2px solid #667eea; padding-bottom: 10px;">%s</h2>
				<div style="background: white; padding: 20px; border-radius: 8px; border-left: 4px solid #667eea; margin: 20px 0;">
					<p style="margin: 0; font-size: 16px; line-height: 1.8;">%s</p>
				</div>
				<hr style="border: none; border-top: 1px solid #e0e0e0; margin: 30px 0;">
				<p style="color: #666; font-size: 12px; text-align: center; margin: 0;">
					<small>Este é um email automático gerado pelo sistema Go Pub/Sub Integration</small>
				</p>
			</div>
		</body>
		</html>
	`, e.Subject, e.Subject, e.Body)
}

// VerificationEmailPayload represents the structure of a verification email message
type VerificationEmailPayload struct {
	To        string `json:"to"`
	Username  string `json:"username"`
	Token     string `json:"token,omitempty"`      // Optional: for backward compatibility
	Code      string `json:"code,omitempty"`       // Verification code
	VerifyURL string `json:"verify_url,omitempty"` // Optional: for backward compatibility
}

// Validate validates the verification email payload
func (v *VerificationEmailPayload) Validate() error {
	if v.To == "" {
		return ErrMissingRecipient
	}
	if v.Username == "" {
		return &ValidationError{Field: "username", Message: "username is required"}
	}
	// Either code or verify_url must be provided (or both for backward compatibility)
	if v.Code == "" && v.VerifyURL == "" {
		return &ValidationError{Field: "code_or_url", Message: "either verification code or verify_url is required"}
	}
	return nil
}

// ToJSON converts the verification payload to JSON bytes
func (v *VerificationEmailPayload) ToJSON() ([]byte, error) {
	return json.Marshal(v)
}

// GenerateSubject generates the email subject for verification
func (v *VerificationEmailPayload) GenerateSubject() string {
	return "Confirme sua conta - Verificação de Email"
}

// GenerateBody generates the HTML email body for verification
func (v *VerificationEmailPayload) GenerateBody() string {
	return fmt.Sprintf(`
		<html>
		<body>
			<h2>Bem-vindo, %s!</h2>
			<p>Para confirmar sua conta, clique no link abaixo:</p>
			<p><a href="%s" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Verificar Email</a></p>
			<p>Ou copie e cole este link no seu navegador:</p>
			<p>%s</p>
			<br>
			<p><small>Se você não criou uma conta, ignore este email.</small></p>
		</body>
		</html>
	`, v.Username, v.VerifyURL, v.VerifyURL)
}
