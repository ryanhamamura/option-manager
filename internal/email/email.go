// internal/email/email.go
package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// Client handles the low-level email sending functionality
type Client struct {
	sesClient *ses.Client
	sender    string
}

// EmailContent represents the content of an email
type EmailContent struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
}

// NewClient creates a new email client
func NewClient(awsRegion, sender string) (*Client, error) {
	if awsRegion == "" {
		return nil, fmt.Errorf("AWS region is required")
	}
	if sender == "" {
		return nil, fmt.Errorf("sender email is required")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return &Client{
		sesClient: ses.NewFromConfig(cfg),
		sender:    sender,
	}, nil
}

// Send sends an email using AWS SES
func (c *Client) Send(ctx context.Context, content *EmailContent) error {
	if content.To == "" {
		return fmt.Errorf("recipient email is required")
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
