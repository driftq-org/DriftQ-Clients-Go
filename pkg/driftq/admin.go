package driftq

import "context"

// Topic describes a broker topic. Good for now
type Topic struct {
	Name       string
	Partitions uint32
	Compacted  bool
}

// Admin wraps admin operations (topics, groups, etc.)
type Admin struct{ c *Client }

// ListTopics returns topics (empty in this stub)
func (a *Admin) ListTopics(_ context.Context) ([]Topic, error) {
	return []Topic{}, nil
}

// CreateTopic creates a topic (no-op in this stub)
func (a *Admin) CreateTopic(_ context.Context, _ Topic) error {
	return nil
}
