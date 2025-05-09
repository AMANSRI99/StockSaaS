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
func (r *PostgresBasketRepo) Save(ctx context.Context, basket *model.Basket, userID uuid.UUID) error {
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
	basketQuery := `INSERT INTO baskets (id, user_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.ExecContext(ctx, basketQuery, basket.ID, userID, basket.Name, basket.CreatedAt, basket.CreatedAt)
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
func (r *PostgresBasketRepo) FindAll(ctx context.Context, userID uuid.UUID) ([]model.Basket, error) {
	queryBaskets := `SELECT id, name, created_at, updated_at FROM baskets WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, queryBaskets, userID)
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

	// Fetch items only for the user's baskets found
	queryItems := `SELECT basket_id, symbol, quantity FROM basket_items WHERE basket_id = $1` // Still fetch by basket_id
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
func (r *PostgresBasketRepo) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Basket, error) {
	queryBasket := `SELECT id, name, created_at, updated_at FROM baskets WHERE id = $1 AND user_id = $2`
	row := r.db.QueryRowContext(ctx, queryBasket, id, userID)

	var b model.Basket
	err := row.Scan(&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrBasketNotFound // Use the custom error
		}
		return nil, fmt.Errorf("failed to query/scan basket %s for user %s: %w", id, userID, err)
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

// DeleteByID removes a basket from the database by its ID.
// It returns ErrBasketNotFound if no rows are affected.
func (r *PostgresBasketRepo) DeleteByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM baskets WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		// Handle potential database errors (e.g., connection issues)
		return fmt.Errorf("failed to execute delete for basket %s, user %s: %w", id, userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This might happen in rare cases, good to log
		log.Printf("Warning: could not get rows affected after delete for basket ID %s: %v", id, err)
		return fmt.Errorf("failed to check rows affected for basket %s, user %s: %w", id, userID, err)
	}

	if rowsAffected == 0 {
		// No rows were deleted, meaning the basket ID didn't exist
		return repository.ErrBasketNotFound
	}

	// If rowsAffected is 1 (or potentially more if ID wasn't unique, but it's PK), success
	return nil // Success
}

// Update modifies an existing basket and its items within a transaction.
// It assumes basket.ID is set and basket contains the new Name and Stocks.
// Returns ErrBasketNotFound if the basket ID doesn't exist.
func (r *PostgresBasketRepo) Update(ctx context.Context, basket *model.Basket, userID uuid.UUID) (err error) { // Use named return for easier defer rollback
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Defer rollback if any error occurs; commit otherwise
	defer func() {
		if p := recover(); p != nil { // Catch potential panics
			log.Printf("Update recovered panic: %v", p)
			rbErr := tx.Rollback()
			if rbErr != nil {
				log.Printf("Error rolling back transaction after panic: %v", rbErr)
			}
			// Re-panic if desired, or set error
			err = fmt.Errorf("internal error during update (panic recovered)")
		} else if err != nil {
			// Explicit error occurred, rollback
			log.Printf("Rolling back transaction for basket %s due to error: %v", basket.ID, err)
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Error rolling back transaction: %v", rbErr)
				// Combine errors? Or just log? Prioritize original error.
				err = fmt.Errorf("original error: %w; rollback error: %v", err, rbErr)
			}
		} else {
			// No error, commit
			err = tx.Commit()
			if err != nil {
				err = fmt.Errorf("failed to commit transaction: %w", err)
			}
		}
	}() // Execute the deferred function

	// 1. Update the baskets table (name and updated_at via trigger)
	// Note: We rely on the DB trigger to update `updated_at`.
	updateBasketQuery := `UPDATE baskets SET name = $1 WHERE id = $2 AND user_id = $3`
	result, err := tx.ExecContext(ctx, updateBasketQuery, basket.Name, basket.ID, userID)
	if err != nil {
		return fmt.Errorf("failed to update basket %s for user %s: %w", basket.ID, userID, err)
	}
	rowsAffected, _ := result.RowsAffected() // Error checking for RowsAffected done previously
	if rowsAffected == 0 {
		return repository.ErrBasketNotFound // Basket ID didn't exist
	}

	// 2. Delete old items associated with this basket
	deleteItemsQuery := `DELETE FROM basket_items WHERE basket_id = $1`
	_, err = tx.ExecContext(ctx, deleteItemsQuery, basket.ID)
	if err != nil {
		return fmt.Errorf("failed to delete old items for basket %s: %w", basket.ID, err)
	}

	// 3. Insert new items
	// Ensure basket.Stocks is not nil before ranging
	if len(basket.Stocks) > 0 {
		insertItemQuery := `INSERT INTO basket_items (basket_id, symbol, quantity) VALUES ($1, $2, $3)`
		for _, stock := range basket.Stocks {
			_, err = tx.ExecContext(ctx, insertItemQuery, basket.ID, stock.Symbol, stock.Quantity)
			if err != nil {
				// Handle potential errors like constraint violations
				return fmt.Errorf("failed to insert new item %s for basket %s: %w", stock.Symbol, basket.ID, err)
			}
		}
	} // else: if basket.Stocks is empty, we just deleted all old items, which is correct for PUT replace.

	// If we reach here without error, the deferred function will commit.
	return nil // Success (commit happens in defer)
}
