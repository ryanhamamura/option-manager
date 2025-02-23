package service

import (
	"errors"
	"option-manager/internal/types"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service defines the interface for business logic
type Service interface {
	RegisterUser(email, firstName, lastName, password string) (types.User, error)
}

// Repository is the data access interface (defined here for simplicity)
type Repository interface {
	SaveUser(user types.User) (types.User, error)
}

// service is the concrete implementation
type service struct {
	repo Repository
}

// New creates a new service instance
func New(repo Repository) Service {
	return &service{repo: repo}
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
