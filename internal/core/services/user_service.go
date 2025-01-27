// internal/core/services/user_service.go
package services

import (
	"context"
	"errors"
	"time"

	"github.com/AMANSRI99/StockSaaS/internal/core/ports"
	"github.com/AMANSRI99/StockSaaS/internal/domain"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo ports.UserRepository
	logger   ports.Logger
	jwtKey   []byte
}

type UserService interface {
	Register(ctx context.Context, email, password, name string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	UpdateZerodhaCredentials(ctx context.Context, userID, apiKey, apiSecret, accessToken string) error
	ValidateToken(ctx context.Context, token string) (string, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
}

func NewUserService(userRepo ports.UserRepository, logger ports.Logger, jwtKey []byte) UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
		jwtKey:   jwtKey,
	}
}

func (s *userService) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, errors.New("failed to create user")
	}

	// Create new user
	user := &domain.User{
		ID:             uuid.New().String(),
		Email:          email,
		HashedPassword: string(hashedPassword),
		Name:           name,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save user to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hour expiry
	})

	// Sign the token
	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		s.logger.Error("Failed to generate token", "error", err)
		return "", errors.New("failed to generate token")
	}

	return tokenString, nil
}

func (s *userService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) UpdateZerodhaCredentials(ctx context.Context, userID, apiKey, apiSecret, accessToken string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	user.ZerodhaAPIKey = apiKey
	user.ZerodhaAPISecret = apiSecret
	user.ZerodhaAccessToken = accessToken
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update Zerodha credentials", "error", err)
		return errors.New("failed to update credentials")
	}

	return nil
}

func (s *userService) ValidateToken(ctx context.Context, tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtKey, nil
	})

	if err != nil {
		return "", errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", errors.New("invalid token claims")
		}
		return userID, nil
	}

	return "", errors.New("invalid token")
}

func (s *userService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash new password", "error", err)
		return errors.New("failed to update password")
	}

	// Update password
	user.HashedPassword = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update password", "error", err)
		return errors.New("failed to update password")
	}

	return nil
}
