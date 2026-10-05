# P2.2 External Run #001

## Objective

Execute the first independent external evaluation using the P2.1 Pilot Execution Package.

## Fixed boundary

- Repository: NAEOS-foundation/naeos
- Workflow: specification-to-service generation
- Control boundary: invalid policy configuration rejection
- Canonical path: `examples/demo-cli/run-demo.sh`
- Commit SHA: `55ec8ab069f5a5e3f27468a58838b1bb67212c05`
- Evaluation status: **PENDING EXECUTION**

> This record pins the current `main` state before execution. No PASS/FAIL or reproducibility claim is recorded until the validation workflow produces evidence.

## Evaluator

- Name / handle: automated external-validation workflow
- Organization: NAEOS Foundation repository CI
- Contact: repository maintainers
- Date: pending execution
- OS / architecture: pending execution
- Go version: pending execution

## Commands

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
git checkout 55ec8ab069f5a5e3f27468a58838b1bb67212c05

go version
go build -o naeos ./cmd/naeos

rm -rf /tmp/naeos-pilot
NAEOS_BIN="$PWD/naeos" NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-pilot ./examples/demo-cli/run-demo.sh

find /tmp/naeos-pilot -maxdepth 3 -type f | sort
```

## Acceptance record

| Check | Result | Evidence / notes |
|---|---|---|
| Clean checkout | PENDING | |
| CLI build | PENDING | |
| Canonical demo | PENDING | |
| Specification | PENDING | |
| Derived NEIR | PENDING | |
| Validation | PENDING | |
| Policy rejection | PENDING | |
| AI context | PENDING | |
| Traceability | PENDING | |
| Generated artifacts | PENDING | |
| Evidence summary | PENDING | |

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

- [ ] REPRODUCED
- [ ] REPRODUCED WITH DEVIATION
- [ ] NOT REPRODUCED
- [x] PENDING EXECUTION

## Observations

Pending execution. This section must be populated from observed workflow evidence, not from repository claims.

## Next action

Run the repository's `External Validation` workflow against this fixed commit and record the generated validation artifact and factual determination.
