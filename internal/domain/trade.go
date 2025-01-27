package domain

import (
	"errors"
	"time"
)

type OrderType string
type TradeType string
type OrderStatus string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"

	TradeTypeBuy  TradeType = "BUY"
	TradeTypeSell TradeType = "SELL"

	OrderStatusPending  OrderStatus = "PENDING"
	OrderStatusComplete OrderStatus = "COMPLETE"
	OrderStatusFailed   OrderStatus = "FAILED"
	OrderStatusCanceled OrderStatus = "CANCELED"
)
var (
    ErrTradeCannotBeCanceled = errors.New("trade cannot be canceled")
)
type Trade struct {
	ID            string      `json:"id"`
	UserID        string      `json:"user_id"`
	Symbol        string      `json:"symbol"`
	Quantity      int         `json:"quantity"`
	OrderType     OrderType   `json:"order_type"`
	Price         float64     `json:"price,omitempty"`
	TradeType     TradeType   `json:"trade_type"`
	Status        OrderStatus `json:"status"`
	ErrorMessage  string      `json:"error_message,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	ExecutedPrice float64     `json:"executed_price,omitempty"`
}

func (t *Trade) Validate() error {
	if t.Symbol == "" {
		return errors.New("symbol is required")
	}
	if t.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if t.OrderType == OrderTypeLimit && t.Price <= 0 {
		return errors.New("price is required for limit orders")
	}
	if t.TradeType != TradeTypeBuy && t.TradeType != TradeTypeSell {
		return errors.New("invalid trade type")
	}
	return nil
}
