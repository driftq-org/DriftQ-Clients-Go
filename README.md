# DriftQ-Clients-Go

Go SDK for **DriftQ** — produce and consume messages, and call admin APIs.

- Minimal, idiomatic Go
- Resilient connections with retries
- Works with `driftq-core` via gRPC
- No `buf` required for end users (generated stubs are vendored)

---

## Requirements
- Go 1.23+
- (Maintainers only) `buf` if you run `make proto`

---

## Install
```bash
go get github.com/BehnamAxo/driftq-clients-go@latest
```

---

## Usage
For complete, runnable examples, see the **`/examples`** directory in this repository:
- `examples/basic-produce-consume/`
- `examples/admin-list/`

These examples are built in CI to prevent drift.

---

## Configuration
See the godoc for `Config`, `Producer`, `Consumer`, and `Admin` types on pkg.go.dev.

---

## Generated Code
The SDK vendors generated stubs under `internal/gen/`.
Maintainers can refresh them with:
```bash
make proto   # requires buf; pulls latest from driftq-proto
```

---

## Versioning
- Follows Semantic Versioning (SemVer).
- Breaking API changes → new **minor** version while < v1.

---

## License
Licensed under the [Apache License 2.0](./LICENSE).
