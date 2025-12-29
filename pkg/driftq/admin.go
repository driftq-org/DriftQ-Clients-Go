package driftq

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Topic (v1 minimal)
type Topic struct {
	Name string `json:"name"`
}

type ListTopicsResponse struct {
	Topics []Topic `json:"topics"`
}

type Admin struct{ c *Client }

func (c *Client) Admin() *Admin { return &Admin{c: c} }

func (t *Topic) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}

		t.Name = s
		return nil
	}

	// supports: {"name":"topic-name"}
	type topicObj struct {
		Name string `json:"name"`
	}

	var o topicObj
	if err := json.Unmarshal(b, &o); err != nil {
		return err
	}

	t.Name = o.Name
	return nil
}

func (a *Admin) ListTopics(ctx context.Context) (ListTopicsResponse, error) {
	var out ListTopicsResponse
	err := a.c.doJSON(ctx, http.MethodGet, "/v1/topics", nil, nil, &out)
	return out, err
}

func (a *Admin) CreateTopic(ctx context.Context, name string) error {
	q := url.Values{}
	q.Set("name", name)
	return a.c.doJSON(ctx, http.MethodPost, "/v1/topics", q, nil, nil)
}
