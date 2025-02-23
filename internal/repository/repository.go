package repository

import (
	"database/sql"
	"fmt"
	"option-manager/internal/service"
	"option-manager/internal/types"
	"time"

	_ "github.com/lib/pq"
)

// postgresRepo is the PostgreSQL implementaiton
type postgresRepo struct {
	db *sql.DB
}

// New creates a new PostgreSQL repository
func New(dataSourceName string) service.Repository {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		panic(fmt.Errorf("Failed to connect to database: %v", err))
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("Failed to ping database: %v", err))
	}

	return &postgresRepo{db: db}
}

func (r *postgresRepo) SaveUser(user types.User) (types.User, error) {
	query := `INSERT INTO users (id, email, first_name, last_name, password_hash) 
						VALUES ($1, $2, $3, $4, $5)
						RETURNING created_at, updated_at`
	var createdAt, updatedAt time.Time
	_, err := r.db.Exec(query, user.ID, user.Email, user.FirstName, user.LastName, user.PasswordHash)
	if err != nil {
		return types.User{}, fmt.Errorf("failed to insert user: %w", err)
	}
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	return user, nil
}

// Close shuts down the database connection (optional, for cleanup)
func (r *postgresRepo) Close() error {
	return r.db.Close()
}
