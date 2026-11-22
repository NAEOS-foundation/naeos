# Evidence Verifier — Replay Protection Contract

Status: implementation contract; durable adapter not yet provided
Related: #870, PR #862, PR #871

## 1. Purpose

The verifier accepts a caller-generated 128-bit request identifier and returns a receipt bound to that identifier. Validation and echoing alone do not provide replay protection. Before the verifier queries an authoritative evidence source, it must atomically claim the identifier in durable storage.

This contract protects the verifier request/receipt boundary. It does not provide exactly-once provider observation or prove that an external side effect occurred once.

## 2. Identity and scope

A replay key is the tuple:

- verifier service scope: a stable configured identifier shared by every replica serving the same logical verifier;
- request ID: the normalized 128-bit hexadecimal identifier from the request.

The service scope must not be a process ID, pod name, or ephemeral deployment instance. Separate logical verifier services may use separate scopes only when their receipt trust domains are intentionally distinct.

The request ID is an opaque nonce, not an authorization credential. Authentication and all request binding validation happen before it is claimed.

## 3. Atomic claim semantics

The storage boundary must provide an operation equivalent to:

```go
type EvidenceVerificationReplayGuard interface {
    Claim(ctx context.Context, scope, requestID string) (claimed bool, err error)
}
```

Required behavior:

1. Claim is an atomic insert-if-absent on the unique pair (scope, request ID).
2. Exactly one concurrent caller may receive claimed=true across all replicas.
3. An existing key returns claimed=false without changing the original claim.
4. Storage errors, timeouts, and ambiguous commit outcomes return an error. The service must not call the evidence source after any such error.
5. A claim is never released because the source timed out, returned an inconclusive result, the caller disconnected, or the response write failed. Releasing it could allow a retry to be mistaken for a new verification attempt.
6. The claim is recorded before calling AuthoritativeExecutionEvidenceSource.ObserveExecution.
7. A duplicate is rejected with a non-confirming conflict response; it must never produce a VERIFIED receipt.

The uniqueness constraint is the serialization primitive. A preflight SELECT, process-local mutex, in-memory map, or cache is insufficient.

## 4. Retention

No replay-retention duration is currently approved for this protocol. Therefore an implementation must not silently expire claims. The initial durable adapter must preserve uniqueness for the lifetime of the verifier scope.

A future bounded-retention policy requires an explicit protocol decision covering maximum request acceptance age, receipt validity, caller retry horizon, clock skew, and tombstone retention. Expiring a claim must never make an old request ID valid again inside the same receipt trust domain.

## 5. Failure mapping

| Condition | Required behavior |
|---|---|
| Invalid request/authentication | Reject before claim; source is not called |
| First successful claim | Continue to authoritative observation |
| Duplicate claim | Conflict; source is not called |
| Claim storage unavailable | 503 / UNKNOWN; source is not called |
| Claim result ambiguous | 503 / UNKNOWN; source is not called |
| Provider inconclusive or unavailable after claim | Non-confirming result; claim remains consumed |
| Process restart / another replica | Existing claim remains effective |

A 503 or UNKNOWN result must not be interpreted as proof that the provider outcome was committed or not committed.

## 6. Implementation boundary

The HTTP service depends on the interface above. Unit tests may use a synchronized in-memory fake to test service behavior, but that fake is test-only and cannot satisfy production readiness.

A production adapter requires a dedicated durable store with a unique constraint on (scope, request_id), bounded transaction deadlines, typed conflict versus infrastructure errors, and integration tests for concurrent replicas, restart persistence, and ambiguous commit handling.

Do not repurpose the existing REST API v2 idempotency table, message outbox, or control-plane decision ledger. The action-ledger architecture specifies a separate invocation store and distinct record semantics. Any reuse of shared PostgreSQL infrastructure must still use a separately reviewed table, repository boundary, and least-privilege role.

## 7. Acceptance tests

- Two concurrent claims for the same scope and request ID yield exactly one winner.
- The same request ID in distinct explicitly configured scopes does not collide.
- A duplicate request is rejected before the evidence source is called.
- A storage error or ambiguous claim result is fail-closed and the source is not called.
- A provider error after a successful claim does not release the key.
- A claim remains effective after adapter/process restart.
- A production readiness claim is blocked until a durable adapter passes PostgreSQL integration and failure-injection tests.
