package ports

import (
    "context"
    "github.com/AMANSRI99/StockSaaS/internal/domain"
)

type TradeService interface {
    PlaceTrade(ctx context.Context, trade *domain.Trade) error
    GetTradeHistory(ctx context.Context, userID string) ([]domain.Trade, error)
    CancelTrade(ctx context.Context, tradeID string) error
    GetTradeByID(ctx context.Context, tradeID string) (*domain.Trade, error)
}

type UserService interface {
    CreateUser(ctx context.Context, user *domain.User, password string) error
    AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error)
    GetUserByID(ctx context.Context, userID string) (*domain.User, error)
    UpdateZerodhaCredentials(ctx context.Context, userID string, apiKey, apiSecret, accessToken string) error
    ValidateToken(ctx context.Context, token string) (string, error)
}

type PortfolioService interface {
    GetPortfolio(ctx context.Context, userID string) (*domain.Portfolio, error)
    UpdatePortfolio(ctx context.Context, userID string) error
}
