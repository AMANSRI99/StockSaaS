package services

import (
	"context"
	"time"

	"github.com/AMANSRI99/StockSaaS/internal/core/ports"
	"github.com/AMANSRI99/StockSaaS/internal/domain"
	"github.com/google/uuid"
)

type tradeService struct {
    tradeRepo    ports.TradeRepository
    userRepo     ports.UserRepository
    zerodhaClient ports.ZerodhaClient
    logger       ports.Logger
}

func NewTradeService(
    tradeRepo ports.TradeRepository,
    userRepo ports.UserRepository,
    zerodhaClient ports.ZerodhaClient,
    logger ports.Logger,
) ports.TradeService {
    return &tradeService{
        tradeRepo:     tradeRepo,
        userRepo:      userRepo,
        zerodhaClient: zerodhaClient,
        logger:        logger,
    }
}

func (s *tradeService) PlaceTrade(ctx context.Context, trade *domain.Trade) error {
    if err := trade.Validate(); err != nil {
        return err
    }

    // Get user's Zerodha credentials
    user, err := s.userRepo.GetByID(ctx, trade.UserID)
    if err != nil {
        return err
    }

    // Initialize trade
    trade.ID = uuid.New().String()
    trade.Status = domain.OrderStatusPending
    trade.CreatedAt = time.Now()
    trade.UpdatedAt = time.Now()

    // Store trade in database
    if err := s.tradeRepo.Create(ctx, trade); err != nil {
        return err
    }

    // Place trade with Zerodha
    go func() {
        err := s.zerodhaClient.PlaceTrade(ctx, trade, user.ZerodhaAPIKey, user.ZerodhaAccessToken)
        if err != nil {
            trade.Status = domain.OrderStatusFailed
            trade.ErrorMessage = err.Error()
        } else {
            trade.Status = domain.OrderStatusComplete
        }
        trade.UpdatedAt = time.Now()
        s.tradeRepo.Update(ctx, trade)
    }()

    return nil
}

func (s *tradeService) GetTradeHistory(ctx context.Context, userID string) ([]domain.Trade, error) {
    return s.tradeRepo.GetByUserID(ctx, userID)
}

func (s *tradeService) CancelTrade(ctx context.Context, tradeID string) error {
    trade, err := s.tradeRepo.GetByID(ctx, tradeID)
    if err != nil {
        return err
    }

    if trade.Status != domain.OrderStatusPending {
        return domain.ErrTradeCannotBeCanceled
    }

    user, err := s.userRepo.GetByID(ctx, trade.UserID)
    if err != nil {
        return err
    }

    if err := s.zerodhaClient.CancelTrade(ctx, trade, user.ZerodhaAPIKey, user.ZerodhaAccessToken); err != nil {
        return err
    }

    trade.Status = domain.OrderStatusCanceled
    trade.UpdatedAt = time.Now()
    return s.tradeRepo.Update(ctx, trade)
}

func (s *tradeService) GetTradeByID(ctx context.Context, tradeID string) (*domain.Trade, error) {
    return s.tradeRepo.GetByID(ctx, tradeID)
}