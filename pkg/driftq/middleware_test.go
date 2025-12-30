package driftq

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestRetryMiddleware_RetriesTransientGET(t *testing.T) {
	var hits int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/healthz" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		c := atomic.AddInt32(&hits, 1)
		if c <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "UNAVAILABLE", Message: "try again"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthzResponse{Status: "ok"})
	}))
	defer srv.Close()

	cli, err := Dial(context.Background(), Config{BaseURL: srv.URL, Retry: RetryConfig{MaxAttempts: 3}})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	resp, err := cli.Healthz(context.Background())
	if err != nil {
		t.Fatalf("Healthz: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("unexpected status: %#v", resp)
	}
	if atomic.LoadInt32(&hits) != 3 {
		t.Fatalf("expected 3 attempts, got %d", atomic.LoadInt32(&hits))
	}
}

func TestDeadlineMiddleware_EnforcesDefaultTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthzResponse{Status: "ok"})
	}))
	defer srv.Close()

	cli, err := Dial(context.Background(), Config{BaseURL: srv.URL, Timeout: 50 * time.Millisecond, Retry: RetryConfig{MaxAttempts: 1}})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	_, err = cli.Healthz(context.Background())
	if err == nil {
		t.Fatalf("expected timeout error")
	}
}

func TestDeadlineMiddleware_DoesNotApplyToConsumeStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/consume" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		fl, _ := w.(http.Flusher)
		time.Sleep(150 * time.Millisecond)
		_, _ = w.Write([]byte(`{"partition":0,"offset":1,"attempts":1,"key":"k","value":"v"}` + "\n"))
		if fl != nil {
			fl.Flush()
		}

		time.Sleep(150 * time.Millisecond)
	}))
	defer srv.Close()

	cli, err := Dial(context.Background(), Config{BaseURL: srv.URL, Timeout: 50 * time.Millisecond, Retry: RetryConfig{MaxAttempts: 1}})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgs, errs, err := cli.ConsumeStream(ctx, ConsumeOptions{Topic: "t", Group: "g", Owner: "o"})
	if err != nil {
		t.Fatalf("ConsumeStream: %v", err)
	}

	select {
	case m := <-msgs:
		if m.Value != "v" {
			t.Fatalf("unexpected message: %#v", m)
		}
		cancel()
	case e := <-errs:
		if e != nil {
			t.Fatalf("unexpected stream error: %v", e)
		}
	case <-time.After(700 * time.Millisecond):
		t.Fatalf("timed out waiting for stream message")
	}
}

func TestTracingMiddleware_InjectsTraceparent(t *testing.T) {
	old := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	defer otel.SetTextMapPropagator(old)

	var gotTraceparent atomic.Value

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceparent.Store(r.Header.Get("traceparent"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthzResponse{Status: "ok"})
	}))
	defer srv.Close()

	cli, err := Dial(context.Background(), Config{BaseURL: srv.URL, Retry: RetryConfig{MaxAttempts: 1}})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	// Seed ctx with a valid parent SpanContext
	parent := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		SpanID:     trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), parent)

	_, err = cli.Healthz(ctx)
	if err != nil {
		t.Fatalf("Healthz: %v", err)
	}

	v := gotTraceparent.Load()
	if v == nil || v.(string) == "" {
		t.Fatalf("expected traceparent header to be injected")
	}
}
