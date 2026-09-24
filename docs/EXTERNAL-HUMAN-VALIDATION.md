# NAEOS External Human Validation Packet

This packet is for a technical evaluator who is **not part of the NAEOS implementation team** and wants to independently reproduce and inspect the canonical NAEOS engineering-control-plane path.

## 1. Evaluation objective

Answer this question from the evidence produced by the repository:

> Can an external engineer reproduce the documented NAEOS Golden Path from a clean checkout, observe the specification → NEIR → validation → policy → context → execution → artifacts chain, and record the result without relying on undocumented maintainer claims?

The evaluator should make the PASS/FAIL determination from the observed run.

## 2. Independence requirement

For this record to count as **External Human Validation**:

- the evaluator should not have implemented the behavior being evaluated;
- the evaluator should use a clean checkout;
- the evaluator should record the exact commit SHA under test;
- the evaluator should record environment/toolchain details;
- the evaluator should preserve deviations instead of silently changing acceptance criteria;
- the evaluator should submit the completed record to a GitHub issue, discussion, pull request, or other durable review record.

A GitHub Actions run maintained by the NAEOS repository is useful reproducibility evidence, but it is **not** by itself an independent human review.

## 3. Fixed evaluation inputs

| Field | Value |
|---|---|
| Repository | NAEOS-foundation/naeos |
| Commit SHA | `<commit SHA>` |
| Reviewer | `<name or handle>` |
| Organization | `<organization or independent>` |
| Date/time (UTC) | `<timestamp>` |
| OS / architecture | `<value>` |
| Go/toolchain | `<output of go version>` |

Do not mix artifacts from different commits without explicitly recording the difference.

## 4. Clean-checkout procedure

Use a fresh clone or clean working tree.

~~~bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
git checkout <commit SHA>

go version
go build -o naeos ./cmd/naeos

rm -rf /tmp/naeos-human-validation
NAEOS_BIN="$PWD/naeos" NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-human-validation ./examples/demo-cli/run-demo.sh
~~~

Expected result: the build succeeds and the canonical demo exits successfully.

The default demo does not require an LLM API key. Do not configure one merely to complete this validation unless the evaluation specifically intends to test the optional AI compilation path.

## 5. Evidence inspection

~~~bash
find /tmp/naeos-human-validation -maxdepth 3 -type f | sort
~~~

At minimum, verify:

- `spec.yaml` — declared engineering intent
- `inspect.json` — derived NEIR representation
- `validate.json` — validation result
- `invalid-policy.log` — deliberate policy rejection
- `context.md` and `context.json` — generated AI context
- `run.json` — execution and traceability record
- `generated/` — generated project artifacts
- `summary.md` — evidence summary

The evaluator should inspect the actual contents rather than treating file existence alone as proof.

## 6. Traceability checks

Open `run.json` and verify these anchors:

~~~text
run_id
specification_hash
neir_hash
validation metadata
policy metadata
context metadata
audit metadata
stage metadata
~~~

| Anchor | Observed value |
|---|---|
| run_id | `<value>` |
| specification_hash | `<value>` |
| neir_hash | `<value>` |

The hashes connect the execution record to the evaluated engineering input and derived representation. They do not establish that every external dependency is trustworthy.

## 7. Policy-boundary check

Inspect `invalid-policy.log`.

Mark PASS only if the deliberately invalid policy configuration is rejected and the rejection is observable in the evidence output.

This demonstrates the policy boundary exercised by this scenario. It is not a blanket security claim about every policy, deployment, or integration.

## 8. Generated-artifact check

Verify at least these representative artifacts:

~~~text
generated/README.md
generated/go.mod
generated/package.json
~~~

Record the generated artifact count from `summary.md`.

## 9. Human validation acceptance matrix

| Check | Required evidence | Result |
|---|---|---|
| Clean checkout | repository + exact commit SHA | PASS / FAIL |
| Environment recorded | OS/architecture + Go version | PASS / FAIL |
| CLI build | successful build | PASS / FAIL |
| Canonical demo | successful exit status | PASS / FAIL |
| Specification observable | `spec.yaml` | PASS / FAIL |
| NEIR observable | `inspect.json` | PASS / FAIL |
| Validation observable | `validate.json` | PASS / FAIL |
| Invalid policy rejected | `invalid-policy.log` | PASS / FAIL |
| AI context generated | `context.md`, `context.json` | PASS / FAIL |
| Run traceable | `run.json` anchors | PASS / FAIL |
| Generated artifacts present | `generated/` | PASS / FAIL |
| Evidence summary present | `summary.md` | PASS / FAIL |

### Overall determination

Use one of these exact conclusions:

- **REPRODUCED** — every required check passed and no reproducibility-blocking deviation was observed.
- **NOT REPRODUCED** — one or more required checks failed.
- **REPRODUCED WITH DEVIATION** — the documented path completed, but a material deviation was observed and does not invalidate the evaluator's reproduction.

Do not use a positive conclusion when a required check failed.

## 10. Deviation record

For every deviation, record the exact command or step, expected result, observed result, commit SHA, relevant run identifier or artifact, environment/toolchain, and whether it blocks reproducibility.

~~~text
Deviation #1
Step:
Expected:
Observed:
Commit SHA:
Run/artifact:
Environment:
Blocks reproducibility: Yes / No
~~~

Do not change the acceptance criteria during the same evaluation to turn a failure into a pass.

## 11. Final external validation record

~~~text
NAEOS External Human Validation

Repository: NAEOS-foundation/naeos
Commit SHA:
Reviewer:
Organization:
Date/time (UTC):
OS / architecture:
Go/toolchain:

Command:
NAEOS_BIN="$PWD/naeos" NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-human-validation ./examples/demo-cli/run-demo.sh

Exit status:

run_id:
specification_hash:
neir_hash:
Generated artifact count:
Invalid policy rejection: Observed / Not observed

Acceptance:
[ ] Clean checkout
[ ] Environment recorded
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

Deviations:
- None / <describe>

Conclusion:
- REPRODUCED / NOT REPRODUCED / REPRODUCED WITH DEVIATION

Reviewer notes:
<observations>
~~~

## 12. Relationship to existing evidence

~~~text
Golden Path
    ↓
Reference Demo
    ↓
External Validation Runbook
    ↓
GitHub-hosted reproducibility record
    ↓
External Human Validation
~~~

The repository's automated run demonstrates that the documented path can execute on a clean GitHub-hosted runner. This packet adds the missing human layer: an external engineer independently inspects the behavior and records a determination.

## 13. Evidence boundaries

This packet does **not** establish production deployment readiness, customer adoption, enterprise compliance, performance or scalability targets, security of every external AI provider or agent, correctness of arbitrary specifications, or safety of every consequential external action.

**The purpose of this packet is reproducibility and independent inspection, not a blanket product certification.**