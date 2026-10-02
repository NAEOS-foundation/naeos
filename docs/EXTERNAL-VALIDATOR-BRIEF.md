# NAEOS External Validator Program — Cohort #001

## Validator Brief

NAEOS is an open-source, vendor-neutral engineering layer for AI coding agents, focused on architecture, policy, execution, evidence, and verification.

We are looking for engineers outside the NAEOS development process to independently test whether documented workflows can actually be executed, observed, and evidenced.

> **Try to verify it. Try to break it. Document what actually happens.**

## Who We Are Looking For

- Software Engineers
- AI Engineers
- Platform / Infrastructure Engineers
- DevOps / SRE
- Security Engineers
- Software Architects
- Developer Experience practitioners
- Technical Researchers

Prior NAEOS experience is not required.

## What You Will Test

**Mission 001 — Verified Golden Path**
Execute the documented Golden Path independently.

**Mission 002 — Policy Change During Execution**
Observe behavior when policy or authorization changes after execution starts, using a supported scenario available at the evaluated commit.

**Mission 003 — Evidence & Handoff Verification**
Determine whether another engineer can reconstruct what happened from the resulting evidence and supported handoff artifacts.

## Commitment

Three missions over approximately 14 days. The workload depends on the execution environment and findings discovered.

## What You Submit

- environment;
- NAEOS version/commit;
- steps performed;
- expected behavior;
- actual behavior;
- evidence;
- result classification;
- findings and reproduction steps where applicable.

## Result Categories

`VERIFIED` · `PARTIAL` · `FAILED` · `BLOCKED` · `AMBIGUOUS` · `NOT TESTED`

There are no scores or rankings. Failures and blockers are valid findings.

## Independence

The first attempt must be performed without maintainer walkthrough or technical assistance. If assistance is needed, record the independent attempt first; assistance can then be provided and the workflow re-tested.

## Recognition

This is a zero-budget OSS community program. No financial compensation is promised.

With consent, validators may receive External Validator recognition, a public contribution record, attribution in the validation report, direct interaction with maintainers, and eligibility for future cohorts.

## Start Here

1. Read [`EXTERNAL-VALIDATION.md`](./EXTERNAL-VALIDATION.md).
2. Review [`EXTERNAL-VALIDATION-MISSIONS.md`](./EXTERNAL-VALIDATION-MISSIONS.md).
3. Use the exact commit designated for the cohort.
4. Submit results through the External Validation Result issue template.

> **Evidence determines the result.**