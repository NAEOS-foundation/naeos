# NAEOS Verified Golden Path

The Verified Golden Path is the canonical public onboarding spine for evaluating NAEOS as an engineering control plane for AI coding agents.

It connects the existing proof surfaces without introducing a second implementation path:

```text
Public Control Plane
        ↓
Golden Path
        ↓
Reference Demo
        ↓
Evidence inspection
        ↓
Independent verification
        ↓
Challenge / contribution
```

## 1. Purpose

Answer one concrete question:

> Can an engineer reproduce an NAEOS-controlled workflow and inspect its evidence without trusting an agent's claim?

The default path is local-first and does not require an external AI API key.

## 2. Control-flow model

```text
Intent
  ↓
Specification
  ↓
NEIR
  ↓
Validation
  ↓
Policy
  ↓
Agent Context
  ↓
Authorized Execution
  ↓
Observation
  ↓
Evidence
  ↓
Independent Verification
```

The public browser demo exposes the authorization boundary. The local Golden Path demonstrates the reproducible specification-to-artifact workflow. P1.11 provides a separate read-only verification boundary for canonical EvidenceBundle records.

## 3. Inspect the public boundary

Open the public [NAEOS Control Plane](https://naeos.dev/control-plane/).

Use it to inspect authorization decisions without triggering a consequential external side effect.

The boundary is:

```text
agent request → policy decision → authorized / denied → runtime boundary
```

A decision is not execution.

## 4. Run the canonical Golden Path

From a clean checkout:

```bash
go build -o naeos ./cmd/naeos
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh
```

The acceptance contract is [docs/GOLDEN-PATH.md](GOLDEN-PATH.md).

The run demonstrates:

- specification validation;
- NEIR materialization and inspection;
- deterministic policy rejection;
- AI context generation;
- authorized generation;
- generated artifacts;
- traceability metadata, including artifact accounting.

Inspect the resulting evidence:

```bash
find /tmp/naeos-demo -maxdepth 3 -type f | sort
```

Key evidence includes `spec.yaml`, `inspect.json`, `validate.json`, `invalid-policy.log`, `context.md`, `context.json`, `run.json`, `generated/`, and `summary.md`.

## 5. Verify independently

For a canonical exported `EvidenceBundle`:

```bash
naeos evidence verify-bundle --input-file evidence.json
naeos evidence verify-bundle --input-file evidence.json --output json
```

The verifier is read-only. It does not contact the control plane, re-run policy, or execute an action. It validates decision/execution bindings and recomputes the SHA-256 evidence digest.

See [P1.11 — Independent Verifier CLI](control-plane/p1-11-independent-verifier-cli.md).

The verifier is a separate proof boundary. The Golden Path's `run.json` is traceability evidence and should not be represented as an EvidenceBundle unless it was produced through the canonical evidence contract.

## 6. Test governance under change

Run:

```bash
go run ./experiments/governance-lifecycle
```

This exercises:

```text
agent intent → policy decision → execution → observation → evidence → independent verification
```

The experiment deliberately distinguishes an agent's claim from observed evidence.

## 7. Test handoff boundaries

Run:

```bash
go run ./experiments/handoff-governance
```

The handoff experiment covers:

- valid signed handoff acceptance;
- capability widening rejection;
- downstream authority escalation rejection;
- replay protection.

The Handoff Contract binds authorization context to the handoff instead of allowing a downstream component to invent authority.

See [experiments/handoff-governance/README.md](../experiments/handoff-governance/README.md).

## 8. What success means

A successful path provides repository-backed evidence that the documented workflow can be reproduced at a recorded repository state and inspected through its declared evidence surfaces.

It does **not** by itself establish production readiness, customer adoption, enterprise compliance, performance targets, security of every external AI provider, correctness of arbitrary specifications, or safety of every consequential external action.

## 9. Evaluation record

For an external evaluation, record:

```text
Repository: NAEOS-foundation/naeos
Commit SHA:
Go/toolchain:
Reviewer:
Date/time (UTC):

Golden Path exit status:
run_id:
specification_hash:
neir_hash:
Logical artifact count:
Materialized file count:
Artifact path collisions:

Independent verification:
PASS / FAIL / NOT RUN

Governance experiment:
PASS / FAIL / NOT RUN

Handoff experiment:
PASS / FAIL / NOT RUN

Deviations:
- None / <describe>
```

Do not mix evidence from different commits without recording the difference.

## 10. Evaluator → contributor

If the path exposes a failure or ambiguity, record:

1. exact command;
2. expected result;
3. observed result;
4. commit SHA;
5. relevant run identifier or artifact;
6. environment/toolchain;
7. whether the deviation blocks reproducibility.

Then open a focused issue or discussion.

The adoption loop is:

```text
Run → Inspect → Verify → Challenge → Contribute
```

## 11. Proof principle

**Do not ask the reviewer to trust the claim when the repository can expose the evidence.**

The Verified Golden Path therefore treats:

```text
Artifact → Evidence → Verification → Reproduction → Feedback
```

as the primary adoption loop.

The individual implementation contracts remain authoritative in their respective specifications, control-plane documents, experiments, and code.
