package repository

import (
	// Use your actual module path
	"github.com/AMANSRI99/StockSaaS/internal/app/model"

	"context" // Use context for cancellation/deadlines
	"errors"  // For defining custom errors
)

// ErrBasketNotFound is returned when a basket is not found.
var ErrBasketNotFound = errors.New("basket not found")

// BasketRepository defines the interface for basket data operations.
// Any struct that implements these methods satisfies the interface.
type BasketRepository interface {
	// Save creates a new basket or updates an existing one.
	// For simplicity now, we'll assume it only creates.
	Save(ctx context.Context, basket *model.Basket) error

	// FindAll retrieves all baskets.
	FindAll(ctx context.Context) ([]model.Basket, error)

	// FindByID retrieves a single basket by its ID.
	// We'll implement the handler for this later.
	// FindByID(ctx context.Context, id uuid.UUID) (*model.Basket, error)
}