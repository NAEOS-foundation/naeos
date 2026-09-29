# NAEOS Live Control Plane Evidence

This document records the public production proof for the NAEOS live Control Plane demo.

It complements the repository Golden Path and Reference Demo. The Golden Path proves the local engineering workflow; this document records the externally reachable policy-evaluation path.

## 1. Production surface

Public page:

- `https://www.naeos.dev/control-plane/`

Production Control Plane endpoint:

- `https://naeos-control-plane-production.up.railway.app`

The public page sends authorization-evaluation requests to the production Control Plane. The live demo evaluates authorization only; it does not execute a side effect.

## 2. Demonstrated scenarios

The live demo exposes two explicit scenarios:

| Scenario | Capability | Expected decision |
|---|---|---|
| ALLOW | `repository.read` | `ALLOW` |
| DENY | `production.deploy` | `DENY` |

Default demo identity:

- Agent: `agent-payment-01`
- Artifact hash: `sha256:demo-artifact`

These scenarios correspond to the Control Plane policy boundary used by the demo. `repository.read` is an allowed capability, while `production.deploy` is a protected capability.

## 3. Production smoke acceptance

A live smoke check is considered successful when all of the following are observable from the public page:

1. The Control Plane endpoint is shown as ready.
2. The `ALLOW · repository.read` scenario can be selected.
3. Evaluation returns `ALLOW`.
4. The `DENY · production.deploy` scenario can be selected.
5. Evaluation returns `DENY`.
6. The result exposes the decision metadata returned by the Control Plane.
7. No side effect is executed by the demo.

The production smoke check performed on 2026-09-29 confirmed the two interactive scenarios were present and functioning in the public UI.

## 4. Evidence boundary

This smoke check establishes that the published demonstration can exercise the production authorization-evaluation path.

It does **not** establish:

- authorization for arbitrary capabilities;
- security of every production integration;
- distributed transaction semantics;
- enterprise compliance;
- scalability or availability targets;
- safety of arbitrary consequential actions.

Those properties require separate evidence and verification.

## 5. Reviewer procedure

A reviewer can reproduce the public check without relying on repository internals:

1. Open the public Control Plane page.
2. Confirm the endpoint status is ready.
3. Select **ALLOW · repository.read**.
4. Run the evaluation.
5. Confirm the returned decision is `ALLOW`.
6. Select **DENY · production.deploy**.
7. Run the evaluation.
8. Confirm the returned decision is `DENY`.
9. Inspect the returned policy, reason, execution status, and evidence information where available.

For a deeper implementation-level review, use:

- [Golden Path](GOLDEN-PATH.md)
- [Reference Demo & Evidence Story](REFERENCE-DEMO.md)
- [External Validation Runbook](EXTERNAL-VALIDATION.md)

## 6. Milestone status

This evidence closes the functional acceptance criterion for **P1.6 — Live Control Plane Demo**:

```text
P1.6
 ├─ Public live endpoint       PASS
 ├─ ALLOW scenario             PASS
 ├─ DENY scenario              PASS
 ├─ Production UI integration  PASS
 └─ No demo side effect        PASS
```

The next engineering work should build on this proof rather than changing the P1.6 demo path without a new acceptance requirement.

**Evidence principle:** the public demo should make the policy boundary observable, not merely described.
