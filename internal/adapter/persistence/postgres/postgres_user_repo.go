package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings" // For case-insensitive email comparison if needed elsewhere

	// Use your actual module path
	"github.com/AMANSRI99/StockSaaS/internal/app/model"
	"github.com/AMANSRI99/StockSaaS/internal/app/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn" // Import for checking specific Postgres errors
)

const (
	pgUniqueViolationCode = "23505" // Postgres error code for unique constraint violation
)

// PostgresUserRepo is a PostgreSQL implementation of UserRepository.
type PostgresUserRepo struct {
	db *sql.DB
}

// NewPostgresUserRepo creates a new user repository instance.
func NewPostgresUserRepo(db *sql.DB) repository.UserRepository {
	return &PostgresUserRepo{db: db}
}

// Save implements repository.UserRepository.Save
func (r *PostgresUserRepo) Save(ctx context.Context, user *model.User) error {
	query := `
        INSERT INTO users (id, email, password_hash, created_at, updated_at)
        VALUES ($1, LOWER($2), $3, $4, $5)
    `
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email, // Store lowercase email for consistency with index/lookup
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		// Check if the error is a Postgres unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
			// Check if the violation is on the email index
			// Note: The specific constraint name might vary depending on your DB/migration.
			// idx_users_lower_email is the index name we created. Check pgErr.ConstraintName if needed.
			if strings.Contains(pgErr.Message, "idx_users_lower_email") || strings.Contains(pgErr.ConstraintName, "users_email_key") || strings.Contains(pgErr.ConstraintName, "idx_users_lower_email") {
				return repository.ErrUserEmailExists
			}
		}
		// Wrap other errors
		return fmt.Errorf("failed to save user %s: %w", user.Email, err)
	}

	return nil // Success
}

// FindByEmail implements repository.UserRepository.FindByEmail
func (r *PostgresUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
        SELECT id, email, password_hash, created_at, updated_at
        FROM users
        WHERE LOWER(email) = LOWER($1)
    `
	row := r.db.QueryRowContext(ctx, query, email)

	var user model.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound // Use the custom error
		}
		// Wrap other scan/query errors
		return nil, fmt.Errorf("failed to find user by email %s: %w", email, err)
	}

	return &user, nil // Success
}

// FindByID implements repository.UserRepository.FindByID
func (r *PostgresUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
        SELECT id, email, password_hash, created_at, updated_at
        FROM users
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	var user model.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound // Use the custom error
		}
		// Wrap other scan/query errors
		return nil, fmt.Errorf("failed to find user by ID %s: %w", id, err)
	}

	return &user, nil // Success
}
