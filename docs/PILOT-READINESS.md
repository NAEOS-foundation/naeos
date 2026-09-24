# NAEOS Pilot Readiness

## Purpose

This document turns the NAEOS Golden Path and Reference Demo into a repeatable engineering evaluation for an external design partner.

The first pilot should test one narrow engineering workflow end-to-end. The goal is to observe reproducibility, traceability, policy behavior, generated artifacts, and operator experience — not to establish production readiness or customer adoption.

## Pilot flow

Golden Path → Reference Demo → External Validation → Human Evaluation → Pilot Feedback → Evidence-backed backlog

## Recommended first use case

**Specification-to-service generation with policy and evidence inspection.**

The evaluator starts from a declared service specification, runs the canonical NAEOS pipeline, verifies the derived NEIR and validation result, exercises the deliberately invalid policy case, inspects AI context and execution metadata, and reviews generated Go/TypeScript project artifacts.

## Prerequisites

- Clean checkout of the repository.
- Go version recorded from go.mod.
- No LLM API key required for the canonical baseline.
- Git available.
- A clean temporary output directory.

## Canonical execution

    git clone https://github.com/NAEOS-foundation/naeos.git
    cd naeos
    git checkout <commit SHA>

    go version
    go build -o naeos ./cmd/naeos

    rm -rf /tmp/naeos-pilot
    NAEOS_BIN="$PWD/naeos" NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-pilot ./examples/demo-cli/run-demo.sh

## Evidence to inspect

- spec.yaml
- inspect.json
- validate.json
- invalid-policy.log
- context.md
- context.json
- run.json
- generated/
- summary.md

The traceability contract requires run_id, specification_hash, neir_hash, validation metadata, policy metadata, context metadata, audit metadata, and stage metadata.

## Pilot evidence contract

| Field | Required |
|---|---|
| Repository + commit SHA | Yes |
| Reviewer / organization | Yes |
| UTC timestamp | Yes |
| OS / architecture | Yes |
| Go/toolchain version | Yes |
| Scenario / specification | Yes |
| Demo command | Yes |
| Exit status | Yes |
| run_id | Yes |
| specification_hash | Yes |
| neir_hash | Yes |
| Invalid-policy rejection | Yes |
| Generated artifact count | Yes |
| Deviations | Yes |
| Reviewer observations | Yes |
| Next action | Yes |

## Evaluation questions

Record observations, not scores.

1. Can an engineer reproduce the workflow from a clean checkout without private context?
2. Is the specification-to-NEIR transition understandable from the evidence?
3. Is validation behavior inspectable?
4. Is the invalid-policy rejection observable and unambiguous?
5. Does run.json provide enough information to trace the execution?
6. Are generated artifacts easy to locate and inspect?
7. Which evidence artifacts are useful in a real engineering workflow?
8. Where does the documented workflow differ from actual operator experience?
9. What additional integration or workflow would make the pilot useful in the evaluator's environment?
10. What should be changed before another engineer repeats the pilot?

## Deviation handling

A deviation is not silently corrected. Record the expected behavior, observed behavior, exact command/input, environment, evidence artifact or log, whether it blocks the pilot, and proposed follow-up.

## Determination

- REPRODUCED — documented workflow completed and acceptance evidence was observed.
- REPRODUCED WITH DEVIATION — workflow completed with documented deviations.
- NOT REPRODUCED — required acceptance evidence could not be established.

Do not convert a successful pilot into a production-readiness, compliance, security, scalability, or adoption claim.

## Feedback → v3.7.0 intake

A pilot observation becomes a v3.7.0 candidate only when:
1. The observation is reproducible or clearly documented.
2. It affects an identifiable engineering workflow.
3. The proposed change has a defined acceptance criterion.
4. The change does not duplicate an existing capability.
5. The need is supported by external evaluation evidence or repeated internal evidence.
6. The scope fits the v3.7.0 release train.

## Pilot record template

NAEOS Pilot Evaluation

repository=
commit_sha=
reviewer=
organization=
utc_timestamp=
os_architecture=
go_version=

scenario=
demo_command=
exit_status=

run_id=
specification_hash=
neir_hash=
generated_artifact_count=
invalid_policy_rejection=

determination=
deviations=
observations=
next_action=

## Related evidence

- docs/GOLDEN-PATH.md
- docs/REFERENCE-DEMO.md
- docs/EXTERNAL-VALIDATION.md
- docs/EXTERNAL-HUMAN-VALIDATION.md
- Issue #244 — first reproducibility record
- Issue #247 — Pilot Readiness milestone

## Boundaries

This pilot package does not establish customer adoption, production readiness, enterprise compliance, performance/scalability, security of every external AI provider or agent, arbitrary specification correctness, or safety of every consequential external action.
