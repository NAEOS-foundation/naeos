# Phase 4.2 — Evidence Query UX

## Purpose

Make existing NAEOS evidence query capability easy to use as a verification workflow.

This is a UX and contract layer over the existing P1.10 Evidence Query & Audit API and append-only control-plane ledger. It does not introduce a second evidence store, query engine, or verification model.

## Canonical query

```http
GET /api/control-plane/evidence
```

Supported correlation filters:

| Filter | Meaning |
|---|---|
| `agent_id` | Events attributed to one agent |
| `request_id` | Events belonging to one request |
| `decision_id` | Events associated with one authorization decision |
| `execution_id` | Events associated with one execution |
| `event_type` | One ledger event type |

Filters are composable. An omitted filter means no restriction on that dimension.

## Recommended evaluator flow

### 1. Start from a decision

Use the decision identifier exposed by the control-plane workflow:

```bash
curl -sS "https://<control-plane>/api/control-plane/evidence?decision_id=<DECISION_ID>"
```

### 2. Inspect the correlation events

The `events` array is the raw ledger correlation surface. Use it to establish:

- request identity;
- agent identity;
- decision identity;
- execution identity;
- event sequence.

Do not treat a raw authorization event as proof that execution occurred.

### 3. Inspect the verifier-facing bundle

The `evidence` array contains canonical `EvidenceBundle` records reconstructed from the decision and matching execution evidence.

The response also exposes:

- `evidence_total`;
- `verified`;
- `persistence_error` when persistence reports an error.

### 4. Verify the claim boundary

Interpret `verified=true` only in the context of the returned bundles.

An ALLOW decision without a matching `EXECUTION_ALLOWED` event remains unverified. This is intentional: authorization is not execution.

A DENY decision can verify without an execution event when the decision identity and evidence digest are internally consistent.

## Query examples

### By request

```bash
curl -sS "https://<control-plane>/api/control-plane/evidence?request_id=<REQUEST_ID>"
```

Use this when reconstructing one end-to-end control-plane request.

### By execution

```bash
curl -sS "https://<control-plane>/api/control-plane/evidence?execution_id=<EXECUTION_ID>"
```

Use this when tracing what execution evidence is attached to an authorized transition.

### By agent and event type

```bash
curl -sS "https://<control-plane>/api/control-plane/evidence?agent_id=<AGENT_ID>&event_type=EXECUTION_ALLOWED"
```

Use this for focused operational inspection.

### By decision and execution

```bash
curl -sS "https://<control-plane>/api/control-plane/evidence?decision_id=<DECISION_ID>&execution_id=<EXECUTION_ID>"
```

Use this when independently checking that a specific execution belongs to a specific authorization decision.

## Minimal response shape

```json
{
  "events": [],
  "total": 0,
  "evidence": [],
  "evidence_total": 0,
  "verified": true
}
```

A persistence failure may add:

```json
{
  "persistence_error": "..."
}
```

Clients should preserve unknown response fields for forward compatibility.

## Evidence interpretation rules

The query UX follows the NAEOS trust model:

```text
Intent
  ↓
Decision
  ↓
Execution authorization
  ↓
Observation
  ↓
Evidence
  ↓
Independent verification
```

Therefore:

1. **Decision ≠ execution.**
2. **Event presence ≠ independently verified claim.**
3. **EvidenceBundle verification is the verifier-facing integrity boundary.**
4. **Raw ledger events remain useful for correlation and chronology.**
5. **A missing execution event must not be silently inferred.**
6. **A persistence error must remain visible to the evaluator.**
7. **Evidence from different repository commits must not be mixed without recording the commit boundary.**

## Reproducible evaluator record

For a technical evaluation, record:

```text
Repository: NAEOS-foundation/naeos
Commit SHA:
Query URL / parameters:
Reviewer:
Date/time (UTC):

Decision ID:
Request ID:
Execution ID:
Agent ID:

Event count:
EvidenceBundle count:
verified:
persistence_error:

Independent verifier:
PASS / FAIL / NOT RUN

Deviation:
None / <describe>
```

## Non-goals

Phase 4.2 does not:

- replace the append-only control-plane ledger;
- create a new evidence database;
- introduce a second query engine;
- change policy evaluation;
- change authorization semantics;
- reinterpret ALLOW as execution;
- replace the independent verifier;
- add dashboard complexity before the query contract is stable.

## Acceptance criteria

- [x] Existing P1.10 query endpoint remains the canonical API.
- [x] Existing correlation filters are documented as the public query vocabulary.
- [x] Raw events and verifier-facing EvidenceBundles are explicitly distinguished.
- [x] ALLOW-without-execution remains visibly unverified.
- [x] Persistence errors remain visible.
- [x] Reproducible evaluator recording format is defined.
- [x] Query examples cover request, decision, execution, agent, and event-type correlation.
- [ ] Fresh-checkout execution is verified on the current release.
- [ ] External evaluator completes one recorded query-and-verification run.

## Relationship to the next phase

Phase 4.3 should consume the same evidence and authorization vocabulary when defining the Agent Handoff Contract.

The resulting proof chain should remain:

```Golden Path
   ↓
Evidence Query
   ↓
Independent Verification
   ↓
Agent Handoff
   ↓
Re-verification
```

The objective is one coherent trust chain, not a collection of disconnected features.
