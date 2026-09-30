# P2.2 External Run #001

## Objective

Execute the first independent external evaluation using the P2.1 Pilot Execution Package.

## Fixed boundary

- Repository: NAEOS-foundation/naeos
- Workflow: specification-to-service generation
- Control boundary: invalid policy configuration rejection
- Canonical path: `examples/demo-cli/run-demo.sh`
- Commit SHA: <fill before run>

## Evaluator

- Name / handle:
- Organization:
- Contact:
- Date:
- OS / architecture:
- Go version:

## Commands

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
git checkout <COMMIT_SHA>

go version
go build -o naeos ./cmd/naeos

rm -rf /tmp/naeos-pilot
NAEOS_BIN="$PWD/naeos" NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-pilot ./examples/demo-cli/run-demo.sh

find /tmp/naeos-pilot -maxdepth 3 -type f | sort
```

## Acceptance record

| Check | Result | Evidence / notes |
|---|---|---|
| Clean checkout | PASS / FAIL | |
| CLI build | PASS / FAIL | |
| Canonical demo | PASS / FAIL | |
| Specification | PASS / FAIL | |
| Derived NEIR | PASS / FAIL | |
| Validation | PASS / FAIL | |
| Policy rejection | PASS / FAIL | |
| AI context | PASS / FAIL | |
| Traceability | PASS / FAIL | |
| Generated artifacts | PASS / FAIL | |
| Evidence summary | PASS / FAIL | |

## Evidence anchors

```
run_id=
specification_hash=
neir_hash=
generated_artifact_count=
invalid_policy_rejection=
```

Metadata groups observed:

```
validation=
policy=
context=
audit=
stages=
```

## Independent verification

If a canonical serialized EvidenceBundle exists:

```bash
naeos evidence verify-bundle --input-file evidence.json
```

Record:

```
independent_verification=
independent_verification_result=
```

If no canonical bundle exists:

```
independent_verification=not_applicable
reason=
```

## Deviations

For every deviation:

| Field | Value |
|---|---|
| Step / command | |
| Expected | |
| Observed | |
| Commit SHA | |
| Run / artifact | |
| Environment | |
| Reproducible | YES / NO |
| Blocks pilot | YES / NO |
| Follow-up | |

## Determination

Select exactly one:

- [ ] REPRODUCED
- [ ] REPRODUCED WITH DEVIATION
- [ ] NOT REPRODUCED

## Observations

<Concise factual observations.>

## Next action

<Concrete action with a testable acceptance criterion.>
