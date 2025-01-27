// cmd/api/main.go
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/AMANSRI99/StockSaaS/internal/core/services"
	"github.com/AMANSRI99/StockSaaS/internal/handlers"
	"github.com/AMANSRI99/StockSaaS/internal/repository/postgres"
	"github.com/AMANSRI99/StockSaaS/internal/zerodha"
	"github.com/AMANSRI99/StockSaaS/pkg/logger"
	"github.com/gorilla/mux"
)

func initLogger() logger.Logger {
	// Get environment from ENV, default to development
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	return logger.NewLogger(env)
}

func main() {
	// Initialize logger
	log := initLogger()
	// Initialize database
	db, err := initDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database", "error", err)
	}
	defer db.Close()

	// Initialize repositories
	tradeRepo := postgres.NewTradeRepository(db)
	userRepo := postgres.NewUserRepository(db)

	// Initialize Zerodha client
	zerodhaClient := zerodha.NewZerodhaClient()

	// Initialize services
	tradeService := services.NewTradeService(tradeRepo, userRepo, zerodhaClient, log)
	userService := services.NewUserService(userRepo, log)

	// Initialize handlers
	tradeHandler := handlers.NewTradeHandler(tradeService, log)

	// Initialize router
	router := mux.NewRouter()

	// Setup API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Setup middleware
	authMiddleware := initAuthMiddleware(userService, log)

	// Setup routes
	handlers.SetupTradeRoutes(apiRouter, tradeHandler, authMiddleware)

	// Setup server
	srv := &http.Server{
		Handler:      router,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Create shutdown context with 15 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shutdown server
	srv.Shutdown(ctx)
	logger.Info("Server shutdown complete")
}

// Available API Endpoints:
// POST   /api/v1/trades           - Place a new trade
// GET    /api/v1/trades           - Get trade history
// GET    /api/v1/trades/{id}      - Get specific trade
// POST   /api/v1/trades/{id}/cancel - Cancel a trade

// Example trade request:
/*
{
    "symbol": "RELIANCE",
    "quantity": 1,
    "order_type": "MARKET",
    "trade_type": "BUY",
    "price": 2500.00  // Optional, required for LIMIT orders
}
*/
