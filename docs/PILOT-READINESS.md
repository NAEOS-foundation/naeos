# NAEOS Pilot Readiness Pack

This pack turns the Golden Path, Reference Demo, External Validation, and Human Validation flow into one repeatable engineering evaluation.

## 1. Pilot objective

Evaluate one narrow engineering workflow end to end:

> Can an external engineer take a declared engineering intent through NAEOS and independently inspect validation, policy enforcement, generated AI context, generated artifacts, and traceability evidence?

The first pilot uses the existing CLI Reference Demo. No new product feature or alternate execution path is required.

## 2. Prerequisites

- clean Git checkout;
- Go version compatible with the repository's `go.mod`;
- no LLM API key required for the default path;
- isolated output directory;
- evaluator outside the implementation work for the human-validation record.

## 3. Canonical setup

~~~bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
git checkout <commit SHA>

go version
go build -o naeos ./cmd/naeos
rm -rf /tmp/naeos-pilot
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-pilot ./examples/demo-cli/run-demo.sh
~~~

Record the commit SHA before execution.

## 4. Evidence contract

| Field | Required |
|---|---|
| Repository | yes |
| Commit SHA | yes |
| Evaluator/reviewer | yes |
| Date/time UTC | yes |
| Environment/toolchain | yes |
| Scenario | yes |
| Demo command | yes |
| Execution result | yes |
| run_id | when produced |
| specification_hash | when produced |
| neir_hash | when produced |
| Generated artifact count | yes |
| Policy observation | yes |
| Deviations | yes, including explicit none |
| Evaluator feedback | yes |
| Next action | yes |

The contract is additive: feedback is recorded in the evaluation record and does not require changing the machine-readable execution record.

## 5. Reference use case

**Use case:** specification-driven project generation for a small engineering project.

**Input:** the checked-in Reference Demo specification.

**Observed chain:**

~~~text
Specification
  -> NEIR
  -> Validation
  -> Policy rejection test
  -> AI Context
  -> Authorized Generation
  -> Generated project artifacts
  -> Traceable run evidence
~~~

**Evaluation questions:**

1. Can an engineer understand the declared intent without undocumented context?
2. Can the derived NEIR be inspected and related back to that intent?
3. Is validation observable before generation?
4. Is the deliberately invalid policy configuration visibly rejected?
5. Can an evaluator inspect the context supplied to an AI agent?
6. Can generated artifacts be traced back to the run?
7. Can another engineer reproduce the path from the recorded commit?
8. Which step creates clarification or reproduction friction?

## 6. Acceptance rule

A pilot run is **Reproduced as documented** only when required checks pass and no blocking deviation is present. Otherwise record **Not reproduced** and preserve the deviation.

Do not infer adoption, production readiness, compliance, security, or general correctness from a successful pilot run.

## 7. Feedback intake

Convert externally observed feedback into engineering backlog only when it contains:

- evaluator observation;
- exact evaluation commit;
- evidence or reproduction details;
- impact on the pilot workflow;
- proposed clarification or change;
- whether the issue blocks reproduction.

Assumptions without external observation remain hypotheses and should not be promoted into the v3.7.0 backlog solely because they are plausible.

## 8. Validation documents

- `docs/EXTERNAL-HUMAN-VALIDATION.md`: independent evaluator record.
- `docs/EXTERNAL-VALIDATION.md`: automated/reproducible validation contract.
- `docs/GOLDEN-PATH.md`: canonical execution and acceptance definition.
- `docs/REFERENCE-DEMO.md`: evidence narrative and reviewer checklist.

## 9. Pilot output

The minimum pilot package is:

1. exact commit SHA;
2. completed human evaluation record;
3. automated validation record, if available for that commit;
4. referenced raw evidence;
5. deviations and evaluator observations;
6. next action.

This package is an engineering evaluation record, not a customer-adoption or production-readiness claim.
