package driftq

import (
	"encoding/json"
	"fmt"
)

type HealthzResponse struct {
	Status string `json:"status"`
}

type VersionResponse struct {
	Version    string `json:"version"`
	Commit     string `json:"commit"`
	WalEnabled bool   `json:"wal_enabled"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("driftq api error: status=%d code=%q message=%q", e.Status, e.Code, e.Message)
}

// ---- Topics (Admin API) ----

// Topic supports BOTH server encodings:
// 1) "demo"
// 2) {"name":"demo"}
type Topic struct {
	Name string `json:"name"`
}

func (t *Topic) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		t.Name = s
		return nil
	}

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

type TopicsListResponse struct {
	Topics []Topic `json:"topics"`
}

type TopicsCreateRequest struct {
	Name       string `json:"name"`
	Partitions int    `json:"partitions,omitempty"`
}

type TopicsCreateResponse struct {
	Status     string `json:"status"`
	Name       string `json:"name"`
	Partitions int    `json:"partitions"`
}
