// internal/repository/postgres/repository.go
package postgres

import (
	"database/sql"
	"option-manager/internal/repository"
)

func NewRepository(db *sql.DB) *repository.Repository {
	return &repository.Repository{
		User:    NewUserRepo(db),
		Session: NewSessionRepo(db),
	}
}
