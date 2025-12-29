package driftq

import "context"

// A minimal DriftQ message for now
type Message struct {
	Key     string
	Value   []byte
	Headers map[string]string
}

// Note to myself: This Producer sends messages to a topic
type Producer struct {
	topic string
	c     *Client
}

// Send publishes a message (no-op in this stub)
func (p *Producer) Send(_ context.Context, _ *Message) error {
	// TODO: call gRPC Produce when stubs exist
	return nil
}
