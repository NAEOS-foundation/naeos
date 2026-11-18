# NAEOS Authorization Contract Test Matrix

**Status:** Static test inventory / review; test execution is not claimed.
**Review date:** 2026-11-09
**Baseline:** `main` at `ed10af58ccd8566c238bdd03fd62772b15977ebd`
**Related:** [Issue #711](https://github.com/NAEOS-foundation/naeos/issues/711), [Issue #713](https://github.com/NAEOS-foundation/naeos/issues/713), [caller map](./NAEOS-AUTHORIZATION-RUNTIME-CALLER-MAP.md)

## Purpose and scope

This document records test evidence visible in the repository for the authorization/runtime contracts. It is a static inventory of named tests and inspected assertions, not a report of tests run in this session. It does not prove deployment composition or exhaustive coverage.

## Contract matrix

| Contract / path | Existing test evidence (source) | Static assessment |
|---|---|---|
| Runtime gateway evaluator failure is a non-decision and blocks sandbox | `internal/runtime/gateway/gateway_test.go`: `TestGatewayEvaluationFailureIsNotRecordedAsDecisionOrExecuted` | Asserts returned error, failed status, no decision value, structured failure metadata, zero sandbox calls, and history as failed/non-decision. |
| Governance/runtime canonical allow-deny and policy freshness | `internal/runtime/gateway/canonical_controlplane_invariants_test.go`: `TestCanonicalControlPlaneDenialBlocksSideEffect`, `TestCanonicalControlPlaneAllowPermitsSideEffect`, `TestCanonicalControlPlaneRevalidationBlocksPolicyChange`; malformed policy test also present | Focused invariant coverage exists for side-effect blocking, allow path, malformed policy activation, and revalidation on policy change. |
| Adapter grant boundary | `internal/runtime/gateway/gateway_test.go`: `TestAdapterPrivilegeBoundaryDeniesWithoutGrant`, `TestAdapterPrivilegeBoundaryAllowsExplicitGrant`; `cmd/naeos/agent_request_cmd_test.go`: `TestAgentRequestCommandRequiresExplicitAdapterPolicy` | Tests cover missing grant denial and explicit grant allowance, plus command-level requirement for explicit adapter policy. |
| Durable/shared replay protection | `internal/runtime/gateway/replay_store_durable_test.go`: `TestDatabaseInvocationStoreSurvivesStoreRestart`, `TestDatabaseInvocationStoreConcurrentSharedClaim`, `TestDatabaseInvocationStoreSharedAcrossGatewayInstances`, `TestAmbiguousExecutionIsUnknownAndCannotBeReplayedAfterRestart`; bridge test `TestRuntimeBridgeGatewayRejectsReplayBeforeSecondWrite` | Restart, concurrency, cross-instance and ambiguous-replay scenarios are named and present. CLI defaults/configuration differ by entrypoint; see caller map. |
| Runtime bridge evidence and side-effect lifecycle | `cmd/naeos/runtime_bridge_cmd_test.go`: `TestRuntimeBridgeEvidenceRejectsMutatedInvocation`, `TestRuntimeBridgeRejectsUnsupportedToolBeforeExecution`, `TestAppendRuntimeEvidenceSyncsJSONLRecord`, `TestWriteAtomicRuntimeReceiptPersistsCompleteRecord`; `cmd/naeos/runtime_bridge_durable_allow_test.go`: `TestRuntimeBridgeDurableAllowLifecycleWithRealSideEffectAndRecovery` | Evidence mutation, unsupported tools, sync/receipt, and durable recovery cases are present. |
| Evidence persistence failure after runtime side effect | `cmd/naeos/runtime_bridge_cmd_test.go`: `TestRuntimeBridgeSideEffectPersistsWhenEvidencePersistenceFails` | This test explicitly documents that the completed side effect remains after the subsequent evidence append fails; replay protection denies a blind retry. This is a known ordering/durability limitation, not atomic rollback. Consequential use cases requiring evidence-before-effect or transactional coupling need a separately scoped design and ADR. |
| Controlplane indeterminate evaluator failure, no decision or side effect | `internal/controlplane/issue_711_indeterminate_boundary_test.go`: `TestDecisionGateway_IndeterminateEvaluationIsAuditedAndCannotExecute` | Tests evaluator unavailable and unresolved policy version; asserts INDETERMINATE, no DecisionID, no side-effect callback, no execution receipt, and a policy-evaluation-failed ledger event. PR #773 adds structured failure metadata assertions. |
| Controlplane ledger persistence failure | `internal/controlplane/ledger_test.go`: `TestLedgerAppendDoesNotCommitWhenPersistenceFails`; plus `TestLedgerPersistence`, `TestLedgerPersistenceRejectsTampering` | Ledger append is asserted not committed when persistence fails. This package-level test does not by itself prove every caller checks persistence errors before consequential side effects. |
| Approval binding, reuse, and stale policy | `internal/controlplane/gateway_test.go`: `TestDecisionGateway_RequiresMatchingApprovalArtifactAndConsumesApproval`, `TestDecisionGateway_BlocksStaleAuthorizationAfterPolicyChange`, `TestDecisionGateway_ExecuteAtomicRejectsStaleAuthorizationBeforeSideEffect` | Relevant focused tests exist for approval artifact matching/consumption and stale policy blocks. Review of each caller’s usage remains necessary. |
| CLI command-level invocation and evidence store | `cmd/naeos/agent_request_cmd_test.go` and `cmd/naeos/evidence_cmd_test.go` | Agent request tests cover replay-store configuration and explicit adapter policy; evidence tests cover request ID persistence and persisted hash tampering. The sampled test inventory does not show a corresponding dedicated `runtime exec` command integration test. |

## Remaining verification work

1. Run focused package tests and record exact commands/results:
   - `go test ./internal/runtime/gateway`
   - `go test ./internal/controlplane`
   - `go test ./cmd/naeos`
2. Run repository-required CI and security gates; do not bypass failures.
3. Trace each command composition root to its tests and verify its replay/revalidation flags against the caller map.
4. Decide whether evidence persistence failure after a completed side effect is acceptable for each supported bridge deployment. If not, propose an explicit transactional/outbox or evidence-before-effect contract; do not imply the current test demonstrates atomicity.
5. Inspect investor-demo route ownership and confirm which gate is authoritative for each consequential action; maintain separate models until an ADR authorizes integration.

## Disposition

The static inventory indicates focused tests exist for the principal fail-closed, revalidation, replay, approval, ledger persistence, and evidence integrity contracts. It does **not** justify closing Issue #711 or #713 on its own: package tests, required CI/security checks, per-caller verification, and the documented post-side-effect evidence-persistence limitation still require resolution or explicit acceptance.

No production code, policy semantics, or gates are changed by this document.
