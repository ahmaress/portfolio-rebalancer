package elasticsearch

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
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

func New(opts Options, logger *slog.Logger) (*Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{opts.URL},
		Username:  opts.Username,
		Password:  opts.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("creating elasticsearch client: %w", err)
	}

	c := &Client{
		ES: es,
		Indices: Indices{
			Portfolios:   opts.IndexPrefix + IndexPortfolios,
			Transactions: opts.IndexPrefix + IndexTransactions,
			Rebalances:   opts.IndexPrefix + IndexRebalances,
		},
		logger: logger,
	}

	wait := opts.ReadyWait
	if wait == 0 {
		wait = 30 * time.Second
	}
	ctx := context.Background()
	if err := c.waitForReady(ctx, wait); err != nil {
		return nil, err
	}
	if err := c.ensureIndices(ctx); err != nil {
		return nil, err
	}
	return c, nil
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

func (c *Client) mappings() map[string]string {
	return map[string]string{
		c.Indices.Portfolios: `{
			"properties": {
				"user_id":           {"type": "keyword"},
				"allocation":        {"type": "object"},
				"drift_threshold":   {"type": "float"},
				"min_trade_percent": {"type": "float"},
				"created_at":        {"type": "date"},
				"updated_at":        {"type": "date"}
			}
		}`,
		c.Indices.Transactions: `{
			"properties": {
				"id":                {"type": "keyword"},
				"rebalance_id":      {"type": "keyword"},
				"user_id":           {"type": "keyword"},
				"action":            {"type": "keyword"},
				"asset":             {"type": "keyword"},
				"rebalance_percent": {"type": "float"},
				"amount":            {"type": "double"},
				"created_at":        {"type": "date"}
			}
		}`,
		c.Indices.Rebalances: `{
			"properties": {
				"id":                {"type": "keyword"},
				"user_id":           {"type": "keyword"},
				"status":            {"type": "keyword"},
				"new_allocation":    {"type": "object", "enabled": false},
				"portfolio_value":   {"type": "double"},
				"idempotency_key":   {"type": "keyword"},
				"transaction_count": {"type": "integer"},
				"max_drift":         {"type": "float"},
				"message":           {"type": "text", "index": false},
				"error":             {"type": "text", "index": false},
				"attempts":          {"type": "integer"},
				"created_at":        {"type": "date"},
				"updated_at":        {"type": "date"},
				"completed_at":      {"type": "date"}
			}
		}`,
	}
}

// ensureIndices creates missing indices and adds any new fields to existing
// ones (additive mapping updates are safe; conflicting ones are reported).
func (c *Client) ensureIndices(ctx context.Context) error {
	for index, properties := range c.mappings() {
		res, err := c.ES.Indices.Exists([]string{index}, c.ES.Indices.Exists.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("checking index %s: %w", index, err)
		}
		res.Body.Close()

		if res.StatusCode == http.StatusOK {
			putRes, err := c.ES.Indices.PutMapping(
				[]string{index},
				strings.NewReader(properties),
				c.ES.Indices.PutMapping.WithContext(ctx),
			)
			if err != nil {
				return fmt.Errorf("updating mapping for %s: %w", index, err)
			}
			body, _ := io.ReadAll(putRes.Body)
			putRes.Body.Close()
			if putRes.IsError() {
				return fmt.Errorf("updating mapping for %s [%s]: %s (drop the index if it holds only dev data)",
					index, putRes.Status(), body)
			}
			c.logger.Info("index mapping up to date", "index", index)
			continue
		}

		createBody := `{"mappings": ` + properties + `}`
		createRes, err := c.ES.Indices.Create(
			index,
			c.ES.Indices.Create.WithBody(strings.NewReader(createBody)),
			c.ES.Indices.Create.WithContext(ctx),
		)
		if err != nil {
			return fmt.Errorf("creating index %s: %w", index, err)
		}
		body, _ := io.ReadAll(createRes.Body)
		createRes.Body.Close()
		// A concurrent instance may have created it first.
		if createRes.IsError() && !strings.Contains(string(body), "resource_already_exists_exception") {
			return fmt.Errorf("elasticsearch error creating index %s [%s]: %s", index, createRes.Status(), body)
		}
		c.logger.Info("index created", "index", index)
	}
	return nil
}
