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

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

type PortfolioRepository interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Portfolio, error)
	Update(ctx context.Context, portfolio *domain.Portfolio) error
}
