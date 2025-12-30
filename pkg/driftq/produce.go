package driftq

import (
	"context"
	"net/http"
)

func (c *Client) Produce(ctx context.Context, req ProduceRequest) (ProduceResponse, error) {
	var out ProduceResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/produce", nil, req, &out)
	return out, err
}
