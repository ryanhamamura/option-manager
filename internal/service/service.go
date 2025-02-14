// internal/service/service.go
package service

import (
	"option-manager/internal/email"
	"option-manager/internal/repository"
)

type Services struct {
	Auth *AuthService
	User *UserService
}

func NewServices(repo *repository.Repository, emailService *email.EmailService) *Services {
	return &Services{
		Auth: NewAuthService(repo.User, repo.Session),
		User: NewUserService(repo.User, emailService),
	}
}
