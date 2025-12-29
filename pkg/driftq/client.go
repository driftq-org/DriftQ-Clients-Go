package driftq

import (
	"context"
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

// Config holds client connection options.
type Config struct {
	Address string
	TLS     *TLSConfig
	Retry   RetryPolicy
	Timeout time.Duration
}

// Client is a lightweight handle to DriftQ services.
type Client struct {
	cfg Config
	// TODO: hold gRPC conn, interceptors, etc.
}

// Dial creates a new client. No network is opened in this stub.
func Dial(_ context.Context, cfg Config) (*Client, error) {
	return &Client{cfg: cfg}, nil
}

// Close shuts down resources (no-op in this stub).
func (c *Client) Close() error { return nil }

// Producer returns a producer bound to a topic.
func (c *Client) Producer(topic string) *Producer {
	return &Producer{topic: topic, c: c}
}

// Consumer returns a consumer bound to a topic/group.
func (c *Client) Consumer(topic, group string) *Consumer {
	return &Consumer{topic: topic, group: group, c: c}
}

// Admin returns an admin client wrapper.
func (c *Client) Admin() *Admin { return &Admin{c: c} }
