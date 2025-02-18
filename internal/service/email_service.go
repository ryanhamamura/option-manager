// internal/service/email_service.go
package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"option-manager/internal/email"
)

// EmailService handles all email-related operations
type EmailService struct {
	baseURL string
	client  *email.Client
}

// NewEmailService creates a new EmailService
func NewEmailService(emailClient *email.Client, baseURL string) (*EmailService, error) {
	if emailClient == nil {
		return nil, fmt.Errorf("email client is required")
	}
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	return &EmailService{
		client:  emailClient,
		baseURL: baseURL,
	}, nil
}

// VerificationEmailData holds data for verification email template
type VerificationEmailData struct {
	FirstName        string
	VerificationLink string
}

// SendVerificationEmail sends an email verification link to the user
func (s *EmailService) SendVerificationEmail(recipient, firstName, verificationToken string) error {
	data := VerificationEmailData{
		FirstName:        firstName,
		VerificationLink: fmt.Sprintf("%s/verify?token=%s", s.baseURL, verificationToken),
	}

	// Parse and execute template
	htmlContent, err := s.executeTemplate(verificationEmailTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to generate email content: %v", err)
	}

	// Generate plain text version
	textContent := fmt.Sprintf("Please verify your email by visiting: %s", data.VerificationLink)

	// Send email using the client
	content := &email.EmailContent{
		To:       recipient,
		Subject:  "Verify Your Email - Options Manager",
		HTMLBody: htmlContent,
		TextBody: textContent,
	}

	if err := s.client.Send(context.Background(), content); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// executeTemplate is a helper function to execute HTML templates
func (s *EmailService) executeTemplate(tmpl string, data interface{}) (string, error) {
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Email templates
const verificationEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Welcome to Options Manager!</h2>
        <p>Hello {{.FirstName}},</p>
        <p>Thank you for registering. Please verify your email address by clicking the button below:</p>
        <p style="text-align: center;">
            <a href="{{.VerificationLink}}" 
               style="display: inline-block; padding: 12px 24px; background-color: #3b82f6; color: white; 
                      text-decoration: none; border-radius: 4px; font-weight: bold;">
                Verify Email Address
            </a>
        </p>
        <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
        <p>{{.VerificationLink}}</p>
        <p>This link will expire in 24 hours.</p>
        <p>If you didn't create an account, you can safely ignore this email.</p>
        <br>
        <p>Best regards,<br>The Options Manager Team</p>
    </div>
</body>
</html>`
