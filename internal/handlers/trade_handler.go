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

// internal/middleware/auth.go
package middleware

import (
    "context"
    "net/http"
    "strings"
    "github.com/AMANSRI99/StockSaaS/internal/core/ports"
)

type AuthMiddleware struct {
    userService ports.UserService
    logger      ports.Logger
}

func NewAuthMiddleware(userService ports.UserService, logger ports.Logger) *AuthMiddleware {
    return &AuthMiddleware{
        userService: userService,
        logger:      logger,
    }
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }

        tokenParts := strings.Split(authHeader, " ")
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
            return
        }

        userID, err := m.userService.ValidateToken(r.Context(), tokenParts[1])
        if err != nil {
            m.logger.Error("Invalid token", "error", err)
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Add user ID to context
        ctx := context.WithValue(r.Context(), "userID", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}