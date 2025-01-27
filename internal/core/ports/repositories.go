package ports

import (
	"context"

	"github.com/AMANSRI99/StockSaaS/internal/domain"
)

type TradeRepository interface {
	Create(ctx context.Context, trade *domain.Trade) error
	GetByID(ctx context.Context, id string) (*domain.Trade, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.Trade, error)
	Update(ctx context.Context, trade *domain.Trade) error
}
