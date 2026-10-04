# Day 01 — Can an AI Agent Action Be Independently Verified?

**Campaign:** NAEOS-CAM-001
**Experiment:** Governance Lifecycle v1
**Question:** Can an engineering action be traced from agent intent through policy, execution, observation, evidence, and independent verification?

## The experiment
The experiment exercises the NAEOS governance lifecycle:

`AGENT INTENT → POLICY DECISION → AUTHORIZED ACTION → EXECUTION → OBSERVATION → EVIDENCE → INDEPENDENT VERIFICATION`

It uses deterministic local components rather than an LLM or network service.

## Scenarios
1. Complete lifecycle — trace a permitted action end-to-end.
2. Agent claim vs observation — demonstrate that an agent-reported success is not itself proof of the side effect.
3. Artifact mutation after approval — bind approval to an exact artifact hash.
4. Stale authorization replay — prevent an authorization issued under an old policy version from surviving a policy change.

## Acceptance criteria
- The experiment exits with code `0` when all assertions pass.
- Output is machine-readable JSON.
- Evidence can be independently verified.
- A characterization result is reported as a characterization result, not as a blanket security claim.

## Reproduction
From the NAEOS repository root:

```bash
go run ./experiments/governance-lifecycle
```

## Publication evidence
Record the exact command, observed exit code, relevant JSON output, commit SHA, environment, and timestamp used for the published run.

Do not publish this page as a claim of production security, regulatory compliance, or universal AI-agent safety. The experiment tests specific governance invariants.

## Discussion
**If an AI coding agent says an action succeeded, what independent evidence should an engineering system require before accepting that claim?**

## Primary CTA
**Run the experiment and compare your observed result.**
