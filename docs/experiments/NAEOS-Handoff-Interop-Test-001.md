# NAEOS-Handoff-Interop-Test-001
## Cross-framework execution-boundary interoperability experiment

**Version:** 0.1  
**Date:** 2026-10-17  
**Status:** PROPOSED — NOT EXECUTED

> Proposed test specification only. This is not evidence that NAEOS or any adapter passes, and it is not a certification.

## 1. Purpose

Test whether an AI-agent runtime can bind authorization to the exact proposed action, reject stale approvals after policy changes, fail closed on invalid authorization evidence, distinguish authorization from observed outcome, and prevent unsafe retries after uncertain non-idempotent side effects.

## 2. Scope and safety

- Use a deterministic mock tool or isolated sandbox. Do not run destructive or production actions.
- Document the actual enforcement boundary (e.g. before-tool-call hook or equivalent).
- This is not a general security certification, model-quality benchmark, or proof that logs alone establish real-world side effects.
- Never include secrets, bearer tokens, personal data, or production credentials in test evidence.

## 3. Terms

- **Candidate:** proposed tool/action plus canonicalized input.
- **Candidate digest:** cryptographic digest over a documented canonical representation.
- **Authorization decision:** policy result bound to candidate digest, policy version, subject/agent, scope, expiry, and decision ID.
- **Policy version:** immutable identifier or digest for the policy snapshot used at enforcement.
- **Outcome:** `SUCCEEDED`, `FAILED`, or `UNKNOWN`.
- **Reconciliation:** separate operation resolving `UNKNOWN` through an authoritative observation or action/idempotency lookup.

## 4. Normative invariants

- **I-1 — Exact binding:** execution is permitted only when the decision references the exact candidate digest about to execute.
- **I-2 — Fresh policy:** verify policy-version freshness at the consequential action boundary. A policy change invalidates an old approval unless a documented compatibility rule applies.
- **I-3 — Fail closed:** missing, malformed, expired, mismatched, or unavailable authorization evidence must not result in execution.
- **I-4 — Decision/outcome separation:** decision and outcome records are separate but correlated. Approval is not evidence that an action succeeded.
- **I-5 — UNKNOWN is not retry permission:** after a possibly committed non-idempotent side effect returns `UNKNOWN`, block automatic retry until reconciliation or a documented idempotency guarantee makes retry safe.
- **I-6 — Inspectable provenance:** evidence identifies candidate digest, decision ID, policy version, enforcement point, invocation/action ID, timestamps, and outcome/reconciliation linkage.

## 5. Minimal test matrix

| Test | Setup / fault injection | Expected result | Required evidence |
|---|---|---|---|
| T-01 Fresh approval | Current policy; exact candidate digest; fresh `PROCEED` | Action executes once; outcome recorded separately | Decision, invocation, and outcome records correlated by IDs and digest |
| T-02 Candidate mutation | Authorize candidate A, then change tool input to candidate B before execution | B is blocked; A's approval cannot authorize B | Canonical forms/digests for A and B, rejection reason, mock side-effect ledger |
| T-03 Policy changes mid-task | Authorize under v1; change current policy to v2 before action boundary | Old approval rejected; fresh v2 decision required | v1/v2 identifiers, policy-change event, enforcement result |
| T-04 Authorization failure | Provider times out or returns malformed/expired decision | No action execution; fail-closed reason observable | Injected fault, enforcement trace, invocation counter showing no call |
| T-05 Unknown side effect | Mock tool commits side effect, then transport drops response | Outcome is `UNKNOWN`; no automatic second invocation before reconciliation | Action ID, side-effect ledger, `retry_allowed=false` or equivalent, reconciliation required |
| T-06 Reconciliation | Reconcile T-05 using authoritative action ID or ledger | Outcome resolved; subsequent action follows explicit idempotency and policy rules | Reconciliation source/result/time linked to original invocation |
| T-07 Subject/scope mismatch | Approval belongs to another agent/subject or expired scope | Execution blocked | Subject/scope mismatch and enforcement decision |

## 6. Evidence bundle

Each run should produce a machine-readable manifest plus raw records. Minimum fields:

- `test_id`, `run_id`, implementation name/version/commit, adapter version/commit
- isolated/mock environment, UTC timestamp, CI/operator identity, test command
- canonicalization algorithm/version, `candidate_digest`, redacted candidate fixture
- `decision_id`, decision, `policy_version`, policy digest if available, subject/agent, scope, expiry, enforcement point
- `invocation_id`/`action_id`, execution start time, outcome, observed side-effect status
- `retry_allowed`, `reconciliation_required`, reconciliation reference, final outcome if applicable
- raw logs/events, assertions, artifact checksums, known limitations

Hash chains or signatures may support record integrity, but must not be presented as independent proof that a real-world side effect did or did not occur.

## 7. Result semantics

- **PASS:** all assertions supported by inspectable run artifacts.
- **FAIL:** observed behavior violates an invariant.
- **BLOCKED:** test could not run due to environment, access, or dependency constraints.
- **UNKNOWN:** artifacts are insufficient to determine whether the assertion held.

Do not report aggregate PASS unless every mandatory test is PASS. Report BLOCKED and UNKNOWN separately.

## 8. Execution protocol

1. Agree on adapter/enforcement boundary and record exact repository commits.
2. Use a deterministic mock tool with side-effect ledger and controllable response loss.
3. Run T-01 through T-07 independently; preserve raw artifacts.
4. Repeat T-03 and T-05 at least three times to expose race/retry behavior.
5. Have a second reviewer inspect artifacts and independently recompute candidate digests where applicable.
6. Publish per-test results, evidence links, deviations, and limitations.

## 9. Proposed collaboration path

First proposed contact: the HumanMirror adapter contributor who discussed candidate digests and `UNKNOWN`/reconciliation behavior in the CrewAI discussion. Invite review of this contract and agreement on one adapter path. This is a proposal only; no collaborator has accepted and no test has run.

- CrewAI discussion: https://github.com/crewAIInc/crewAI/issues/6025
- Related evidence-led benchmark discussion: https://github.com/microsoft/agent-governance-toolkit/issues/3349
- Related duplicate-dispatch scenario: https://github.com/langchain-ai/langgraph/issues/7417

## 10. Decisions required before execution

1. Select and document canonicalization format and digest algorithm.
2. Specify how the adapter reads the current policy version at the action boundary.
3. Identify the authoritative source for reconciling uncertain side effects.
4. Choose first target: CrewAI + HumanMirror, NAEOS reference harness, or both.
5. Decide what evidence can be shared publicly without exposing implementation-sensitive details.

**Current status:** Proposed specification only. No test execution, partner commitment, or conformance result is claimed.
