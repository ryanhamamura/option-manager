// internal/email/email.go
package email

import (
	"bytes"
	"context"
	"fmt"
	"option-manager/internal/config"
	"path/filepath"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// Client handles the low-level email sending functionality
type Client struct {
	sesClient    *ses.Client
	sender       string
	templatesDir string
	templates    map[string]*template.Template
}

// EmailContent represents the content of an email
type EmailContent struct {
	To           string
	Subject      string
	HTMLBody     string
	TextBody     string
	Template     string
	TemplateData interface{}
}

// Options configures the email client behavior
type Options struct {
	RetryAttempts  int
	RetryDelay     time.Duration
	RequestTimeout time.Duration
}

// DefaultOptions returns sensible defaults for the email client
func DefaultOptions() Options {
	return Options{
		RetryAttempts:  3,
		RetryDelay:     time.Second * 2,
		RequestTimeout: time.Second * 10,
	}
}

// NewClient creates a new email client
func NewClient(cfg config.EmailConfig, awsCfg config.AWSConfig) (*Client, error) {
	return NewClientWithOptions(cfg, awsCfg, DefaultOptions())
}

// NewClientWithOptions creates a new email client with custom options
func NewClientWithOptions(cfg config.EmailConfig, awsCfg config.AWSConfig, opts Options) (*Client, error) {
	if cfg.SenderAddress == "" {
		return nil, fmt.Errorf("sender email address is required")
	}

	// Load AWS configuration
	sdkConfig, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(awsCfg.Region),
		awsconfig.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     awsCfg.AccessKeyID,
				SecretAccessKey: awsCfg.SecretAccessKey,
			}, nil
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := &Client{
		sesClient:    ses.NewFromConfig(sdkConfig),
		sender:       cfg.SenderAddress,
		templatesDir: cfg.TemplatesDir,
		templates:    make(map[string]*template.Template),
	}

	// Load email templates
	if err := client.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load email templates: %w", err)
	}

	return client, nil
}

// loadTemplates loads all email templates from the templates directory
func (c *Client) loadTemplates() error {
	if c.templatesDir == "" {
		return fmt.Errorf("templates directory not configured")
	}

	templates := []string{
		"verification.html",
		"verification.txt",
		"reset_password.html",
		"reset_password.txt",
		// Add more templates as needed
	}

	for _, tmpl := range templates {
		t, err := template.ParseFiles(filepath.Join(c.templatesDir, tmpl))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", tmpl, err)
		}
		c.templates[tmpl] = t
	}

	return nil
}

// Send sends an email using AWS SES
func (c *Client) Send(ctx context.Context, content *EmailContent) error {
	if content.To == "" {
		return fmt.Errorf("recipient email is required")
	}
	if content.Template != "" {
		if err := c.applyTemplate(content); err != nil {
			return fmt.Errorf("failed to apply template: %w", err)
		}
	}
	if content.Subject == "" {
		return fmt.Errorf("email subject is required")
	}
	if content.HTMLBody == "" && content.TextBody == "" {
		return fmt.Errorf("email body (HTML or text) is required")
	}

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{content.To},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: aws.String(content.Subject),
			},
			Body: &types.Body{},
		},
		Source: aws.String(c.sender),
	}

	// Add HTML body if provided
	if content.HTMLBody != "" {
		input.Message.Body.Html = &types.Content{
			Data: aws.String(content.HTMLBody),
		}
	}

	// Add text body if provided
	if content.TextBody != "" {
		input.Message.Body.Text = &types.Content{
			Data: aws.String(content.TextBody),
		}
	}

	_, err := c.sesClient.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// applyTemplate applies the specified template to the email content
func (c *Client) applyTemplate(content *EmailContent) error {
	htmlTemplate := c.templates[content.Template+".html"]
	textTemplate := c.templates[content.Template+".txt"]

	if htmlTemplate == nil && textTemplate == nil {
		return fmt.Errorf("template %s not found", content.Template)
	}

	if htmlTemplate != nil {
		var htmlBuf bytes.Buffer
		if err := htmlTemplate.Execute(&htmlBuf, content.TemplateData); err != nil {
			return fmt.Errorf("failed to execute HTML template: %w", err)
		}
		content.HTMLBody = htmlBuf.String()
	}

	if textTemplate != nil {
		var textBuf bytes.Buffer
		if err := textTemplate.Execute(&textBuf, content.TemplateData); err != nil {
			return fmt.Errorf("failed to execute text template: %w", err)
		}
		content.TextBody = textBuf.String()
	}

	return nil
}

// SendVerificationEmail sends an email verification link to a user
func (c *Client) SendVerificationEmail(recipient, firstName, verificationToken string) error {
	content := &EmailContent{
		To:       recipient,
		Subject:  "Verify Your Email - Options Manager",
		Template: "verification",
		TemplateData: map[string]interface{}{
			"FirstName":        firstName,
			"VerificationLink": verificationToken,
		},
	}

	return c.Send(context.Background(), content)
}

// SendPasswordResetEmail sends a password reset link to a user
func (c *Client) SendPasswordResetEmail(recipient, firstName, resetToken string) error {
	content := &EmailContent{
		To:       recipient,
		Subject:  "Reset Your Password - Options Manager",
		Template: "reset_password",
		TemplateData: map[string]interface{}{
			"FirstName": firstName,
			"ResetLink": resetToken,
		},
	}

	return c.Send(context.Background(), content)
}
