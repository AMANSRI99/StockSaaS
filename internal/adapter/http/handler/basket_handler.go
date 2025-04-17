package handler

import (
	// Use your actual module path
	"errors"

	"github.com/AMANSRI99/StockSaaS/internal/app/model"
	"github.com/AMANSRI99/StockSaaS/internal/app/repository"
	"github.com/google/uuid"

	// "github.com/AMANSRI99/StockSaaS/internal/app/repository" // No longer needed here
	"github.com/AMANSRI99/StockSaaS/internal/app/service" // Import service interface

	"fmt"
	"log"
	"net/http"

	// "time" // No longer needed directly here

	// "github.com/google/uuid" // No longer needed directly here
	"github.com/labstack/echo/v4"
)

// BasketHandler now holds a service interface.
type BasketHandler struct {
	service service.BasketService // Use the service interface type
}

// NewBasketHandler accepts the service interface.
func NewBasketHandler(svc service.BasketService) *BasketHandler {
	return &BasketHandler{
		service: svc,
	}
}

// CreateBasket handles POST requests - delegates to the service.
func (h *BasketHandler) CreateBasket(c echo.Context) error {
	// DTO (Data Transfer Object) for the request binding
	type createBasketRequest struct {
		Name   string        `json:"name"`
		Stocks []model.Stock `json:"stocks"` // Keep using model.Stock for input for now
	}

	req := new(createBasketRequest)
	if err := c.Bind(req); err != nil {
		log.Printf("Handler: Error binding basket: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// --- Delegate to the service ---
	ctx := c.Request().Context()
	// Pass the relevant data from the request to the service method
	createdBasket, err := h.service.CreateBasket(ctx, req.Name, req.Stocks)
	if err != nil {
		log.Printf("Handler: Error calling CreateBasket service: %v", err)
		// Map service errors to HTTP errors (could be more sophisticated)
		// For now, assume most service errors are internal server errors or bad requests if validation fails
		// We might need specific error types from the service later.
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Could not create basket: %v", err))
	}

	log.Printf("Handler: Successfully created basket ID %s via service", createdBasket.ID)
	// Return the basket returned by the service
	return c.JSON(http.StatusCreated, createdBasket)
}

// ListBaskets handles GET requests - delegates to the service.
func (h *BasketHandler) ListBaskets(c echo.Context) error {
	// --- Delegate to the service ---
	ctx := c.Request().Context()
	allBaskets, err := h.service.ListAllBaskets(ctx)
	if err != nil {
		log.Printf("Handler: Error calling ListAllBaskets service: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Could not retrieve baskets: %v", err))
	}

	// Service ensures we get []model.Basket{}, not nil

	log.Printf("Handler: Returning %d baskets from service", len(allBaskets))
	return c.JSON(http.StatusOK, allBaskets)
}

// GetBasketByID handles GET requests to /baskets/:id
func (h *BasketHandler) GetBasketByID(c echo.Context) error {
	// 1. Parse and Validate ID from path parameter
	idStr := c.Param("id")
	basketID, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("Handler: Invalid UUID format for ID '%s': %v", idStr, err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid basket ID format: %s", idStr))
	}

	// 2. Call the Service
	ctx := c.Request().Context()
	log.Printf("Handler: Calling GetBasketByID service for ID %s", basketID)
	basket, err := h.service.GetBasketByID(ctx, basketID)
	if err != nil {
		log.Printf("Handler: Error from GetBasketByID service for ID %s: %v", basketID, err)
		// 3. Handle specific errors (like NotFound)
		if errors.Is(err, repository.ErrBasketNotFound) { // Check for the specific error
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Basket with ID %s not found", basketID))
		}
		// Handle other potential errors
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve basket %s: %v", basketID, err))
	}

	// 4. Return Success Response
	log.Printf("Handler: Returning basket ID %s", basketID)
	return c.JSON(http.StatusOK, basket)
}
