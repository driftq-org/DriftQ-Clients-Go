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

// RetryPolicy controls client retry behavior.
type RetryPolicy struct {
	Attempts   int
	MinBackoff time.Duration
	MaxBackoff time.Duration
}

type Config struct {
	BaseURL string
	TLS     *TLSConfig
	Retry   RetryPolicy
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
		httpc:   &http.Client{Timeout: to},
	}, nil
}

// Close shuts down resources (no-op in this stub)
func (c *Client) Close() error { return nil }

// Producer returns a producer bound to a topic
func (c *Client) Producer(topic string) *Producer {
	return &Producer{topic: topic, c: c}
}

// Consumer returns a consumer bound to a topic/group
func (c *Client) Consumer(topic, group string) *Consumer {
	return &Consumer{topic: topic, group: group, c: c}
}

// Admin returns an admin client wrapper
func (c *Client) Admin() *Admin { return &Admin{c: c} }
