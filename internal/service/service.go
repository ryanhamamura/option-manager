package service

import (
	"context"
	"fmt"
	"option-manager/internal/types"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service defines the interface for business logic
type Service interface {
	RegisterUser(ctx context.Context, email, firstName, lastName, password string) (types.User, error)
	LoginUser(ctx context.Context, email, password string) (types.User, error)
	GetPositions(ctx context.Context, portfolioID string) ([]types.Position, error)
	GetTrades(ctx context.Context, positionID string) ([]types.Trade, error)
}

// Repository is the data access interface (defined here for simplicity)
type Repository interface {
	SaveUser(ctx context.Context, user types.User) (types.User, error)
	GetUserByEmail(ctx context.Context, email string) (types.User, error)
	GetPositions(ctx context.Context, portfolioID string) ([]types.Position, error)
	GetTrades(ctx context.Context, positionID string) ([]types.Trade, error)
}

// service is the concrete implementation
type service struct {
	repo       Repository
	quoteCache map[string]types.Quote
}

// New creates a new service instance
func New(repo Repository) Service {
	return &service{
		repo:       repo,
		quoteCache: make(map[string]types.Quote),
	}
}

// CreateUser creates a new user with the given name
func (s *service) RegisterUser(ctx context.Context, email, firstName, lastName, password string) (types.User, error) {
	if err := validateUserInput(email, firstName, lastName, password); err != nil {
		return types.User{}, err
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return types.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	user := types.User{
		ID:           uuid.New().String(),
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		PasswordHash: string(hashedPassword), // Store the hash
	}
	return s.repo.SaveUser(ctx, user)
}

// LoginUser authenticates a user by email and password.
func (s *service) LoginUser(ctx context.Context, email, password string) (types.User, error) {
	if email == "" || password == "" {
		return types.User{}, fmt.Errorf("email and password are required")
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return types.User{}, fmt.Errorf("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return types.User{}, fmt.Errorf("invalid email or password")
	}
	return user, nil
}

func (s *service) GetPositions(ctx context.Context, portfolioID string) ([]types.Position, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolioID cannot be empty")
	}
	return s.repo.GetPositions(ctx, portfolioID)
}

func (s *service) GetTrades(ctx context.Context, positionID string) ([]types.Trade, error) {
	if positionID == "" {
		return nil, fmt.Errorf("positionID cannot be empty")
	}
	return s.repo.GetTrades(ctx, positionID)
}

func validateUserInput(email, firstName, lastName, password string) error {
	if email == "" || firstName == "" || lastName == "" || password == "" {
		return fmt.Errorf("all fields (email, first name, last name, password) are required")
	}
	return nil
}
