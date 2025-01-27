package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/AMANSRI99/StockSaaS/internal/core/ports"
	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	userService ports.UserService
	logger      *logrus.Logger
}

func NewAuthMiddleware(userService ports.UserService, logger *logrus.Logger) *AuthMiddleware {
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
