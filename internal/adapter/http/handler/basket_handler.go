package handler

import (
	"fmt"

	"github.com/AMANSRI99/StockSaaS/internal/app/model"
	"github.com/AMANSRI99/StockSaaS/internal/app/repository"

	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// BasketHandler holds dependencies for basket endpoints
type BasketHandler struct {
	repo repository.BasketRepository
}

// NewBasketHandler creates a new handler instance (adjust map type)
func NewBasketHandler(repo repository.BasketRepository) *BasketHandler {
	return &BasketHandler{
		repo: repo,
	}
}

// CreateBasket handles POST requests to /baskets
func (h *BasketHandler) CreateBasket(c echo.Context) error {
	// Use a struct for the request body to separate concerns slightly
	type createBasketRequest struct {
		Name   string        `json:"name"`
		Stocks []model.Stock `json:"stocks"`
	}

	req := new(createBasketRequest)

	// Bind the request body to the struct
	if err := c.Bind(req); err != nil {
		log.Printf("Error binding basket: %v", err)
		// Use Echo's HTTPError for standard error responses
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Basic validation (can use Echo's validator later)
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Basket name is required")
	}
	if len(req.Stocks) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Basket must contain at least one stock")
	}
	for i, stock := range req.Stocks {
		if stock.Symbol == "" || stock.Quantity <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid data for stock #%d: symbol and positive quantity required", i+1))
		}
	}

	// Create the new basket model
	newBasket := model.Basket{
		ID:        uuid.New(), // Generate UUID
		Name:      req.Name,
		Stocks:    req.Stocks,
		CreatedAt: time.Now().UTC(),
	}

	//use the repository to save.
	//Get context from requet.
	ctx := c.Request().Context()
	err := h.repo.Save(ctx, &newBasket)
	if err != nil {
		log.Printf("Error saving basket: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not save basket")
	}
	log.Printf("Created basket: ID=%s, Name=%s\n", newBasket.ID, newBasket.Name)

	// Respond with the created basket using Echo's JSON helper
	return c.JSON(http.StatusCreated, newBasket)
}

// ListBaskets handles GET requests to /baskets
// Signature now uses echo.Context and returns error
func (h *BasketHandler) ListBaskets(c echo.Context) error {
	ctx := c.Request().Context()
	allBaskets, err := h.repo.FindAll(ctx)

	if err != nil {
		log.Printf("Error finding all the baskets: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "could not retrieve baskets")
	}

	//if there are no baskets yet.
	if allBaskets == nil {
		allBaskets = []model.Basket{}
	}
	// Use Echo's JSON helper for the response
	return c.JSON(http.StatusOK, allBaskets)
}
