package driftq

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ConsumeStream opens /v1/consume and decodes NDJSON items until ctx is cancelled
// or the server closes the stream.
//
// IMPORTANT: this intentionally does NOT use doJSON.
// Streaming lifetime must be controlled by ctx (or server-side shutdown), not a generic client timeout.
//
// NOTE: We explicitly disable the client's default timeout middleware here.
// If the caller wants a deadline, they should set it on ctx themselves.
func (c *Client) ConsumeStream(ctx context.Context, opt ConsumeOptions) (<-chan ConsumeMessage, <-chan error, error) {
	// Streaming calls must NOT be killed by the client's default timeout middleware.
	// If you want a deadline, set it on ctx yourself.
	ctx = WithNoDefaultTimeout(ctx)

	topic := strings.TrimSpace(opt.Topic)
	group := strings.TrimSpace(opt.Group)
	owner := strings.TrimSpace(opt.Owner)

	if topic == "" || group == "" || owner == "" {
		return nil, nil, errors.New("topic, group, and owner are required")
	}
	if opt.LeaseMS < 0 {
		return nil, nil, errors.New("lease_ms must be >= 0")
	}

	q := url.Values{}
	q.Set("topic", topic)
	q.Set("group", group)
	q.Set("owner", owner)
	if opt.LeaseMS > 0 {
		q.Set("lease_ms", strconv.Itoa(int(opt.LeaseMS)))
	}

	u := c.baseURL + "/v1/consume"
	if enc := q.Encode(); enc != "" {
		u += "?" + enc
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Accept", "application/x-ndjson")
	if ua := c.cfg.UserAgent; ua != "" {
		req.Header.Set("User-Agent", ua)
	}

	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var er ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&er) // best-effort
		return nil, nil, &APIError{
			Status:  resp.StatusCode,
			Code:    er.Error,
			Message: er.Message,
		}
	}

	msgs := make(chan ConsumeMessage)
	errs := make(chan error, 1)

	go func() {
		defer close(msgs)
		defer close(errs)
		defer resp.Body.Close()

		dec := json.NewDecoder(resp.Body)

		for {
			var m ConsumeMessage
			if err := dec.Decode(&m); err != nil {
				if errors.Is(err, io.EOF) || ctx.Err() != nil {
					return
				}
				select {
				case errs <- err:
				default:
				}
				return
			}

			select {
			case msgs <- m:
			case <-ctx.Done():
				return
			}
		}
	}()

	return msgs, errs, nil
}
