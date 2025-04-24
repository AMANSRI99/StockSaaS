package handler

import (
	"log"
	"net/http"
	"time"

	// Use your actual module path
	kiteadapter "github.com/AMANSRI99/StockSaaS/internal/adapter/broker/kiteconnect" // Use alias if needed
	"github.com/AMANSRI99/StockSaaS/internal/app/service"
	"github.com/AMANSRI99/StockSaaS/internal/common/jwtutil"
	"github.com/AMANSRI99/StockSaaS/internal/config"
	"github.com/google/uuid"

	//mw "github.com/AMANSRI99/StockSaaS/internal/adapter/http/middleware"
	"github.com/labstack/echo/v4"
)

// KiteHandler dependencies change: remove oauthStateRepo, add config
type KiteHandler struct {
	kiteAdapter *kiteadapter.Adapter
	kiteService service.KiteService
	cfg         config.AppConfig // Add config
}

// NewKiteHandler updated constructor
func NewKiteHandler(ka *kiteadapter.Adapter, ks service.KiteService, cfg config.AppConfig) *KiteHandler {
	if ka == nil || ks == nil {
		log.Fatal("FATAL: Nil kite adapter or service passed to NewKiteHandler")
	}
	return &KiteHandler{
		kiteAdapter: ka,
		kiteService: ks,
		cfg:         cfg, // Store config
	}
}

// InitiateKiteConnect starts the OAuth flow using a cookie for state.
func (h *KiteHandler) InitiateKiteConnect(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	log.Printf("Handler: Initiating Kite Connect flow for user %s", userID)

	// 1. Generate short-lived JWT state token containing userID
	// Use a short expiry, e.g., 5 or 10 minutes
	stateTokenExpiry := 10 * time.Minute
	// Reusing GenerateToken - assuming it takes expiry duration.
	// We don't need email here, just userID (subject).
	stateToken, err := jwtutil.GenerateToken(userID, "", h.cfg.JWT.SecretKey, stateTokenExpiry)
	if err != nil {
		log.Printf("Handler: Failed to generate state JWT for user %s: %v", userID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initiate connection (state jwt gen)")
	}

	// 2. Set the state JWT as a secure cookie
	stateCookie := new(http.Cookie)
	stateCookie.Name = "kite_oauth_state" // Consistent cookie name
	stateCookie.Value = stateToken
	//stateCookie.Expires = time.Now().Add(stateTokenExpiry)
	stateCookie.MaxAge = 600                    // MaxAge is an alternative to Expires
	stateCookie.Path = "/"                      // Or restrict path if needed e.g., "/api/kite/connect"
	stateCookie.HttpOnly = true                 // Crucial: Prevent JS access
	stateCookie.Secure = true                  // Crucial: Send only over HTTPS (requires HTTPS setup/ngrok locally)
	stateCookie.SameSite = http.SameSiteLaxMode // Recommended for OAuth callbacks

	c.SetCookie(stateCookie)
	log.Printf("Handler: Set state cookie for user %s", userID)

	// 3. Get base login URL (no need to append state to URL anymore)
	loginURL := h.kiteAdapter.GetLoginURL()

	log.Printf("Handler: Redirecting user %s to Kite Login URL: %s", userID, loginURL)
	return c.Redirect(http.StatusTemporaryRedirect, loginURL)
}

// HandleKiteCallback receives redirect, verifies state cookie, exchanges token, saves credentials.
func (h *KiteHandler) HandleKiteCallback(c echo.Context) error {
	log.Printf("Handler: Received Kite callback request URL: %s", c.Request().URL.String())

	frontendSuccessURL := "/?kite_connected=success"
	frontendErrorURLBase := "/?kite_error="
	stateCookieName := "kite_oauth_state" // Use consistent name

	redirectWithError := func(errorCode string, logMsg string, logArgs ...interface{}) error {
		log.Printf("Handler Error: "+logMsg, logArgs...)
		// Clear the state cookie on error too
		clearCookie := &http.Cookie{Name: stateCookieName, MaxAge: -1, Path: "/"}
		c.SetCookie(clearCookie)
		errorURL := frontendErrorURLBase + errorCode
		return c.Redirect(http.StatusTemporaryRedirect, errorURL)
	}

	// 1. Extract query parameters (status, request_token - state is no longer expected here)
	status := c.QueryParam("status")
	requestToken := c.QueryParam("request_token")

	if status != "success" || requestToken == "" {
		errMsg := c.QueryParam("message")
		return redirectWithError("callback_failed", "Kite callback indicated failure or missing request token. Status: '%s', Error: '%s', RequestToken: '%s'", status, errMsg, requestToken)
	}

	// 2. Read and validate the state cookie
	stateCookie, err := c.Cookie(stateCookieName)
	if err != nil {
		// If cookie is missing (http.ErrNoCookie or other error)
		return redirectWithError("missing_state_cookie", "State cookie '%s' not found or failed to read: %v", stateCookieName, err)
	}

	// 3. Validate the state JWT from the cookie
	claims, err := jwtutil.ValidateToken(stateCookie.Value, h.cfg.JWT.SecretKey)
	if err != nil {
		// Covers expired tokens, invalid signatures etc.
		return redirectWithError("invalid_state", "Invalid or expired state cookie: %v", err)
	}

	// 4. Extract UserID from validated claims
	userIDStr := claims.UserID // Or claims.Subject
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return redirectWithError("internal_error", "Failed to parse user ID from state cookie claims ('%s'): %v", userIDStr, err)
	}
	log.Printf("Handler: State cookie validated successfully for user %s", userID)

	// 5. Clear the state cookie immediately after validation prevent reuse
	clearCookie := &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		MaxAge:   -1,  // Tell browser to delete immediately
		Path:     "/", // Use the same path as when setting
		HttpOnly: true,
		Secure:   true, // Match original settings
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(clearCookie)
	log.Printf("Handler: Cleared state cookie for user %s", userID)

	// 6. Call the service to complete authentication, passing userID from cookie
	ctx := c.Request().Context()
	err = h.kiteService.CompleteAuthentication(ctx, userID, requestToken)
	if err != nil {
		return redirectWithError("token_exchange_failed", "Error completing Kite authentication for user %s: %v", userID, err)
	}

	// 7. Redirect user to frontend success page
	log.Printf("Handler: Kite connection successful for user %s. Redirecting to success page: %s", userID, frontendSuccessURL)
	return c.Redirect(http.StatusTemporaryRedirect, frontendSuccessURL)
}
