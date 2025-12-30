package driftq

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type RoundTripperMiddleware func(next http.RoundTripper) http.RoundTripper

// ChainTransport applies middleware in the order provided:
// outermost is the first middleware in the list.
func ChainTransport(base http.RoundTripper, mws ...RoundTripperMiddleware) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	rt := base
	for i := len(mws) - 1; i >= 0; i-- {
		rt = mws[i](rt)
	}
	return rt
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// ---- Deadline middleware ----

type ctxKey int

const ctxKeyNoDefaultTimeout ctxKey = iota

// WithNoDefaultTimeout disables the client's default timeout middleware for this ctx so we
// can use it for long-lived streaming calls where the caller controls lifetime via ctx cancel
func WithNoDefaultTimeout(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeyNoDefaultTimeout, true)
}

func noDefaultTimeout(ctx context.Context) bool {
	v := ctx.Value(ctxKeyNoDefaultTimeout)
	b, _ := v.(bool)
	return b
}

// DeadlineMiddleware applies a default timeout only when:
// - defaultTimeout > 0
// - ctx has no deadline already
// - ctx does NOT opt out via WithNoDefaultTimeout
func DeadlineMiddleware(defaultTimeout time.Duration) RoundTripperMiddleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if defaultTimeout <= 0 {
				return next.RoundTrip(req)
			}
			if noDefaultTimeout(req.Context()) {
				return next.RoundTrip(req)
			}
			if _, ok := req.Context().Deadline(); ok {
				return next.RoundTrip(req)
			}

			ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
			defer cancel()

			return next.RoundTrip(req.Clone(ctx))
		})
	}
}

// ---- Retry middleware ----

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func (c RetryConfig) withDefaults() RetryConfig {
	if c.MaxAttempts == 0 {
		c.MaxAttempts = 3
	}

	if c.MaxAttempts < 0 {
		c.MaxAttempts = 1
	}

	if c.MaxAttempts < 1 {
		c.MaxAttempts = 1
	}

	if c.BaseDelay <= 0 {
		c.BaseDelay = 100 * time.Millisecond
	}

	if c.MaxDelay <= 0 {
		c.MaxDelay = 2 * time.Second
	}

	return c
}

var (
	jitterMu  sync.Mutex
	jitterRng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func jitter(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}

	// +/- 20%
	jitterMu.Lock()
	n := jitterRng.Float64()
	jitterMu.Unlock()

	factor := 0.8 + 0.4*n
	return time.Duration(float64(d) * factor)
}

func retryableStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func canRetryRequest(req *http.Request) bool {
	if isSafeMethod(req.Method) {
		return true
	}

	return req.Header.Get("Idempotency-Key") != ""
}

func isRetryableErr(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Many transport errors are wrapped; be permissive but bounded by MaxAttempts
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	return true
}

func parseRetryAfter(h string) (time.Duration, bool) {
	if h == "" {
		return 0, false
	}

	if secs, err := strconv.Atoi(h); err == nil {
		if secs <= 0 {
			return 0, true
		}
		return time.Duration(secs) * time.Second, true
	}

	if t, err := http.ParseTime(h); err == nil {
		d := time.Until(t)
		if d < 0 {
			d = 0
		}
		return d, true
	}

	return 0, false
}

func backoff(base, max time.Duration, attempt int) time.Duration {
	// attempt is 1 for the first retry delay, 2 for the second, etc.
	exp := float64(attempt - 1)
	d := time.Duration(float64(base) * math.Pow(2, exp))
	if d > max {
		d = max
	}

	return jitter(d)
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// RetryMiddleware retries transient failures for retryable requests.
// - Network errors: retry
// - Status codes: 429, 500, 502, 503, 504
//
// For unsafe HTTP methods, it only retries when Idempotency-Key is present
func RetryMiddleware(cfg RetryConfig) RoundTripperMiddleware {
	cfg = cfg.withDefaults()

	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if cfg.MaxAttempts <= 1 {
				return next.RoundTrip(req)
			}
			if !canRetryRequest(req) {
				return next.RoundTrip(req)
			}

			span := trace.SpanFromContext(req.Context())

			var lastResp *http.Response
			var lastErr error

			for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
				// Re-create body for retries if possible
				r := req
				if attempt > 1 {
					r = req.Clone(req.Context())
					if req.GetBody != nil {
						b, err := req.GetBody()
						if err != nil {
							return nil, err
						}
						r.Body = b
					}
				}

				resp, err := next.RoundTrip(r)
				lastResp, lastErr = resp, err

				if err == nil && resp != nil && !retryableStatus(resp.StatusCode) {
					return resp, nil
				}
				if err != nil && !isRetryableErr(err) {
					return nil, err
				}

				if attempt == cfg.MaxAttempts {
					if err != nil {
						return nil, err
					}
					return resp, nil
				}

				var wait time.Duration
				if resp != nil {
					if ra, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
						wait = ra
					} else {
						wait = backoff(cfg.BaseDelay, cfg.MaxDelay, attempt)
					}
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				} else {
					wait = backoff(cfg.BaseDelay, cfg.MaxDelay, attempt)
				}

				span.AddEvent("driftq.retry", trace.WithAttributes(
					attribute.Int("attempt", attempt),
					attribute.String("method", req.Method),
					attribute.String("url", req.URL.String()),
					attribute.Int64("wait_ms", wait.Milliseconds()),
					attribute.String("error", fmt.Sprintf("%v", err)),
					attribute.Int("status_code", func() int {
						if resp == nil {
							return 0
						}
						return resp.StatusCode
					}()),
				))

				if serr := sleepCtx(req.Context(), wait); serr != nil {
					return nil, serr
				}
			}

			// Unreachable; loop always returns
			_ = lastResp
			return lastResp, lastErr
		})
	}
}

// ---- Tracing middleware ----

type TracingConfig struct {
	Disable    bool
	StartSpans bool
	TracerName string
}

func TracingMiddleware(cfg TracingConfig) RoundTripperMiddleware {
	if cfg.Disable {
		return func(next http.RoundTripper) http.RoundTripper { return next }
	}

	if cfg.TracerName == "" {
		cfg.TracerName = "github.com/driftq-org/DriftQ-Clients-Go"
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			ctx := req.Context()

			var span trace.Span
			if cfg.StartSpans {
				ctx, span = otel.Tracer(cfg.TracerName).Start(ctx, spanName(req))
				defer span.End()
			}

			otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
			req = req.Clone(ctx)
			resp, err := next.RoundTrip(req)

			if cfg.StartSpans {
				span.SetAttributes(
					attribute.String("http.method", req.Method),
					attribute.String("http.url", req.URL.String()),
				)
				if resp != nil {
					span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
					if resp.StatusCode >= 400 {
						span.SetStatus(codes.Error, http.StatusText(resp.StatusCode))
					}
				}
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
				}
			}

			return resp, err
		})
	}
}

func spanName(req *http.Request) string {
	if req == nil || req.URL == nil {
		return "driftq.http"
	}
	return fmt.Sprintf("driftq.http %s %s", req.Method, req.URL.Path)
}
