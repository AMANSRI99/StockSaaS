package ports

import (
	"context"

	"github.com/AMANSRI99/StockSaaS/internal/domain"
)

type TradeService interface {
	PlaceTrade(ctx context.Context, trade *domain.Trade) error
	GetTradeHistory(ctx context.Context, userID string) ([]domain.Trade, error)
	CancelTrade(ctx context.Context, tradeID string) error
}
