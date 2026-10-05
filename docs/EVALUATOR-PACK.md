# NAEOS External Evaluator Pack

This pack is the shortest independent technical evaluation path for NAEOS. It is designed for an evaluator who did not participate in the implementation.

## Evaluation boundary

Evaluate a fixed repository commit. Do not treat repository documentation, CI status, or author claims as substitutes for observed behavior.

This pack evaluates the documented Golden Path and evidence controls. It does not establish production readiness, customer adoption, compliance, scalability, or safety of arbitrary external agent integrations.

## 1. Freeze the evaluation

Record before execution:

| Field | Value |
|---|---|
| Repository | NAEOS-foundation/naeos |
| Commit SHA | <fixed SHA> |
| Evaluator | <name or handle> |
| Date/time UTC | <timestamp> |
| OS | <OS/version> |
| Go/toolchain | <go version> |

Do not mix evidence from different commits without recording the difference.

## 2. Clean checkout

Run:

    git clone https://github.com/NAEOS-foundation/naeos.git
    cd naeos
    git checkout <fixed SHA>
    go version
    go build -o naeos ./cmd/naeos

Record the build exit status.

## 3. Run the canonical path

    rm -rf /tmp/naeos-evaluator
    NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-evaluator ./examples/demo-cli/run-demo.sh

Expected: exit status 0 and an isolated evidence directory.

Inspect:

    find /tmp/naeos-evaluator -maxdepth 3 -type f | sort

## 4. Acceptance checklist

Mark each item from observed evidence only.

- [ ] Clean checkout at the recorded SHA
- [ ] CLI builds successfully
- [ ] Canonical demo exits successfully
- [ ] spec.yaml is present and readable
- [ ] inspect.json exposes the derived NEIR
- [ ] validate.json exposes validation
- [ ] invalid-policy.log shows deliberate policy rejection
- [ ] context.md and context.json are generated
- [ ] run.json contains run_id
- [ ] run.json contains specification_hash
- [ ] run.json contains neir_hash
- [ ] run.json contains validation/policy/context/audit/stage metadata
- [ ] Expected generated artifacts exist
- [ ] summary.md exists and reports the evidence set

## 5. Record the observed anchors

    run_id:
    specification_hash:
    neir_hash:
    generated_artifact_count:

## 6. Adversarial checks

The evaluator should actively look for:

1. undocumented setup requirements;
2. steps that only work from a dirty checkout;
3. evidence that cannot be tied to the evaluated commit;
4. claims in documentation that are not supported by generated artifacts;
5. missing or ambiguous failure behavior;
6. generated evidence that cannot be inspected independently;
7. accidental dependence on an LLM API key for the default path.

Do not repair the repository during the evaluation. Record the deviation instead.

## 7. Deviation report

For every deviation record:

    Step:
    Expected:
    Observed:
    Commit SHA:
    Command/output:
    Environment/toolchain:
    Evidence/artifact:
    Impact on reproducibility:

## 8. Verdict

Choose exactly one:

- REPRODUCED — all required acceptance checks passed with no blocking deviation.
- REPRODUCED_WITH_DEVIATIONS — the documented path ran, but one or more deviations should be investigated.
- NOT_REPRODUCED — a required acceptance check failed or the evaluator could not reproduce the documented path.

## 9. Evaluator report template

    NAEOS Independent Evaluation

    Repository: NAEOS-foundation/naeos
    Commit SHA:
    Evaluator:
    Date/time UTC:
    OS:
    Go/toolchain:

    Build exit status:
    Demo exit status:

    run_id:
    specification_hash:
    neir_hash:
    generated_artifact_count:

    Acceptance checks:
    [ ] Clean checkout
    [ ] CLI build
    [ ] Canonical demo
    [ ] Specification evidence
    [ ] NEIR evidence
    [ ] Validation evidence
    [ ] Policy rejection evidence
    [ ] AI context evidence
    [ ] Traceability evidence
    [ ] Generated artifacts
    [ ] Evidence summary

    Adversarial observations:
    - None / <describe>

    Deviations:
    - None / <describe>

    Verdict:
    REPRODUCED / REPRODUCED_WITH_DEVIATIONS / NOT_REPRODUCED

    Notes:
    <independent observations>

## Evidence boundary

A successful evaluation means an external evaluator reproduced the documented repository behavior at the recorded commit. It is evidence of that bounded behavior only. It is not evidence of customer adoption or production readiness.
