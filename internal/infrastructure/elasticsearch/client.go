package elasticsearch

import (
	"context"
	"fmt"
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

// Ping reports whether Elasticsearch is reachable; used by the readiness probe.
func (c *Client) Ping(ctx context.Context) error {
	res, err := c.ES.Ping(c.ES.Ping.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("elasticsearch ping: %s", res.Status())
	}
	return nil
}

func (c *Client) waitForReady(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		lastErr = c.Ping(pingCtx)
		cancel()
		if lastErr == nil {
			c.logger.Info("elasticsearch is ready")
			return nil
		}
		c.logger.Info("waiting for elasticsearch...", "retry_in", "2s", "error", lastErr)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("elasticsearch not ready after %s: %v", timeout, lastErr)
}
