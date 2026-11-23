package model

import "time"

// Portfolio is a user's target allocation plus the rules used when rebalancing it.
type Portfolio struct {
	UserID     string             `json:"user_id"`
	Allocation map[string]float64 `json:"allocation"`
	// DriftThreshold (percentage points): if no asset has drifted by at least
	// this much, a rebalance produces no trades. 0 means always rebalance.
	DriftThreshold float64 `json:"drift_threshold"`
	// MinTradePercent (percentage points): individual trades smaller than this
	// are skipped to avoid tiny, fee-dominated orders. 0 means no minimum.
	MinTradePercent float64   `json:"min_trade_percent"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UpdatedPortfolio is the rebalance event published to Kafka. RebalanceID
// doubles as the event ID and is what makes processing idempotent.
type UpdatedPortfolio struct {
	RebalanceID    string             `json:"rebalance_id"`
	UserID         string             `json:"user_id"`
	NewAllocation  map[string]float64 `json:"new_allocation"`
	PortfolioValue float64            `json:"portfolio_value,omitempty"`
	IdempotencyKey string             `json:"idempotency_key,omitempty"`
	RequestedAt    time.Time          `json:"requested_at"`
}

type RebalanceTransaction struct {
	ID               string  `json:"id"`
	RebalanceID      string  `json:"rebalance_id"`
	UserID           string  `json:"user_id"`
	Action           string  `json:"action"`
	Asset            string  `json:"asset"`
	RebalancePercent float64 `json:"rebalance_percent"`
	// Amount is the trade size in currency units; only set when the
	// rebalance request included portfolio_value.
	Amount    float64   `json:"amount,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	ActionBuy  = "BUY"
	ActionSell = "SELL"
)
