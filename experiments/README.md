# NAEOS Experiments

Experiments are deterministic, reproducible tests of NAEOS engineering and governance assumptions.

## Flagship governance experiment

[Governance Lifecycle v1](governance-lifecycle/README.md) tests:

`AGENT INTENT → POLICY DECISION → AUTHORIZED ACTION → EXECUTION → OBSERVATION → EVIDENCE → INDEPENDENT VERIFICATION`

Run it with:

```bash
go run ./experiments/governance-lifecycle
```

The experiment uses real NAEOS governance, evidence, and verification components and does not require an LLM or network access.

## Level-3 real side-effect experiment

[Level-3 Evidence v1](level3-evidence/README.md) extends the lifecycle boundary to a real local filesystem side effect:

`AGENT INTENT → POLICY DECISION → GATEWAY → REAL SIDE EFFECT → INDEPENDENT OBSERVATION → EVIDENCE → INDEPENDENT VERIFICATION`

It covers ALLOW, DENY, direct bypass, execution/observation separation, and post-observation tampering.

Run it with:

```bash
go run ./experiments/level3-evidence
```

It uses the existing NAEOS control plane, execution gateway, evidence store, and verification chain. The filesystem sandbox and observer are experiment harnesses, not replacement production subsystems.

## Adversarial governance experiment

[Policy Bypass](policy-bypass/ADVERSARIAL-GOVERNANCE-V1.md) tests concrete enforcement boundaries such as instruction integrity, governance configuration, and policy evaluator semantics.

## Experiment discipline

Each experiment should:

1. state one falsifiable engineering claim;
2. use deterministic inputs where practical;
3. exercise real NAEOS components rather than parallel mock implementations;
4. record expected versus observed behavior;
5. preserve failures as regression scenarios;
6. distinguish a characterization result from a production security claim.

## Agent handoff governance experiment

[Agent Handoff Governance v1](handoff-governance/README.md) tests whether a handoff can widen authority across agents or components. It covers capability widening, downstream escalation, replay, payload integrity, provenance, protocol/canonicalization versioning, expiration, and signature tampering.

Run it with:

```bash
go run ./experiments/handoff-governance
```

## Attack Matrix v1

[Attack Matrix v1](attack-matrix/ATTACK-MATRIX-V1.md) consolidates the current deterministic adversarial coverage across instruction integrity, governance configuration, policy context/evaluator semantics, runtime enforcement, evidence integrity, and agent handoffs. It identifies covered boundaries and explicitly tracks gaps for future experiments.

The matrix is a coverage map, not a vulnerability count or production penetration-test result.

## Attack Matrix v2

[Attack Matrix v2](attack-matrix/ATTACK-MATRIX-V2.md) promotes AM-18 into a deterministic evidence-completeness experiment. It verifies that required evidence cannot be missing, logically reordered, or detached from the consequential run identity while still being treated as VERIFIED.

## Evidence Completion Enforcement v3

Reusable lifecycle enforcement gate for AM-18 evidence completeness.

- Experiment: `evidence-completion-enforcement-v3`
- Attack Matrix: `attack-matrix/ATTACK-MATRIX-V3.md`

## Evidence Integrity & Run Binding v5

[Evidence Integrity & Run Binding v5](evidence-integrity-run-binding-v5/README.md) hardens the completion boundary with explicit run binding, mixed-run rejection, duplicate-contract rejection, predecessor identity, and backing hash-chain verification.

Run it with:

    go run ./experiments/evidence-integrity-run-binding-v5

Attack Matrix: [v5](attack-matrix/ATTACK-MATRIX-V5.md)

## Evidence Runtime Provenance v5.1

[Evidence Runtime Provenance v5.1](evidence-runtime-provenance-v5-1/README.md) binds completion evidence to the expected runtime stage/event and a deterministic provenance digest derived from the observed payload digest.

Run it with:

    go run ./experiments/evidence-runtime-provenance-v5-1

Attack Matrix: [v5.1](attack-matrix/ATTACK-MATRIX-V5-1.md)

## Evidence Runtime Event Binding v5.2

[Evidence Runtime Event Binding v5.2](evidence-runtime-event-binding-v5-2/README.md) requires each lifecycle evidence record to reference an observed runtime event with matching run identity, event type, sequence, and payload digest. It explicitly blocks missing events, cross-run events, payload mismatch, and events after the completion boundary.

Run it with:

    go run ./experiments/evidence-runtime-event-binding-v5-2

Attack Matrix: [v5.2](attack-matrix/ATTACK-MATRIX-V5-2.md)

## Evidence Runtime Event Ledger v5.3

[Evidence Runtime Event Ledger v5.3](evidence-runtime-event-ledger-v5-3/README.md) separates runtime observation from evidence construction through an append-only runtime event ledger and a read-only evidence builder. It blocks missing, cross-run, and late runtime events at the completion boundary.

Run it with:

    go run ./experiments/evidence-runtime-event-ledger-v5-3

Attack Matrix: [v5.3](attack-matrix/ATTACK-MATRIX-V5-3.md)

## Independent Runtime Observer v5.4

[Independent Runtime Observer v5.4](evidence-runtime-observer-v5-4/README.md) introduces an explicit read-only observer boundary between runtime execution and evidence construction. The observer owns event publication while evidence construction can only consume already-observed events.

Run it with:

    go run ./experiments/evidence-runtime-observer-v5-4

Attack Matrix: [v5.4](attack-matrix/ATTACK-MATRIX-V5-4.md)

## Durable Runtime Event Ledger v5.5

[Durable Runtime Event Ledger v5.5](evidence-durable-runtime-ledger-v5-5/README.md) adds durable append-only local persistence and integrity verification for runtime events.

Attack Matrix: [v5.5](attack-matrix/ATTACK-MATRIX-V5-5.md)

## Durable Runtime Receipt & Recovery v5.6

[Durable Runtime Receipt & Recovery v5.6](evidence-durable-receipt-v5-6/README.md) introduces an explicit durable receipt that binds a sealed runtime ledger to run identity, event boundaries, and a canonical ledger digest. It verifies the receipt after ledger reload and blocks foreign-run or tampered receipts.

Run it with:

    go run ./experiments/evidence-durable-receipt-v5-6

Attack Matrix: [v5.6](attack-matrix/ATTACK-MATRIX-V5-6.md)
