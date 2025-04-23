package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AMANSRI99/StockSaaS/internal/app/repository"
	"github.com/google/uuid"
)

type PostgresOAuthStateRepo struct {
	db *sql.DB
}

func NewPostgresOAuthStateRepo(db *sql.DB) repository.OAuthStateRepository {
	return &PostgresOAuthStateRepo{db: db}
}

func (r *PostgresOAuthStateRepo) SaveState(ctx context.Context, stateToken string, userID uuid.UUID, expiry time.Duration) error {
	query := `INSERT INTO oauth_states (state_token, user_id, expires_at) VALUES ($1, $2, $3)`
	expiresAt := time.Now().UTC().Add(expiry)
	_, err := r.db.ExecContext(ctx, query, stateToken, userID, expiresAt)
	if err != nil {
		// Handle potential primary key conflicts? Should be rare with good random state.
		return fmt.Errorf("failed to save oauth state: %w", err)
	}
	return nil
}

func (r *PostgresOAuthStateRepo) VerifyAndConsumeState(ctx context.Context, stateToken string) (userID uuid.UUID, err error) {
	// Use a transaction to SELECT, check expiry, and DELETE atomically (or as close as possible)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to begin state verification transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil || err != nil {
			tx.Rollback() // Ensure rollback on panic or error
		}
	}() // Add basic panic safety

	query := `SELECT user_id, expires_at FROM oauth_states WHERE state_token = $1 FOR UPDATE` // Lock row
	var expiresAt time.Time
	err = tx.QueryRowContext(ctx, query, stateToken).Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = repository.ErrInvalidOrExpiredState // Set error before rollback defer runs
			return uuid.Nil, err
		}
		err = fmt.Errorf("failed to query oauth state: %w", err) // Set error before rollback defer runs
		return uuid.Nil, err
	}

	// Check expiry
	if time.Now().UTC().After(expiresAt) {
		// Attempt to delete the expired token anyway
		_, delErr := tx.ExecContext(ctx, "DELETE FROM oauth_states WHERE state_token = $1", stateToken)
		if delErr != nil { /* log warning */
		}
		tx.Commit()                               // Commit the deletion of expired token
		err = repository.ErrInvalidOrExpiredState // Set error before returning
		return uuid.Nil, err
	}

	// State is valid and not expired, consume (delete) it
	deleteQuery := `DELETE FROM oauth_states WHERE state_token = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, stateToken)
	if err != nil {
		err = fmt.Errorf("failed to delete consumed oauth state: %w", err) // Set error before rollback
		return uuid.Nil, err
	}

	// Commit the transaction (find and delete)
	err = tx.Commit()
	if err != nil {
		err = fmt.Errorf("failed to commit state verification transaction: %w", err)
		return uuid.Nil, err
	}

	return userID, nil // Success
}
