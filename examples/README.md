# Examples

These are runnable examples for `DriftQ-Clients-Go`.

## Prerequisite: run DriftQ-Core
Start DriftQ-Core in a separate terminal so the HTTP API is reachable at:

- `http://localhost:8080`

Quick sanity check (any browser or curl):
- `GET /v1/healthz`

## Run examples

### 1) Produce + consume (one-shot)
```bash
go run ./examples/basic_produce_consume
```

### 2) Consume (stream)
```bash
go run ./examples/basic_consume
```

### 3) Nack example
```bash
go run ./examples/basic_nack
```

### 4) Admin: list topics
```bash
go run ./examples/admin_list
```

### 5) System check
```bash
go run ./examples/system_check
```

### 6) Worker (StepHandler)
This is the “worker loop” example using `Worker` + `StepFunc`:

```bash
go run ./examples/worker_basic
```

What it does:
- opens `/v1/consume` NDJSON stream
- for each message, calls your handler
- handler returns `nil` => `/v1/ack`
- handler returns `error` => `/v1/nack`

Stop with `Ctrl+C`.
