package driftq

import (
	"context"
	"net/http"
)

func (c *Client) Healthz(ctx context.Context) (*HealthzResponse, error) {
	var out HealthzResponse
	if err := c.doJSON(ctx, http.MethodGet, "/v1/healthz", nil, nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (c *Client) Version(ctx context.Context) (*VersionResponse, error) {
	var out VersionResponse
	if err := c.doJSON(ctx, http.MethodGet, "/v1/version", nil, nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}
