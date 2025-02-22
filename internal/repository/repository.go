package repository

import (
	"database/sql"
	"fmt"
	"option-manager/internal/service"
	"option-manager/internal/types"

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

func (r *postgresRepo) SaveUser(user types.User) error {
	query := `INSERT INTO users (id, email, first_name, last_name, password_hash, created_at, updated_at) 
						VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query, user.ID, user.FirstName, user.LastName, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

// Close shuts down the database connection (optional, for cleanup)
func (r *postgresRepo) Close() error {
	return r.db.Close()
}
