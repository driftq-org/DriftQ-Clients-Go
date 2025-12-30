package driftq

import (
	"context"
	"net/http"
)

func (c *Client) Produce(ctx context.Context, req ProduceRequest) (ProduceResponse, error) {
	var out ProduceResponse

	hdr := make(http.Header)
	if req.Envelope != nil {
		if k := req.Envelope.IdempotencyKey; k != "" {
			hdr.Set("Idempotency-Key", k)
		}
	}

	err := c.doJSONWithHeaders(ctx, http.MethodPost, "/v1/produce", nil, hdr, req, &out)
	return out, err
}
