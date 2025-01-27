package services

import (
	"context"

	"github.com/AMANSRI99/StockSaaS/internal/core/ports"
	"github.com/AMANSRI99/StockSaaS/internal/domain"
)

type tradeService struct {
	tradeRepo     ports.TradeRepository
	zerodhaClient ports.ZerodhaClient
	logger        ports.Logger
}

func NewTradeService(
	tradeRepo ports.TradeRepository,
	zerodhaClient ports.ZerodhaClient,
	logger ports.Logger,
) ports.TradeService {
	return &tradeService{
		tradeRepo:     tradeRepo,
		zerodhaClient: zerodhaClient,
		logger:        logger,
	}
}

func (s *tradeService) PlaceTrade(ctx context.Context, trade *domain.Trade) error {
	// Validate trade
	if err := s.validateTrade(trade); err != nil {
		return err
	}

	// Place trade with Zerodha
	if err := s.zerodhaClient.PlaceTrade(ctx, trade); err != nil {
		s.logger.Error("Failed to place trade with Zerodha", "error", err)
		return err
	}

	// Store trade in database
	if err := s.tradeRepo.Create(ctx, trade); err != nil {
		s.logger.Error("Failed to store trade", "error", err)
		return err
	}

	return nil
}

func (s *tradeService) validateTrade(trade *domain.Trade) error {
	// Implement trade validation logic
	return nil
}
