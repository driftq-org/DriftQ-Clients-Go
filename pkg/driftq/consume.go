package driftq

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) ConsumeStream(ctx context.Context, opt ConsumeOptions) (<-chan ConsumedMessage, <-chan error, error) {
	if opt.Topic == "" || opt.Group == "" || opt.Owner == "" {
		return nil, nil, fmt.Errorf("topic, group, and owner are required")
	}

	q := url.Values{}
	q.Set("topic", opt.Topic)
	q.Set("group", opt.Group)
	q.Set("owner", opt.Owner)
	if opt.LeaseMS > 0 {
		q.Set("lease_ms", fmt.Sprint(opt.LeaseMS))
	}

	u := c.baseURL + "/v1/consume?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/x-ndjson")

	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, nil, fmt.Errorf("server returned %s", resp.Status)
	}

	msgCh := make(chan ConsumedMessage)
	errCh := make(chan error, 1)

	go func() {
		defer resp.Body.Close()
		defer close(msgCh)
		defer close(errCh)

		sc := bufio.NewScanner(resp.Body)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 4*1024*1024)

		for sc.Scan() {
			line := sc.Bytes()
			if len(line) == 0 {
				continue
			}

			var m ConsumedMessage
			if err := json.Unmarshal(line, &m); err != nil {
				errCh <- fmt.Errorf("decode NDJSON: %w", err)
				return
			}

			select {
			case msgCh <- m:
			case <-ctx.Done():
				return
			}
		}

		if err := sc.Err(); err != nil && ctx.Err() == nil {
			errCh <- err
		}
	}()

	return msgCh, errCh, nil
}
