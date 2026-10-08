# NAEOS External Validation Runbook

This runbook defines the smallest independent technical evaluation of the NAEOS control-plane path.

It is designed for a reviewer who has not participated in the implementation and wants to verify observable behavior from a clean repository checkout.

## 1. Validation objective

The evaluation asks:

> Can an independent engineer reproduce the canonical NAEOS workflow and connect declared intent, derived representation, validation, policy, execution, artifacts, and traceability evidence without relying on screenshots or undocumented assumptions?

The runbook verifies the existing Golden Path and Reference Demo. It does not introduce a second demo path.

The evaluation should distinguish two proof boundaries: the Golden Path demonstrates the runnable local control-plane workflow; P1.11 provides a separate read-only independent verifier for serialized canonical evidence.

## 2. Fixed evaluation record

Record these values before running:

| Field | Value |
|---|---|
| Repository | NAEOS-foundation/naeos |
| Commit SHA | <commit SHA> |
| Go/toolchain version | <output of go version> |
| Demo command | NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh |
| Start time (UTC) | <timestamp> |
| Reviewer | <name or handle> |

The commit SHA is the version under evaluation. Do not mix evidence from different commits without recording the difference.

## 3. Reproduce from a clean checkout

Use a fresh clone or clean working tree.

~~~bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
git checkout <commit SHA>

go version
go build -o naeos ./cmd/naeos

rm -rf /tmp/naeos-demo
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh
~~~

Expected result: the command exits successfully and creates the isolated evidence directory.

## 4. Verify the evidence chain

Inspect the generated files:

~~~bash
find /tmp/naeos-demo -maxdepth 3 -type f | sort
~~~

Verify at minimum:

- spec.yaml — declared engineering intent
- inspect.json — derived NEIR representation
- validate.json — validation result
- invalid-policy.log — deliberate policy rejection
- context.md and context.json — generated AI context
- run.json — execution and traceability record
- generated/ — generated project artifacts
- summary.md — evidence summary

## 5. Traceability verification

Inspect run.json and confirm these anchors are present:

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

Record the observed run_id, specification_hash, and neir_hash in the evaluation report.

The hashes establish traceability between the execution record and the evaluated engineering inputs. They do not by themselves prove the correctness or trustworthiness of every external dependency.

## 6. Policy boundary verification

Confirm that invalid-policy.log records rejection of the deliberately invalid policy configuration.

Expected result:

- the invalid configuration is rejected;
- the rejection is observable in the evidence output;
- the normal generation path remains independently executable.

This is evidence of the demonstrated local policy boundary, not a blanket security claim.

## 7. Artifact verification

Confirm the expected generated artifacts exist:

~~~text
generated/README.md
generated/go.mod
generated/package.json
~~~

Record the artifact count from summary.md.

## 8. Acceptance matrix

| Check | Evidence | Result |
|---|---|---|
| Clean checkout | repository + commit SHA | PASS / FAIL |
| CLI builds | build command exit status | PASS / FAIL |
| Canonical demo completes | demo exit status | PASS / FAIL |
| Specification is observable | spec.yaml | PASS / FAIL |
| NEIR is observable | inspect.json | PASS / FAIL |
| Validation is observable | validate.json | PASS / FAIL |
| Invalid policy is rejected | invalid-policy.log | PASS / FAIL |
| AI context is generated | context.md, context.json | PASS / FAIL |
| Run is traceable | run.json | PASS / FAIL |
| Generated artifacts exist | generated/ | PASS / FAIL |
| Evidence summary exists | summary.md | PASS / FAIL |

A validation report should not mark the overall run successful if a required check fails. Record deviations rather than silently omitting them.

### What a successful validation proves

It proves that the documented repository workflow was reproduced at the recorded commit and that its required evidence artifacts matched the acceptance matrix. It does not establish production readiness, customer adoption, enterprise compliance, or the safety of every external AI-agent integration.

## 9. Validation report template

Copy this section into an issue, discussion, or partner evaluation record:

~~~text
NAEOS External Validation

Repository: NAEOS-foundation/naeos
Commit SHA:
Go/toolchain:
Reviewer:
Date/time (UTC):

Demo command:
Exit status:

run_id:
specification_hash:
neir_hash:
Generated artifact count:

Checks:
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

Deviations:
- None / <describe>

Conclusion:
- Reproduced as documented / Not reproduced

Notes:
<observations, environment differences, or follow-up questions>
~~~

## 10. Handling deviations

A deviation is any observed difference between the documented acceptance contract and the actual run.

For each deviation record:

1. exact command or step;
2. expected result;
3. observed result;
4. commit SHA;
5. relevant run identifier or artifact;
6. environment/toolchain details;
7. whether the deviation blocks reproducibility.

Do not convert a deviation into a success by changing the acceptance criteria during the same evaluation.

## 11. Runtime boundary validation

For P3.1 external-adopter validation, the reviewer must also exercise the live protocol-neutral bridge. The repository test suite is not sufficient evidence for this gate.

Build the current commit and run the bridge against an isolated filesystem root:

~~~bash
go build -o naeos ./cmd/naeos
rm -rf /tmp/naeos-runtime /tmp/naeos-runtime-evidence.jsonl
mkdir -p /tmp/naeos-runtime

cat > /tmp/naeos-request.jsonl <<'EOF'
{"request_id":"p31-allow-001","invocation_id":"p31-inv-001","actor":"external-validator","tool":"filesystem","action":"write","resource":"filesystem","payload":{"path":"observed.txt","content":"NAEOS-P31"}}
{"request_id":"p31-replay-001","invocation_id":"p31-inv-replay","actor":"external-validator","tool":"filesystem","action":"write","resource":"filesystem","payload":{"path":"replay.txt","content":"one"}}
{"request_id":"p31-replay-001","invocation_id":"p31-inv-replay","actor":"external-validator","tool":"filesystem","action":"write","resource":"filesystem","payload":{"path":"replay.txt","content":"two"}}
EOF

./naeos runtime bridge --filesystem-root /tmp/naeos-runtime --evidence-file /tmp/naeos-runtime-evidence.jsonl < /tmp/naeos-request.jsonl

cat /tmp/naeos-runtime/observed.txt
cat /tmp/naeos-runtime/replay.txt
cat /tmp/naeos-runtime-evidence.jsonl
~~~

Acceptance for this runtime boundary:

- the first ALLOW produces exactly one observable side effect;
- the replayed invocation is denied before a second side effect;
- each successful receipt contains request_id, invocation_id, invocation_digest, policy identity, decision, execution status, and observation identity;
- the observer's digest equals the digest of the exact governed request;
- an independently recomputed receipt digest verifies without the agent process;
- mutation of payload or context changes the invocation digest and invalidates the prior evidence binding;
- durable replay must be exercised with --replay-db when restart/cross-worker recovery is claimed.

The canonical runtime receipt is intentionally distinct from sandbox stdout. A successful stdout message is not evidence unless the independently observed artifact is also bound to the same invocation digest.
## 12. Evidence boundaries

This runbook establishes a reproducible technical evaluation of the repository path. It does not establish:

- production deployment readiness;
- customer adoption;
- enterprise compliance;
- performance or scalability targets;
- security of every external AI provider or agent;
- correctness of arbitrary specifications;
- safety of every consequential external action.

Those claims require additional evidence appropriate to the claim.

## 13. Relationship to the canonical path

There is intentionally one first-run execution path:

~~~text
Golden Path
    ↓
Reference Demo
    ↓
External Validation
    ↓
Evaluation Record
~~~

- docs/GOLDEN-PATH.md defines the runnable acceptance contract.
- docs/REFERENCE-DEMO.md explains the evidence story.
- This document defines how an independent reviewer records and evaluates the observed evidence.

**Evidence should be reproducible before it is persuasive.**
