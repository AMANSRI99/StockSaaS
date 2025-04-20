package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/AMANSRI99/StockSaaS/internal/app/model"
	"github.com/AMANSRI99/StockSaaS/internal/app/repository"
	"github.com/AMANSRI99/StockSaaS/internal/common/jwtutil"
	"github.com/AMANSRI99/StockSaaS/internal/config"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Basic email validation regex
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

// --- Interface Definition ---

// UserService defines the interface for user business logic.
type UserService interface {
	Signup(ctx context.Context, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, error) // Add later
}

// --- Implementation ---

type userService struct {
	userRepo repository.UserRepository
	cfg      config.AppConfig // Store app config for JWT secret/expiry later
}

// NewUserService creates a new user service instance.
func NewUserService(repo repository.UserRepository, cfg config.AppConfig) UserService {
	return &userService{
		userRepo: repo,
		cfg:      cfg, // Store config
	}
}

// Signup validates input, hashes password, and saves a new user.
func (s *userService) Signup(ctx context.Context, email, password string) (*model.User, error) {
	// 1. Basic Input Validation
	email = strings.ToLower(strings.TrimSpace(email))
	if !isEmailValid(email) {
		return nil, fmt.Errorf("invalid email format provided")
	}
	// Add password complexity rules if desired (e.g., length)
	if len(password) < 8 { // Example: minimum 8 characters
		return nil, fmt.Errorf("password must be at least 8 characters long")
	}

	// 2. Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Service: Error hashing password for %s: %v", email, err)
		return nil, fmt.Errorf("failed to process password: %w", err)
	}

	// 3. Create User model
	now := time.Now().UTC()
	user := &model.User{
		ID:           uuid.New(),
		Email:        email, // Use cleaned, lowercased email
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 4. Save user using the repository
	log.Printf("Service: Attempting to save user %s (ID: %s)", user.Email, user.ID)
	err = s.userRepo.Save(ctx, user)
	if err != nil {
		log.Printf("Service: Error saving user %s: %v", user.Email, err)
		// Pass up specific known errors like EmailExists
		if errors.Is(err, repository.ErrUserEmailExists) {
			return nil, err // Return the specific error
		}
		// Wrap other errors
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("Service: Successfully created user %s (ID: %s)", user.Email, user.ID)
	// 5. Return created user (PasswordHash has json:"-" tag, so it won't be serialized)
	return user, nil
}

// Helper function for basic email validation
func isEmailValid(email string) bool {
	return emailRegex.MatchString(email)
}

// Login verifies credentials and returns a JWT token upon success.
func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	log.Printf("Service: Attempting login for email %s", email)
	email = strings.ToLower(strings.TrimSpace(email))

	// 1. Basic Validation
	if !isEmailValid(email) {
		return "", fmt.Errorf("invalid email format provided")
	}
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// 2. Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// If user not found OR other DB error, return generic failure message
		// DO NOT reveal whether the email exists or not for security.
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Printf("Service: Login failed - user not found for email %s", email)
		} else {
			log.Printf("Service: DB error during login for email %s: %v", email, err)
		}
		return "", fmt.Errorf("invalid email or password") // Generic error
	}

	// 3. Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// If passwords don't match (or other bcrypt error)
		log.Printf("Service: Login failed - password mismatch for email %s", email)
		// Use the same generic error message
		return "", fmt.Errorf("invalid email or password")
	}

	// 4. Credentials are valid - Generate JWT
	log.Printf("Service: Credentials valid for user %s (ID: %s). Generating token.", user.Email, user.ID)
	token, err := jwtutil.GenerateToken(user.ID, user.Email, s.cfg.JWT.SecretKey, s.cfg.JWT.ExpiryDuration)
	if err != nil {
		log.Printf("Service: Error generating JWT for user %s: %v", user.Email, err)
		// This is an internal server error
		return "", fmt.Errorf("could not generate authentication token: %w", err)
	}

	log.Printf("Service: Token generated successfully for user %s", user.Email)
	// 5. Return the token string
	return token, nil
}
