package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	// Use your actual module path
	"github.com/AMANSRI99/StockSaaS/internal/common/jwtutil" // Import our JWT helpers

	"github.com/golang-jwt/jwt/v5" // Need for error checking
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ContextKey type for storing user ID in context to avoid collisions
type ContextKey string

const UserIDContextKey ContextKey = "user_id"

// NewJWTAuthMiddleware creates an Echo middleware function for JWT authentication.
// It takes the JWT secret key as a dependency.
func NewJWTAuthMiddleware(jwtSecret string) echo.MiddlewareFunc {
	// Return the actual middleware handler
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		// This inner function is the actual handler executed by Echo
		return func(c echo.Context) error {
			// 1. Get the Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Println("Auth Middleware: Missing Authorization header")
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization header")
			}

			// 2. Check if it's a Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				log.Printf("Auth Middleware: Malformed Authorization header: %s", authHeader)
				return echo.NewHTTPError(http.StatusUnauthorized, "Malformed authorization header (expecting 'Bearer <token>')")
			}
			tokenString := parts[1]

			// 3. Validate the token using our helper
			claims, err := jwtutil.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				log.Printf("Auth Middleware: Token validation failed: %v", err)
				// Check for specific errors like expiration
				if errors.Is(err, jwt.ErrTokenExpired) || strings.Contains(err.Error(), "token has expired") {
					return echo.NewHTTPError(http.StatusUnauthorized, "Token has expired")
				}
				// Other validation errors
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			// 4. Token is valid, extract UserID from claims
			// We stored UserID in the Subject ("sub") or our custom "user_id" claim
			userIDStr := claims.UserID // Assuming UserID field exists in CustomClaims
			if userIDStr == "" {
				userIDStr = claims.Subject // Fallback to Subject if UserID claim wasn't set
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				log.Printf("Auth Middleware: Failed to parse user ID from token claims ('%s'): %v", userIDStr, err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Invalid user identifier in token") // Should not happen if generated correctly
			}

			// 5. Store UserID in context for downstream handlers/services
			log.Printf("Auth Middleware: User %s authenticated successfully.", userID)
			c.Set(string(UserIDContextKey), userID) // Use typed key

			// 6. Call the next handler in the chain
			return next(c)
		}
	}
}
