package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidOrExpiredState = errors.New("invalid or expired OAuth state token")

// OAuthStateRepository manages temporary state tokens for OAuth flows.
type OAuthStateRepository interface {
	// SaveState stores the state token associated with a user ID and expiry.
	SaveState(ctx context.Context, stateToken string, userID uuid.UUID, expiry time.Duration) error

	// VerifyAndConsumeState checks if the state token is valid, returns the associated userID,
	// and deletes the token to prevent reuse. Returns ErrInvalidOrExpiredState if not found or expired.
	VerifyAndConsumeState(ctx context.Context, stateToken string) (userID uuid.UUID, err error)
}
