package jwtutil

import (
	"errors"
	"fmt"
	"log"
	"time"

	// Assuming user model needed for validation later
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// CustomClaims defines the structure for our JWT claims.
// We embed RegisteredClaims and add our own.
type CustomClaims struct {
	UserID string `json:"user_id"` // Using string for easier claim access
	Email  string `json:"email"`   // Optional: include email if useful
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT access token.
func GenerateToken(userID uuid.UUID, email string, secretKey string, expiryDuration time.Duration) (string, error) {
	// Create the claims
	claims := CustomClaims{
		UserID: userID.String(), // Store user ID as string in claim
		Email:  email,           // Optionally include email
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiryDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "stocksaas-api", // Optional: Identify your service
			Subject:   userID.String(), // Standard place for user identifier
			// Audience:  []string{"some_audience"}, // Optional
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken parses and validates a JWT token string.
// Returns the claims if valid, otherwise returns an error.
// This will be used by the authentication middleware in the next phase.
func ValidateToken(tokenString string, secretKey string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key for validation
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Printf("Error parsing token: %v", err)
		// Check for specific errors like expiration
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token has expired: %w", err)
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if the claims can be asserted and the token is valid
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// Optionally check issuer/audience here if set during generation
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
