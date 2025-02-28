package service

import (
	"errors"
	"fmt"
	"option-manager/internal/types"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service defines the interface for business logic
type Service interface {
	RegisterUser(email, firstName, lastName, password string) (types.User, error)
	LoginUser(email, password string) (types.User, error)
	GetPositions(portfolioID string) ([]types.Position, error)
	GetTrades(positionID string) ([]types.Trade, error)
}

// Repository is the data access interface (defined here for simplicity)
type Repository interface {
	SaveUser(user types.User) (types.User, error)
	GetUserByEmail(email string) (types.User, error)
	GetPositions(portfolioID string) ([]types.Position, error)
	GetTrades(positionID string) ([]types.Trade, error)
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
func (s *service) RegisterUser(email, firstName, lastName, password string) (types.User, error) {
	// Basic validation
	if email == "" || firstName == "" || lastName == "" || password == "" {
		return types.User{}, errors.New("all fields (email, first name, last name, password) are required")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return types.User{}, errors.New("failed to hash password: " + err.Error())
	}

	user := types.User{
		ID:           uuid.New().String(),
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		PasswordHash: string(hashedPassword), // Store the hash
	}

	updatedUser, err := s.repo.SaveUser(user)
	if err != nil {
		return types.User{}, errors.New("failed to register user: " + err.Error())
	}

	return updatedUser, nil
}

// LoginUser authenticates a user by email and password.
func (s *service) LoginUser(email, password string) (types.User, error) {
	if email == "" || password == "" {
		return types.User{}, errors.New("email and password are required")
	}

	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		fmt.Printf("GetUserByEmail failed for %s: %v\n", email, err)
		return types.User{}, errors.New("invalid email or password")
	}

	fmt.Printf("Retrieved user: %+v\n", user)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		fmt.Printf("Password check failed: %v\n", err)
		return types.User{}, errors.New("invalid email or password")
	}

	return user, nil
}

func (s *service) GetPositions(portfolioID string) ([]types.Position, error) {
	return s.repo.GetPositions(portfolioID)
}

func (s *service) GetTrades(positionID string) ([]types.Trade, error) {
	return s.repo.GetTrades(positionID)
}
