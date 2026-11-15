package elasticsearch

import (
	"log/slog"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

// Base index names; the configured prefix is prepended to each.
const (
	IndexPortfolios   = "portfolios"
	IndexTransactions = "rebalance_transactions"
	IndexRebalances   = "rebalances"
)

// refreshPolicy makes writes visible to subsequent searches without forcing
// an expensive immediate refresh on every request.
const refreshPolicy = "wait_for"

type Options struct {
	URL         string
	Username    string
	Password    string
	IndexPrefix string
	ReadyWait   time.Duration
}

type Indices struct {
	Portfolios   string
	Transactions string
	Rebalances   string
}

type Client struct {
	ES      *elasticsearch.Client
	Indices Indices
	logger  *slog.Logger
}
