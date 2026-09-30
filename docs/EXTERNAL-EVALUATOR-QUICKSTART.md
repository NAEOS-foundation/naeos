# NAEOS External Evaluator Quickstart

Use this when evaluating NAEOS independently from a clean checkout.

## 1. Fix the evaluation target

Choose one commit and record it before running anything:

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
git checkout <commit SHA>

go version
```

Do not mix evidence from different commits.

## 2. Run the canonical proof

```bash
go build -o naeos ./cmd/naeos

rm -rf /tmp/naeos-evaluation
NAEOS_BIN="$PWD/naeos" NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-evaluation ./examples/demo-cli/run-demo.sh
```

No LLM API key or external service is required for the baseline path.

## 3. Inspect evidence

```bash
find /tmp/naeos-evaluation -maxdepth 3 -type f | sort
```

Confirm at minimum:

- `spec.yaml`
- `inspect.json`
- `validate.json`
- `invalid-policy.log`
- `context.md`
- `context.json`
- `run.json`
- `generated/`
- `summary.md`

From `run.json`, record:

- `run_id`
- `specification_hash`
- `neir_hash`

## 4. Evaluate the control boundary

The baseline proof demonstrates:

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

A successful run does not by itself establish production readiness, customer adoption, enterprise compliance, or security of every external agent integration.

## 5. Record the result

Use [External Validation](EXTERNAL-VALIDATION.md) and [External Human Validation](EXTERNAL-HUMAN-VALIDATION.md) to record:

- exact commit and environment;
- acceptance checks;
- evidence anchors;
- deviations;
- evaluator observations;
- follow-up questions.

Then open an **External Validation Report** issue using the repository issue form.

## 6. Independent verification

For a canonical serialized `EvidenceBundle`, use:

```bash
naeos evidence verify-bundle --input-file evidence.json
```

P1.11 is read-only and independent of the live control plane. It verifies the evidence record; it does not re-run policy or execute an action.

## 7. What makes a useful evaluation

Do not only report whether the demo passed. Report what an independent engineer could understand, reproduce, challenge, or not reproduce.

The most valuable feedback identifies:

- unclear evidence boundaries;
- reproducibility friction;
- stale or missing documentation;
- policy or authorization assumptions;
- evidence that is difficult to inspect;
- workflow gaps that would matter in a real repository.

**Evidence first. Claims second.**
