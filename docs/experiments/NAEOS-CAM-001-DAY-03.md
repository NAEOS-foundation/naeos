# NAEOS Campaign Day 03 — Policy Change Mid-Run

Campaign ID: NAEOS-CAM-001  
Day: 03  
Theme: Policy change mid-run  
Status: Reproducible repository evidence

## Question

What happens when an already-authorized engineering action reaches the execution boundary after the policy that authorized it has been replaced?

## Experiment

This day uses the existing deterministic control-plane experiment:

`examples/control-plane-policy-change/`

Run:

```bash
./examples/control-plane-policy-change/run-demo.sh
```

or:

```bash
go run ./examples/control-plane-policy-change
```

The experiment requires no LLM, network access, credentials, or external service.

## Scenario

1. Policy v1 is active.
2. A `repository.write` authorization receives `ALLOW`.
3. Policy v2 becomes active before execution.
4. The original v1 authorization reaches the execution boundary.
5. The gateway checks authorization freshness against the active policy.
6. The stale authorization is rejected with `stale_policy`.
7. Execution records `EXECUTION_BLOCKED`.
8. The expected side effect does not occur.
9. Session evidence is independently verified.

## Acceptance criteria

A valid run must demonstrate:

- initial decision: `ALLOW`;
- execution decision: `DENY`;
- denial reason: `stale_policy`;
- execution event: `EXECUTION_BLOCKED`;
- no side-effect file;
- original authorization evidence retained;
- blocked execution evidence retained;
- session verification: `PASS`;
- process exits non-zero if an assertion fails.

## Engineering invariant

> An authorization decision is valid only while its bound policy version remains the active policy at the execution boundary.

This is an authorization-freshness property. It is not a claim about a specific AI model or arbitrary agent behavior.

## Publication evidence

Before publishing a Day 03 result, record:

- exact command;
- commit SHA;
- environment;
- UTC timestamp;
- process exit code;
- observed scenario results;
- repository path;
- any failure or deviation.

Do not fabricate runtime output.

## Why it matters

An agent can receive valid authorization and still execute later under a different governance state. The execution boundary therefore needs a freshness check rather than treating an earlier approval as permanently valid.

## Limitations

This experiment does not by itself prove production readiness, comprehensive security, compliance, customer adoption, or correctness of arbitrary AI-agent workflows. It demonstrates one concrete control-plane property: policy-version freshness at the execution boundary.

## Discussion question

If an AI coding agent receives approval to modify a repository and the governing policy changes before the write occurs, should the original approval remain valid?

## CTA

Run the experiment and compare the observed result with the documented invariant.
