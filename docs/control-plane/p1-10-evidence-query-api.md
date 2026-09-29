# P1.10 — Evidence Query & Audit API

## Purpose

P1.10 exposes the canonical control-plane evidence model through the existing HTTP evidence endpoint.

The endpoint keeps raw ledger events for correlation while also returning verifier-facing EvidenceBundle records reconstructed from canonical decision events.

## Query contract

GET /api/control-plane/evidence

Supported correlation filters:

- agent_id
- request_id
- decision_id
- execution_id
- event_type

Response fields:

- events — matching raw ledger events
- total — matching event count
- evidence — canonical EvidenceBundle records
- evidence_total — bundle count
- verified — whether all returned bundles pass independent bundle verification
- persistence_error — present only when ledger snapshot persistence reports an error

## Verification semantics

A bundle is reconstructed from the canonical authorization decision and matching execution evidence.

An ALLOW decision without an EXECUTION_ALLOWED event is intentionally reported as unverified. This prevents the query API from turning an authorization decision into a claim that execution occurred.

A DENY decision can verify without an execution event when the evidence digest and decision identity are internally consistent.

## Scope

P1.10 adds queryability and verifier-facing API exposure. It does not change policy evaluation, authorization semantics, execution authorization, or persistence guarantees.

The endpoint remains backed by the existing append-only control-plane ledger.
