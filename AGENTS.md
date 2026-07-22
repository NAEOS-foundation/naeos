# NAEOS Development Guide

## Build & Test Commands
- Build: `go build ./cmd/naeos/`
- All tests: `go test -race -count=1 -timeout 300s ./...`
- Single package: `go test -race -count=1 -timeout 60s ./internal/broker/...`
- Linter: `golangci-lint run ./...`
- Format: `gofmt -w <file>`

## Code Conventions
- Use standard library patterns
- No external logging libraries
- Error handling: return errors, no panics
- Thread safety: use `sync.Mutex`/`sync.RWMutex` + `atomic.Int64`
- Constructor pattern: `New*()` returns struct pointer (not interface)
- Import ordering: stdlib → third-party → internal
- Config pattern: constructor takes config struct, validates and stores copy
- Logging: structured via middleware, not inline
- Testing: table-driven with `tt` struct, use `t.Parallel()` where safe
