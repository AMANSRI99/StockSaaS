package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrBrokerCredentialsNotFound indicates credentials for a user/broker combo were not found.
var ErrBrokerCredentialsNotFound = errors.New("broker credentials not found for user")

// BrokerRepository defines the interface for storing/retrieving broker credentials.
type BrokerRepository interface {
	// SaveOrUpdateKiteCredentials saves or updates Kite credentials for a user.
	// accessToken should be the raw, unencrypted token. Encryption happens internally.
	SaveOrUpdateKiteCredentials(ctx context.Context, userID uuid.UUID, accessToken []byte, publicToken string, kiteUserID string) error

	// GetKiteAccessToken retrieves the raw, decrypted access token for a user.
	// Returns ErrBrokerCredentialsNotFound if not found.
	GetKiteAccessToken(ctx context.Context, userID uuid.UUID) ([]byte, error)
}