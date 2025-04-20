package repository

import (
	"github.com/AMANSRI99/StockSaaS/internal/app/model"

	"context"
	"errors"

	"github.com/google/uuid"
)

// Define standard errors for user repository operations
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserEmailExists = errors.New("user email already exists")
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	// Save creates a new user record.
	// Returns ErrUserEmailExists if the email is already taken.
	Save(ctx context.Context, user *model.User) error

	// FindByEmail retrieves a user by their email address (case-insensitive).
	// Returns ErrUserNotFound if no user is found.
	FindByEmail(ctx context.Context, email string) (*model.User, error)

	// FindByID retrieves a user by their unique ID.
	// Returns ErrUserNotFound if no user is found.
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}
