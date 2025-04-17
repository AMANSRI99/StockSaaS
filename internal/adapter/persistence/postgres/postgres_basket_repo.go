package postgres

import (
	"context"
	"database/sql"
	"errors" // Required for sql.ErrNoRows comparison
	"fmt"
	"log"

	// Use your actual module path
	"github.com/AMANSRI99/StockSaaS/internal/app/model"
	"github.com/AMANSRI99/StockSaaS/internal/app/repository"

	"github.com/google/uuid"
)

// PostgresBasketRepo implements repository.BasketRepository using PostgreSQL.
type PostgresBasketRepo struct {
	db *sql.DB // Database connection pool
}

// NewPostgresBasketRepo creates a new repository instance.
func NewPostgresBasketRepo(db *sql.DB) repository.BasketRepository {
	return &PostgresBasketRepo{db: db}
}

// Save inserts a new basket and its items within a transaction.
func (r *PostgresBasketRepo) Save(ctx context.Context, basket *model.Basket) error {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil) // Use default transaction options
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Defer rollback in case of error, commit if no error occurs
	defer func() {
		if err != nil {
			log.Printf("Rolling back transaction due to error: %v", err)
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Error rolling back transaction: %v", rbErr)
			}
		}
	}() // Note the final () to call the deferred function

	// 1. Insert into baskets table
	basketQuery := `INSERT INTO baskets (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, basketQuery, basket.ID, basket.Name, basket.CreatedAt, basket.CreatedAt) // Use CreatedAt for UpdatedAt initially
	if err != nil {
		// Check for potential unique constraint violation or other errors
		return fmt.Errorf("failed to insert basket: %w", err)
	}

	// 2. Insert into basket_items table
	itemQuery := `INSERT INTO basket_items (basket_id, symbol, quantity) VALUES ($1, $2, $3)`
	for _, stock := range basket.Stocks {
		_, err = tx.ExecContext(ctx, itemQuery, basket.ID, stock.Symbol, stock.Quantity)
		if err != nil {
			// Check for potential unique constraint violation (basket_id, symbol)
			return fmt.Errorf("failed to insert basket item %s for basket %s: %w", stock.Symbol, basket.ID, err)
		}
	}

	// 3. Commit the transaction if all inserts were successful
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil // Success
}

// FindAll retrieves all baskets and their associated items.
// NOTE: This uses a simple N+1 query approach. Optimize later if needed.
func (r *PostgresBasketRepo) FindAll(ctx context.Context) ([]model.Basket, error) {
	queryBaskets := `SELECT id, name, created_at, updated_at FROM baskets ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, queryBaskets)
	if err != nil {
		return nil, fmt.Errorf("failed to query baskets: %w", err)
	}
	defer rows.Close()

	basketsMap := make(map[uuid.UUID]*model.Basket) // Temp map to hold baskets while fetching items
	basketOrder := []uuid.UUID{}                    // Keep track of order

	for rows.Next() {
		var b model.Basket
		if err := rows.Scan(&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan basket row: %w", err)
		}
		b.Stocks = []model.Stock{} // Initialize empty slice
		basketsMap[b.ID] = &b
		basketOrder = append(basketOrder, b.ID)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating basket rows: %w", err)
	}

	if len(basketsMap) == 0 {
		return []model.Basket{}, nil // Return empty slice if no baskets found
	}

	// Now fetch items for all found baskets (This is the N+1 part, one query per basket)
	// Optimization needed for many baskets (e.g., use JOIN or WHERE basket_id IN (...))
	queryItems := `SELECT basket_id, symbol, quantity FROM basket_items WHERE basket_id = $1`
	for basketID := range basketsMap {
		itemRows, err := r.db.QueryContext(ctx, queryItems, basketID)
		if err != nil {
			// Log error but potentially continue to fetch for other baskets? Or fail all?
			log.Printf("Warning: failed to query items for basket %s: %v", basketID, err)
			continue // Skip items for this basket on error
		}

		for itemRows.Next() {
			var item model.Stock
			var bID uuid.UUID // Need to scan basket_id to map back, though we know it here
			if err := itemRows.Scan(&bID, &item.Symbol, &item.Quantity); err != nil {
				itemRows.Close() // Close inner rows on error
				return nil, fmt.Errorf("failed to scan basket item row for basket %s: %w", basketID, err)
			}
			// Append item to the correct basket in the map
			if basket, ok := basketsMap[bID]; ok {
				basket.Stocks = append(basket.Stocks, item)
			}
		}
		if err := itemRows.Err(); err != nil {
			itemRows.Close()
			return nil, fmt.Errorf("error iterating item rows for basket %s: %w", basketID, err)
		}
		itemRows.Close() // IMPORTANT: Close inner rows loop
	}

	// Convert map back to slice respecting original order
	result := make([]model.Basket, len(basketOrder))
	for i, id := range basketOrder {
		result[i] = *basketsMap[id]
	}

	return result, nil
}

// FindByID retrieves a single basket and its items.
// NOTE: Also uses N+1 approach for items.
func (r *PostgresBasketRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Basket, error) {
	queryBasket := `SELECT id, name, created_at, updated_at FROM baskets WHERE id = $1`
	row := r.db.QueryRowContext(ctx, queryBasket, id)

	var b model.Basket
	err := row.Scan(&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrBasketNotFound // Use the custom error
		}
		return nil, fmt.Errorf("failed to query or scan basket by ID %s: %w", id, err)
	}

	// Fetch items for this basket
	queryItems := `SELECT symbol, quantity FROM basket_items WHERE basket_id = $1`
	itemRows, err := r.db.QueryContext(ctx, queryItems, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query items for basket %s: %w", id, err)
	}
	defer itemRows.Close()

	b.Stocks = []model.Stock{} // Initialize empty slice
	for itemRows.Next() {
		var item model.Stock
		if err := itemRows.Scan(&item.Symbol, &item.Quantity); err != nil {
			return nil, fmt.Errorf("failed to scan basket item row for basket %s: %w", id, err)
		}
		b.Stocks = append(b.Stocks, item)
	}
	if err = itemRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating item rows for basket %s: %w", id, err)
	}

	return &b, nil
}
