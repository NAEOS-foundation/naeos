# NAEOS Trust-Boundary Map

**Document status:** Proposed canonical architecture map  
**Scope:** AI-assisted engineering control plane  
**Authority:** This document is subordinate to the NAEOS Constitution, NAEOS Reference Architecture (NAEOS-NRA-001), Master Technical Specification, and applicable Engineering Specifications/ADRs.

## 1. Purpose

NAEOS separates what an AI coding agent **proposes** from what the engineering system **authorizes**, what the runtime **executes**, what the system **observes**, and what independent verification can **establish**.

The trust boundary is therefore not the model prompt. It is the sequence of controllable boundaries between intent, authorization, execution, observation, evidence, verification, and external side effects.

## 2. Canonical control flow

```text
Engineering Specification
        |
        v
      NEIR
        |
        v
Validation + Policy Evaluation
        |
        +----> Policy Decision / Authorization
        |
        v
Agent Context / Agent Intent
        |
        v
Handoff Contract
        |
        v
Final Boundary Revalidation
        |
        v
Authorized Runtime Execution
        |
        v
Observation
        |
        v
Evidence / Receipt
        |
        v
Independent Verification
```

An agent may propose an action, but an agent-generated proposal is never itself an authorization.

## 3. Boundary matrix

| Boundary | Actor / component | Input | Authority | Authorization | Untrusted data | Observation / evidence | Fail-closed / revalidation |
|---|---|---|---|---|---|---|---|
| B1 — Specification → NEIR | Specification parser/compiler | Human-authored specification | Normative specification hierarchy | Schema/semantic validation | Free-form specification values | Parse/validation result, canonical NEIR and hashes | Reject invalid or incompatible input |
| B2 — NEIR → Policy | Validator / Policy engine | Canonical NEIR + policy context | Applicable policy and governance | Policy decision | Agent/user supplied values that are not authoritative policy | Policy decision and decision metadata | Deny on policy failure, ambiguity, or version mismatch |
| B3 — Policy → Agent Context | Runtime/context builder | Authorized constraints + task context | Policy decision + specification | Capability scope derived from policy | Model-generated text and proposed actions | Generated context, policy/version identifiers | Do not expand capability from model output |
| B4 — Agent Intent → Handoff | Agent/runtime | Proposed action + requested capability | Existing authorization state | Handoff Contract must bind requested and authorized capability | Agent-generated intent, arguments, downstream inputs | Signed/canonical handoff data and HandoffProbe observations | Reject malformed, unsigned, replayed, or mismatched handoffs |
| B5 — Handoff → Runtime | Runtime executor | Authorized handoff | Policy decision + runtime contract | Revalidate at the final controllable boundary | External inputs and downstream responses | Execution event, receipt, exit status | Revalidate immediately before side effect; deny on mismatch |
| B6 — Runtime → Observation | Runtime / OpsWatch | Execution events and external results | Runtime instrumentation contract | Observation does not grant authorization | Provider responses, logs, network data | Immutable/auditable observations | Missing or malformed observations do not become evidence |
| B7 — Observation → Evidence | Evidence subsystem | Observations + artifact/receipt metadata | Evidence schema and experiment contract | Evidence acceptance rules | Raw logs and externally supplied claims | Canonical evidence record, digests, receipts | Reject incomplete, tampered, non-canonical, or version-incompatible evidence |
| B8 — Evidence → Verification | Independent verifier | Evidence package | Verification contract | Independent verification procedure | Claims contained in the evidence package | Verification result and verification metadata | Verification must not silently substitute for missing evidence |
| B9 — Runtime → External Side Effect | Authorized runtime / GO-gate | Deployment, publication, commit, or other public action | Explicit authorization + applicable policy | GO-gate / final authorization | Provider state and external responses | External receipt: commit/artifact/provider ID, health, rollback state | No external action without explicit authorization; revalidate before execution |

## 4. Trust rules

### 4.1 Intent is not authority

The model, agent, or generated context may express intent. It cannot promote that intent into permission.

```text
Agent proposal != Policy decision
Agent proposal != Authorization
Evidence != Authorization
Observation != Authorization
```

### 4.2 Policy decision is separate from execution

A policy engine determines whether a requested capability is authorized. Runtime execution consumes that authorization but does not reinterpret model output as policy.

### 4.3 Handoffs carry constrained authority

A Handoff Contract must make the security-relevant distinction between:

- requested capability;
- authorized capability;
- untrusted input;
- downstream capability.

The downstream component must not infer broader authority from untrusted handoff content.

### 4.4 Revalidate at the last controllable boundary

An earlier policy decision is not sufficient if security-relevant inputs can change before an external side effect. The final controllable boundary must revalidate the contract, capability, policy/version context, and relevant integrity identifiers.

### 4.5 Evidence is not a decision

Evidence records what was observed and what verification established. It does not authorize an action retroactively.

### 4.6 Version identity is part of trust

Where canonicalization, contracts, schemas, policies, algorithms, or profiles affect interpretation, their version/profile identifiers must remain bound to the decision or evidence. A security-relevant version mismatch must fail closed.

## 5. External side effects

External/public actions include, as applicable:

- pushing or publishing source changes;
- creating or deploying artifacts;
- invoking production or third-party systems;
- publishing documentation or releases;
- other actions that cross the repository or runtime trust boundary.

For these actions, NAEOS should preserve a complete chain:

```text
request
  -> authorization
  -> final revalidation
  -> execution
  -> observation
  -> external receipt
  -> verification
```

The GO-gate represents explicit human approval where the applicable NAEOS workflow requires it. It is an authorization control, not merely a user-interface confirmation.

## 6. Failure semantics

A boundary is considered fail-closed when the system does not proceed across it after detecting a security-relevant condition such as:

- missing authorization;
- policy denial or ambiguity;
- invalid signature;
- replay detection;
- contract/schema mismatch;
- canonicalization mismatch;
- version/profile mismatch;
- capability mismatch;
- missing required evidence;
- verification failure.

Failure handling must preserve enough diagnostic and audit information to distinguish **decision**, **execution**, **observation**, and **verification** outcomes.

## 7. Relationship to NAEOS architecture

This map does not introduce a second architecture. It connects two existing views:

1. **Engineering model:** Specification → NEIR → validation → policy → AI context → generation.
2. **Control-plane model:** Agent intent → policy decision → authorized execution → observation → evidence → independent verification.

The first describes how engineering intent becomes a machine-interpretable engineering representation. The second describes how that representation and agent activity are constrained when actions cross trust boundaries.

## 8. Normative status and claim discipline

This document is an architectural map and boundary model. It must not be read as proof that every listed control is implemented in every NAEOS execution path.

Implementation status must be established by the relevant source code, Engineering Specifications/ADRs, deterministic tests, and experiments.

Experiments may provide evidence for named components and scenarios but do not override the normative architecture.

## 9. Related authority

- [Documentation Authority Model](DOCUMENTATION-AUTHORITY.md)
- [NAEOS Reference Architecture](Reference%20Architecture/NAEOS-NRA-001.md)
- [Master Technical Specification](NAEOS-MTS-001.md)
- [Architecture Overview](ARCHITECTURE-OVERVIEW.md)

**Issue:** #207
