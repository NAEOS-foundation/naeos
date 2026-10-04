# Evidence & Verification Invariants v1

**Status:** Normative  
**Version:** 1.0.0  
**Scope:** Evidence evaluation and verification outcomes across the NAEOS control plane

## Purpose

This document defines the minimum invariants that govern when NAEOS may treat an execution outcome as verified.

The invariants are derived from the Evidence Verification Boundary Matrix experiment and apply across policy, runtime, observation, evidence, and verification boundaries.

This document does **not** introduce a second policy engine or prescribe a concrete storage schema.

## Normative invariants

### 1. No unsupported confirmation

NAEOS MUST NOT produce a `CONFIRMED` outcome unless the available evidence satisfies the verification policy for the decision class.

An agent claim, execution narrative, or convenient single signal MUST NOT be treated as confirmation by itself.

### 2. Evidence provenance

Every verification outcome MUST retain the evidence identifiers and provenance that justified the outcome.

Evidence provenance MUST identify the authority boundary relevant to the claim.

### 3. External-effect authority boundary

A runtime execution record MAY be authoritative for effects owned by that runtime.

A runtime execution record MUST NOT be treated as proof of an external side effect merely because the record is internally authoritative.

External confirmation requires evidence from an authority capable of observing or acknowledging the external effect.

### 4. Effect-derived minimum verification

The verification floor MUST be derived from the effect characteristics, including reversibility and blast radius.

A decision-specific verification policy MAY require stronger evidence than the minimum floor.

A verification policy MUST NOT lower the minimum floor required by the effect.

### 5. Freshness

Evidence MUST be evaluated against its freshness requirements.

Evidence that is no longer current MUST NOT satisfy a current verification requirement. The verifier MUST return `REVERIFY` or another policy-defined non-confirming state.

### 6. Contradictory evidence

Conflicting observations MUST NOT be collapsed into `CONFIRMED`.

When policy cannot resolve the conflict, the verifier MUST preserve `UNKNOWN` and retain the conflicting evidence.

### 7. Estimated evidence

Estimated or derived evidence MUST remain distinguishable from observed or authoritative evidence.

Estimated evidence MUST NOT masquerade as authoritative observation and MUST NOT satisfy a verification floor that requires authoritative evidence.

### 8. Missing verification policy

A missing verification policy MUST be visible and fail closed.

The default outcome for an ungoverned verification decision MUST be `ESCALATE`, unless a higher-level normative policy explicitly defines another safe outcome.

### 9. Unknown is first-class

`UNKNOWN` is a valid verification outcome.

Insufficient, stale, contradictory, or authority-inadequate evidence MUST NOT be coerced into success or failure merely to avoid an unknown state.

## Reference lifecycle

```
MODEL / AGENT
     |
     v
POLICY / AUTHORIZATION
     |
     v
CAPABILITY
     |
     v
RUNTIME
     |
     v
EFFECT
     |
     v
OBSERVATION
     |
     v
EVIDENCE
     |
     v
VERIFICATION
     |
     v
OUTCOME
```

The trust boundary is explicit: execution produces an effect; observation establishes evidence about that effect; verification determines what that evidence supports.

## Relationship to experiments

The Evidence Verification Boundary Matrix is the executable evidence for these invariants.

Promotion into production implementation requires separate compatibility, schema, and migration review. This document therefore defines the normative contract without forcing the experiment's standalone data model into the production kernel.

## Required regression coverage

Implementations claiming conformance SHOULD maintain executable coverage for at least:

- runtime-only evidence for an external effect remains non-confirming;
- authoritative external/provider evidence can confirm when policy permits;
- stale evidence requires re-verification;
- contradictory evidence remains unknown;
- estimated evidence cannot satisfy an authoritative verification floor;
- missing verification policy fails closed;
- every confirming outcome records the evidence used to justify it.
