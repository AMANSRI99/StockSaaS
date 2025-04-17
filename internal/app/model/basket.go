package model

import (
	"time"

	"github.com/google/uuid"
)

// Basket represents a collection of stocks.
type Basket struct {
	ID        uuid.UUID `json:"id"`        // Use UUID for unique identifier
	Name      string    `json:"name"`      // User-defined name for the basket
	Stocks    []Stock   `json:"stocks"`    // List of stocks in the basket
	CreatedAt time.Time `json:"createdAt"` // Keep track of creation time
	UpdatedAt time.Time `json:"updatedAt"`
}

// Add a helper function maybe? (optional)
func NewBasket(name string, stocks []Stock) *Basket {
	now := time.Now().UTC()
	return &Basket{
		ID:        uuid.New(), // Generate UUID on creation
		Name:      name,
		Stocks:    stocks,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
