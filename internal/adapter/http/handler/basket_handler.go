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
	mw "github.com/AMANSRI99/StockSaaS/internal/adapter/http/middleware"
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

// Helper function to get userID from context
func getUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	userIDCtx := c.Get(string(mw.UserIDContextKey)) // Use the key defined in middleware
	userID, ok := userIDCtx.(uuid.UUID)
	if !ok {
		log.Printf("Handler: Failed to get user ID from context or type assertion failed")
		// This indicates a programming error (middleware not run or wrong type set)
		return uuid.Nil, echo.NewHTTPError(http.StatusInternalServerError, "Could not identify user from context")
	}
	return userID, nil
}

// CreateBasket handles POST requests - delegates to the service.
func (h *BasketHandler) CreateBasket(c echo.Context) error {

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

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
	log.Printf("Handler: Calling CreateBasket service for user %s", userID)
	// Pass userID to service
	createdBasket, err := h.service.CreateBasket(ctx, req.Name, req.Stocks, userID)
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

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	// --- Delegate to the service ---
	ctx := c.Request().Context()
	log.Printf("Handler: Calling ListAllBaskets service for user %s", userID)

	allBaskets, err := h.service.ListAllBaskets(ctx, userID)
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
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}
	// 1. Parse and Validate ID from path parameter
	idStr := c.Param("id")
	basketID, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("Handler: Invalid UUID format for ID '%s': %v", idStr, err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid basket ID format: %s", idStr))
	}

	// 2. Call the Service
	ctx := c.Request().Context()
	log.Printf("Handler: Calling GetBasketByID service for user %s, basket %s", userID, basketID)
	basket, err := h.service.GetBasketByID(ctx, basketID, userID)
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

// DeleteBasketByID handles DELETE requests to /baskets/:id
func (h *BasketHandler) DeleteBasketByID(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}
	// 1. Parse and Validate ID from path parameter
	idStr := c.Param("id")
	basketID, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("Handler: Invalid UUID format for ID '%s': %v", idStr, err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid basket ID format: %s", idStr))
	}

	// 2. Call the Service
	ctx := c.Request().Context()
	log.Printf("Handler: Calling DeleteBasketByID service for user %s, basket %s", userID, basketID)

	err = h.service.DeleteBasketByID(ctx, basketID, userID)
	log.Printf("Handler: Calling DeleteBasketByID service for user %s, basket %s", userID, basketID)
	// No basket returned on delete
	if err != nil {
		log.Printf("Handler: Error from DeleteBasketByID service for ID %s: %v", basketID, err)
		// 3. Handle specific errors (like NotFound)
		if errors.Is(err, repository.ErrBasketNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Basket with ID %s not found", basketID))
		}
		// Handle other potential errors
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to delete basket %s: %v", basketID, err))
	}

	// 4. Return Success Response (No Content)
	log.Printf("Handler: Successfully deleted basket ID %s", basketID)
	return c.NoContent(http.StatusNoContent) // Use 204 No Content for successful DELETE
}

// UpdateBasket handles PUT requests to /baskets/:id
func (h *BasketHandler) UpdateBasket(c echo.Context) error {

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}
	// 1. Parse and Validate ID
	idStr := c.Param("id")
	basketID, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("Handler: Invalid UUID format for ID '%s': %v", idStr, err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid basket ID format: %s", idStr))
	}

	// 2. Bind Request Body
	// Use a specific struct for update request validation/binding
	type updateBasketRequest struct {
		Name   string        `json:"name"`
		Stocks []model.Stock `json:"stocks"` // Expects full list of stocks for replacement
	}
	req := new(updateBasketRequest)
	if err := c.Bind(req); err != nil {
		log.Printf("Handler: Error binding update basket request for ID %s: %v", basketID, err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	// 3. Basic Request Body Validation (Service layer might do more detailed business logic validation)
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Basket name is required")
	}
	// Allow empty stocks array for PUT (means delete all items)
	for i, stock := range req.Stocks {
		if stock.Symbol == "" || stock.Quantity <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid data for stock #%d: symbol and positive quantity required", i+1))
		}
	}

	// 4. Call the Service
	ctx := c.Request().Context()
	log.Printf("Handler: Calling UpdateBasket service for user %s, basket %s", userID, basketID)
	updatedBasket, err := h.service.UpdateBasket(ctx, basketID, req.Name, req.Stocks, userID)
	if err != nil {
		log.Printf("Handler: Error from UpdateBasket service for ID %s: %v", basketID, err)
		// 5. Handle specific errors
		if errors.Is(err, repository.ErrBasketNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Basket with ID %s not found", basketID))
		}
		// Check for potential validation errors from service (though basic ones handled above)
		// if errors.Is(err, service.ErrValidation) { return echo.NewHTTPError(http.StatusBadRequest, err.Error()) }

		// Handle other potential errors
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to update basket %s: %v", basketID, err))
	}

	// 6. Return Success Response
	log.Printf("Handler: Successfully updated basket ID %s", basketID)
	return c.JSON(http.StatusOK, updatedBasket) // Return the updated basket details
}

// Define validateCreateRequest, validateUpdateRequest helpers if needed
type createBasketRequest struct {
	Name   string        `json:"name"`
	Stocks []model.Stock `json:"stocks"`
}
type updateBasketRequest struct {
	Name   string        `json:"name"`
	Stocks []model.Stock `json:"stocks"`
}
