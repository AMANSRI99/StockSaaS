package domain

type Portfolio struct {
	UserID     string          `json:"user_id"`
	Holdings   []PortfolioItem `json:"holdings"`
	TotalValue float64         `json:"total_value"`
}

type PortfolioItem struct {
	Symbol       string  `json:"symbol"`
	Quantity     int     `json:"quantity"`
	AveragePrice float64 `json:"average_price"`
	CurrentPrice float64 `json:"current_price"`
	CurrentValue float64 `json:"current_value"`
	ProfitLoss   float64 `json:"profit_loss"`
}
