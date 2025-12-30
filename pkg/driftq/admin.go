package driftq

import (
	"context"
	"net/http"
)

type Admin struct{ c *Client }

func (c *Client) Admin() *Admin { return &Admin{c: c} }

func (a *Admin) ListTopics(ctx context.Context) (TopicsListResponse, error) {
	var out TopicsListResponse
	err := a.c.doJSON(ctx, http.MethodGet, "/v1/topics", nil, nil, &out)
	return out, err
}

// partitions: pass 0 to let server default to 1
func (a *Admin) CreateTopic(ctx context.Context, name string, partitions int) (TopicsCreateResponse, error) {
	in := TopicsCreateRequest{Name: name, Partitions: partitions}
	var out TopicsCreateResponse
	err := a.c.doJSON(ctx, http.MethodPost, "/v1/topics", nil, in, &out)
	return out, err
}
