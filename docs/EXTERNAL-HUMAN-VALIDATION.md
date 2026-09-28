# NAEOS External Human Validation

This packet is for a technical evaluator outside the implementation work. It records human observation separately from automated GitHub Actions validation.

## Boundary

Automated External Validation proves reproducible execution on a clean runner. This packet records an evaluator's own observation. Neither record implies customer adoption, production readiness, compliance, security of every integration, or correctness of arbitrary specifications.

## Fixed evaluation target

Record before execution:

| Field | Value |
|---|---|
| Repository | NAEOS-foundation/naeos |
| Commit SHA | <exact commit under evaluation> |
| Evaluator | <name or handle> |
| Date/time UTC | <timestamp> |
| Environment | <OS, architecture, tool versions> |
| Toolchain | <go version> |

Do not mix evidence from different commits without recording the difference.

## Canonical procedure

From a clean checkout:

~~~bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
git checkout <commit SHA>

go version
go build -o naeos ./cmd/naeos

rm -rf /tmp/naeos-human-validation
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-human-validation ./examples/demo-cli/run-demo.sh
~~~

Inspect the evidence:

~~~bash
find /tmp/naeos-human-validation -maxdepth 3 -type f | sort
~~~

## Human observation checklist

Mark each item only from observed repository output.

- [ ] Clean checkout matches the recorded commit SHA.
- [ ] CLI builds successfully.
- [ ] Canonical demo completes successfully.
- [ ] `spec.yaml` exposes the declared engineering intent.
- [ ] `inspect.json` exposes the derived NEIR.
- [ ] `validate.json` records specification validation.
- [ ] `invalid-policy.log` shows deliberate policy rejection.
- [ ] `context.md` and `context.json` expose generated AI context.
- [ ] `run.json` contains `run_id`, `specification_hash`, and `neir_hash`.
- [ ] Expected generated artifacts are present.
- [ ] `summary.md` records the run evidence.
- [ ] Observed behavior matches `docs/GOLDEN-PATH.md`.

## Evaluation record

~~~text
NAEOS External Human Validation

Repository: NAEOS-foundation/naeos
Commit SHA:
Evaluator:
Date/time (UTC):
Environment:
Toolchain:

Demo command:
Exit status:

run_id:
specification_hash:
neir_hash:
Generated artifact count:
Observed policy rejection:

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

Result:
- Reproduced as documented / Not reproduced

Deviations:
- None / <describe>

Evaluator observations:
<what was clear, unclear, surprising, or difficult to reproduce>

Follow-up questions:
<questions or evidence gaps>

Next action:
<repeat / clarify documentation / open engineering issue / no action>
~~~

## Deviation discipline

For each deviation record the exact step, expected result, observed result, commit SHA, environment/toolchain, relevant run identifier or artifact, and whether it blocks reproducibility.

Do not alter acceptance criteria during an evaluation to turn a deviation into a pass.

## Evidence boundary

The evaluator's report is an observation record. Attach it to a GitHub issue, discussion, or pilot report together with the exact commit and generated evidence. Human validation is not a substitute for security review, production qualification, compliance assessment, or customer evidence.
