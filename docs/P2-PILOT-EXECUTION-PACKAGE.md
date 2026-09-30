# P2.1 Pilot Execution Package

## Purpose

This package operationalizes the first NAEOS external technical pilot without introducing a second demo path.

Canonical flow:

    Clean checkout → fixed commit → CLI build → canonical demo / Golden Path
    → policy rejection evidence → traceability evidence
    → independent EvidenceBundle verification → evaluator record

The objective is to establish whether an engineer who did not build NAEOS can reproduce and inspect the control-plane workflow from repository evidence.

This pilot does not establish production readiness, customer adoption, enterprise compliance, scalability, or security of arbitrary external AI-agent integrations.

## 1. Fixed pilot boundary

| Boundary | Definition |
|---|---|
| Repository | NAEOS-foundation/naeos |
| Workflow | Specification-to-service generation |
| Control boundary | Deliberately invalid policy configuration must be rejected before the normal generation path |
| Evidence boundary | spec.yaml, inspect.json, validate.json, invalid-policy.log, context artifacts, run.json, generated artifacts, summary.md |
| Independent verification | naeos evidence verify-bundle --input-file evidence.json when a canonical serialized EvidenceBundle is available |
| External dependency | None for the baseline canonical run |
| Optional AI stage | Only when NAEOS_LLM_API_KEY is explicitly configured |

The pilot must not expand its acceptance boundary during execution.

## 2. Evaluation identity

Before running the pilot, record:

| Field | Value |
|---|---|
| Repository | NAEOS-foundation/naeos |
| Commit SHA | <commit SHA> |
| Reviewer | <name or handle> |
| Organization | <organization> |
| UTC start time | <timestamp> |
| OS / architecture | <value> |
| Go/toolchain | <output of go version> |

Use one fixed commit for the complete baseline run. Do not combine evidence from different commits without explicitly recording the difference.

## 3. Clean execution

From a clean checkout:

    git clone https://github.com/NAEOS-foundation/naeos.git
    cd naeos
    git checkout <commit SHA>

    go version
    go build -o naeos ./cmd/naeos

    rm -rf /tmp/naeos-pilot
    NAEOS_BIN="$PWD/naeos" NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-pilot ./examples/demo-cli/run-demo.sh

Expected result:
- CLI build succeeds.
- Canonical demo exits successfully.
- Evidence is isolated under /tmp/naeos-pilot.
- No LLM API key is required for the baseline path.

The canonical script is the only first-run execution path for this pilot.

## 4. Baseline acceptance matrix

Inspect generated output:

    find /tmp/naeos-pilot -maxdepth 3 -type f | sort

| Check | Required evidence | Result |
|---|---|---|
| Clean checkout | repository + commit SHA | PASS / FAIL |
| CLI build | build exit status | PASS / FAIL |
| Canonical demo | demo exit status | PASS / FAIL |
| Specification | spec.yaml | PASS / FAIL |
| Derived NEIR | inspect.json | PASS / FAIL |
| Validation | validate.json | PASS / FAIL |
| Policy rejection | invalid-policy.log | PASS / FAIL |
| AI context | context.md, context.json | PASS / FAIL |
| Traceability | run.json | PASS / FAIL |
| Generated artifacts | generated/ | PASS / FAIL |
| Evidence summary | summary.md | PASS / FAIL |

A required failed check means the baseline was not reproduced as documented. Record the deviation; do not silently change the acceptance criteria.

## 5. Evidence anchors

From run.json, record:

    run_id=
    specification_hash=
    neir_hash=

Also confirm these metadata groups exist:

    validation
    policy
    context
    audit
    stages

Minimum traceability relationship:

    declared specification
           ↓
    derived NEIR
           ↓
    validation
           ↓
    policy decision
           ↓
    execution
           ↓
    artifacts + trace metadata

The hashes establish traceability to the evaluated inputs and derived representation. They do not independently prove correctness of external dependencies.

## 6. Policy control-boundary test

The canonical demo creates invalid-naeos.yaml with a deliberately invalid policy condition and expects the run to fail with a policy-evaluation failure.

Evaluator acceptance:
1. The invalid policy configuration is rejected.
2. The rejection is observable in invalid-policy.log.
3. The normal generation path remains executable.
4. The evaluator records observed behavior rather than inferring it.

This is evidence of the demonstrated local policy boundary, not a blanket security claim.

## 7. Generated-artifact check

Confirm at minimum:

    generated/README.md
    generated/go.mod
    generated/package.json

Record the artifact count from summary.md.

The exact generated artifact set may evolve. The canonical script remains authoritative for required-file validation.

## 8. Independent verification boundary

The external evaluator may independently verify a serialized canonical EvidenceBundle:

    naeos evidence verify-bundle --input-file evidence.json

This verifier is read-only and independent of the live control plane.

Important distinction:
- The canonical demo proves the runnable local workflow.
- The EvidenceBundle verifier checks a serialized evidence record.
- Verification does not re-run policy or execute an action.

If no canonical EvidenceBundle is produced by the scenario under evaluation, record:

    independent_verification=not_applicable
    reason=<why no serialized EvidenceBundle was available>

Do not fabricate an evidence bundle solely to satisfy the checklist.

## 9. Pilot determination

Use exactly one factual determination:
- REPRODUCED — documented workflow completed and required acceptance evidence was observed.
- REPRODUCED WITH DEVIATION — workflow completed but one or more documented deviations were observed.
- NOT REPRODUCED — required acceptance evidence could not be established.

The determination describes the observed pilot result. It is not a production-readiness or commercialization judgment.

## 10. Deviation protocol

For every deviation record:

| Field | Value |
|---|---|
| Step / command | <exact step> |
| Expected | <documented behavior> |
| Observed | <actual behavior> |
| Commit SHA | <SHA> |
| Run / artifact | <reference> |
| Environment | <OS/toolchain> |
| Reproducible | YES / NO |
| Blocks pilot | YES / NO |
| Follow-up | <issue or action> |

Do not repair the repository or alter acceptance criteria during the same evaluation and then report the original run as clean.

If a fix is required, preserve the original evidence and perform a new run against the new commit.

## 11. Evaluator record

Copy this record into the external evaluation issue:

    NAEOS P2.1 Pilot Evaluation

    repository=NAEOS-foundation/naeos
    commit_sha=
    reviewer=
    organization=
    utc_start=
    os_architecture=
    go_version=

    scenario=specification-to-service generation
    control_boundary=invalid policy configuration rejection
    demo_command=NAEOS_BIN="$PWD/naeos" NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-pilot ./examples/demo-cli/run-demo.sh
    exit_status=

    run_id=
    specification_hash=
    neir_hash=
    generated_artifact_count=
    invalid_policy_rejection=

    independent_verification=
    independent_verification_result=

    determination=
    deviations=
    observations=
    next_action=

## 12. Evidence-to-backlog rule

A pilot finding becomes an engineering backlog candidate only when:
1. The observation is reproducible or precisely documented.
2. It affects an identifiable engineering workflow.
3. The proposed change has a testable acceptance criterion.
4. It does not duplicate an existing capability.
5. The need is supported by external evaluation evidence or repeated internal evidence.

Classify follow-up findings as:
- Documentation / onboarding
- Product / control-plane behavior
- Integration / workflow
- Security / governance
- Evidence / verification

Do not turn an isolated preference into a product requirement without evidence.

## 13. Canonical references

This package composes the existing evidence contracts; it does not replace them:
- docs/GOLDEN-PATH.md — runnable acceptance contract
- docs/REFERENCE-DEMO.md — evidence narrative
- docs/EXTERNAL-EVALUATOR-QUICKSTART.md — independent quickstart
- docs/EXTERNAL-VALIDATION.md — validation runbook
- docs/EXTERNAL-EVALUATOR-SCORECARD.md — evaluator evidence record
- docs/PILOT-READINESS.md — pilot boundary and feedback contract
- examples/demo-cli/README.md — canonical CLI demo instructions

## 14. P2.1 exit gate

P2.1 is complete when:
- this package is merged to main;
- an evaluator can identify one fixed commit and one canonical execution path;
- the control boundary and acceptance evidence are explicit;
- deviations have a defined recording protocol;
- the evaluator record captures the minimum traceability anchors;
- the next run can be executed without creating a competing demo workflow.

The next step is P2.2 — evaluator intake and first external run, using this package as the execution contract.

**Evidence first. Claims second.**