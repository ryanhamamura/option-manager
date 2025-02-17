// internal/service/service.go
package service

import (
	"fmt"
	"option-manager/internal/email"
	"option-manager/internal/repository"
)

type Services struct {
	Auth  *AuthService
	User  *UserService
	Email *EmailService
}

func NewServices(repo *repository.Repository, emailClient *email.Client, baseURL string) (*Services, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository is required")
	}
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	// Create EmailService first since other services depend on it
	emailService, err := NewEmailService(emailClient, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create email service: %w", err)
	}

	// Create AuthService
	authService, err := NewAuthService(repo.User, repo.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth service: %w", err)
	}

	// Create UserService
	userService, err := NewUserService(repo.User, emailService)
	if err != nil {
		return nil, fmt.Errorf("failed to create user service: %w", err)
	}
	return &Services{
		Auth:  authService,
		User:  userService,
		Email: emailService,
	}, nil
}
