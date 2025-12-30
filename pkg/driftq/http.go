package driftq

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func (c *Client) doJSON(ctx context.Context, method, path string, q url.Values, in any, out any) error {
	// Apply per-request timeout ONLY for non-stream JSON calls
	if c.cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.cfg.Timeout)
		defer cancel()
	}

	u := c.baseURL + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}

	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var er ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&er) // best-effort
		return &APIError{
			Status:  resp.StatusCode,
			Code:    er.Error,
			Message: er.Message,
		}
	}

	// Important: ACK/NACK returns 204 No Content
	if resp.StatusCode == http.StatusNoContent || out == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
