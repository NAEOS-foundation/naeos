# P2 Runtime Integration Contract — Supabase Control Plane

- Status: implementation sequencing note
- Scope: Issue #786, production runtime integration only
- Related ADR: [Production Control Plane Runtime and Supabase Persistence](../architecture/ADR-PRODUCTION-CONTROL-PLANE-SUPABASE-PERSISTENCE.md)

## Verified repository baseline

The PostgreSQL persistence components are present on `main`:

- `PostgresPolicyRepository`, `PostgresGrantRepository`, and `PostgresApprovalRepository` use `database/sql`.
- `PostgresLedger` uses `pgxpool` and exposes transactional `Append(ctx, event) (event, sequence, error)` plus integrity-checking `Load(ctx)`.
- The reusable `DecisionGateway` currently has concrete fields `*Evaluator`, `*Ledger`, and `*ApprovalStore`; its decision/execution flows call synchronous in-memory ledger APIs (`Append(event)`, `Decision(id)`, `HasExecution(id)`, and event query helpers).
- The demo composition root `newControlPlaneGateway` seeds in-memory policy state and constructs `NewLedger()`. The demo API router also contains demo reset/scenario routes.
- `internal/api.JWKSValidator` exists for the generic API path; it has not yet been wired to a dedicated production Control Plane router.

## Compatibility gap (do not paper over)

`PostgresLedger` is not a drop-in replacement for `*Ledger`: method signatures differ, it is context-aware, and its current API does not provide the gateway's canonical decision/execution query methods. The gateway also relies on evaluator access to an in-process policy store and execution locking. Directly substituting the PostgreSQL ledger or sharing the demo router would therefore be unsafe and incomplete.

## P2 implementation sequence

1. Define narrow, context-aware interfaces for production policy/grant/approval reads and ledger operations; keep existing demo-facing concrete APIs unchanged during migration.
2. Add production adapters for the interface boundary, including ledger lookups for canonical authorization decisions and prior execution state. A repository append error must produce an indeterminate result, not an ALLOW.
3. Build a dedicated production composition root that opens the single `DATABASE_URL` contract, applies TLS and bounded pool settings, pings the database, and validates required schema compatibility before binding the listener.
4. Load the active immutable policy and grant records from PostgreSQL for every authorization request; reject stale policy-version grants and missing/revoked/expired or scope-mismatched grants.
5. Register only explicitly allowlisted production routes. Do not mount investor demo reset, scripted scenario, demo execution, or seed routes.
6. Add liveness and readiness handlers. Readiness must be false if DB connectivity, schema compatibility, or required identity trust configuration is unavailable. Shut down the server and pools cleanly on cancellation.

## Required tests for this integration slice

- Interface conformance tests for the in-memory test doubles and PostgreSQL implementations.
- Authorization negative cases: absent grant, revoked/expired grant, agent mismatch, capability scope mismatch, and stale policy version.
- Persistence fail-closed: append/query failure never returns a usable ALLOW decision.
- Router allowlist: demo-only routes are not registered in the production mux.
- Startup failure: missing/invalid DATABASE_URL, DB ping failure, or missing/incompatible schema prevents listener startup.
- Readiness transition: database dependency failure flips readiness unhealthy; liveness remains independent.
- Shutdown: cancellation closes listener and database pools without leaking goroutines.

## Verification and rollout boundaries

- This note does not assert that runtime composition has been implemented.
- No production Supabase migration, role/credential change, Railway variable edit, or deployment is authorized by this change.
- Keep the PR draft until the integration code and CI evidence satisfy the acceptance criteria in Issue #786.
