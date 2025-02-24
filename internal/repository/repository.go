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
	err := r.db.QueryRow(query, user.ID, user.Email, user.FirstName, user.LastName, user.PasswordHash).Scan(&createdAt, &updatedAt)
	if err != nil {
		fmt.Printf("SaveUser failed: %v\n", err)
		return types.User{}, fmt.Errorf("failed to insert user: %w", err)
	}
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	return user, nil
}

// GetUserByEmail retrieves User from the database with email
func (r *postgresRepo) GetUserByEmail(email string) (types.User, error) {
	var user types.User
	query := `SELECT id, email, first_name, last_name, password_hash, created_at, updated_at
						FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		fmt.Printf("GetUserByEmail: no user found for email %s\n", email)
		return types.User{}, fmt.Errorf("user not found: %w", err)
	}
	if err != nil {
		fmt.Printf("GetUserByEmail failed: %v\n", err)
		return types.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	fmt.Printf("GetUserByEmail: succeeded: found user %+v\n", user)
	return user, nil
}

// Close shuts down the database connection (optional, for cleanup)
func (r *postgresRepo) Close() error {
	return r.db.Close()
}
