package service

import (
	"context"
	"fmt"
	"option-manager/internal/config"
	"option-manager/internal/email"
	"option-manager/internal/repository"
)

type Services struct {
	Auth  AuthService
	User  UserService
	Email EmailService
}

// EmailService defines email-related operations
type EmailService interface {
	SendVerificationEmail(ctx context.Context, recipient string, firstName string, verificationToken string) error
	SendPasswordResetEmail(ctx context.Context, recipient string, firstName string, resetToken string) error
	SendWelcomeEmail(ctx context.Context, recipient string, firstName string) error
	IsEmailConfigured() bool
	GetSenderAddress() string
}

// UserService defines user-related operations
type UserService interface {
	RegisterUser(ctx context.Context, input RegistrationInput) (*repository.User, error)
	ValidateRegistration(input RegistrationInput) error
	GenerateVerificationToken(ctx context.Context, userID int) (string, error)
	VerifyEmail(ctx context.Context, token string) error
	RequestPasswordReset(ctx context.Context, email string) error
}

type AuthService interface {
	Authenticate(ctx context.Context, email string, password string) (*AuthenticateResponse, error)
	CreateSession(ctx context.Context, userID int, rememberMe bool) (*repository.Session, error)
	GetSession(ctx context.Context, sessionID string) (*repository.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

func NewServices(repo *repository.Repository, emailClient *email.Client, cfg *config.Config) (*Services, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository is required")
	}
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	// Create EmailService first since other services depend on it
	emailService, err := NewEmailService(emailClient, cfg.Email, cfg.App.BaseURL)
	if err != nil {
		// Log warning but continue - application can work without email
		fmt.Printf("Warning: failed to initialize email service: %v\n", err)
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

// MockServices creates a set of mock services for testing
func MockServices() *Services {
	return &Services{
		Auth:  NewMockAuthService(),
		User:  NewMockUserService(),
		Email: NewMockEmailService(),
	}
}

// Mock service constructors for testing
func NewMockAuthService() AuthService {
	return &mockAuthService{}
}

func NewMockUserService() UserService {
	return &mockUserService{}
}

// Mock implementations (minimal examples - expand as needed)
type mockAuthService struct{}
type mockUserService struct{}

// Implement minimal mock methods...
func (m *mockAuthService) Authenticate(ctx context.Context, email, password string) (*AuthenticateResponse, error) {
	return &AuthenticateResponse{}, nil
}

func (m *mockAuthService) CreateSession(ctx context.Context, userID int, rememberMe bool) (*repository.Session, error) {
	return &repository.Session{}, nil
}

func (m *mockAuthService) GetSession(ctx context.Context, sessionID string) (*repository.Session, error) {
	return &repository.Session{}, nil
}

func (m *mockAuthService) DeleteSession(ctx context.Context, sessionID string) error {
	return nil
}

func (m *mockUserService) RegisterUser(ctx context.Context, input RegistrationInput) (*repository.User, error) {
	return &repository.User{}, nil
}

func (m *mockUserService) ValidateRegistration(input RegistrationInput) error {
	return nil
}

func (m *mockUserService) GenerateVerificationToken(ctx context.Context, userID int) (string, error) {
	return "mock-token", nil
}

func (m *mockUserService) VerifyEmail(ctx context.Context, token string) error {
	return nil
}

func (m *mockUserService) RequestPasswordReset(ctx context.Context, email string) error {
	return nil
}
