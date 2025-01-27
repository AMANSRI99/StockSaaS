package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/AMANSRI99/StockSaaS/internal/core/ports"
	"github.com/AMANSRI99/StockSaaS/internal/domain"
)

type TradeHandler struct {
	tradeService ports.TradeService
	logger       ports.Logger
}

func NewTradeHandler(tradeService ports.TradeService, logger ports.Logger) *TradeHandler {
	return &TradeHandler{
		tradeService: tradeService,
		logger:       logger,
	}
}

func (h *TradeHandler) PlaceTrade(w http.ResponseWriter, r *http.Request) {
	var trade domain.Trade
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(string)
	trade.UserID = userID

	if err := h.tradeService.PlaceTrade(r.Context(), &trade); err != nil {
		h.logger.Error("Failed to place trade", "error", err)
		http.Error(w, "Failed to place trade", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(trade)
}
