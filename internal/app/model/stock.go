package model

// Stock represents a single stock within a basket.
type Stock struct {
	Symbol   string `json:"symbol"`   // e.g., "RELIANCE", "INFY"
	Quantity int    `json:"quantity"` // Number of shares
}
