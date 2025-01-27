package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AMANSRI99/StockSaaS/internal/domain"
)

type tradeRepository struct {
	db *sql.DB
}

func NewTradeRepository(db *sql.DB) *tradeRepository {
	return &tradeRepository{db: db}
}

func (r *tradeRepository) Create(ctx context.Context, trade *domain.Trade) error {
	query := `
        INSERT INTO trades (
            id, user_id, symbol, quantity, order_type,
            price, trade_type, status, error_message,
            created_at, updated_at, executed_price
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `

	_, err := r.db.ExecContext(ctx, query,
		trade.ID,
		trade.UserID,
		trade.Symbol,
		trade.Quantity,
		trade.OrderType,
		trade.Price,
		trade.TradeType,
		trade.Status,
		trade.ErrorMessage,
		trade.CreatedAt,
		trade.UpdatedAt,
		trade.ExecutedPrice,
	)

	if err != nil {
		return fmt.Errorf("error creating trade: %w", err)
	}

	return nil
}

func (r *tradeRepository) GetByID(ctx context.Context, id string) (*domain.Trade, error) {
	query := `
        SELECT 
            id, user_id, symbol, quantity, order_type,
            price, trade_type, status, error_message,
            created_at, updated_at, executed_price
        FROM trades 
        WHERE id = $1
    `

	trade := &domain.Trade{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&trade.ID,
		&trade.UserID,
		&trade.Symbol,
		&trade.Quantity,
		&trade.OrderType,
		&trade.Price,
		&trade.TradeType,
		&trade.Status,
		&trade.ErrorMessage,
		&trade.CreatedAt,
		&trade.UpdatedAt,
		&trade.ExecutedPrice,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("trade not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting trade: %w", err)
	}

	return trade, nil
}

func (r *tradeRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Trade, error) {
	query := `
        SELECT 
            id, user_id, symbol, quantity, order_type,
            price, trade_type, status, error_message,
            created_at, updated_at, executed_price
        FROM trades 
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying trades: %w", err)
	}
	defer rows.Close()

	var trades []domain.Trade
	for rows.Next() {
		var trade domain.Trade
		err := rows.Scan(
			&trade.ID,
			&trade.UserID,
			&trade.Symbol,
			&trade.Quantity,
			&trade.OrderType,
			&trade.Price,
			&trade.TradeType,
			&trade.Status,
			&trade.ErrorMessage,
			&trade.CreatedAt,
			&trade.UpdatedAt,
			&trade.ExecutedPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning trade: %w", err)
		}
		trades = append(trades, trade)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trades: %w", err)
	}

	return trades, nil
}

func (r *tradeRepository) Update(ctx context.Context, trade *domain.Trade) error {
	query := `
        UPDATE trades 
        SET 
            status = $1,
            error_message = $2,
            updated_at = $3,
            executed_price = $4
        WHERE id = $5
    `

	result, err := r.db.ExecContext(ctx, query,
		trade.Status,
		trade.ErrorMessage,
		time.Now(),
		trade.ExecutedPrice,
		trade.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating trade: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("trade not found")
	}

	return nil
}
