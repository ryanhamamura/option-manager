package email

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

type EmailService struct {
	sesClient *ses.Client
	sender    string
}

func NewEmailService(awsRegion, sender string) (*EmailService, error) {
	if awsRegion == "" {
		return nil, fmt.Errorf("AWS_REGION is required")
	}
	if sender == "" {
		return nil, fmt.Errorf("EMAIL_SENDER is required")
	}

	log.Printf("Attempting to load AWS config with region: %s", awsRegion)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	log.Printf("AWS config loaded successfully, creating SES client")
	client := ses.NewFromConfig(cfg)

	log.Printf("SES client created, initializing email service with sender: %s", sender)
	return &EmailService{
		sesClient: client,
		sender:    sender,
	}, nil
}
