package service

import (
	"context"
	"fmt"
	"log"

	// Use your actual module path
	kiteadapter "github.com/AMANSRI99/StockSaaS/internal/adapter/broker/kiteconnect"
	"github.com/AMANSRI99/StockSaaS/internal/app/repository"
	"github.com/AMANSRI99/StockSaaS/internal/common/encryptutil"
	"github.com/AMANSRI99/StockSaaS/internal/config"

	"github.com/google/uuid"
)

// --- Interface Definition ---

// KiteService defines the interface for Kite Connect related business logic.
type KiteService interface {
	// CompleteAuthentication exchanges the request token from the callback,
	// encrypts the access token, and saves the broker credentials for the user.
	CompleteAuthentication(ctx context.Context, userID uuid.UUID, requestToken string) error
}

// --- Implementation ---

type kiteService struct {
	kiteAdapter *kiteadapter.Adapter
	brokerRepo  repository.BrokerRepository
	cfg         config.AppConfig // Need APISecret and EncryptionKey
}

// NewKiteService creates a new KiteService instance.
func NewKiteService(ka *kiteadapter.Adapter, br repository.BrokerRepository, cfg config.AppConfig) KiteService {
	return &kiteService{
		kiteAdapter: ka,
		brokerRepo:  br,
		cfg:         cfg,
	}
}

// CompleteAuthentication handles the final step of the OAuth flow.
func (s *kiteService) CompleteAuthentication(ctx context.Context, userID uuid.UUID, requestToken string) error {
	log.Printf("Service: Completing Kite authentication for user %s", userID)

	// 1. Exchange request_token for access_token using the adapter
	log.Printf("Service: Exchanging request token for user %s", userID)
	session, err := s.kiteAdapter.GenerateSession(requestToken, s.cfg.Kite.APISecret)
	if err != nil {
		log.Printf("Service: Failed to generate Kite session for user %s: %v", userID, err)
		// Check for specific Kite errors if needed, otherwise wrap generally
		return fmt.Errorf("broker session generation failed: %w", err)
	}
	// Log success but be careful not to log sensitive parts of the session object
	log.Printf("Service: Session generated successfully for user %s (Kite UserID: %s, Public Token: %s)",
		userID, session.UserID, session.PublicToken)

	// Basic validation of received data
	if session.AccessToken == "" || session.PublicToken == "" || session.UserID == "" {
		log.Printf("Service: Incomplete session data received from Kite for user %s", userID)
		return fmt.Errorf("incomplete session data received from broker")
	}

	// 2. Encrypt the received access token
	accessTokenBytes := []byte(session.AccessToken)
	encryptedAccessToken, err := encryptutil.Encrypt(accessTokenBytes, s.cfg.EncryptionKey)
	if err != nil {
		log.Printf("Service: Failed to encrypt access token for user %s: %v", userID, err)
		// This is an internal error, don't expose details directly
		return fmt.Errorf("internal security error processing credentials")
	}

	// 3. Save the credentials using the broker repository
	log.Printf("Service: Saving broker credentials for user %s", userID)
	err = s.brokerRepo.SaveOrUpdateKiteCredentials(
		ctx,
		userID,
		encryptedAccessToken, // Pass encrypted bytes
		session.PublicToken,
		session.UserID, // This is Kite's User ID
	)
	if err != nil {
		log.Printf("Service: Failed to save broker credentials for user %s: %v", userID, err)
		return fmt.Errorf("failed to store broker credentials")
	}

	log.Printf("Service: Kite authentication completed and credentials saved for user %s", userID)
	return nil // Success
}
