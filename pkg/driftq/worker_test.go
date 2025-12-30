package driftq

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorker_AcksOnSuccess(t *testing.T) {
	var ackCount int32
	var nackCount int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/consume":
			w.Header().Set("Content-Type", "application/x-ndjson")
			w.WriteHeader(http.StatusOK)

			// one message then close stream
			_, _ = w.Write([]byte(`{"partition":0,"offset":1,"attempts":1,"key":"k","value":"v"}` + "\n"))
			return

		case "/v1/ack":
			atomic.AddInt32(&ackCount, 1)

			var req AckRequest
			_ = json.NewDecoder(r.Body).Decode(&req)

			if req.Topic != "demo" || req.Group != "demo" || req.Owner != "worker-1" {
				t.Fatalf("unexpected ack identity: %#v", req)
			}

			if req.Partition != 0 || req.Offset != 1 {
				t.Fatalf("unexpected ack position: %#v", req)
			}

			w.WriteHeader(http.StatusNoContent)
			return

		case "/v1/nack":
			atomic.AddInt32(&nackCount, 1)
			w.WriteHeader(http.StatusNoContent)
			return

		default:
			http.NotFound(w, r)
			return
		}
	}))
	defer srv.Close()

	c, err := Dial(context.Background(), Config{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	wk, err := NewWorker(WorkerConfig{
		Client: c,
		Consume: ConsumeOptions{
			Topic: "demo",
			Group: "demo",
			Owner: "worker-1",
		},
		Handler: StepFunc(func(ctx context.Context, msg ConsumeMessage) error {
			return nil // Ack
		}),
	})

	if err != nil {
		t.Fatalf("NewWorker: %v", err)
	}

	if err := wk.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := atomic.LoadInt32(&ackCount); got != 1 {
		t.Fatalf("expected 1 ack, got %d", got)
	}

	if got := atomic.LoadInt32(&nackCount); got != 0 {
		t.Fatalf("expected 0 nack, got %d", got)
	}
}

func TestWorker_NacksOnHandlerError(t *testing.T) {
	var ackCount int32
	var nackCount int32
	var lastNack NackRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/consume":
			w.Header().Set("Content-Type", "application/x-ndjson")
			w.WriteHeader(http.StatusOK)

			_, _ = w.Write([]byte(`{"partition":2,"offset":99,"attempts":1,"key":"k","value":"v"}` + "\n"))
			return

		case "/v1/ack":
			atomic.AddInt32(&ackCount, 1)
			w.WriteHeader(http.StatusNoContent)
			return

		case "/v1/nack":
			atomic.AddInt32(&nackCount, 1)

			_ = json.NewDecoder(r.Body).Decode(&lastNack)
			w.WriteHeader(http.StatusNoContent)
			return

		default:
			http.NotFound(w, r)
			return
		}
	}))
	defer srv.Close()

	c, err := Dial(context.Background(), Config{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	wk, err := NewWorker(WorkerConfig{
		Client: c,
		Consume: ConsumeOptions{
			Topic: "demo",
			Group: "demo",
			Owner: "worker-1",
		},
		Handler: StepFunc(func(ctx context.Context, msg ConsumeMessage) error {
			return errors.New("boom") // Nack
		}),
	})

	if err != nil {
		t.Fatalf("NewWorker: %v", err)
	}

	if err := wk.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := atomic.LoadInt32(&ackCount); got != 0 {
		t.Fatalf("expected 0 ack, got %d", got)
	}

	if got := atomic.LoadInt32(&nackCount); got != 1 {
		t.Fatalf("expected 1 nack, got %d", got)
	}

	if lastNack.Topic != "demo" || lastNack.Group != "demo" || lastNack.Owner != "worker-1" {
		t.Fatalf("unexpected nack identity: %#v", lastNack)
	}

	if lastNack.Partition != 2 || lastNack.Offset != 99 {
		t.Fatalf("unexpected nack position: %#v", lastNack)
	}

	if lastNack.Reason != "boom" {
		t.Fatalf("expected nack reason 'boom', got %q", lastNack.Reason)
	}
}

func TestWorker_HonorsEnvelopeDeadline(t *testing.T) {
	deadline := time.Now().Add(200 * time.Millisecond).UTC()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/consume":
			w.Header().Set("Content-Type", "application/x-ndjson")
			w.WriteHeader(http.StatusOK)

			m := ConsumeMessage{
				Partition: 0,
				Offset:    1,
				Attempts:  1,
				Key:       "k",
				Value:     "v",
				Envelope:  &Envelope{Deadline: &deadline},
			}
			b, _ := json.Marshal(m)
			_, _ = w.Write(append(b, '\n'))
			return

		case "/v1/ack":
			w.WriteHeader(http.StatusNoContent)
			return

		default:
			http.NotFound(w, r)
			return
		}
	}))
	defer srv.Close()

	c, err := Dial(context.Background(), Config{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	parentCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wk, err := NewWorker(WorkerConfig{
		Client: c,
		Consume: ConsumeOptions{
			Topic: "demo",
			Group: "demo",
			Owner: "worker-1",
		},
		Handler: StepFunc(func(ctx context.Context, msg ConsumeMessage) error {
			dl, ok := ctx.Deadline()
			if !ok {
				t.Fatalf("expected handler ctx to have a deadline")
			}

			if !dl.Equal(deadline) {
				t.Fatalf("expected deadline %v, got %v", deadline, dl)
			}
			return nil
		}),
	})
	if err != nil {
		t.Fatalf("NewWorker: %v", err)
	}

	if err := wk.Run(parentCtx); err != nil {
		t.Fatalf("Run: %v", err)
	}
}
