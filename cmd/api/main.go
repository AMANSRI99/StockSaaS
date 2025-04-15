package main

import (
	// Use your actual module path here for internal imports
	"github.com/AMANSRI99/StockSaaS/internal/adapter/http/handler"
	"github.com/AMANSRI99/StockSaaS/internal/adapter/persistence/memory"

	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// --- Echo instance ---
	e := echo.New()

	// --- Middleware ---
	// Standard middleware: Logger, Recover (prevents crashes from panics)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// --- Initializing repository ---
	basketRepo := memory.NewInMemoryBasketRepo()

	// --- Initializing Handler ---
	// Injecting repository(as the interface type) into the handler.
	basketHandler := handler.NewBasketHandler(basketRepo)

	// --- Routes ---
	basketGroup := e.Group("/baskets")
	{
		basketGroup.POST("", basketHandler.CreateBasket) // POST /baskets
		basketGroup.GET("", basketHandler.ListBaskets)   // GET /baskets
		// We'll add GET /baskets/:id, PUT /baskets/:id, DELETE /baskets/:id later
	}

	// Simple health check endpoint
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Basket Trader API is running!")
	})

	// --- Start Server ---
	port := ":8080"
	log.Printf("Starting Echo server on port %s\n", port)

	// Start the server using Echo's built-in logger for fatal errors
	e.Logger.Fatal(e.Start(port))
}
