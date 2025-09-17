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

// GetWelcomeEmailHTML returns the HTML template for welcome emails
func (r *ResendService) GetWelcomeEmailHTML(username, companyName string) string {
	template := `<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Bem-vindo</title>
  <style>
    body,table,td {font-family: Arial, Helvetica, sans-serif; margin:0; padding:0;}
    img {border:0; display:block;}
    a {color:#ffffff; text-decoration:none}

    .wrapper {width:100%; background:#f0f2f5; padding:30px 0;}
    .content {max-width:600px; background:#ffffff; margin:0 auto; border-radius:10px; overflow:hidden; box-shadow:0 4px 12px rgba(0,0,0,0.08)}

    .header {background:#1a73e8; padding:30px; text-align:center; color:#fff;}
    .header h1 {margin:0; font-size:24px;}

    .body {padding:30px; color:#333; line-height:1.6;}
    .body h2 {margin-top:0; color:#1a73e8;}

    .btn {display:inline-block; background:#1a73e8; padding:12px 20px; border-radius:6px; font-weight:bold;}

    .footer {background:#f7f7f7; padding:20px; font-size:12px; text-align:center; color:#666;}

    @media only screen and (max-width:480px) {
      .header h1 {font-size:20px;}
      .body h2 {font-size:18px;}
    }
  </style>
</head>
<body>
  <table role="presentation" class="wrapper" width="100%" cellspacing="0" cellpadding="0">
    <tr>
      <td align="center">
        <table role="presentation" class="content" width="100%" cellspacing="0" cellpadding="0">
          
          <!-- Header -->
          <tr>
            <td class="header">
              <h1>Bem-vindo(a) à ` + companyName + ` 🎉</h1>
            </td>
          </tr>

          <!-- Body -->
          <tr>
            <td class="body">
              <h2>Estamos muito felizes em ter você conosco!</h2>
              <p>Agora você faz parte da nossa comunidade e terá acesso a todas as vantagens que preparamos para você.</p>

              <p>Para começar, recomendamos:</p>
              <ul>
                <li>Completar seu perfil;</li>
                <li>Explorar os recursos principais;</li>
                <li>Ativar notificações para não perder nenhuma novidade.</li>
              </ul>

              <p style="margin:20px 0; text-align:center;">
                <a href="https://seusite.exemplo/entrar" target="_blank" class="btn">Acessar minha conta</a>
              </p>

              <p>Se precisar de ajuda, nossa equipe está à disposição. Basta responder este e-mail ou acessar nossa central de suporte.</p>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td class="footer">
              <p>Você recebeu este e-mail porque se cadastrou em ` + companyName + `.</p>
              <p>Endereço da empresa • Cidade • Estado</p>
              <p><a href="#">Cancelar inscrição</a></p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

	return template
}
