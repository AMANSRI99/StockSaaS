package ports

import (
	"context"

	"github.com/AMANSRI99/StockSaaS/internal/domain"
)

type ZerodhaClient interface {
	PlaceTrade(ctx context.Context, trade *domain.Trade, apiKey, accessToken string) error
	CancelTrade(ctx context.Context, trade *domain.Trade, apiKey, accessToken string) error
	GetTradeStatus(ctx context.Context, tradeID string, apiKey, accessToken string) (*domain.Trade, error)
}
