# ADR-005: WebSocket Communication

> Status: Accepted
> Date: 2026-07-20

## Context

NAEOS pipeline execution can be long-running (multi-minute code generation, compilation, cloud provisioning). The CLI and API clients need real-time progress updates without polling. Key requirements:

- Push pipeline lifecycle events (started, step-progress, completed, failed) to connected clients
- Support multiple concurrent observers per pipeline run
- Graceful handling of client disconnects and server shutdown
- Minimal overhead for single-client scenarios

## Decision

Use **WebSocket** as the real-time communication channel, bridged to the pipeline via a `PipelineObserver` interface.

The observer receives typed events (`PipelineStartedEvent`, `StepProgressEvent`, `PipelineCompletedEvent`, etc.). An `EventBroadcaster` holds a set of WebSocket connections and serializes events to JSON before broadcasting.

The broadcaster is thread-safe, uses a `sync.Mutex` for concurrent write access, and drains connections gracefully on server shutdown via a `done` channel.

## Consequences

### Positive

- Clients receive real-time progress without polling overhead
- The observer pattern decouples pipeline execution from the transport layer
- Multiple clients can observe the same pipeline run (dashboard + CLI simultaneously)
- WebSocket connections are lightweight; the server handles hundreds of concurrent observers

### Negative

- WebSocket requires persistent connections; load-balanced deployments need a sticky session or a shared event bus (e.g., Redis pub/sub)
- Clients behind restrictive proxies may not support WebSocket; a fallback SSE endpoint is needed
- Connection state management adds complexity: heartbeat pings, reconnection logic, backlog on reconnect

### Mitigations

- An SSE (Server-Sent Events) fallback endpoint is available at `/api/v1/events/stream`
- The observer pattern allows replacing the WebSocket transport with a shared event bus for clustered deployments
- A `draining` mechanism ensures no writes are attempted after shutdown begins, preventing panics on closed connections
- Heartbeat pings at 30s intervals detect dead connections; stale clients are cleaned up after 3 missed pings

## Notes

- Implementation: `internal/websocket/` — broadcaster, connection pool, observer bridge
- Integrated into pipeline via `pipelinemiddleware/` middleware chain
- Related NES document: NES-043 (WebSocket Real-time Communication)
