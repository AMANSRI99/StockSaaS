package main

import (
	// Use your actual module path
	"net/http"

	"github.com/AMANSRI99/StockSaaS/internal/adapter/http/handler"
	"github.com/AMANSRI99/StockSaaS/internal/adapter/persistence/postgres"
	"github.com/AMANSRI99/StockSaaS/internal/app/service"
	"github.com/AMANSRI99/StockSaaS/internal/config"

	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	//Setting up db connection
	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()
	// --- Echo instance ---
	e := echo.New()

	// --- Middleware ---
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// --- Initializing Dependencies (Repo -> Service -> Handler) ---

	// 1. Repository
	basketRepo := postgres.NewPostgresBasketRepo(db)

	// 2. Service (inject repository)
	basketSvc := service.NewBasketService(basketRepo)

	// 3. Handler (inject service)
	basketHandler := handler.NewBasketHandler(basketSvc) // Pass service interface

	// --- Routes ---
	basketGroup := e.Group("/baskets")
	{
		basketGroup.POST("", basketHandler.CreateBasket)
		basketGroup.GET("", basketHandler.ListBaskets)
		basketGroup.GET("/:id", basketHandler.GetBasketByID)
		basketGroup.DELETE("/:id", basketHandler.DeleteBasketByID)
		basketGroup.PUT("/:id", basketHandler.UpdateBasket)
	}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Basket Trader API is running!")
	})

	// --- Start Server ---
	port := ":8080"
	log.Printf("Starting Echo server on port %s\n", port)
	e.Logger.Fatal(e.Start(port))
}
