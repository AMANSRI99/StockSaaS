package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	// Use your actual module path
	"github.com/AMANSRI99/StockSaaS/internal/app/repository" // Need repository errors
	"github.com/AMANSRI99/StockSaaS/internal/app/service"

	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication related endpoints.
type AuthHandler struct {
	userService service.UserService
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(userSvc service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userSvc,
	}
}

// Signup handles user registration requests.
func (h *AuthHandler) Signup(c echo.Context) error {
	// Define request structure for binding
	type signupRequest struct {
		Email    string `json:"email" validate:"required,email"`    // Add validation tags later
		Password string `json:"password" validate:"required,min=8"` // Add validation tags later
	}

	req := new(signupRequest)

	// Bind request body
	if err := c.Bind(req); err != nil {
		log.Printf("Handler: Error binding signup request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Basic validation (can enhance with Echo validator later)
	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email and password are required")
	}

	// Call the signup service
	ctx := c.Request().Context()
	log.Printf("Handler: Calling Signup service for email %s", req.Email)
	createdUser, err := h.userService.Signup(ctx, req.Email, req.Password)
	if err != nil {
		log.Printf("Handler: Error from Signup service for email %s: %v", req.Email, err)
		// Map service/repository errors to HTTP status codes
		if errors.Is(err, repository.ErrUserEmailExists) {
			return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("Email '%s' is already registered", req.Email))
		}
		// Check for specific validation errors returned by the service
		if strings.Contains(err.Error(), "password must be") || strings.Contains(err.Error(), "invalid email format") {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		// Handle other errors
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create user: %v", err))
	}

	log.Printf("Handler: Successfully created user %s (ID: %s)", createdUser.Email, createdUser.ID)
	// Return the created user details (password hash excluded due to json tag)
	return c.JSON(http.StatusCreated, createdUser)
}

// Login handles user login requests.
func (h *AuthHandler) Login(c echo.Context) error {
	// Define request structure
	type loginRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	req := new(loginRequest)

	// Bind request body
	if err := c.Bind(req); err != nil {
		log.Printf("Handler: Error binding login request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email and password are required")
	}

	// Call the login service
	ctx := c.Request().Context()
	log.Printf("Handler: Calling Login service for email %s", req.Email)
	token, err := h.userService.Login(ctx, req.Email, req.Password)
	if err != nil {
		log.Printf("Handler: Error from Login service for email %s: %v", req.Email, err)
		// Check if the error indicates invalid credentials
		if strings.Contains(err.Error(), "invalid email or password") { // Check for the generic service error
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password") // Return 401
		}
		// Check for other specific errors if needed (e.g., account locked)

		// Handle other internal errors (like token generation failure)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Login failed: %v", err))
	}

	// Return the token on successful login
	log.Printf("Handler: Login successful for email %s, returning token.", req.Email)
	return c.JSON(http.StatusOK, echo.Map{
		"token": token, // Wrap token in a JSON object
	})
}
