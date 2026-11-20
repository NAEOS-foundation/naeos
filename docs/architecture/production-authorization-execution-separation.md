# Stage 2 Design — Authorization Decision / Execution Lifecycle Separation

**Status:** Draft companion to ADR-001; design-only  
**Date:** 2026-11-20  
**Parent:** PR #850 / Issue #848  
**Scope:** Define service responsibilities and invariants before changing runtime code.

## 1. Objective

Separate the act of deciding whether a specific request is authorized from the lifecycle that durably accepts, dispatches, observes, and reconciles that request.

This is a separation of authority and side effects, not a second policy engine and not a claim that the current runtime APIs already provide these boundaries.

## 2. Existing-code constraints

The current code has three distinct concepts that must not be conflated:

1. `internal/governance/control.ControlPlane.Evaluate` evaluates governance policy and produces a `DecisionRecord`. `ValidateDecision` re-evaluates and verifies grant digest, expiry, revocation, and policy identity. This is the best existing policy decision contract for the runtime gateway path.
2. `internal/runtime/gateway.ExecutionGateway.Authorize` currently performs both authorization and sandbox execution in one method. It also performs final revalidation, adapter grant checks in adapter entry points, replay claim, sandbox dispatch, and observation. This is a guarded execution path, but its method name does not mean it is decision-only.
3. `internal/controlplane.DecisionGateway.Authorize` records a Control Plane decision; `ExecuteDecision`/ `ExecuteAtomic` are separate execution-authorization/intent callback primitives. They use a different action/grant model and must not silently become a second policy authority for the governance runtime path.

Do not change the semantics of existing public methods in place without tracing all callers and compatibility tests. The refactor should introduce explicit internal service contracts and migrate callers deliberately.

## 3. Responsibility boundaries

### A. Authorization Decision Service

Responsibilities:
- Accept an authenticated principal context supplied by trusted middleware, never an actor identity asserted by the request body.
- Validate the normalized action/resource/environment/capability request and canonicalization version.
- Evaluate exactly one authoritative governance policy contract.
- Resolve the active policy and applicable grant/approval references.
- Issue an immutable, expiring decision bound to the canonical request digest and verified principal/tenant.
- Revalidate that exact decision immediately before a dispatch is authorized.
- Return ALLOW, DENY, REQUIRE_APPROVAL, or INDETERMINATE with reason codes and evidence references.

Must not:
- invoke an adapter, sandbox, tool, queue consumer, or external provider;
- create a dispatch attempt;
- claim execution completed or observed;
- mutate the action lifecycle beyond writing authorization evidence.

Proposed internal interface (conceptual; names are not committed API):

```go
type AuthorizationDecisionService interface {
    Evaluate(ctx context.Context, req AuthorizationRequest) (AuthorizationDecision, error)
    Revalidate(ctx context.Context, decision AuthorizationDecision, req AuthorizationRequest) (AuthorizationDecision, error)
}
```

A transport/API layer may expose these operations later, but endpoint paths, HTTP methods, and status mapping remain unspecified until identity and deployment topology are approved.

### B. Execution Lifecycle Service

Responsibilities:
- Accept an immutable, validated execution intent and a stable idempotency binding.
- Persist the intent and idempotency claim atomically before dispatch.
- Invoke Authorization Decision Service for final exact-digest revalidation at the last controllable boundary.
- Dispatch only through the guarded runtime gateway / registered and explicitly granted adapter.
- Persist dispatch attempt, executor receipt, independent observation, and lifecycle transitions.
- Return the canonical existing lifecycle for duplicate same-key/same-digest submissions.
- Reject same-key/different-digest submissions as conflict.
- Mark ambiguous post-dispatch outcomes UNKNOWN and enqueue/record reconciliation without blind side-effect retry.
- Permit reconciliation only from authenticated authoritative evidence or a governed operator-resolution workflow.

Must not:
- implement policy matching independently;
- infer ALLOW from registration, approval existence, cached state, or an earlier decision;
- bypass adapter/tool grants or call raw delegates;
- rewrite terminal history or convert UNKNOWN to success without evidence.

Proposed internal interface (conceptual; names are not committed API):

```go
type ExecutionLifecycleService interface {
    SubmitIntent(ctx context.Context, intent ExecutionIntent) (ExecutionSnapshot, error)
    Dispatch(ctx context.Context, invocationID string) (ExecutionSnapshot, error)
    Get(ctx context.Context, invocationID string) (ExecutionSnapshot, error)
    Reconcile(ctx context.Context, invocationID string) (ExecutionSnapshot, error)
}
```

The implementation may split worker dispatch and reconciliation into separate components. These operations are not required to be synchronous HTTP calls.

## 4. Trust and call sequence

1. Transport middleware authenticates workload credentials and produces a verified principal/tenant context.
2. Request validation/canonicalization creates a request digest bound to principal, tenant, action, tool/adapter version, arguments, scope, environment, and relevant context.
3. Authorization Decision Service evaluates the request and records decision evidence. A DENY, REQUIRE_APPROVAL, or INDETERMINATE result is terminal for that attempt and must not reach dispatch.
4. Execution Lifecycle Service atomically persists the immutable intent and idempotency binding. Persistence failure means no dispatch.
5. Immediately before dispatch, Execution Lifecycle Service calls decision revalidation with the exact persisted digest and verifies decision expiry, active policy, grant, approval, principal/tenant, target, adapter version, and lifecycle state.
6. The guarded runtime gateway checks its own explicit adapter/tool/action grant and dispatches once using a stable dispatch-attempt identity.
7. Executor receipt and independent observation are persisted. If the process cannot establish whether dispatch occurred or whether the side effect completed, the invocation is UNKNOWN.
8. Reconciliation uses the provider/runtime's authoritative lookup or signed receipt; absent conclusive evidence, UNKNOWN remains.

The decision record is an input to lifecycle authorization, not an execution command or bearer capability that can be widened or replayed.

## 5. State ownership

| Concern | Owner | Persistence |
|---|---|---|
| Policy evaluation result | Authorization Decision Service | Existing decision/evidence ledger, subject to verified durability |
| Policy/grant/approval freshness | Authorization Decision Service | Authoritative policy/grant/approval repositories |
| Idempotency key claim | Execution Lifecycle Service | New durable invocation store with unique constraint |
| Immutable intent | Execution Lifecycle Service | Same transaction as idempotency claim |
| Dispatch attempt | Execution Lifecycle Service / guarded executor | Durable lifecycle event/record |
| Side-effect receipt | Executor/provider, ingested by lifecycle service | Authenticated receipt reference and digest |
| Observation | Independent observer | Durable observation evidence |
| UNKNOWN resolution | Reconciliation component | Append-only linked evidence |

The authorization decision ledger and execution lifecycle ledger may share a physical PostgreSQL cluster but must retain distinct record semantics and least-privilege access boundaries. Do not conflate `EXECUTION_ALLOWED`, `EXECUTION_STARTED`, sandbox return, provider receipt, and independently observed effect.

## 6. Required invariants

- Decision evaluation has zero side effects.
- A non-ALLOW or operationally indeterminate decision has zero dispatch attempts.
- A decision is valid only for its exact immutable request digest and verified principal/tenant.
- Authorization is revalidated at the last controllable dispatch boundary.
- Durable intent and idempotency claim commit before dispatch.
- Same key + same digest yields one canonical invocation and at most one dispatch attempt unless a provider-specific idempotency guarantee proves safe retry.
- Same key + different digest is a conflict.
- No cross-process correctness claim is based on a process-local mutex.
- Any ambiguous post-dispatch outcome is UNKNOWN; no blind retry.
- UNKNOWN is resolved only by authoritative evidence; inconclusive reconciliation remains UNKNOWN.
- All adapters are explicitly granted and can reach execution only through the guarded gateway.
- Policy administration must not be implicitly granted to the runtime identity.

## 7. Migration strategy

1. Keep current public runtime gateway behavior unchanged while introducing tests that characterize current authorization, revalidation, adapter grants, replay, sandbox, and observer semantics.
2. Extract a decision-only service around the governance control plane. It must not call the sandbox or mutate invocation state.
3. Introduce an execution lifecycle coordinator that accepts only immutable intent plus decision reference/digest and has no policy evaluator of its own.
4. Add a storage-backed invocation repository with transactional idempotency and intent creation. This is a later implementation stage; no schema is defined in this design note.
5. Route the coordinator through the existing guarded runtime gateway. If the gateway API must be split, add an explicit decision-only method and a separately named execution method, preserve compatibility wrappers, and migrate call sites with tests.
6. Prove final revalidation and policy-update concurrency against the selected database/worker topology.
7. Add process-crash and provider-response-loss integration tests before enabling external dispatch.

## 8. Blocking security and topology gates

The parent Issue #848 contains findings that must be resolved before runtime implementation or staging deployment:

- The current `naeos_control_plane_runtime` database group can INSERT policy snapshots and mutate the active policy pointer. This is policy-authoring authority, not merely runtime evaluation. Decide and implement a separate policy-administration role or document an explicit threat-model exception with narrowly scoped service identities. Preferred direction: runtime identity is read-only for policy snapshots and active pointer; policy administration uses a separate trusted control path.
- Grant and approval write permissions on the runtime group also require explicit ownership, endpoint authorization, and audit-attribution review.
- Staging Supabase schema is installed, but a dedicated staging LOGIN runtime role and a dedicated Railway staging service are not demonstrated. PR preview is not staging.
- Railway Control Plane runtime variables/healthcheck and live service readiness remain unverified. Do not use production or a PR preview to claim staging evidence.
- Production Supabase schema is absent; no production migration is authorized by this document.

No DDL, role provisioning, secret changes, Railway configuration, deployment, or production operation is authorized by this design note.

## 9. Acceptance tests for Stage 2

1. Calling Evaluate creates a decision/evidence record but does not invoke sandbox, adapter, provider, or execution store.
2. DENY / REQUIRE_APPROVAL / INDETERMINATE produce zero invocation/dispatch records.
3. A decision bound to request digest A cannot authorize digest B.
4. A decision for principal/tenant A cannot be used by principal/tenant B.
5. Expired, revoked, stale-policy, stale-grant, substituted-approval decisions fail final revalidation.
6. A policy mutation between evaluation and dispatch prevents dispatch.
7. A failed durable intent/idempotency transaction prevents dispatch.
8. Adapter registration without explicit tool/action grant prevents dispatch.
9. Existing public gateway behavior remains compatible during migration.
10. No test treats an authorization decision, dispatch start, sandbox output, or provider response alone as independent observation.

## 10. Exit criteria

Stage 2 is complete only when the service responsibilities and interfaces are approved, package ownership is unambiguous, compatibility strategy is documented, and tests demonstrate decision-only behavior with zero side effects. Stage 3 (persistent action ledger implementation) must not begin until the storage transaction model and the Stage 2 boundary are approved.
