package main

import (
	// Use your actual module path
	"net/http"

	kiteAdapter "github.com/AMANSRI99/StockSaaS/internal/adapter/broker/kiteconnect"
	"github.com/AMANSRI99/StockSaaS/internal/adapter/http/handler"
	httpMw "github.com/AMANSRI99/StockSaaS/internal/adapter/http/middleware"
	"github.com/AMANSRI99/StockSaaS/internal/adapter/persistence/postgres"
	"github.com/AMANSRI99/StockSaaS/internal/app/service"
	"github.com/AMANSRI99/StockSaaS/internal/config"

	"log"

	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

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

	e := echo.New()
	e.Use(echoMw.Logger())
	e.Use(echoMw.Recover())

	// --- Initialize Repositories ---
	basketRepo := postgres.NewPostgresBasketRepo(db)
	userRepo := postgres.NewPostgresUserRepo(db)
	brokerRepo := postgres.NewPostgresBrokerRepo(db, cfg.EncryptionKey)

	kiteAdpt := kiteAdapter.NewAdapter(cfg.Kite.APIKey)
	// --- Initialize Services ---
	basketSvc := service.NewBasketService(basketRepo)
	userSvc := service.NewUserService(userRepo, *cfg)
	kiteSvc := service.NewKiteService(kiteAdpt, brokerRepo, *cfg)

	// --- Initialize Handlers ---
	basketHandler := handler.NewBasketHandler(basketSvc) // Pass basket service
	authHandler := handler.NewAuthHandler(userSvc)       // <-- Instantiate Auth Handler
	kiteHandler := handler.NewKiteHandler(kiteAdpt, kiteSvc, *cfg)

	//Initialising auth middleware
	authMiddleware := httpMw.NewJWTAuthMiddleware(cfg.JWT.SecretKey)
	// --- Routes ---
	// Group API routes (good practice)
	apiGroup := e.Group("/api")
	{
		// Auth routes (no auth middleware needed)
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.POST("/signup", authHandler.Signup) // <-- Register Signup Route
			authGroup.POST("/login", authHandler.Login)
		}

		kiteGroup := apiGroup.Group("/kite", authMiddleware) // Group for authenticated kite actions
		{
			// Endpoint to start the connection flow
			kiteGroup.GET("/connect/initiate", kiteHandler.InitiateKiteConnect, authMiddleware)
			// Callback does NOT need user logged in via JWT (comes from external redirect)
			kiteGroup.GET("/connect/callback", kiteHandler.HandleKiteCallback) // <-- Register Callback Route

		}

		// Basket routes (will add auth middleware later)
		basketGroup := apiGroup.Group("/baskets", authMiddleware)
		{
			basketGroup.POST("", basketHandler.CreateBasket)
			basketGroup.GET("", basketHandler.ListBaskets)
			basketGroup.GET("/:id", basketHandler.GetBasketByID)
			basketGroup.DELETE("/:id", basketHandler.DeleteBasketByID)
			basketGroup.PUT("/:id", basketHandler.UpdateBasket)
		}
	}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Basket Trader API is running!")
	})

	serverPort := ":" + cfg.ServerPort
	log.Printf("Starting Echo server on port %s\n", serverPort)
	e.Logger.Fatal(e.Start(serverPort))
}
