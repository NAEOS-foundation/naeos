# NAEOS Reference Demo & Evidence Story

The Reference Demo is the externally consumable evidence narrative built on the canonical Golden Path. It answers one question:

> Can an engineer independently inspect how one declared engineering intent moves through NAEOS and becomes a generated artifact with traceable execution metadata?

This document is intentionally evidence-oriented. It describes what a reviewer can observe, how to reproduce it, and what the run does **not** prove.

## 1. Run the reference demo

From a clean checkout:

```bash
go build -o naeos ./cmd/naeos
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh
```

The default path does not require an LLM API key. The optional AI compilation stage runs only when `NAEOS_LLM_API_KEY` is configured.

Inspect the resulting evidence set:

```bash
find /tmp/naeos-demo -maxdepth 3 -type f | sort
```

The canonical script fails if a required stage or expected artifact is missing.

## 2. Evidence chain

| Question | Evidence | Reviewer check |
|---|---|---|
| What was intended? | `spec.yaml` | Read the declared engineering intent |
| What representation did NAEOS derive? | `inspect.json` | Inspect the derived NEIR |
| Was the specification validated? | `validate.json` | Confirm validation output |
| Can policy reject a disallowed configuration? | `invalid-policy.log` | Confirm the deliberate policy failure |
| What context reaches an agent? | `context.md`, `context.json` | Compare generated context with the specification |
| What was executed? | `run.json` | Inspect run identity, stages, validation, policy, context, and audit metadata |
| What artifacts resulted? | `generated/` | Inspect generated project files |
| Can the result be summarized after execution? | `summary.md` | Confirm the evidence summary and artifact count |

## 3. Traceability contract

A successful reference run must expose:

- `run_id`
- `specification_hash`
- `neir_hash`
- validation metadata
- policy metadata
- context metadata
- audit metadata
- stage metadata

The hashes and run identifier are the minimum traceability anchors connecting the execution record to the declared engineering input and derived representation.

This is an inspectable contract, not a claim that every external system is automatically trustworthy.

## 4. Policy evidence

The demo intentionally submits an invalid policy configuration before the normal generation run.

The expected behavior is rejection with a policy-evaluation failure. This demonstrates the control-plane property that a policy decision can prevent a disallowed configuration from proceeding.

The test is deliberately local and deterministic; it should not be interpreted as proof of every policy boundary in every deployment.

## 5. Generation evidence

The successful run verifies the presence of representative generated files, including:

- `generated/README.md`
- `generated/go.mod`
- `generated/package.json`

The generated artifact set is counted and recorded in `summary.md`.

The important evidence relationship is:

```text
specification
     ↓
validated / represented
     ↓
policy checked
     ↓
context derived
     ↓
run executed
     ↓
artifacts + metadata
```

## 6. Independent reviewer checklist

A technical evaluator can use this sequence without trusting the documentation narrative:

1. Start from a clean checkout of the repository.
2. Build the CLI from source.
3. Run the canonical demo with an isolated output directory.
4. Confirm the script exits successfully.
5. Inspect `spec.yaml` and `inspect.json`.
6. Inspect the deliberate policy rejection in `invalid-policy.log`.
7. Inspect `run.json` and verify the required traceability fields.
8. Inspect `context.md` and `context.json`.
9. Inspect the generated project files.
10. Compare the observed result with the acceptance criteria in `docs/GOLDEN-PATH.md`.

The evaluator should treat the repository files and generated outputs as the primary evidence, rather than relying on screenshots or claims in this document.

## 7. CI relationship

The Golden Path is already exercised by repository automation and the CLI test suite. That makes the reference demo a regression contract as well as a presentation path.

When a change modifies a control-plane stage or its output contract, the corresponding Golden Path acceptance criteria and evidence map should be reviewed in the same change.

## 8. Scope and non-claims

The Reference Demo establishes a reproducible engineering workflow and inspectable evidence artifacts.

It does **not** by itself establish:

- production deployment readiness;
- customer adoption;
- enterprise compliance;
- performance or scalability targets;
- security of every external AI provider or agent;
- correctness of arbitrary specifications;
- safety of every consequential external action.

Those properties require separate tests, deployment evidence, threat models, operational controls, and independent verification.

## 9. Partner-pilot use

For a technical pilot, use the same repository path and ask the evaluator to record:

- repository commit SHA;
- Go/toolchain version;
- demo command;
- demo exit status;
- generated artifact count;
- `run_id`;
- `specification_hash`;
- `neir_hash`;
- observed policy rejection;
- any deviations from the acceptance criteria.

This produces a compact, reproducible evidence record that can be attached to a GitHub issue, discussion, or partner evaluation report.

## 10. Evidence principle

**Do not ask the reviewer to trust the claim when the repository can expose the evidence.**

The Reference Demo therefore prioritizes:

```text
Declared intent
    ↓
Inspectable representation
    ↓
Explicit policy decision
    ↓
Observable execution
    ↓
Traceable artifacts
    ↓
Independent review
```

The Golden Path proves the runnable sequence. The Reference Demo makes the resulting evidence legible to an external engineer.
