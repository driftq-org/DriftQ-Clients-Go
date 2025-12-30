package driftq

import "errors"

var (
	// These are some common typed errors (expand as real APIs land)
	ErrTopicNotFound     = errors.New("topic not found")
	ErrBrokerUnavailable = errors.New("broker unavailable")
)
