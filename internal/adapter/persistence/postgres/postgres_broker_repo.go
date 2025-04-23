package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	// Use your actual module path
	"github.com/AMANSRI99/StockSaaS/internal/app/repository"
	"github.com/AMANSRI99/StockSaaS/internal/common/encryptutil" // Import your encrypt util

	"github.com/google/uuid"
)

// PostgresBrokerRepo implements repository.BrokerRepository.
type PostgresBrokerRepo struct {
	db            *sql.DB
	encryptionKey []byte // Key for encrypting/decrypting tokens
}

// NewPostgresBrokerRepo creates a new broker repository instance.
func NewPostgresBrokerRepo(db *sql.DB, encryptionKey []byte) repository.BrokerRepository {
	if len(encryptionKey) != 32 {
		// Or return an error? Panicking is harsh but indicates severe config issue.
		panic("Encryption key must be 32 bytes for PostgresBrokerRepo")
	}
	return &PostgresBrokerRepo{
		db:            db,
		encryptionKey: encryptionKey,
	}
}

// SaveOrUpdateKiteCredentials implements repository.BrokerRepository.SaveOrUpdateKiteCredentials
func (r *PostgresBrokerRepo) SaveOrUpdateKiteCredentials(ctx context.Context, userID uuid.UUID, accessToken []byte, publicToken string, kiteUserID string) error {
	// 1. Encrypt the access token
	encryptedAccessToken, err := encryptutil.Encrypt(accessToken, r.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token for user %s: %w", userID, err)
	}

	// 2. Use INSERT ON CONFLICT (UPSERT)
	query := `
        INSERT INTO user_broker_credentials
            (user_id, broker, kite_user_id, public_token, access_token_encrypted, created_at, updated_at)
        VALUES
            ($1, $2, $3, $4, $5, NOW(), NOW())
        ON CONFLICT (user_id, broker) DO UPDATE SET
            kite_user_id = EXCLUDED.kite_user_id,
            public_token = EXCLUDED.public_token,
            access_token_encrypted = EXCLUDED.access_token_encrypted,
            updated_at = NOW()
    `

	_, err = r.db.ExecContext(ctx, query,
		userID,
		"kite", // Broker identifier
		kiteUserID,
		publicToken,
		encryptedAccessToken, // Store the encrypted bytes
	)

	if err != nil {
		return fmt.Errorf("failed to save/update kite credentials for user %s: %w", userID, err)
	}

	return nil // Success
}

// GetKiteAccessToken implements repository.BrokerRepository.GetKiteAccessToken
func (r *PostgresBrokerRepo) GetKiteAccessToken(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	query := `
        SELECT access_token_encrypted
        FROM user_broker_credentials
        WHERE user_id = $1 AND broker = 'kite'
    `
	var encryptedToken []byte
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&encryptedToken)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrBrokerCredentialsNotFound
		}
		return nil, fmt.Errorf("failed to query kite access token for user %s: %w", userID, err)
	}

	if len(encryptedToken) == 0 {
		// Should not happen if row exists, but good practice
		return nil, fmt.Errorf("retrieved empty encrypted token for user %s", userID)
	}

	// Decrypt the token
	decryptedToken, err := encryptutil.Decrypt(encryptedToken, r.encryptionKey)
	if err != nil {
		// Log the decryption error but maybe return a generic error to caller?
		// Or return ErrBrokerCredentialsNotFound if decryption fails implies data corruption?
		return nil, fmt.Errorf("failed to decrypt access token for user %s: %w", userID, err)
	}

	return decryptedToken, nil
}
