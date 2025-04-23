package kiteconnect

import (
	"fmt"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

// Adapter wraps the Kite Connect client.
type Adapter struct {
	client *kiteconnect.Client
}

// NewAdapter creates a new Kite Connect client instance.
func NewAdapter(apiKey string) *Adapter {
	client := kiteconnect.New(apiKey)
	return &Adapter{
		client: client,
	}
}

// GetLoginURL generates the Kite Connect login URL for user redirection.
// The redirectURL passed here should be the one registered in your Kite app settings.
func (a *Adapter) GetLoginURL() string {
	// Use the library's built-in method
	return a.client.GetLoginURL()
	// We will append the state parameter manually in the handler before redirecting.
}

// GenerateSession exchanges a request token for an access token and user session details.
// (This method remains unchanged and matches the example's usage)
func (a *Adapter) GenerateSession(requestToken, apiSecret string) (kiteconnect.UserSession, error) {
	userSession, err := a.client.GenerateSession(requestToken, apiSecret)
	if err != nil {
		return kiteconnect.UserSession{}, fmt.Errorf("kite connect generate session failed: %w", err)
	}
	return userSession, nil
}

// --- We will add other methods later for placing orders, getting profile etc. ---
// func (a *Adapter) PlaceOrder(...) (...)
// func (a *Adapter) GetUserProfile(...) (...)
