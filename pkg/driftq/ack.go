package driftq

import (
	"context"
	"net/http"
)

func (c *Client) Ack(ctx context.Context, req AckRequest) error {
	return c.doJSON(ctx, http.MethodPost, "/v1/ack", nil, req, nil)
}

func (c *Client) Nack(ctx context.Context, req NackRequest) error {
	return c.doJSON(ctx, http.MethodPost, "/v1/nack", nil, req, nil)
}
