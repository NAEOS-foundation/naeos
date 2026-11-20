# ADR-001 — Production Authorization & Execution Gateway

**Status:** Draft — design review required; not approved for implementation or deployment  
**Date:** 2026-11-20  
**Related:** Issue #848 (authoritative remote execution contract), architecture audit #713, PR #847

## 1. Decision context

The repository has two complementary but non-equivalent paths:

- `internal/governance/control` + `internal/runtime/gateway` is the existing tool/runtime enforcement path. It evaluates governance policy, supports second-phase decision revalidation, enforces adapter grants, and can report uncertain outcomes. Its default invocation store and execution history are in-memory; the inspected gateway does not establish a durable pre-dispatch intent protocol.
- `internal/controlplane` contains decision, grant, approval, and hash-chain ledger behavior. `DecisionGateway.ExecuteAtomic` validates bindings and revalidates under an in-process policy-store lock, writes an execution-start event before invoking its callback, and preserves callback errors as UNKNOWN. The approval store is in-memory, the lock is not proven cross-process, and the JSON snapshot ledger is not a distributed transactional store.

Neither subsystem alone is established as a production remote execution service. Calling the decision HTTP API and then dispatching directly through a runtime adapter is rejected because it creates a check-to-dispatch gap.

## 2. Proposed architecture

1. **Identity boundary:** authenticate the workload and validate issuer, audience, subject, tenant, credential status, and required scopes.
2. **Request contract:** validate a versioned schema; canonicalize the request using an explicit canonicalization version; calculate a digest bound to the authenticated principal and the complete execution target.
3. **Authorization boundary:** use the existing governance policy contract as the single authoritative tool/runtime policy authority. Resolve grants and approvals under that contract. Do not introduce a second independent policy engine.
4. **Durable intent boundary:** atomically claim the idempotency binding and persist the immutable invocation intent before any side effect can be dispatched.
5. **Final dispatch boundary:** revalidate the exact request digest, decision/policy version, grant, approval, expiry, tenant/principal binding, adapter/tool/action scope, and replay state immediately before dispatch.
6. **Execution boundary:** route only through `internal/runtime/gateway` and its registered, explicitly granted adapter/tool path. No raw delegate is exposed to remote callers.
7. **Observation boundary:** persist execution receipts and independent observation evidence. Sandbox return values alone do not prove external effect.
8. **Reconciliation boundary:** resolve UNKNOWN only from an authoritative provider/runtime receipt or an explicitly governed operator resolution with evidence. Never blindly retry a non-idempotent side effect.

The remote orchestration layer is a narrow contract and persistence boundary, not a third policy engine. The precise package composition remains subject to code-level review and proof of storage concurrency semantics.

## 3. Contract envelope (proposed schema, not a committed endpoint)

### Request

- `contract_version`
- `request_id` — stable logical request correlation ID
- `invocation_id` — unique execution attempt identity
- `idempotency_key` — caller-provided stable key, scoped by tenant and authenticated principal
- `canonicalization_version`
- `principal` — derived from verified workload credentials; caller-supplied identity fields are not authoritative
- `tenant_id` — verified from trusted identity/binding, never trusted from an unverified body field
- `target` — tool, adapter identity/version, action, resource, environment, and destination
- `arguments` — schema-validated arguments; secrets must be references, not ledger plaintext
- `capability_scope` — requested capability and resource scope
- `context` — bounded, typed context relevant to policy evaluation

The canonical request digest must bind contract and canonicalization versions, verified principal and tenant, target, action, arguments, capability/resource scope, environment/context, and adapter identity/version. A retry using the same idempotency key with a different digest is a conflict.

### Authorization decision

- `decision_id`
- `request_id`, `invocation_id`, and `request_digest`
- `principal_id` and `tenant_id`
- `decision`: ALLOW, DENY, REQUIRE_APPROVAL, or INDETERMINATE
- `policy_id`, immutable `policy_version`
- `grant_id` and grant/version or digest
- optional `approval_id` and approval binding
- `issued_at`, `expires_at`
- `reason_codes`
- `evidence_ref`

An ALLOW is a time-bounded authorization assertion for the exact digest; it is not permission to mutate the request, bypass final revalidation, or claim execution completion.

### Execution receipt

- `execution_id`, `request_id`, `invocation_id`, `decision_id`
- `request_digest`, `dispatch_attempt_id`
- `state`, `started_at`, `finished_at`
- `adapter/tool` identity and version
- provider/runtime receipt reference and digest, where available
- observation reference and verification status
- outcome classification and redacted reason code

Do not persist raw credentials or unnecessary raw arguments in ledger events.

## 4. Lifecycle and state semantics

The durable state machine is:

`RECEIVED → AUTHENTICATED → EVALUATING → DENIED | AUTHORIZED`

For an allowed request:

`AUTHORIZED → INTENT_DURABLE → DISPATCHING → COMPLETED → OBSERVED`

Failure and recovery branches:

- `FAILED_PRE_DISPATCH`: proven no dispatch occurred; safe to retry only under the same idempotency contract.
- `UNKNOWN`: dispatch may have occurred, but authoritative outcome is not available. No automatic side-effect retry.
- `RECONCILIATION_REQUIRED`: UNKNOWN has been durably queued for reconciliation.
- `RECONCILED_SUCCEEDED` / `RECONCILED_FAILED`: terminal resolution backed by verified authoritative evidence.
- `DENIED`: explicit policy denial; no dispatch.
- `INDETERMINATE`: evaluator, identity, policy, or authorization dependency unavailable/invalid; no dispatch.

State transitions are append-only and validated. Terminal states cannot be silently overwritten. A later reconciliation is a new event linked to the prior UNKNOWN, not mutation of history.

## 5. Idempotency, replay, and concurrency

- Unique storage constraint on (tenant, authenticated principal, idempotency key).
- Store the canonical request digest and invocation binding with the key.
- Same key + same digest returns the canonical existing lifecycle/receipt; it never creates a second dispatch.
- Same key + different digest returns conflict and never dispatches.
- A distinct invocation ID does not bypass an existing idempotency binding.
- Claim and durable intent creation must be one storage transaction/atomic operation.
- Cross-process concurrency must be enforced by database constraints/transactions or a proven distributed primitive, not a Go mutex.
- Retention and expiry must not permit replay of an action inside its defined replay-protection window.
- For external tools, exactly-once side effects are not assumed. Provider-native idempotency may be used only when its scope and guarantees are documented and verified.

## 6. Persistence and UNKNOWN rules

- No dispatch unless the durable intent and idempotency claim have committed successfully.
- A persistence error before dispatch is `FAILED_PRE_DISPATCH`; the tool body must not be called.
- A crash after durable intent but before known dispatch is recovered by inspecting dispatch-attempt evidence; absence of proof is not proof of non-dispatch.
- A timeout, lost response, process crash after dispatch, or receipt persistence failure is `UNKNOWN` unless an authoritative receipt proves the result.
- UNKNOWN blocks automatic replay of non-idempotent work.
- Reconciliation queries the authoritative provider/runtime by stable dispatch-attempt or idempotency reference and records the evidence digest, source, time, and resolver.
- If the provider cannot establish outcome, retain UNKNOWN; do not manufacture success or failure.

The durable action ledger must be append-only, tenant-isolated, integrity-verifiable, and backed by storage with demonstrated transactional durability. The existing JSON snapshot ledger is not accepted as the production multi-process ledger without separate evidence and an explicit architecture decision.

## 7. Security invariants

- Authentication and audience validation precede authorization.
- Verified identity is server-derived; no trust in body-supplied actor/tenant.
- Exact digest binding at authorization and dispatch.
- Final policy/grant/approval/expiry revalidation at the last controllable boundary.
- Fail closed on unavailable policy, malformed/unsupported versions, persistence failure, stale decision, or missing adapter grant.
- Explicit least-privilege tool/action/environment/capability grants; registration alone grants no execution privilege.
- No direct database access from adapters and no raw tool delegate escape.
- Tenant/principal isolation applies to idempotency lookup, ledger query, receipt retrieval, and reconciliation.
- Redact credentials and sensitive arguments from logs and receipts.
- Audit event integrity and ordering must be verifiable.

## 8. Required verification matrix

P0 integration tests must cover:

1. Unauthenticated, wrong audience, revoked workload, and cross-tenant access.
2. DENY / REQUIRE_APPROVAL / INDETERMINATE never invoke the tool body.
3. Exact digest ALLOW; argument, target, principal, tenant, adapter, or context mutation rejected.
4. Stale policy, revoked/expired grant, substituted/expired/replayed approval rejected at dispatch.
5. Duplicate same-key/same-digest requests cause at most one dispatch; same-key/different-digest conflicts.
6. Concurrent duplicate requests across independent processes cause at most one dispatch.
7. Intent persistence failure prevents dispatch.
8. Crash before intent, after intent/before dispatch, and after dispatch/before receipt recovery.
9. Timeout/response loss/observer failure produces UNKNOWN and suppresses blind retry.
10. Authoritative reconciliation produces linked, immutable evidence; inconclusive reconciliation remains UNKNOWN.
11. Adapter without explicit grant and raw delegate bypass attempts are blocked.
12. Policy mutation concurrent with dispatch is serialized by the actual deployed storage primitive.
13. Ledger tampering, unsupported contract/canonicalization versions, and secret leakage are rejected/detected.

Tests must exercise the real Go gateway composition and the selected persistent store. Mocks may supplement but cannot replace integration evidence.

## 9. Rollout gates

1. Approve this ADR and the versioned request/decision/receipt schemas.
2. Implement storage migrations and repository interfaces in an isolated staging database; prove constraints, transaction isolation, and multi-process duplicate handling.
3. Implement authorization/intent lifecycle without enabling external dispatch.
4. Integrate the guarded runtime gateway and test the complete failure matrix.
5. Deploy to staging only after CI/security gates pass; use synthetic tools and non-production credentials.
6. Capture redacted evidence for deployment identity, schema version, health/readiness, authorization outcomes, duplicate suppression, UNKNOWN/reconciliation, and rollback.
7. Production remains blocked until staging evidence is reviewed and production-specific identity, storage, and operational controls are approved.

## 10. Explicit non-goals and approvals

This draft does not authorize production endpoints, production migrations, secret changes, production deployment, policy bypasses, or merging. No endpoint path or HTTP method is fixed by this ADR; those are to be specified in the separately versioned API schema after the authority boundary is approved.
