# NAEOS Flagship Golden Path

This is the smallest externally consumable runtime proof of the NAEOS execution-control boundary.

## Flow

```text
AI Agent / Intent
      ↓
NAEOS Policy
      ↓
Authorization
      ↓
Consequential Execution
      ↓
Observation
      ↓
Evidence Receipt
      ↓
Independent Verification
```

The negative path is deliberately interleaved:

```text
ALLOW(v1)
   ↓
Policy v2 becomes active
   ↓
stale authorization
   ↓
EXECUTION_BLOCKED
   ↓
no side effect
   ↓
evidence
```

## Run

From the repository root:

```bash
go run ./examples/flagship-golden-path
```

The demo uses only local deterministic state and writes its evidence to:

```text
/tmp/naeos-flagship-golden-path/evidence.json
```

No screenshots, external service, or AI API key is required.

## What to inspect

The evidence receipt records:

- request/run identity;
- authorization decision;
- policy version;
- execution outcome;
- observed side effect state;
- evidence event types;
- independent verification result.

The positive scenario creates `positive-side-effect.json`. The stale-policy scenario must **not** create `stale-side-effect.json`.

## Acceptance

A successful run must demonstrate:

1. v1 authorization returns ALLOW;
2. the consequential positive action produces a real local side effect;
3. the side effect is recorded as an observation;
4. evidence contains authorization, execution, and observation events;
5. independent verification returns PASS;
6. after policy v2 becomes active, the original v1 authorization is rejected at execution;
7. the stale path records `EXECUTION_BLOCKED`;
8. the stale path produces no side effect;
9. the evidence receipt is written without relying on screenshots or agent claims.

## Evidence boundary

This demo establishes the tested local execution-control behavior. It does **not** establish production readiness, exactly-once semantics for arbitrary external side effects, blanket security, enterprise compliance, or safety of every external AI-agent integration.

For the broader evidence record, see [External Validation Report #001](EXTERNAL-VALIDATION-REPORT-001.md), [Golden Path](GOLDEN-PATH.md), and [External Validation Runbook](EXTERNAL-VALIDATION.md).

**Principle:** evidence should be reproducible before it is persuasive.
