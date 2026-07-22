# ADR-006: Distributed Task Execution

> Status: Accepted
> Date: 2026-07-20

## Context

NAEOS pipeline tasks (code generation, compilation, validation) can be independently executed in parallel. For large projects with multiple modules and targets, a single process becomes a bottleneck. The project needs a distributed execution model where tasks are dispatched across workers on different machines. Key requirements:

- Priority-aware scheduling (critical path tasks first)
- Worker registration and health tracking via heartbeats
- Result aggregation with partial failure handling
- Graceful draining on shutdown

## Decision

Adopt a **Coordinator–Worker** architecture with an in-memory priority queue for task scheduling.

The `Coordinator` maintains:
- A `PriorityQueue` (implemented via `container/heap`) ordering tasks by priority and submission time
- An `AgentRegistry` tracking live workers, their capacities, and last heartbeat timestamps
- Task dispatch via a channel-based work queue

The `Worker` executes tasks and reports results back to the coordinator. Results are aggregated by the coordinator; any worker failure triggers task retry on an alternate worker.

A `LoadBalancer` selects the least-loaded available worker for each dispatch.

## Consequences

### Positive

- Pipeline execution scales horizontally by adding workers
- Priority scheduling ensures critical-path tasks (e.g., dependency resolution) execute before downstream tasks
- Heartbeat mechanism detects worker failures within 15s (3 missed beats at 5s interval)
- `Drain()` blocks until in-flight tasks complete, ensuring no work is lost during shutdown

### Negative

- The current coordinator is a single point of failure; there is no leader election or failover
- Task state is held in memory; coordinator restart loses all queued tasks
- Network latency between coordinator and workers adds overhead for fine-grained tasks

### Mitigations

- The coordinator is designed for colocated deployment (same LAN / same pod) where latency is sub-millisecond
- Tasks are designed to be coarse-grained (module-level code generation, not function-level) to amortize dispatch overhead
- A future leader-election mechanism (e.g., etcd, Redis) can make the coordinator highly available
- Pipeline result caching (`internal/pipelinecache/`) reduces the need to re-execute tasks across restarts

## Notes

- Implementation: `internal/distributed/` — coordinator, worker, priority queue, agent registry, load balancer
- Heartbeat interval: 5s with 3-strike timeout (15s)
- Related NES document: NES-045 (Distributed Task Execution)
