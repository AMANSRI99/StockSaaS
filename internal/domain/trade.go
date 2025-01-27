package domain

import "time"

type OrderType string
type TradeType string

const (
    OrderTypeMarket OrderType = "MARKET"
    OrderTypeLimit  OrderType = "LIMIT"
    
    TradeTypeBuy  TradeType = "BUY"
    TradeTypeSell TradeType = "SELL"
)

type Trade struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Symbol    string    `json:"symbol"`
    Quantity  int       `json:"quantity"`
    OrderType OrderType `json:"order_type"`
    Price     float64   `json:"price,omitempty"`
    TradeType TradeType `json:"trade_type"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}