# NAEOS Campaign Day 04 — Stale Authorization

Campaign ID: NAEOS-CAM-001  
Day: 04  
Theme: Stale authorization  
Status: Reproducible repository evidence

## Question

Can an authorization issued under an earlier policy version silently survive after governance state changes?

## Experiment

Day 04 reuses the deterministic policy-change control-plane experiment:

`examples/control-plane-policy-change/`

Run:

```bash
./examples/control-plane-policy-change/run-demo.sh
```

or:

```bash
go run ./examples/control-plane-policy-change
```

## Test path

- Policy v1 is active.
- A `repository.write` request is authorized under v1.
- Policy v2 replaces v1 before execution.
- The old authorization is presented at the execution boundary.
- Freshness validation compares the authorization's bound policy version with the active policy.
- The stale authorization is denied rather than replayed.
- No consequential side effect is allowed through the stale authorization.

## Acceptance criteria

The experiment must demonstrate:

- v1 authorization initially receives `ALLOW`;
- active policy changes to v2;
- v1 authorization is rejected at execution;
- denial reason is `stale_policy`;
- execution records `EXECUTION_BLOCKED`;
- the protected side effect does not occur;
- evidence for the original authorization remains distinguishable from blocked execution evidence;
- verification passes;
- a failed assertion causes a non-zero process exit.

## Engineering invariant

> A previously valid authorization must not silently become an evergreen permission after the policy state that produced it has changed.

This is a specific authorization-freshness property, not a general security or AI-agent safety claim.

## Why this matters

Long-running AI coding tasks create time between authorization and execution. During that interval, policy, repository state, risk posture, or required controls may change. Treating an old approval as permanently valid can create a governance gap.

## Publication evidence

Record only actual execution results:

- exact command;
- commit SHA;
- environment;
- UTC timestamp;
- exit code;
- observed denial reason;
- observed execution event;
- side-effect check;
- verification result.

Do not fabricate runtime output.

## Limitations

This experiment does not prove production readiness, comprehensive security, compliance, customer adoption, or correctness of arbitrary AI-agent workflows. It demonstrates one tested stale-authorization boundary.

## Discussion question

Should an AI agent be allowed to continue using an approval after the policy version that produced that approval is no longer active?

## CTA

Run the experiment and inspect the stale-authorization behavior.
