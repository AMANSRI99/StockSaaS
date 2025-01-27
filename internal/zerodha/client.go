package zerodha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AMANSRI99/StockSaaS/internal/domain"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: "https://api.kite.trade",
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

type OrderResponse struct {
	Status  string `json:"status"`
	OrderID string `json:"order_id"`
	Message string `json:"message,omitempty"`
}

func (c *Client) PlaceTrade(ctx context.Context, trade *domain.Trade, apiKey, accessToken string) error {
	endpoint := fmt.Sprintf("%s/orders/regular", c.baseURL)

	payload := map[string]interface{}{
		"tradingsymbol":    trade.Symbol,
		"exchange":         "NSE",
		"transaction_type": string(trade.TradeType),
		"order_type":       string(trade.OrderType),
		"quantity":         trade.Quantity,
		"product":          "CNC", // For equity delivery
		"validity":         "DAY",
	}

	if trade.OrderType == domain.OrderTypeLimit {
		payload["price"] = trade.Price
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling order payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Kite-API-Key", apiKey)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	var orderResp OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("zerodha error: %s", orderResp.Message)
	}

	trade.ID = orderResp.OrderID
	return nil
}

func (c *Client) CancelTrade(ctx context.Context, trade *domain.Trade, apiKey, accessToken string) error {
	endpoint := fmt.Sprintf("%s/orders/regular/%s", c.baseURL, trade.ID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Kite-API-Key", apiKey)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	var orderResp OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error canceling order: %s", orderResp.Message)
	}

	return nil
}

func (c *Client) GetTradeStatus(ctx context.Context, tradeID string, apiKey, accessToken string) (*domain.Trade, error) {
	endpoint := fmt.Sprintf("%s/orders/%s", c.baseURL, tradeID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Kite-API-Key", apiKey)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	var orderResp struct {
		Status   string  `json:"status"`
		OrderID  string  `json:"order_id"`
		Price    float64 `json:"price"`
		Quantity int     `json:"quantity"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &domain.Trade{
		ID:            orderResp.OrderID,
		Status:        domain.OrderStatus(orderResp.Status),
		ExecutedPrice: orderResp.Price,
		Quantity:      orderResp.Quantity,
	}, nil
}
