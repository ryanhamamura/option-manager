// internal/service/email_service.go
package service

import (
	"context"
	"fmt"
	"net/url"
	"option-manager/internal/config"
	"option-manager/internal/email"
	"path"
)

// EmailService handles all email-related operations
type EmailService struct {
	client  *email.Client
	config  config.EmailConfig
	baseURL string
}

// NewEmailService creates a new EmailService
func NewEmailService(emailClient *email.Client, cfg config.EmailConfig, baseURL string) (*EmailService, error) {
	if emailClient == nil {
		return nil, fmt.Errorf("email client is required")
	}
	if cfg.SenderAddress == "" {
		return nil, fmt.Errorf("sender address is required")
	}
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	// Validate base URL format
	_, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &EmailService{
		client:  emailClient,
		config:  cfg,
		baseURL: baseURL,
	}, nil

}

// SendVerificationEmail sends an email verification link to the user
func (s *EmailService) SendVerificationEmail(ctx context.Context, recipient, firstName, verificationToken string) error {
	if recipient == "" {
		return fmt.Errorf("recipient email is required")
	}
	if firstName == "" {
		return fmt.Errorf("first name is required")
	}
	if verificationToken == "" {
		return fmt.Errorf("verification token is required")
	}

	// Build verification URL
	verificationURL, err := url.Parse(s.baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}
	verificationURL.Path = path.Join(verificationURL.Path, "verify")
	q := verificationURL.Query()
	q.Set("token", verificationToken)
	verificationURL.RawQuery = q.Encode()

	content := &email.EmailContent{
		To:       recipient,
		Subject:  "Verify Your Email - Options Manager",
		Template: "verification",
		TemplateData: map[string]interface{}{
			"FirstName":        firstName,
			"VerificationLink": verificationURL.String(),
		},
	}

	if err := s.client.Send(ctx, content); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// SendPasswordResetEmail sends a password reset link to a user
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, recipient, firstName, resetToken string) error {
	if recipient == "" {
		return fmt.Errorf("recipient email is required")
	}
	if firstName == "" {
		return fmt.Errorf("first name is required")
	}
	if resetToken == "" {
		return fmt.Errorf("reset token is required")
	}

	// Build reset URL
	resetURL, err := url.Parse(s.baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}
	resetURL.Path = path.Join(resetURL.Path, "reset-password")
	q := resetURL.Query()
	q.Set("token", resetToken)
	resetURL.RawQuery = q.Encode()

	content := &email.EmailContent{
		To:       recipient,
		Subject:  "Reset Your Password - Options Manager",
		Template: "reset_password",
		TemplateData: map[string]interface{}{
			"FirstName": firstName,
			"ResetLink": resetURL.String(),
		},
	}

	if err := s.client.Send(ctx, content); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// SendWelcomeEmail sends a welcome email to a newly verified user
func (s *EmailService) SendWelcomeEmail(ctx context.Context, recipient, firstName string) error {
	if recipient == "" {
		return fmt.Errorf("recipient email is required")
	}
	if firstName == "" {
		return fmt.Errorf("first name is required")
	}

	content := &email.EmailContent{
		To:       recipient,
		Subject:  "Welcome to Options Manager",
		Template: "welcome",
		TemplateData: map[string]interface{}{
			"FirstName": firstName,
			"LoginURL":  fmt.Sprintf("%s/login", s.baseURL),
		},
	}

	if err := s.client.Send(ctx, content); err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	return nil
}

// IsEmailConfigured returns true if email sending is properly configured
func (s *EmailService) IsEmailConfigured() bool {
	return s.config.SenderAddress != "" && s.client != nil
}

// GetSenderAddress returns the configured sender address
func (s *EmailService) GetSenderAddress() string {
	return s.config.SenderAddress
}

// For testing and development
type MockEmailService struct {
	SentEmails []MockEmail
}

type MockEmail struct {
	To           string
	Subject      string
	Template     string
	TemplateData map[string]interface{}
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SentEmails: make([]MockEmail, 0),
	}
}

func (m *MockEmailService) SendVerificationEmail(ctx context.Context, recipient, firstName, verificationToken string) error {
	m.SentEmails = append(m.SentEmails, MockEmail{
		To:       recipient,
		Subject:  "Verify Your Email - Options Manager",
		Template: "verification",
		TemplateData: map[string]interface{}{
			"FirstName":        firstName,
			"VerificationLink": verificationToken,
		},
	})
	return nil
}

func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, recipient, firstName, resetToken string) error {
	m.SentEmails = append(m.SentEmails, MockEmail{
		To:       recipient,
		Subject:  "Reset Your Password - Options Manager",
		Template: "reset_password",
		TemplateData: map[string]interface{}{
			"FirstName": firstName,
			"ResetLink": resetToken,
		},
	})
	return nil
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, recipient, firstName string) error {
	m.SentEmails = append(m.SentEmails, MockEmail{
		To:       recipient,
		Subject:  "Welcome to Options Manager",
		Template: "welcome",
		TemplateData: map[string]interface{}{
			"FirstName": firstName,
		},
	})
	return nil
}
