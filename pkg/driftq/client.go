package driftq

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Placeholder for future TLS options.
type TLSConfig struct {
	// TODO: add fields (CA, cert/key, InsecureSkipVerify, etc.)
}

type Client struct {
	cfg     Config
	baseURL string
	httpc   *http.Client
}

type Config struct {
	BaseURL   string
	TLS       *TLSConfig
	Timeout   time.Duration
	Retry     RetryConfig
	Tracing   TracingConfig
	UserAgent string
	Transport http.RoundTripper
}

func Dial(ctx context.Context, cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, fmt.Errorf("base_url is required")
	}

	if _, err := url.Parse(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("invalid base_url: %w", err)
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	if cfg.Timeout < 0 {
		cfg.Timeout = 0
	}

	if strings.TrimSpace(cfg.UserAgent) == "" {
		cfg.UserAgent = "driftq-go/" + Version
	}

	// Middleware stack (outer -> inner):
	// Deadline -> Tracing -> Retry -> base transport
	baseTransport := cfg.Transport
	transport := ChainTransport(
		baseTransport,
		DeadlineMiddleware(cfg.Timeout),
		TracingMiddleware(cfg.Tracing),
		RetryMiddleware(cfg.Retry),
	)

	httpc := &http.Client{Transport: transport}

	return &Client{
		cfg:     cfg,
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		httpc:   httpc,
	}, nil
}

// Close shuts down resources (no-op in this stub)
func (c *Client) Close() error { return nil }
