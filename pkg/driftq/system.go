package driftq

import (
	"context"
	"net/http"
)

func (c *Client) Healthz(ctx context.Context) (HealthzResponse, error) {
	var out HealthzResponse
	err := c.doJSON(ctx, http.MethodGet, "/v1/healthz", nil, nil, &out)
	return out, err
}

func (c *Client) Version(ctx context.Context) (VersionResponse, error) {
	var out VersionResponse
	err := c.doJSON(ctx, http.MethodGet, "/v1/version", nil, nil, &out)
	return out, err
}
