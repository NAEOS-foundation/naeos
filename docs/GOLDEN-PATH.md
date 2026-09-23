# NAEOS Golden Path

The Golden Path is the repository's smallest reproducible proof of the NAEOS engineering control plane. It is designed for a first-time engineer, evaluator, or integration partner to run locally without an external service or AI API key.

## Objective

Prove this sequence end to end:

```text
Specification
    ↓
NEIR
    ↓
Validation
    ↓
Policy
    ↓
AI Context
    ↓
Authorized Generation
    ↓
Artifacts
    ↓
Traceable Evidence
```

The Golden Path is implemented by the checked-in CLI demo at [`examples/demo-cli/`](../examples/demo-cli/).

## Five-minute run

From a clean checkout:

```bash
go build -o naeos ./cmd/naeos
./examples/demo-cli/run-demo.sh
```

No LLM API key is required for the default path. AI compilation is an optional stage when `NAEOS_LLM_API_KEY` is configured.

To isolate output outside the repository:

```bash
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh
```

## Acceptance criteria

A Golden Path run is successful only when all of these are true:

1. `spec.yaml` is accepted by the specification validator.
2. The CLI can inspect the derived NEIR.
3. A deliberately invalid policy configuration is rejected.
4. AI context is generated as both Markdown and JSON.
5. The normal generation pipeline completes.
6. `run.json` contains `run_id`, `specification_hash`, `neir_hash`, validation, policy, context, audit, and stage metadata.
7. Expected generated project files exist.
8. A summary is written under the isolated run directory.

The repository CI executes the same demo script with an isolated temporary output directory, so the Golden Path is a regression gate rather than documentation-only. The CLI test suite also invokes the canonical script.

## Evidence map

| Stage | Repository input/output | What it demonstrates |
|---|---|---|
| Specification | `examples/demo-cli/spec.yaml` | Engineering intent is declared explicitly |
| NEIR | `examples/demo-cli/.run/inspect.json` | Intent is materialized into the canonical engineering representation |
| Validation | `validate.json` | The specification is checked before generation |
| Policy | `invalid-policy.log` | A disallowed configuration fails closed |
| AI Context | `context.md`, `context.json` | Agent-facing context is derived from the engineering model |
| Generation | `generated/` | The same specification drives project artifacts |
| Traceability | `run.json` | Run identity and specification/NEIR hashes connect execution to intent |
| Evidence | `summary.md` plus run metadata | The result is inspectable after the process completes |

## What the Golden Path does not claim

The demo is a reproducible engineering path, not a production deployment or benchmark. It does not by itself establish customer adoption, enterprise compliance, performance characteristics, or the safety of every external AI-agent integration.

Those claims require separate evidence and verification.

## Reuse

Use this path as the baseline for:

- technical evaluations and partner pilots;
- README and website onboarding;
- demonstrations of specification-driven AI engineering;
- regression testing of the control-plane contract;
- collecting reproducible evidence in GitHub issues or discussions.

When a feature changes the control-plane stages or their contracts, update the Golden Path and its acceptance criteria in the same change.

**Golden Path principle:** one engineering intent, one reproducible run, inspectable evidence from specification to generated artifact.
