# External Reproducible Case Study #001

## Purpose

This case study records the first independently executed NAEOS Golden Path validation as a concrete, reproducible technical observation.

It is intentionally narrower than a product case study: the evidence describes what an independent evaluator reproduced at a fixed commit. It does not claim production readiness, customer adoption, enterprise compliance, scalability, or blanket security.

## Evaluation record

| Field | Recorded value |
|---|---|
| Repository | `NAEOS-foundation/naeos` |
| Validated commit | `6e3b3a4cbd57404c191382b37801712ca74aef32` |
| Evaluator | Manus evaluator |
| Environment | Ubuntu 24.04, linux/amd64, Sandbox |
| Go | 1.26.6 |
| run_id | `pipe-1790911759522032327` |
| specification_hash | `3704afb986b588ef` |
| neir_hash | `a43775fcfc1c13d6` |
| Canonical demo exit | 0 |
| Human checklist | 12/12 PASS |
| Logical artifacts | 61 |
| Materialized files | 55 |
| Path collisions | 6 |

## Reproduction boundary

The evaluator reproduced the documented control-plane sequence:

```text
Specification
  → NEIR
  → Validation
  → Policy rejection
  → AI Context
  → Authorized Generation
  → Artifacts
  → Traceable Evidence
```

The baseline workflow did not require an LLM API key or an external service.

## Observed evidence

The independent report recorded:

- the canonical demo completed successfully;
- all 12 human validation checks passed;
- invalid policy configuration was rejected while the normal generation path remained executable;
- the evidence/run metadata contained the recorded traceability anchors;
- 61 logical artifacts were reported;
- 55 files were materialized;
- the six-file difference was explicitly reconciled as path collisions rather than silently treated as missing output.

The evaluator also independently checked generated Go license headers:

- `examples/demo-cli/.run/generated/api/package.go`
- `examples/demo-cli/.run/generated/auth/package.go`

Both contained the expected Apache-2.0 header.

## Deviation

Optional AI compilation was not executed because `NAEOS_LLM_API_KEY` was not supplied.

The evaluator classified this as non-blocking for the documented default path because the baseline Golden Path does not require an LLM API key.

## Factual determination

**REPRODUCED AS DOCUMENTED** for the documented baseline path at the fixed commit recorded above.

This determination is commit-scoped. It must not be silently carried forward to later commits.

In particular, a later `main` commit requires a fresh independent run before this case study can be described as validation of that later commit.

## Reproduction procedure

Start from a clean checkout and pin the exact validated commit:

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
git checkout 6e3b3a4cbd57404c191382b37801712ca74aef32

go version
go build -o naeos ./cmd/naeos

rm -rf /tmp/naeos-evaluation
NAEOS_BIN="$PWD/naeos" \
  NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-evaluation \
  ./examples/demo-cli/run-demo.sh

find /tmp/naeos-evaluation -maxdepth 3 -type f | sort
```

For the complete evaluator procedure and acceptance matrix, see:

- [External Evaluator Quickstart](EXTERNAL-EVALUATOR-QUICKSTART.md)
- [External Validation Runbook](EXTERNAL-VALIDATION.md)
- [Golden Path](GOLDEN-PATH.md)

## Evidence boundary

This case study demonstrates one independently reproduced repository workflow at one fixed commit.

It does not establish:

- production deployment readiness;
- customer adoption;
- enterprise or regulatory compliance;
- performance or scalability;
- security of every integration;
- correctness of arbitrary specifications;
- safety of every consequential external action.

Those claims require separate evidence.

## Relationship to current main

The repository may advance after an external validation run. Therefore every validation record must remain anchored to its exact commit.

At the time this case study was created, current `main` had advanced beyond the validated commit. A fresh validation against the newer `main` is a separate evaluation task.

**Evidence first. Claims second.**
