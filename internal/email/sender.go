package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

func (s *EmailService) SendVerificationEmail(recipient, firstName, verificationToken string, baseURL string) error {
	verificationLink := fmt.Sprintf("%s/verify?token=%s", baseURL, verificationToken)

	// Prepare email data
	data := VerificationEmailData{
		FirstName:        firstName,
		VerificationLink: verificationLink,
	}

	// Parse and execute template
	tmpl, err := template.New("verification").Parse(verificationEmailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var bodyBuf bytes.Buffer
	if err := tmpl.Execute(&bodyBuf, data); err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	// Prepare SES input
	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{recipient},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(bodyBuf.String()),
				},
				Text: &types.Content{
					Data: aws.String(fmt.Sprintf("Please verify your email by visiting: %s", verificationLink)),
				},
			},
			Subject: &types.Content{
				Data: aws.String("Verify Your Email - Options Manager"),
			},
		},
		Source: aws.String(s.sender),
	}

	// Send email
	_, err = s.sesClient.SendEmail(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
