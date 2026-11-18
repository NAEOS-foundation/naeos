# NAEOS Authorization and Runtime Caller Map — Follow-up #001

- **Status:** Targeted static source review supplement
- **Review date:** 2026-11-15
- **Repository:** `NAEOS-foundation/naeos`
- **Baseline reference:** Main branch as inspected after merge of PR #810 (commit `73975c0cebc662f8e8d2a12c4adf2e5d8181b060`)
- **Parent map:** [NAEOS-AUTHORIZATION-RUNTIME-CALLER-MAP.md](NAEOS-AUTHORIZATION-RUNTIME-CALLER-MAP.md)
- **Tracking:** [Issue #713](https://github.com/NAEOS-foundation/naeos/issues/713), [NARP-001 #717](https://github.com/NAEOS-foundation/naeos/issues/717)

## Purpose

This supplement records subsequent caller/configuration and authorization-failure semantics evidence that post-dates the parent map. It does not claim exhaustive repository call-graph coverage, production deployment validation, or runtime execution. No package consolidation is proposed.

## Updated observations

### 1. Entrypoint-level runtime guarantees

The parent map's caller inventory should be interpreted per composition root, not as uniform runtime guarantees. The follow-up source survey recorded in [Issue #713](https://github.com/NAEOS-foundation/naeos/issues/713#issuecomment-6739750379) identifies:

| CLI entrypoint | Governance control + runtime gateway | Required decision revalidation | Replay protection | Observer / evidence qualification |
|---|---|---:|---:|---|
| `cmd/naeos/runtime_cmd.go` | Yes | Enabled | Not configured in this composition | Do not infer durable replay from gateway presence |
| `cmd/naeos/agent_request_cmd.go` | Yes | Enabled | Enabled; process-local unless `--replay-db` is supplied | Adapter policy grants are explicitly scoped |
| `cmd/naeos/runtime_bridge_cmd.go` | Yes | Enabled | Enabled | Filesystem observer and durable evidence reconciliation are configured in this bridge |
| `cmd/naeos/evidence_cmd.go` (`evidence log`) | Yes | Enabled | Not configured in this composition | Appends to the separate `internal/evidence` store after execution; not the same persistence path as runtime bridge ledger evidence |
| `cmd/naeos/control_cmd.go` (`control evaluate`) | Governance control only | Not an execution gateway | N/A | Standalone decision evaluation, no side effect |

These rows are configuration-level findings from source survey and must not be promoted to deployment-wide claims. Replay, revalidation, observation, and evidence durability are separate properties.

### 2. Evaluation failure semantics have changed since the parent review

The parent map's initial comparison described an evaluator-failure semantic mismatch. Subsequent security work addressed the scoped paths:

- **Governance control / CLI runtime:** PR [#714](https://github.com/NAEOS-foundation/naeos/pull/714) introduces typed internal `EvaluationFailure`. The error is not an authorization decision, does not mint a grant digest, and the runtime gateway must not reach the sandbox on indeterminate evaluation.
- **Control-plane evaluator / decision gateway:** PR [#716](https://github.com/NAEOS-foundation/naeos/pull/716), tracked by [Issue #715](https://github.com/NAEOS-foundation/naeos/issues/715), introduces `DecisionIndeterminate` / `ReasonEvaluationFailed`, records a separate `POLICY_EVALUATION_FAILED` event, and blocks normal/atomic execution on indeterminate evaluation. The source-level gateway inspection also shows that a ledger persistence failure prevents returning a usable authorization decision.

This is a scoped implementation update, not a statement that every caller or every deployed service has been validated. Test and CI results must be tied to the exact candidate commit before reporting verified runtime behavior.

### 3. Persistence/deployment remains a separate open proof boundary

The accepted [Supabase persistence ADR](ADR-PRODUCTION-CONTROL-PLANE-SUPABASE-PERSISTENCE.md) and open [Issue #786](https://github.com/NAEOS-foundation/naeos/issues/786) explicitly distinguish the investor-demo composition from a dedicated database-backed production Control Plane runtime. The ADR states that the current Railway service runs `naeos demo`, with policy/grant/approval state in memory and ledger durability dependent on optional snapshot configuration; `DATABASE_URL` alone is not evidence of Control Plane persistence.

Therefore:
- Do not label the current investor-demo HTTP service as the production persistence implementation.
- Do not infer Supabase persistence, multi-replica safety, durable approval lifecycle, or production readiness until the dedicated runtime and integration evidence are delivered.
- Keep CRM service/database outside this workstream, as required by #786.

## Architecture disposition

Retain the current bounded package responsibilities pending explicit integration objective and ADR:
- `internal/governance/control` remains the policy-decision contract in the observed CLI runtime path.
- `internal/runtime/gateway` remains the general CLI agent/tool execution boundary.
- `internal/controlplane` remains its separate action/grant/approval/ledger domain for the investor-demo bridge and focused examples until an approved migration establishes otherwise.

No package merge, deletion, policy gate relaxation, or production configuration change is proposed in this supplement.

## Verification status and limitations

- **Source inspection:** targeted file review and repository issue/PR records.
- **Clean-checkout Golden Path execution:** NOT RUN in this review.
- **Runtime integration/deployment verification:** NOT RUN in this review.
- **Exhaustive call-graph generation:** NOT RUN in this review.

This supplement is evidence of what the inspected source and recorded change descriptions establish; it is not a substitute for executable test evidence. The next verification action remains a clean-checkout Golden Path and, separately, runtime/deployment evidence for the relevant production boundary.
