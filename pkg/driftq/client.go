package driftq

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// Placeholder for future TLS options.
type TLSConfig struct {
	// TODO: add fields (CA, cert/key, InsecureSkipVerify, etc.)
}

type Config struct {
	BaseURL string
	TLS     *TLSConfig
	Timeout time.Duration
}

type Client struct {
	cfg     Config
	baseURL string
	httpc   *http.Client
}

func Dial(_ context.Context, cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080"
	}
	base := strings.TrimRight(cfg.BaseURL, "/")

	to := cfg.Timeout
	if to <= 0 {
		to = 10 * time.Second
	}

	return &Client{
		cfg:     cfg,
		baseURL: base,
		httpc:   &http.Client{},
	}, nil
}

// Close shuts down resources (no-op in this stub)
func (c *Client) Close() error { return nil }
