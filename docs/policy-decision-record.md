# Policy Decision Record

**Status:** Proposed  
**Version:** 1.0.0  
**Issue:** #266

## Purpose

A Policy Decision Record (PDR) binds an already-issued NAEOS policy decision to the evidence required to explain and independently verify that decision.

A PDR is **not an authorization mechanism** and must never grant authority retroactively.

## Authority chain

```
Model / Agent proposal
        ↓
Policy classification + decision
        ↓
Verification gates
        ↓
Durable evidence
        ↓
Policy Decision Record
        ↓
Independent verification
```

The existing policy engines remain authoritative. PDR records their output and binds it to durable evidence.

## Required bindings

A PDR binds:
- policy ID and policy version;
- PDR schema version;
- decision ID and optional run ID;
- risk/criticality classification and policy decision;
- required verification gates;
- evidence identifiers and SHA-256 digests;
- final outcome and verification state.

## Fail-closed rules

PDR validation must reject:
- missing required evidence;
- malformed evidence digests;
- policy/schema version mismatch;
- unsupported record version;
- evidence references that cannot be independently resolved;
- a PDR that attempts to convert an agent proposal into authorization.

## Relationship to existing NAEOS components

- **Change Risk Governance** classifies repository changes and decides the required governance path.
- **Dependency Risk Governance** classifies dependency changes.
- **EvidenceStore** stores tamper-evident evidence records.
- **Control-plane Ledger** records decision/execution events.
- **Verification** establishes what the evidence supports.
- **PDR** binds these artifacts into a durable decision record.

No second policy engine is introduced.

## Claim discipline

A valid PDR proves traceability and integrity of the decision/evidence relationship. It does not by itself prove that the underlying execution was safe, nor does it authorize an action.

## Versioning

The PDR schema version is part of the trust boundary. Unsupported versions fail closed.
