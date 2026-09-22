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
