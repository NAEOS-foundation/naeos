# Production Action Ledger — Persistence Contract

Status: proposed design for Production Authorization & Execution Gateway
Parent: #848
Related: PR #850 — Production Authorization & Execution Gateway

## 1. Purpose and authority

The action ledger is the durable source of truth for an execution invocation and its lifecycle. It is distinct from:

- the authorization decision record, which answers whether a specific action is permitted;
- the message outbox, which schedules delivery to workers;
- runtime observations, which report what an external executor actually did.

A committed authorization decision is not an execution intent. A committed execution intent is not proof that the external side effect happened. A timeout after dispatch is not proof of failure.

The ledger must support recovery across process restarts and concurrent gateway replicas. In-process mutexes and process-local maps are not correctness mechanisms.

## 2. Existing repository evidence and compatibility

The current repository provides useful but incomplete foundations:

- `internal/controlplane.LedgerRepository` persists canonical append-only control-plane evidence.
- `PostgresLedger.Append` serializes the hash-chain head update through the database append function and commits before returning.
- `PostgresLedger.Load` verifies sequence continuity, predecessor hashes, event hashes, and the authoritative head in a repeatable-read snapshot.
- The existing ledger explicitly routes reconciliation events through a dedicated workflow; generic append rejects `EXECUTION_RECONCILED`.
- `internal/api/v2_idempotency.go` is HTTP response replay protection, not a safe execution ledger. It uses a globally scoped key, TTL deletion, and its claim path does not reliably distinguish uniqueness conflicts from database errors.
- `internal/messagequeue/outbox.go` is a delivery queue. Its unique key and row leases are useful patterns, but outbox state must not become the authoritative execution outcome.

Do not repurpose the existing API idempotency table or outbox as the gateway action ledger. Preserve existing public contracts until callers and compatibility tests are traced.

## 3. Logical data model

The proposed model has two logical records. They may share the existing private PostgreSQL cluster, but must have separate tables, repository methods, and least-privilege grants from policy administration.

### 3.1 Invocation aggregate

One row per canonical invocation:

| Field | Meaning |
|---|---|
| invocation_id | Server-generated immutable identifier; primary key |
| tenant_id | Authenticated tenant boundary |
| principal_id | Authenticated requesting principal/service identity |
| idempotency_key | Client-supplied opaque key, scoped to tenant and principal |
| request_digest | Digest of canonical action intent and target |
| canonicalization_version | Exact canonicalization algorithm/version |
| action_type | Registered action/tool identifier |
| target_digest | Digest or stable identifier of the target; avoid secret-bearing raw values |
| decision_id | Immutable reference to the committed authorization decision |
| decision_policy_id/version | Policy identity and version used by the decision |
| state | Current authoritative lifecycle state |
| state_version | Monotonic optimistic-concurrency version |
| created_at / updated_at | Database-generated UTC timestamps |
| dispatch_attempt | Monotonic dispatch attempt counter |
| evidence_reference | Optional reference to separately stored evidence; not a substitute for the event stream |

Required uniqueness:

- Primary key on `invocation_id`.
- Unique `(tenant_id, principal_id, idempotency_key)`.
- Unique execution binding where the policy requires at most one invocation for a decision; do not assume decision ID alone is globally unique across tenants.
- Foreign-key or validated immutable reference to the canonical decision record, according to the existing ledger schema and transaction boundary.

### 3.2 Append-only lifecycle events

One row per lifecycle transition/evidence event:

| Field | Meaning |
|---|---|
| event_id | Immutable event identifier; primary key |
| invocation_id | Parent invocation |
| sequence | Per-invocation monotonic sequence |
| event_type | Typed lifecycle event |
| prior_state / next_state | Explicit transition |
| actor_type / actor_id | Authenticated service or principal responsible |
| attempt_id | Stable dispatch attempt identifier, when applicable |
| payload_digest | Digest of canonical, non-secret event payload |
| occurred_at | Database-generated UTC timestamp |
| correlation_id | Trace/correlation reference |
| evidence_digest/reference | Evidence reference for observation/reconciliation |

Required uniqueness: `(invocation_id, sequence)` and `event_id`. Events are append-only; application roles must not UPDATE or DELETE them. A correction is a new compensating/reconciliation event, never mutation of historical evidence.

Raw credentials, bearer tokens, authorization headers, private keys, and secret-bearing payloads must never be stored in either record. Store bounded, redacted metadata and cryptographic digests where appropriate.

## 4. Idempotent intent-creation transaction

The repository operation should accept a fully authenticated tenant/principal scope, idempotency key, canonical request digest and version, committed decision reference, and immutable invocation metadata.

Within one PostgreSQL transaction:

1. Validate required fields and supported canonicalization version before opening the transaction.
2. Insert the invocation aggregate with its initial state and unique scoped idempotency binding.
3. Append the initial `INTENT_RECORDED` event at sequence 1.
4. Commit both records atomically.
5. Return the canonical committed invocation only after successful commit.

Concurrent duplicate handling:

- If the scoped unique idempotency constraint wins, return the new committed invocation.
- If it conflicts, load the existing canonical invocation from the authoritative database.
- If the stored request digest and canonicalization version match, return the existing invocation and its current lifecycle state; do not dispatch again.
- If either differs, return a typed idempotency conflict. Never replace the original request or decision binding.
- If the database cannot be read after a conflict, return an infrastructure/indeterminate error. Do not guess that the request is new.
- Do not delete idempotency bindings merely because a response cache TTL elapsed. Retention/archival is a separate reviewed policy and must not permit a key to authorize a second side effect.

The uniqueness constraint—not a preflight SELECT and not an in-process mutex—is the cross-replica serialization primitive.

## 5. State transition and concurrency contract

All state transitions must be transactional compare-and-set operations against the expected `state_version` and expected prior state. The transition and append-only event insertion commit together.

Initial lifecycle vocabulary (final enum remains subject to Stage 4/5 review):

- `INTENT_RECORDED`
- `READY_TO_DISPATCH`
- `DISPATCHING`
- `SUCCEEDED`
- `FAILED_CONFIRMED`
- `UNKNOWN`
- `RECONCILING`
- `RECONCILED_SUCCEEDED`
- `RECONCILED_FAILED`
- `CANCELLED_BEFORE_DISPATCH`

Terminal state rules and legal transition graph must be explicit in code and database constraints where feasible. A stale writer must receive a conflict and must not append an event that claims a transition it did not win.

A lease may coordinate workers, but lease expiry alone must never prove that a side effect did not happen. A worker recovering an expired `DISPATCHING` invocation must move it to `UNKNOWN` or a reconciliation-required state, not silently re-dispatch.

## 6. Proposed repository boundary

The gateway repository should expose operations with context-aware, typed errors, conceptually:

```go
type ActionLedgerRepository interface {
    CreateIntent(ctx context.Context, input CreateIntentInput) (Invocation, error)
    GetInvocation(ctx context.Context, tenantID, invocationID string) (Invocation, error)
    AppendTransition(ctx context.Context, input TransitionInput) (Invocation, error)
    ClaimDispatch(ctx context.Context, input DispatchClaimInput) (DispatchClaim, error)
    AppendObservation(ctx context.Context, input ObservationInput) (Invocation, error)
    LoadEvents(ctx context.Context, tenantID, invocationID string) ([]InvocationEvent, error)
}
```

This is a design sketch, not an approved exported Go API. Exact package placement and types must be reconciled with the Stage 2 service split and current gateway callers before implementation.

Repository errors must distinguish at least: not found, idempotency conflict, state conflict, unavailable, integrity failure, and unsupported schema/canonicalization version. Infrastructure or integrity errors must never be converted into authorization DENY, an empty result, or success.

## 7. Authorization and execution boundary

- Authorization service remains the only authority that evaluates policy and emits a canonical decision.
- Intent creation accepts only a committed decision reference and verifies that the decision is bound to the exact tenant, principal, action, target digest, policy version, and validity window.
- Execution service owns dispatch lifecycle, not policy authoring.
- Immediately before dispatch, revalidate the decision/grant/revocation/policy binding through the approved authorization contract.
- Persist the dispatch claim/attempt before crossing the external side-effect boundary.
- Persist observations only from a trusted executor/observer path; agent/model claims are not execution evidence.
- UNKNOWN blocks automatic retry until an authoritative reconciliation workflow resolves the original invocation.
- The ledger's database role must not be able to create or activate policy versions, modify grants, or manufacture authorization decisions.

## 8. Security and operational requirements

- Tenant and principal scope are mandatory on every read/write API; do not rely on caller-supplied invocation IDs as authorization.
- Runtime DB role receives only the minimum required INSERT/SELECT/transition privileges on action-ledger tables and required narrowly scoped functions.
- No UPDATE/DELETE on append-only event rows for the runtime role.
- Policy administration, grant administration, reconciliation authority, and runtime dispatch identities remain separately authorized.
- Use database-generated timestamps and bounded transaction/request deadlines.
- Use parameterized SQL and explicit transaction boundaries.
- Avoid logging idempotency keys, raw payloads, tokens, or secret-bearing target values.
- Add indexes for tenant-scoped invocation lookup, scoped idempotency lookup, and ordered per-invocation event reads.
- Define retention, archival, backup, restore, and tamper-evidence requirements before production rollout.
- Do not hold a database transaction open across arbitrary network calls or external side effects.

## 9. Required verification before implementation is accepted

PostgreSQL integration tests must use an ephemeral isolated database and exercise:

1. Two concurrent identical CreateIntent requests produce one invocation and one initial event.
2. Same scoped key with a different request digest/version returns conflict and leaves original binding unchanged.
3. Same key under a different tenant or principal does not collide.
4. Database failure during event insert rolls back the invocation insert.
5. Commit ambiguity is recovered by querying the scoped idempotency binding; no blind second dispatch.
6. Concurrent state transitions have exactly one winner and one corresponding event.
7. Expired worker lease does not trigger automatic redispatch of a possibly executed side effect.
8. Runtime role cannot update/delete lifecycle events or mutate policy/grants.
9. Tenant-scoped reads cannot disclose another tenant's invocation.
10. Restart/reconnect preserves invocation state and event ordering.
11. Tampered payload, sequence gap, invalid transition, unsupported canonicalization version, and missing decision reference fail closed.
12. Cancellation and timeout preserve a typed indeterminate outcome where commit status cannot be established.

Memory adapters may support unit tests, but cannot satisfy durability, concurrency, or cross-process acceptance criteria.

## 10. Rollout and explicit non-goals

This stage is a persistence contract only. It does not authorize:

- creating or applying SQL migrations;
- changing Supabase schemas or database roles;
- provisioning Railway services, variables, or secrets;
- changing production or staging traffic;
- enabling external tool execution;
- weakening or bypassing existing security gates.

Before DDL or implementation, review the current private schema and migration conventions, confirm staging database ownership/role topology, map the existing decision and reconciliation transactions, and approve the exact privilege matrix. Then implement behind integration tests and a staging-only rollout plan.

## 11. Acceptance decision

Stage 3 design is ready for architecture review when reviewers can confirm:

- scoped uniqueness is the authoritative idempotency boundary;
- intent plus initial event are atomic;
- lifecycle writes are compare-and-set and append-only;
- authorization records and execution records remain separate authorities;
- UNKNOWN cannot be retried without authoritative reconciliation;
- least-privilege database roles prevent policy authoring through runtime credentials.

No production readiness claim follows from this design document.
