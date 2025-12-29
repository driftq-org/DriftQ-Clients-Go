package driftq

import "fmt"

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
