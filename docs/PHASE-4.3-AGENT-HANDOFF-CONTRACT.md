# Phase 4.3 — Agent Handoff Contract v1

## Status

Proposed normative contract for agent-to-agent or component-to-component handoff.

The implementation baseline is the existing HandoffValidator and deterministic handoff-governance experiment. This document makes the trust boundary explicit without introducing a second authorization system.

## Contract thesis

> A handoff transfers context, not authority.

A valid handoff may delegate only authority already granted by the upstream boundary. The recipient must independently re-verify the contract before acting on it.

## Canonical contract

    {
      "contract_version": "1.0",
      "canonical_version": "1",
      "initiator": "<agent-or-component-id>",
      "recipient": "<agent-or-component-id>",
      "requested_capability": "<capability>",
      "authorized_capabilities": ["<capability>"],
      "payload_digest": "<canonical-payload-digest>",
      "payload": {},
      "policy_id": "<policy-id>",
      "policy_version": 1,
      "provenance": {
        "source": "<initiator>",
        "destination": "<recipient>"
      },
      "created_at": "<RFC3339 timestamp>",
      "expires_at": "<RFC3339 timestamp>",
      "replay_protection": {
        "nonce": "<unique nonce>",
        "timestamp": "<RFC3339 timestamp>"
      },
      "signature": "<contract-signature>"
    }

Nested delegation uses the same contract shape under downstream_handoff.

## Normative invariants

### 1. Authority is non-escalating

The requested capability MUST exist in authorized_capabilities.

For a nested handoff:
- the downstream requested capability MUST be authorized by the parent;
- every downstream authorized capability MUST already exist in the parent's authorized set;
- a downstream contract MUST NOT create a broader grant by listing additional capabilities.

Formally:

Requested(downstream) ∈ Authorized(parent)

and

Authorized(downstream) ⊆ Authorized(parent)

### 2. Identity and provenance are bound

initiator identifies the authority source.

recipient identifies the receiving boundary.

If provenance declares source or destination, they MUST match the corresponding identity fields.

A missing or inconsistent provenance record fails closed.

### 3. Policy identity is immutable within the contract

The handoff binds policy_id and policy_version.

A consumer MUST NOT treat a handoff as portable across policy versions without a new authorization decision.

### 4. Canonicalization is explicit

canonical_version is part of the signed contract.

Unsupported canonicalization versions MUST be rejected rather than interpreted heuristically.

### 5. Integrity is bound to the payload

payload_digest MUST match the canonical payload.

Changing the payload after signing MUST invalidate verification.

The parent signature MUST bind the canonical downstream contract fields when nested delegation is present.

### 6. Signature is mandatory

Unsigned contracts MUST fail closed.

The current implementation uses HMAC-SHA256 for the demo boundary. Production deployments MUST source signing material from an appropriate secure key or attestation mechanism; the demo key is not a production credential.

### 7. Replay protection is mandatory

Every contract requires a nonce.

A previously consumed nonce MUST be rejected.

Replay state MUST be consumed only after all other validation succeeds so malformed or unauthorized contracts cannot poison the replay ledger.

### 8. Lifetime is bounded

The contract MUST have an expiry.

A downstream contract MUST NOT outlive its parent authorization.

Expired contracts MUST fail closed.

### 9. Re-verification is local to the receiving boundary

The recipient MUST validate the handoff before using the delegated authority.

Upstream validation is not sufficient evidence for downstream execution.

The recipient MUST establish at least:
1. signature validity;
2. contract and canonicalization version support;
3. expiry validity;
4. replay validity;
5. payload integrity;
6. capability non-escalation;
7. provenance and identity consistency.

## Verification semantics

A successful handoff validation means:

> The receiving boundary verified that this signed contract is internally consistent and does not widen the authority encoded by its parent boundary.

It does NOT mean:
- the requested action was executed;
- the downstream component is trustworthy in general;
- the underlying policy is permanently valid;
- the payload is safe for every possible consumer.

Execution remains a separate authorization/execution boundary and must emit its own evidence.

## Evidence contract

A handoff implementation SHOULD record enough evidence to reconstruct:
- contract identity/version;
- initiator and recipient;
- policy identity/version;
- requested capability;
- authorized capability set;
- payload digest;
- nonce and replay state;
- validation result;
- rejection reasons;
- relevant decision and execution correlation identifiers.

A verifier MUST be able to distinguish HANDOFF_VALIDATED from EXECUTION_ALLOWED.

Validation of a handoff MUST NOT be interpreted as proof that execution occurred.

## Failure modes

The validator MUST fail closed for:
- missing signature;
- invalid signature;
- unsupported contract version;
- unsupported canonicalization version;
- missing nonce;
- replayed nonce;
- expired contract;
- payload digest mismatch;
- missing recipient;
- provenance mismatch;
- requested capability outside the authorized set;
- downstream capability escalation;
- downstream version mismatch;
- downstream expiry beyond parent expiry;
- downstream payload digest mismatch.

## Reproducible test matrix

The handoff-governance experiment covers:

| Scenario | Expected invariant |
|---|---|
| Valid signed handoff | Accepted |
| Capability widening | Rejected |
| Downstream escalation | Rejected |
| Replay | Rejected |
| Payload tampering | Rejected |
| Provenance mismatch | Rejected |
| Contract version mismatch | Rejected |
| Canonicalization mismatch | Rejected |
| Expired handoff | Rejected |
| Signature tampering | Rejected |

These scenarios are deterministic at the scenario level and do not require an LLM or network service.

## Trust chain

The Phase 4 trust chain is:

    Intent
      ↓
    Policy / Authorization Decision
      ↓
    Evidence
      ↓
    Handoff Contract
      ↓
    Recipient Re-verification
      ↓
    Execution Authorization
      ↓
    Execution Evidence

The critical boundary is non-escalation: downstream components can receive context and a bounded authority envelope, but cannot mint new authority through the handoff itself.

## Relationship to Phase 4.2

Phase 4.2 provides the evidence query surface.

Phase 4.3 consumes the same decision and evidence identifiers and keeps the distinction explicit:

authorization ≠ handoff validation ≠ execution

A future evaluator should be able to query the evidence chain and correlate:

decision_id → handoff validation → execution evidence

without relying on agent claims or screenshots.

## Acceptance criteria

- [x] Contract fields and versions are explicit.
- [x] Non-escalating capability rule is normative.
- [x] Nested delegation constraints are normative.
- [x] Provenance, replay, expiry, canonicalization, and signature boundaries are explicit.
- [x] Validation is separated from execution authorization.
- [x] Existing deterministic handoff experiment is identified as the reproducible test matrix.
- [ ] Fresh-checkout handoff experiment is independently recorded against a fixed commit.
- [ ] Evidence correlation from decision to handoff validation to execution is recorded by an external evaluator.

## Non-goals

This phase does not introduce:
- a new authorization engine;
- a second policy model;
- a new evidence store;
- automatic trust of downstream agents;
- implicit capability inheritance;
- production key-management implementation.
