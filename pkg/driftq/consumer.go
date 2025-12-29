package driftq

import "context"

// This handler processes a single message; return nil to ACK :)
type Handler func(*Message) error

// Consumer receives messages from a topic/group.
type Consumer struct {
	topic string
	group string
	c     *Client
}

// Start begins consumption using the provided handler
// Returns a stop function that halts consumption (no-op in this stub)
func (co *Consumer) Start(_ context.Context, _ Handler) (func() error, error) {
	stop := func() error { return nil }
	// TODO: open stream, deliver to handler, handle ACKs
	return stop, nil
}
