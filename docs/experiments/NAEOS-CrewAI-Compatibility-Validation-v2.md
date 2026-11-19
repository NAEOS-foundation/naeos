# NAEOS CrewAI Compatibility Validation v2

**Version:** 0.1  
**Status:** PARTIAL RUNTIME EVIDENCE RECORDED — NOT A COMPLETE V2 CONFORMANCE RESULT  
**Branch:** `feat/crewai-v2-harness-foundation`  
**Scope:** NAEOS external validation against CrewAI runtime and adapter execution paths.

> This document defines the v2 validation plan. It is not a conformance result, security certification, or claim that any listed adapter is currently covered.

## 1. Objective

Establish reproducible evidence that NAEOS authorization is enforced at the actual consequential execution boundary—not merely by calling a hook dispatcher directly—and that the decision is bound to the exact candidate, current policy, subject, scope, and expiry.

The validation must distinguish:

- authorization decision from execution outcome;
- framework hook invocation from actual tool-body execution;
- blocked dispatch from independently observed absence of side effects;
- known failure from an uncertain outcome requiring reconciliation.

## 2. Version matrix

Run the same mandatory scenarios against these pinned CrewAI versions:

| Track | CrewAI version | Purpose |
|---|---:|---|
| Baseline | 1.15.16 | Reproduce the version used by the existing v1 workflow |
| Current target | 1.15.44 | Verify against the latest release identified when this plan was prepared |

Do not silently float dependency versions. Record Python version, lockfile/dependency resolution, CrewAI version, NAEOS commit, test command, and UTC timestamp in every evidence bundle.

If a pinned version cannot install or run on the selected Python/runtime, mark that track BLOCKED with the exact failure. Do not substitute another version and report it as the requested track.

## 3. Execution-path inventory

**Static source inspection is recorded separately:** [NAEOS CrewAI Source Execution-Path Inventory v2](./NAEOS-CrewAI-Source-Path-Inventory-v2.md). The inventory identifies hook-mediated and suspected bypass paths in tags 1.15.16 and 1.15.44. It is source inspection only; all runtime paths remain NOT EXECUTED until the harness produces evidence.

The harness must identify and exercise, where supported by the pinned CrewAI version:

1. ReAct/text-based CrewAI agent tool execution through Crew/Task.
2. Native function/tool-calling execution through Crew/Task.
3. Flow declarative `ToolAction` execution.
4. OpenAI Agents tool adapter conversion and invocation.
5. LangGraph tool adapter conversion and invocation.
6. Any additional tool dispatch path identified by source inspection or runtime tracing.

For each path record:

- path name and CrewAI version;
- public API or adapter entry point used;
- before/after hook registration mechanism;
- whether the NAEOS enforcement callback was observed;
- whether the tool body was entered;
- whether the side-effect ledger changed;
- source file / symbol or upstream commit used to identify the path.

A path that is unavailable in a pinned version is BLOCKED/NOT_APPLICABLE with evidence and rationale; it must not be counted as PASS. A path that reaches the tool body without enforcement is FAIL.

## 3A. Current runtime evidence — DENY boundary probes

The current implementation has executed a limited DENY boundary probe for the following CrewAI paths and adapter invocation boundaries. These results are not the V2-01 through V2-12 scenario suite and do not establish exact candidate binding, policy freshness, replay control, UNKNOWN reconciliation, or production readiness.

| Execution path | CrewAI 1.15.16 | CrewAI 1.15.44 | Runtime evidence |
|---|---|---|---|
| ReAct via public `Crew.kickoff()` | PASS — hook observed; zero body entries and side effects | PASS — hook observed; zero body entries and side effects | [Run 44596426405](https://github.com/NAEOS-foundation/naeos/actions/runs/44596426405) |
| Native tool calling via public `Crew.kickoff()` | PASS — native path selected; hook observed; zero body entries and side effects | PASS — native path selected; hook observed; zero body entries and side effects | [Run 44596426405](https://github.com/NAEOS-foundation/naeos/actions/runs/44596426405) |
| Flow ToolAction via public `Flow.kickoff()` | FAIL — hook not dispatched; one body entry and side effect | FAIL — hook not dispatched; one body entry and side effect | [Run 44596426405](https://github.com/NAEOS-foundation/naeos/actions/runs/44596426405) |
| OpenAI Agents Runner via the CrewAI adapter | FAIL — hook not dispatched; one body entry and side effect | FAIL — hook not dispatched; one body entry and side effect | [Run 44596426405](https://github.com/NAEOS-foundation/naeos/actions/runs/44596426405) |
| LangGraph ToolNode via the CrewAI adapter | UNKNOWN — returned unawaited coroutine; no hook/body/side effect observed | UNKNOWN — returned unawaited coroutine; no hook/body/side effect observed | [Run 44596426405](https://github.com/NAEOS-foundation/naeos/actions/runs/44596426405) |

The Flow and OpenAI Agents adapter FAIL results are enforcement bypasses for the tested paths: the tool body and independent side-effect ledger were entered under DENY. LangGraph remains UNKNOWN because the converted async wrapper was returned as an unawaited coroutine; no policy enforcement conclusion can be drawn. The initial full ToolNode attempt was separately BLOCKED by a LangGraph config error.

The v2 harness and raw JSON artifacts are maintained in PR [#847](https://github.com/NAEOS-foundation/naeos/pull/847). The PR remains draft. Do not aggregate these path-level results into an overall PASS.

## 4. Normative invariants

- **V2-I1 — Exclusive enforcement:** every tested dispatch path reaches the declared NAEOS enforcement point before the tool body. No ungoverned fallback path is acceptable.
- **V2-I2 — Exact candidate binding:** authorization is bound to a versioned canonical representation of tool identity, action, input, subject, and relevant context. Any mutation invalidates the authorization.
- **V2-I3 — Fresh policy:** policy identity/version and decision expiry are checked at the final enforcement boundary.
- **V2-I4 — Fail closed:** missing, malformed, expired, mismatched, unavailable, or unverifiable decision evidence blocks execution.
- **V2-I5 — Decision/outcome separation:** authorization does not imply execution or success. Decision and outcome are separate records correlated by stable identifiers.
- **V2-I6 — Single-use/replay control:** replay of a consumed decision or duplicate invocation is rejected according to a documented idempotency contract.
- **V2-I7 — UNKNOWN handling:** when a non-idempotent side effect may have committed but its response is lost, record UNKNOWN and prohibit automatic retry until authoritative reconciliation or a proven idempotency mechanism permits it.
- **V2-I8 — Independent observation:** side-effect assertions use a test-owned ledger or isolated artifact inspected independently of the agent's returned text.

## 5. Mandatory scenario matrix

| ID | Scenario | Injection / setup | Required result |
|---|---|---|---|
| V2-01 | Fresh PROCEED | Exact candidate, current policy, valid unexpired decision | Tool body executes once; outcome recorded separately |
| V2-02 | Candidate mutation | Change tool input after decision issuance | Block before tool body; no side effect |
| V2-03 | Tool identity mutation | Change tool/action while retaining prior decision | Block before tool body |
| V2-04 | Policy freshness | Issue under policy v1, change active policy to v2 | Old decision rejected at final boundary |
| V2-05 | Expired decision | Expire decision before dispatch | Block before tool body |
| V2-06 | Provider unavailable | Timeout, connection error, malformed response | Fail closed; no tool body entry |
| V2-07 | Subject/scope mismatch | Decision belongs to another actor or scope | Block before tool body |
| V2-08 | Duplicate/replay | Replay the same decision/invocation identifier | Duplicate side effect prevented; behavior recorded |
| V2-09 | NEEDS_REVIEW | Require review for candidate A | No execution until explicit approval; approval must bind to A |
| V2-10 | Approval substitution | Approve A, mutate candidate to B, then resume | B remains blocked; approval for A cannot authorize B |
| V2-11 | Unknown side effect | Commit mock side effect, drop response | Outcome UNKNOWN; automatic retry prohibited |
| V2-12 | Reconciliation | Resolve V2-11 through authoritative test ledger/action ID | Resolution linked to original invocation; retry policy explicit |

Every mandatory scenario must be run independently for each supported execution path and version. Where a scenario is not meaningful for a path, document the reason and do not silently omit it.

## 6. Harness design

Use deterministic mock tools only. The harness must expose at least:

- a tool-body entry counter;
- an append-only side-effect ledger with unique action IDs;
- a controllable response-loss fault after side-effect commit;
- deterministic policy provider with version changes;
- deterministic decision provider supporting PROCEED, NEEDS_REVIEW, and SILENCE;
- a controllable clock for expiry tests;
- invocation/decision identifiers and candidate digest;
- raw framework events sufficient to establish which dispatch path ran.

No production credentials, customer data, external write APIs, or destructive operations are permitted.

### Canonical candidate

Define and version canonicalization before execution. At minimum, the canonical candidate must cover:

- framework/runtime identity and version;
- adapter/dispatch path;
- tool name and action;
- normalized tool input;
- actor/agent identity;
- resource and scope;
- relevant execution context that can alter the effect.

Record the canonicalization algorithm/version and digest algorithm in every run. Avoid hashing ambiguous concatenated strings. Use a deterministic structured encoding and publish test vectors for equivalent and non-equivalent inputs.

## 7. Evidence bundle

Each run must produce machine-readable JSON plus raw logs. Minimum fields:

- schema version, test ID, run ID, UTC timestamp;
- NAEOS repository commit and CrewAI version;
- Python version and dependency lock/hash;
- execution path and adapter version/commit, if applicable;
- canonicalization version and candidate digest;
- decision ID, decision literal, policy ID/version/digest, subject, scope, expiry;
- invocation ID and action ID;
- enforcement point observed and hook/adapter event trace;
- tool-body entry count and side-effect ledger snapshot/checksum;
- outcome: SUCCEEDED, FAILED, or UNKNOWN;
- retry_allowed and reconciliation_required;
- reconciliation reference/result/time where applicable;
- test assertion results and known limitations.

Evidence integrity hashes are useful for detecting modification, but a hash alone does not prove that a real-world side effect did or did not occur. This harness only makes claims about its isolated mock environment.

## 8. Result semantics

- **PASS:** all assertions for that scenario are supported by inspectable artifacts.
- **FAIL:** observed behavior violates an invariant, including an ungoverned path reaching a tool body.
- **BLOCKED:** the scenario/path could not be exercised because of environment, dependency, API, or adapter availability constraints.
- **UNKNOWN:** artifacts cannot establish whether the assertion held.

Do not aggregate BLOCKED or UNKNOWN into PASS. The version-level result is PASS only if every mandatory supported path/scenario passes and all exclusions are explicitly justified. Any FAIL makes the version-level result FAIL.

## 9. Implementation sequence

1. **Completed:** inspect CrewAI 1.15.16 and 1.15.44 source/API surfaces and map ReAct, native, Flow, and adapter dispatch paths.
2. **Partially completed:** pinned-version GitHub Actions jobs exist for the tested CrewAI paths and optional adapter dependencies.
3. **Partially completed:** an independent append-only side-effect ledger and deterministic DENY probes exist; policy freshness, decision lifecycle, controllable clock, and response-loss injection remain outstanding.
4. **Partially completed:** ReAct, native, Flow, and adapter boundary probes have been added; full adapter orchestration and remaining scenario coverage are outstanding.
5. **Not completed:** V2-01 through V2-12 have not been executed as a complete scenario suite; current results are limited to DENY boundary probes.
6. Independently inspect the ledger and recompute candidate digests.
7. Publish a version-by-version report with PASS/FAIL/BLOCKED/UNKNOWN and exact limitations.

## 10. Relationship to existing evidence

- `scripts/external_validation_crewai_boundaries.py` and `docs/EXTERNAL-VALIDATION-REPORT-001.md` remain the immutable v1 evidence for mutation and policy freshness scenarios tested by that harness.
- `docs/experiments/NAEOS-Handoff-Interop-Test-001.md` remains a proposed cross-framework contract until its scenarios are actually executed.
- This v2 plan expands coverage; it does not retroactively upgrade the scope or result of v1.

## 11. Explicit non-claims

Even a successful v2 run does not establish production readiness, security of arbitrary CrewAI applications, safety of external services, enterprise compliance, or correctness of every tool implementation. It establishes only the observed behavior of the pinned versions, tested paths, controlled environment, and recorded scenarios.
