// internal/repository/postgres/session.go
package postgres

import (
	"context"
	"database/sql"
	"option-manager/internal/repository"
)

type SessionRepo struct {
	db *sql.DB
}

func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, session *repository.Session) error {
	query := `
        INSERT INTO sessions (id, user_id, expires_at)
        VALUES ($1, $2, $3)
        RETURNING created_at`

	return r.db.QueryRowContext(
		ctx,
		query,
		session.ID,
		session.UserID,
		session.ExpiresAt,
	).Scan(&session.CreatedAt)
}

func (r *SessionRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
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

func (r *SessionRepo) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at <= NOW()`

	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *SessionRepo) FindByID(ctx context.Context, id string) (*repository.Session, error) {
	session := &repository.Session{}
	query := `
        SELECT id, user_id, expires_at, created_at
        FROM sessions
        WHERE id = $1 AND expires_at > NOW()`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return session, nil
}
