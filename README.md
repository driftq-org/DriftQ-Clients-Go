# DriftQ-Clients-Go

Go SDK for **DriftQ** (the DriftQ-Core HTTP API).

What you get:
- **Produce** messages
- **Consume** via NDJSON streaming (`/v1/consume`)
- **Ack/Nack** for safe processing workflows
- **Admin** endpoints (topics, version, healthz, etc.)
- Client middleware for **retries**, **default deadlines**, and **trace context propagation**
- A minimal **Worker** loop (`Consume -> Handle -> Ack/Nack`) via `StepHandler`

---

## Requirements
- Go 1.25+ (see `go.mod`)
- DriftQ-Core running locally (or reachable over HTTP)

---

## Install
```bash
go get github.com/driftq-org/DriftQ-Clients-Go@latest
```

---

## Quick start

Dial a client:
```go
c, err := driftq.Dial(ctx, driftq.Config{
  BaseURL: "http://localhost:8080",
})
```

Produce:
```go
resp, err := c.Produce(ctx, driftq.ProduceRequest{
  Topic: "demo",
  Value: "hello",
  Envelope: &driftq.Envelope{
    // If set, Produce will attach Idempotency-Key so POST retries are safe.
    IdempotencyKey: "demo-001",
  },
})
_ = resp
```

Consume (stream):
```go
msgs, errs, err := c.ConsumeStream(ctx, driftq.ConsumeOptions{
  Topic: "demo",
  Group: "demo",
  Owner: "worker-1",
  LeaseMS: 30000,
})
if err != nil { /* ... */ }

for {
  select {
  case m := <-msgs:
    // process m ...
    _ = c.Ack(ctx, driftq.AckRequest{
      Topic: "demo", Group: "demo", Owner: "worker-1",
      Partition: m.Partition, Offset: m.Offset,
    })
  case err := <-errs:
    // stream-level error (unexpected read/decode) or nil on clean shutdown
    _ = err
  case <-ctx.Done():
    return
  }
}
```

---

## Client middleware

### Default deadlines (timeouts)
Non-streaming calls get a default timeout (`Config.Timeout`) **only if** the request context has no deadline.

- If your context has a deadline: that deadline wins.
- `ConsumeStream` explicitly opts out so streaming lifetime is controlled by `ctx` cancellation.

Disable the default timeout entirely:
```go
c, _ := driftq.Dial(ctx, driftq.Config{
  BaseURL: "http://localhost:8080",
  Timeout: -1,
})
```

### Retries
The client retries transient failures (network errors + `429/500/502/503/504`) with backoff.

Rules:
- **GET/HEAD/OPTIONS**: retries automatically.
- **POST/PUT/PATCH/DELETE**: retries only if `Idempotency-Key` is set.
  - `Produce` sets `Idempotency-Key` automatically when `envelope.idempotency_key` is present.

Configure:
```go
c, _ := driftq.Dial(ctx, driftq.Config{
  BaseURL: "http://localhost:8080",
  Retry: driftq.RetryConfig{
    MaxAttempts: 3,
    BaseDelay:   100 * time.Millisecond,
    MaxDelay:    2 * time.Second,
  },
})
```

### Tracing context propagation (OpenTelemetry)
If your app uses OpenTelemetry, this SDK will inject trace context into outgoing HTTP headers (e.g. `traceparent`).

- Propagation works by default.
- Optional: create a client span per request with `StartSpans: true`.

```go
c, _ := driftq.Dial(ctx, driftq.Config{
  BaseURL: "http://localhost:8080",
  Tracing: driftq.TracingConfig{
    Disable:    false,
    StartSpans: false,
  },
})
```

---

## Worker API (StepHandler)

Use `Worker` when you want the “classic worker” model:
**Consume -> run user code -> Ack/Nack**.

```go
wk, _ := driftq.NewWorker(driftq.WorkerConfig{
  Client: c,
  Consume: driftq.ConsumeOptions{
    Topic: "demo", Group: "demo", Owner: "worker-1",
    LeaseMS: 30000,
  },
  Concurrency: 4,
  Handler: driftq.StepFunc(func(ctx context.Context, msg driftq.ConsumeMessage) error {
    // return nil => Ack
    // return error => Nack(reason=err.Error())
    return nil
  }),
})

_ = wk.Run(ctx)
```

Notes:
- If a message includes `envelope.deadline`, the worker applies it as the handler context deadline.
- Handler errors are “expected” and result in Nack.
- Stream/transport errors are reported via `WorkerConfig.OnError` (if set) and will stop the run.

---

## Examples
See `examples/README.md` for how to run everything end-to-end.

Included examples:
- `examples/basic_produce_consume`
- `examples/basic_consume`
- `examples/basic_nack`
- `examples/admin_list`
- `examples/system_check`
- `examples/worker_basic`

---

## Versioning
- Follows Semantic Versioning (SemVer).
- Breaking API changes → new **minor** version while < v1.

---

## License
Licensed under the [Apache License 2.0](./LICENSE).
