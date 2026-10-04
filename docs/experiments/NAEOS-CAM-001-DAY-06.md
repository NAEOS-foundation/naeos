# NAEOS-CAM-001 — Day 06: Real Local Side Effect

**Campaign:** NAEOS-CAM-001  
**Day:** 06  
**Experiment:** NAEOS Level-3 Evidence Experiment v1  
**Status:** Publication-ready experiment brief

## Question

Can NAEOS demonstrate that a policy decision is connected to a real local side effect, independent observation, durable evidence, and independent verification?

## Reproduction

```bash
go run ./experiments/level3-evidence
```

Expected exit code: `0`.

## Scenarios

1. **ALLOW:** the execution gateway authorizes a filesystem write; an independent observer reads the resulting file; evidence records the observed artifact digest; verification passes.
2. **DENY:** the gateway denies the request; the sandbox is not invoked; the observer confirms no side effect.
3. **Direct bypass:** a file is written outside the gateway; the observer detects the side effect while gateway history contains no authorized execution. This is recorded as a governance failure.
4. **Execution/observation separation:** reported completion without persistence is detected by the independent observer.
5. **Tamper after observation:** changing the observed artifact after evidence capture causes independent verification to fail.

## Acceptance

The experiment passes only when every expected scenario outcome is observed, including expected detection of governance failures.

The key invariant is:

> A policy decision is not proof that an unauthorized side effect was prevented.

## Evidence boundary

This experiment demonstrates a concrete local filesystem side effect and independent observation under a controlled harness. It does **not** prove production security, OS-level isolation, deployment-wide enforcement, or that an external agent cannot bypass the execution boundary.

The direct-bypass scenario is intentionally retained because detecting an out-of-band side effect is different from preventing every possible out-of-band write.

## Publication rule

Publish only the exact runtime result from the evidence workflow, including commit SHA, workflow run, exit code, and artifact digest. Do not generalize the result into a production-security claim.
