# Go Package Boundary Audit — Governance and Runtime

**Status:** Preliminary source inspection  
**Baseline:** `main` at `d6f1a55440e3b54303e432b6b35238071572b2f8`  
**Purpose:** Record concrete responsibility boundaries before any package moves or consolidation. This is not a complete import-graph analysis and does not authorize runtime changes.

## Findings from sampled source

| Package | Observed responsibility | Direct imports observed in sampled files | Boundary guidance |
|---|---|---|---|
| `internal/controlplane` | Action authorization gateway, grants/policy inputs, decision lifecycle, ledger events, verifier-facing evidence bundle | Standard library in sampled `gateway.go` and `evidence.go` | Treat as a cohesive execution-authorization domain until callers and contract overlap are mapped. |
| `internal/governance/control` | Policy-backed decisions, decision records, grant identity/digest, grant expiry and revocation-related state | `internal/errors`, `internal/governance/policy`, standard library | Treat as the policy/governance decision domain. Do not assume it duplicates `internal/controlplane` solely because both expose authorization-related concepts. |
| `internal/evidence` | Evidence records linked to policy decisions and execution outcomes; runtime evidence built from independently observed ledger events | `internal/errors`, `internal/governance/control`, standard library | Preserve the observed dependency direction from evidence to governance contracts unless a specific contract inversion is justified. |
| `internal/runtime/gateway` | Governed tool execution, request/invocation identity, observation and side-effect verification | `internal/errors`, `internal/governance/control`, standard library | Keep the execution adapter separate from policy evaluation; audit the full call path before extracting interfaces. |
| `internal/audit` | Application audit event records and file-backed audit logging | `internal/errors`, standard library in sampled `audit.go` | Do not merge with `internal/evidence` based on the shared word “audit”; persistence and integrity contracts may differ. |
| `internal/runtime/engine` | Runtime engine orchestration (per prior reconnaissance) | `internal/errors` and standard library in the sampled source | Keep separate from gateway pending full call-graph and test review. |

## Preliminary risks

1. **Conceptual overlap is not proof of duplication.** The control-plane and governance-control packages both model authorization, but the sampled files expose different request models and responsibilities. Consolidation without caller mapping could change policy semantics or break contracts.
2. **Evidence is not interchangeable with audit logging.** Evidence records bind decisions to execution and observations; the sampled audit package persists application audit events. Their integrity, retention, and verification requirements need separate treatment.
3. **Execution must remain downstream of authorization.** Runtime gateway imports governance-control contracts. Any extraction must preserve fail-closed behavior, grant revalidation, invocation identity, observation verification, and evidence provenance.
4. **Internal-only visibility does not eliminate compatibility risk.** CLI handlers, tests, examples, and integration workflows may consume these packages directly.
5. **The import graph is not yet reviewed.** Counts, fan-in/fan-out, and actual edge distribution must come from running `scripts/internal-import-graph.sh` against the intended commit.

## Next analysis sequence

1. Run `bash -n scripts/internal-import-graph.sh`.
2. Run `bash scripts/internal-import-graph.sh > /tmp/naeos-import-graph.tsv` from a checkout.
3. Compute fan-in and fan-out from the generated TSV and inspect all inbound/outbound edges for the five packages above.
4. Search direct callers and tests for exported types and constructors before proposing any boundary change.
5. Choose one low-risk refactor only when a concrete coupling issue is supported by graph and call-site evidence.
6. Add targeted regression tests and run package tests, full Go CI, and relevant integration checks.

## Decision

**No package move or consolidation is proposed by this audit.** Current evidence supports preserving the existing boundaries while completing the import-graph and caller analysis. Runtime behavior and public contracts must remain unchanged in the first code-refactor PR.

## Caller reconnaissance (source-search pass)

The following direct consumers were found in the baseline source. This is a targeted search result, not an exhaustive fan-in count.

| Package | Confirmed consumers | Refactor implication |
|---|---|---|
| `internal/governance/control` | `cmd/naeos/control_cmd.go`, `cmd/naeos/runtime_cmd.go`, `experiments/policy-bypass/layer_control.go`, `internal/evidence/runtime_evidence_builder.go`; governance-lifecycle and level-3 evidence experiments also name it as a core component | A change to decision or grant contracts crosses CLI, execution, evidence construction, and security experiments. Keep the policy decision API stable during any first extraction. |
| `internal/controlplane` | `internal/investordemo/controlplane_bridge.go` (identified in repository reconnaissance), control-plane golden-path and policy-change examples, plus the live-proof API documentation | Preserve investor-demo and control-plane proof flows; validate the demo and atomic-execution regression tests if this package is ever changed. |
| `internal/evidence` | Governance-lifecycle, level-3 evidence, evidence-tampering, and evidence-integrity experiments; referenced by the verification package's documented workflows | Evidence contracts are part of independent verification experiments, not just storage internals. Run targeted evidence and verifier suites after any contract change. |
| `internal/audit` | `internal/demoobs/siem.go`, `internal/observability/siem.go`, `internal/compliance/framework.go`, `internal/api/server.go`, and CLI handlers including compliance/observability commands | This is a cross-cutting application audit-event API with multiple consumers. Do not fold it into execution evidence without proving event-schema and retention compatibility. |

### What this changes

- The caller search confirms that `internal/governance/control` is a high-consequence boundary: it is consumed by CLI entry points, runtime execution, and evidence generation.
- `internal/audit` is consumed by API, compliance, observability, and demo-observability code, which supports keeping it separate from hash-chained execution evidence.
- `internal/controlplane` and `internal/governance/control` have distinct observed consumers and workflows. No duplicate API or safe consolidation has been established.
- Exact fan-in/fan-out counts remain unknown until the import-graph script can run in a Go checkout. The current environment could not clone GitHub, so neither script execution nor Go tests are claimed as complete.
