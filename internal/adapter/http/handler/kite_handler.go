package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	// Use your actual module path
	kiteadapter "github.com/AMANSRI99/StockSaaS/internal/adapter/broker/kiteconnect" // Use alias if needed
	"github.com/AMANSRI99/StockSaaS/internal/app/repository"
	"github.com/AMANSRI99/StockSaaS/internal/app/service"

	//mw "github.com/AMANSRI99/StockSaaS/internal/adapter/http/middleware"
	//"github.com/google/uuid"

	// "github.com/google/uuid" // Needed for state generation later

	"github.com/labstack/echo/v4"
)

// KiteHandler now holds KiteService as well
type KiteHandler struct {
	kiteAdapter    *kiteadapter.Adapter
	oauthStateRepo repository.OAuthStateRepository
	kiteService    service.KiteService // <-- Add KiteService dependency
}

// NewKiteHandler updated constructor
func NewKiteHandler(ka *kiteadapter.Adapter, osr repository.OAuthStateRepository, ks service.KiteService) *KiteHandler {
	return &KiteHandler{
		kiteAdapter:    ka,
		oauthStateRepo: osr,
		kiteService:    ks, // <-- Inject KiteService
	}
}

// InitiateKiteConnect starts the OAuth flow by redirecting the user to Kite login.
// (Ensure this is updated to use oauthStateRepo correctly as shown previously)
func (h *KiteHandler) InitiateKiteConnect(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}
	log.Printf("Handler: Initiating Kite Connect flow for user %s", userID)

	// 1. Generate secure random state token
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		log.Printf("Handler: Failed to generate state token: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initiate connection (state gen)")
	}
	stateToken := hex.EncodeToString(stateBytes)

	// 2. Save state token with user ID and expiry (e.g., 10 minutes)
	ctx := c.Request().Context()
	stateExpiry := 10 * time.Minute
	err = h.oauthStateRepo.SaveState(ctx, stateToken, userID, stateExpiry)
	if err != nil {
		log.Printf("Handler: Failed to save state token for user %s: %v", userID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initiate connection (state save)")
	}

	// 3. Get base login URL and append state
	loginURLBase := h.kiteAdapter.GetLoginURL() // Use corrected adapter method
	parsedURL, err := url.Parse(loginURLBase)
	if err != nil {
		log.Printf("Handler: Failed to parse base login URL '%s': %v", loginURLBase, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to construct login URL")
	}
	query := parsedURL.Query()
	query.Set("state", stateToken) // Add state parameter
	parsedURL.RawQuery = query.Encode()
	finalLoginURL := parsedURL.String()

	log.Printf("Handler: Redirecting user %s to Kite Login URL with state: %s", userID, finalLoginURL)
	return c.Redirect(http.StatusTemporaryRedirect, finalLoginURL)
}

// HandleKiteCallback receives the redirect from Kite, verifies state, exchanges token, saves credentials.
func (h *KiteHandler) HandleKiteCallback(c echo.Context) error {
	log.Printf("Handler: Received Kite callback request URL: %s", c.Request().URL.String())

	// Define frontend redirect URLs (Load from config ideally)
	// TODO: Load these from configuration
	frontendSuccessURL := "/?kite_connected=success" // Example relative URL for SPA handling
	frontendErrorURLBase := "/?kite_error="          // Example relative URL base for SPA handling

	// Helper function for error redirects
	redirectWithError := func(errorCode string, logMsg string, logArgs ...interface{}) error {
		log.Printf("Handler Error: "+logMsg, logArgs...)
		errorURL := frontendErrorURLBase + errorCode
		// Maybe add context like url.QueryEscape(logMsg)? Be careful not to expose sensitive info.
		return c.Redirect(http.StatusTemporaryRedirect, errorURL)
	}

	// 1. Extract query parameters
	status := c.QueryParam("status")
	requestToken := c.QueryParam("request_token")
	stateToken := c.QueryParam("state") // Get the state token

	// 2. Check status and request token from Kite
	if status != "success" || requestToken == "" {
		errMsg := c.QueryParam("message")
		return redirectWithError("callback_failed", "Kite callback indicated failure or missing request token. Status: '%s', Error: '%s', RequestToken: '%s'", status, errMsg, requestToken)
	}

	// 3. Check if state token is present (as it was missing in user's previous trace)
	if stateToken == "" {
		return redirectWithError("missing_state", "Kite callback response missing 'state' parameter.")
	}

	// 4. Verify and consume the state token using the repository
	ctx := c.Request().Context()
	userID, err := h.oauthStateRepo.VerifyAndConsumeState(ctx, stateToken)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidOrExpiredState) {
			return redirectWithError("invalid_state", "Invalid or expired state token received: %s", stateToken)
		}
		// Otherwise, internal error reading/deleting from DB
		return redirectWithError("state_verification_failed", "State token verification failed for state '%s': %v", stateToken, err)
	}
	log.Printf("Handler: State token verified successfully for user %s", userID)

	// 5. Call the service to complete authentication and save tokens
	err = h.kiteService.CompleteAuthentication(ctx, userID, requestToken)
	if err != nil {
		// Service layer already logged details
		return redirectWithError("token_exchange_failed", "Error completing Kite authentication for user %s: %v", userID, err)
	}

	// 6. Redirect user to a frontend success page
	log.Printf("Handler: Kite connection successful for user %s. Redirecting to success page: %s", userID, frontendSuccessURL)
	return c.Redirect(http.StatusTemporaryRedirect, frontendSuccessURL)
}
