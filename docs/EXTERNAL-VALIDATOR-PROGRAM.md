# NAEOS External Validator Program

**Version:** 1.0.0  
**Status:** Proposed  
**Program Owner:** NAEOS OSS  
**Budget:** Zero-budget / community-driven  
**Initial Cohort:** Cohort #001  
**Duration:** 14 days

## Purpose

The NAEOS External Validator Program is an independent validation program for engineers outside the NAEOS development process.

The program tests whether documented NAEOS workflows can be independently exercised, observed, and evidenced.

> **Try to verify it. Try to break it. Document what actually happens.**

This is an evidence program, not a testimonial program.

## Relationship to the Existing Validation Runbook

The existing [`docs/EXTERNAL-VALIDATION.md`](./EXTERNAL-VALIDATION.md) remains the canonical first-run technical evaluation runbook.

This program adds the cohort, independence, evidence-submission, review, and public-reporting layer around that runbook.

The program intentionally does not introduce a competing Golden Path.

```text
Golden Path
    ↓
Reference Demo
    ↓
External Validation Runbook
    ↓
Validator Cohort
    ↓
Evidence Submission
    ↓
Maintainer Review
    ↓
Remediation / Re-test
    ↓
Public Validation Report
```

## Objectives

1. Validate the NAEOS Golden Path through external engineers.
2. Identify failures, ambiguities, documentation gaps, and unexpected behavior.
3. Determine whether generated evidence is reproducible.
4. Test behavior when policy or execution state changes.
5. Produce a public validation record that can be independently reviewed.

## Validator Profile

The program is open to engineers with experience in Software Engineering, AI Engineering, Platform Engineering, DevOps/SRE, Security Engineering, Software Architecture, Developer Experience, or Technical Research.

Prior NAEOS experience is not required.

## Independence

During the first attempt, validators use the public repository and documentation, do not receive a maintainer walkthrough or technical solution, record blockers and ambiguities, and preserve evidence from the actual execution.

If assistance is required, the independent attempt must be recorded first.

## Cohort #001

| Item | Target |
|---|---|
| Validators | 5 |
| Missions | 3 |
| Duration | 14 days |
| Budget | $0 |
| Public report | 1 |

The three missions are defined in [`EXTERNAL-VALIDATION-MISSIONS.md`](./EXTERNAL-VALIDATION-MISSIONS.md).

## Result Classification

Results are observations, not scores:

- **VERIFIED** — acceptance criteria satisfied with sufficient evidence.
- **PARTIAL** — some acceptance criteria satisfied.
- **FAILED** — acceptance criteria not satisfied.
- **BLOCKED** — execution could not continue because of an identified blocker.
- **AMBIGUOUS** — expected behavior or requirement was not sufficiently clear.
- **NOT TESTED** — mission was not executed.

Failures and blockers are valid findings. The program provides no incentive for positive results.

## Evidence Requirements

Each mission submission should include Validator ID, Mission ID, NAEOS version/commit, environment, preconditions, steps performed, expected behavior, actual behavior, evidence, result, findings, reproduction steps where applicable, and recommendations where applicable.

Do not silently replace or remove evidence. If an artifact changes, preserve the revision trail.

## Security

Validators must stay within the mission scope.

Do not exploit third-party systems, access unauthorized data, perform denial-of-service attacks, perform destructive testing, or use credentials belonging to another person.

Security-sensitive findings should follow the repository's responsible disclosure process rather than being posted publicly in an issue.

## Recognition

This is a zero-budget community program. No financial compensation is promised.

With validator consent, participants may receive NAEOS External Validator recognition, a public validation/contribution record, attribution in the public validation report, direct interaction with maintainers, and eligibility for future validation cohorts.

Attribution is opt-in.

## Completion Criteria

A cohort is complete when missions have been completed or formally closed, evidence has been submitted, findings reviewed, remediation documented where applicable, re-testing performed where appropriate, final results classified, and the public report can be published.

## Public Reporting

The final report should distinguish **Observed Fact**, **Validator Interpretation**, and **NAEOS Maintainer Response**.

Recommended sections: Program Methodology, Validator Cohort, Tested Version, Environment Matrix, Mission Results, Verified Paths, Failures, Ambiguities, Documentation Gaps, Remediation, Re-test Results, Open Findings, and Limitations.

## Core Principle

The validator is not testing whether NAEOS looks correct.

The validator is testing whether documented behavior can be independently observed and evidenced.

> **Evidence determines the result.**