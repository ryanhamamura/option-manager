// internal/repository/postgres/user.go
package postgres

import (
	"context"
	"database/sql"
	"option-manager/internal/repository"
	"time"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *repository.User) error {
	query := `
        INSERT INTO users (
            email, password_hash, first_name, last_name, 
            email_verified, verification_token, verification_expires_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.EmailVerified,
		user.VerificationToken,
		user.VerificationExpiry,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepo) FindByID(ctx context.Context, id int) (*repository.User, error) {
	user := &repository.User{}
	query := `
        SELECT 
            id, email, password_hash, first_name, last_name,
            email_verified, verification_token, verification_expires_at,
            created_at, updated_at
        FROM users 
        WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.EmailVerified,
		&user.VerificationToken,
		&user.VerificationExpiry,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) UpdateVerificationStatus(ctx context.Context, userID int, verified bool) error {
	query := `
        UPDATE users 
        SET email_verified = $2,
            verification_token = NULL,
            verification_expires_at = NULL,
            updated_at = NOW()
        WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID, verified)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *UserRepo) SetVerificationToken(ctx context.Context, userID int, token string, expiry time.Time) error {
	query := `
        UPDATE users 
        SET verification_token = $2,
            verification_expires_at = $3,
            updated_at = NOW()
        WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID, token, expiry)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*repository.User, error) {
	user := &repository.User{}
	query := `
        SELECT 
            id, email, password_hash, first_name, last_name,
            email_verified, verification_token, verification_expires_at,
            created_at, updated_at
        FROM users 
        WHERE email = $1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.EmailVerified,
		&user.VerificationToken,
		&user.VerificationExpiry,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
