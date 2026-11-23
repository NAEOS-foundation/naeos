# NAEOS Master Development Plan v1.0
**Audit date:** 2026-11-21  
**Scope:** NAEOS Core, CRM, Intelligence, Control Plane, staging, security and pilot readiness.  
**Status:** Phase 0 inventory established; live CI run conclusions are intentionally limited to verified metadata and PR descriptions.

## Executive decision
Follow the gated sequence **Stabilize → Secure → Integrate → Verify → Pilot → Early Access**. Avoid new feature expansion until critical identity, execution, audit, staging and E2E gates are closed. No production deployment is authorized by this plan.

## Phase 0 goals
1. Inventory open PRs across Core, CRM and Intelligence.
2. Categorize by critical-path dependencies, blockers, mergeability and scope.
3. Establish a canonical milestone/task sequence and release gates.
4. Identify CI or validation evidence gaps.
5. Preserve security gates; do not bypass required checks.

## Repository snapshot
### NAEOS Core — `NAEOS-foundation/naeos`
Open PRs in priority focus:
- #830 — production API identity contract (draft; mergeable on refresh; CI/Compliance/PostgreSQL integration fail)
- #831 — preserve JWT subject/client identity separately (draft; mergeable; CI/Compliance/PostgreSQL integration fail)
- #832 — persisted principal-agent binding resolver (draft; not mergeable on refresh; External Validation/Compliance/PostgreSQL integration/CI fail)
- #841 — align P2 contract with runtime (draft; mergeability currently false; checks not refreshed)
- #842 — Docker local staging runtime (draft; mergeable; current Compose Validation and CI pass, but DCO and Sign-off checks fail)
- #851 — brand governance stabilization (draft; mergeable)
- #852 — asset validation tooling (draft; mergeable)
- #853 — Social Media asset reference audit (draft; mergeable)
Other open PRs include #823, #820, #818, #817, #815, #794, #696, #692, #676, #675, #670 and older work. These require a dedicated stale/dependency review before disposition.

### NAEOS CRM — `NAEOS-foundation/naeos-crm`
- #69 — database-backed integration test isolation (draft; mergeable)
- #68 — atomic audit prototype for Policy Rule updates (draft; mergeable)
- #67 — audit durability findings/remediation plan (draft; mergeable)
- #66 — company deletion safeguard (draft; mergeable)
- #65 — company access and relationship validation (draft; mergeable)
- #64 — Go-Gate requester self-rejection safeguard (draft; mergeable)
- #63 — CRM read-only Control Plane dashboard (draft; mergeable)
- #62 — audit hash metadata v2 (draft; mergeable)
- #60 — legacy audit chain backfill regression reproduction (draft; mergeable)
- #37 — Prisma/outbox lifecycle stabilization (open; older; requires status/dependency review)

### NAEOS Intelligence — `NAEOS-foundation/naeos-intelligence`
- #78 — Supabase ES256 JWT verification (draft; mergeable)
- #77 — isolated partner deployment baseline (non-draft; mergeable)
- #74 — Intelligence DB runtime role matrix (draft; mergeable)
- #73 — unified DB migration plan (draft; mergeable; migration/cutover remains unapproved)
- #72 — separate JWT subject from client ownership (draft; mergeable)
- #71 — read-only Core repository context integration (non-draft; mergeable on refresh; CI fails)
- #69 — Core/GitHub context resolver integration (open; older verification limits noted in PR)
- #67 — read-only API and M1 governance alignment (open; review its exact current state before closing)

## Phase 0 live CI evidence (refreshed 2026-11-21)
| Repo / PR | Head (short) | Current observed checks | Phase 0 disposition |
|---|---|---|---|
| Core #830 | `355db86f7` | DCO, docs, deployment gate, gofmt, sign-off, policy bypass, secrets scan, browser E2E, CodeQL pass; NAEOS Compliance, PostgreSQL integration and CI fail | Revise; debug required failing gates |
| Core #831 | `2c3678901` | Similar passing security checks; NAEOS Compliance, PostgreSQL integration and CI fail | Revise; keep identity patch narrow, resolve gates |
| Core #832 | `f6c662358` | Browser E2E, docs, migration replay, CodeQL, DCO, sign-off pass; External Validation, PostgreSQL integration, Compliance and CI fail | Hold/revise; blocked by identity contract/test failures; not mergeable |
| Core #842 | `425f63253` | Compose Validation, CI, Compliance, CodeQL, Secrets scan and deployment gate pass; DCO and Sign-off fail | Revise sign-off; Compose config now verified by current run, despite older PR description |
| CRM #69 | `4eb9af8d6` | CI pass | Merge candidate after review/dependency check; draft status and review still required |
| CRM #68 | `9faa47f9a` | CI pass | Keep draft prototype; review transaction semantics and broaden remediation |
| CRM #67 | `cbbc7a9f9` | CI pass | Review/merge documentation findings plan; use as remediation anchor |
| CRM #63 | `591ea0b7f` | CI pass | Hold integration merge until scoped Core API dependency and review confirmed |
| Intelligence #78 | `d74932b1d` | CI pass | Merge candidate after IAM review and base/dependency checks |
| Intelligence #77 | `498a5ae41` | CI pass | Merge candidate; isolated deployment baseline only, not IAM closure |
| Intelligence #72 | `93d38d17e` | CI pass | Merge candidate after review/base alignment |
| Intelligence #71 | `05cb9bd6e` | CI fail | Revise; integration remains read-only and blocked on Core contract alignment |

**Limitations:** only PR-triggered workflow runs returned by the connector are shown; this may not include every required check or review state. Refresh exact head/checks/reviews immediately before merge. Merge candidates are not authorization to merge.

## Phase 0 live CI evidence (refreshed 2026-11-21)
| Repo / PR | Head (short) | Current observed checks | Phase 0 disposition |
|---|---|---|---|
| Core #830 | `355db86f7` | DCO, docs, deployment gate, gofmt, sign-off, policy bypass, secrets scan, browser E2E and CodeQL pass; Compliance, PostgreSQL integration and CI fail | Revise; resolve gates |
| Core #831 | `2c3678901` | Supporting checks pass; Compliance, PostgreSQL integration and CI fail | Revise; preserve identity semantics and resolve failures |
| Core #832 | `f6c662358` | Browser E2E/docs/migration replay/CodeQL/DCO/sign-off pass; External Validation, PostgreSQL integration, Compliance and CI fail; not mergeable | Hold/revise; binding depends on identity contract |
| Core #842 | `425f63253` | Compose Validation, CI, Compliance, CodeQL, Secrets scan and deployment gate pass; DCO and Sign-off fail | Revise sign-off; current Compose check now passes despite older PR description |
| Core #841 | `880a65f42` | Compliance, CodeQL, docs and other checks pass; DCO and Sign-off fail | Revise sign-off; align contract after runtime proof |
| Core #851 | `beefc54c9` | CI, Compliance, CodeQL and security checks pass; Sign-off fails | Revise sign-off |
| Core #852 | `7dd19a71b` | Asset Validation/security/docs gates pass; CI cancelled | Rerun CI |
| Core #853 | `6ec28d4bb` | Render QA and asset validation pass; DCO and Sign-off fail | Revise sign-off |
| Core #817 | `59462aa43` | CI, PostgreSQL integration, CodeQL, Browser E2E and Compliance pass; DCO/Sign-off fail; not mergeable | Hold/revise sign-off and branch alignment |
| Core #815 | `038ce0555` | Major checks pass; not mergeable | Hold; resolve base/mergeability and dependency review |
| Core #794 | `8bb7ea89a` | Major checks pass; not mergeable | Hold; check supersession/base conflict |
| Core #676 | `c5425b8dc` | CI/docs/DCO/CodeQL pass; Secrets scan and Browser E2E fail; not mergeable | Revise; security-related preview isolation |
| CRM #69 | `4eb9af8d6` | CI pass | Review candidate; verify base/dependencies/review |
| CRM #68 | `9faa47f9a` | CI pass | Draft prototype; review transaction/audit semantics |
| CRM #67 | `cbbc7a9f9` | CI pass | Review/merge remediation plan after review |
| CRM #63 | `591ea0b7f` | CI pass | Hold for scoped Core API dependency and review |
| CRM #64 | `513a44ca2` | CI pass | Review candidate after dependency/base checks |
| CRM #65 | `2e84d0e9c` | CI pass | Review candidate after dependency/base checks |
| CRM #66 | `7df8f81c5` | CI pass | Review candidate after dependency/base checks |
| CRM #62 | `51096b748` | CI pass | Hold/review audit hash version dependency |
| Intelligence #78 | `d74932b1d` | CI pass | Review candidate after IAM review |
| Intelligence #77 | `498a5ae41` | CI pass | Review candidate; baseline only, not IAM closure |
| Intelligence #72 | `93d38d17e` | CI pass | Review candidate after ownership/IAM review |
| Intelligence #74 | `16fbe947d` | CI pass | Draft; review DB role matrix |
| Intelligence #73 | `9c71f908b` | CI fail; not mergeable | Hold; migration remains unapproved, resolve CI |
| Intelligence #71 | `05cb9bd6e` | CI fail; mergeable metadata refreshed | Revise; keep integration read-only |
| Intelligence #69 | `e3627f2a9` | CI fail; not mergeable | Revise/hold; reconcile build/dependencies |
| Intelligence #67 | `3ab461430` | CI fail; not mergeable | Revise; maintain read-only API contract |

**Evidence limitations:** connector returns PR-triggered runs only and may omit other required checks/reviews. Refresh head SHA, checks and review state immediately before merge. Merge candidates are not authorization to merge.

## Phase 0 review-state audit
- Core #830/#831/#832 have maintainer technical self-review comments only; these are **not independent approvals**.
- Core #842 has no review submissions recorded.
- CRM #69/#68/#67/#63 have no review submissions recorded in the connector snapshot.
- Intelligence #78 has a maintainer assessment comment; #72/#71/#77 have no review submissions recorded in the connector snapshot.
- Roadmap PR #856 is DCO-signed; its CI, DCO and required checks were pending/queued at last refresh. PR #855 was closed as duplicate/superseded due to missing DCO sign-off.
- All review/check data is a snapshot. Refresh immediately before merge.


## Dependency-aware critical path
### P0 — security and operational safety
1. **CORE-IDENTITY:** #831 preserve `iss/sub/client_id` distinctions; then align/ratify #830 contract.
2. **CORE-BINDING:** #832 exact issuer/principal-kind/subject → active agent binding; fail closed on missing, ambiguous, inactive or mismatch.
3. **CORE-SCOPES:** verify route-specific scopes and authorization tests; ensure ledger/run/evidence queries are agent/principal scoped.
4. **CORE-STAGING:** #842 local staging Compose validation; current fetched run shows Compose Validation and CI passing, but DCO and Sign-off checks failing. Resolve sign-off first; then build/preflight/readiness.
5. **CRM-TEST-SAFETY:** #69 local-only database integration guard and safe cleanup.
6. **CRM-AUDIT:** #67 findings plan + #68 prototype; expand atomic mutation/audit remediation to all consequential mutation paths. Fail closed on audit persistence failures.
7. **INT-IAM:** #78 ES256 verifier + #72 subject/ownership; reconcile scopes, issuer/audience, resource ownership, deployment/partner binding. #77 accepted isolated-partner model is documentation evidence, not complete IAM closure.
8. **INT-READONLY:** #71 Core context integration, after API identity/scope contract is stable; no external side effects in pilot Intelligence.

### P1 — integration and proof
9. Integrate CRM operations dashboard (#63) against scoped read-only Control Plane API.
10. Complete staging migration rehearsal only after explicit review/approval; unified DB migration (#73) is not authorized by this plan.
11. E2E critical scenario: identity → context → plan → policy/authorization → approval boundary → guarded action (only when enabled) → observation → reconciliation → evidence → independent verification.
12. Failure-injection and recovery: unauthorized, wrong scope/binding, revoked principal, duplicate/replay, timeout before/after side effect, executor/observation/evidence failure, reconciliation, concurrent requests.
13. Backup/restore and rollback rehearsal; deployment/monitoring runbook.

### P2 — pilot and Early Access
14. Build a bounded read-only engineering governance pilot (repository analysis, risk, plan, PR review, evidence).
15. Pilot report and KPI baseline. Suggested acceptance: zero unauthorized execution; zero duplicate side effects; critical evidence durable; 100% critical/high finding verification; 100% critical E2E scenario pass.
16. Early Access release package only after gates pass: security review, staging evidence, operational runbook, onboarding, support/escalation, release artifacts, known limitations and rollback.

## PR disposition rules (Phase 0)
- **Merge candidate:** all required CI/checks pass, review and DCO complete, no unresolved blocking comments, dependency/base is merged, change scope is understood.
- **Revise:** security/CI/validation gaps, incomplete tests, stale base, conflict, or unmet acceptance criteria.
- **Hold:** depends on critical contract/binding/runtime work not yet merged.
- **Close:** superseded, duplicate, no longer relevant, or unsafe; only after reviewing issue/PR history and confirming replacement.
- Never bypass required security, dependency-risk, DCO or CI gates.

## CI and verification evidence policy
- PR descriptions and mergeability metadata are not a substitute for current check-run results.
- Record exact head SHA and latest check conclusion before merge decisions.
- Current fetched runs: #830/#831 CI, NAEOS Compliance Gate and Control Plane PostgreSQL Integration failed; #832 External Validation, PostgreSQL Integration, Compliance and CI failed. #842 Compose Validation and CI pass, but DCO and Sign-off fail (PR description is stale regarding Compose validation runner failure). CRM #69/#68/#67/#63 CI pass. Intelligence #78/#77/#72 CI pass; #71 CI fails.
- Where tool/API access does not provide check-run detail, mark **UNVERIFIED** and request/trigger a proper rerun rather than infer success.
- PR head SHA and mergeability should be refreshed immediately before any merge.

## Phase 0 action checklist
- [x] Inventory open PRs in Core, CRM, Intelligence.
- [x] Record initial mergeability/draft metadata for critical PR subset.
- [x] Establish dependency-aware critical path and safety rules.
- [x] Refresh current checks/head SHA for critical PRs in batches (sampled; see CI evidence below).
- [x] Review submitted review-state snapshot for the critical PR subset (independent approvals absent in sampled PRs; see notes).\n- [ ] Review all PR comments/reviews and identify stale or conflicting PRs (full repository sweep remains).
- [ ] Confirm base branches and stacked dependencies (full sweep remains).
- [x] Create master roadmap/backlog PR as tracker for P0/P1/P2 items.
- [x] Publish initial Phase 0 disposition recommendations (merge/revise/hold) with check evidence; stale/close dispositions remain pending history review.
- [ ] Reconcile plan after PR merges and update status.

## Gate definitions
- **G0 Backlog controlled:** inventory and dependencies reconciled; no unsafe bypass.
- **G1 Identity secure:** identity contract, exact binding, scopes and ownership tests pass.
- **G2 Execution governed:** durable action ledger, idempotency/replay protection, UNKNOWN outcome and authoritative reconciliation.
- **G3 CRM durable:** consequential mutation + audit durability/atomicity and safe integration tests proven.
- **G4 Staging proven:** deploy/preflight/readiness, migrations, secrets, backup/restore, rollback and observability tested in isolated staging.
- **G5 E2E verified:** critical scenarios and failure injection pass with evidence.
- **G6 Pilot accepted:** bounded scope, KPI and report; no critical/high unresolved blocker for pilot scope.
- **G7 Early Access:** repeatable isolated deployment, onboarding, support, release and rollback package.

## Safety constraints
- This roadmap does not authorize production deployment, production migration, credential changes, public client access, or execution-capability activation.
- Do not place secrets in repository, issue comments, PR bodies or logs.
- Use staging-only credentials and target verification.
- External side effects remain disabled until authorization/execution gateway readiness and explicit scope approval.

## Current next action
Finish Phase 0 check-run/review/dependency refresh for critical PRs, then begin Phase 1 by resolving identity contract and binding dependencies in order.

## Phase 0 reconciliation refresh — 2026-11-21
**This section supersedes stale or duplicated PR snapshot statements above.** Open inventory refreshed across three repositories: Core 23 PRs, CRM 10 PRs, Intelligence 8 PRs (41 total). Exact check-run evidence is available only for PRs explicitly listed below; all other checks remain UNVERIFIED until fetched. Mergeability metadata is not merge authorization.

### Core disposition snapshot
| PR | Current observed checks | Mergeability | Disposition |
|---|---|---|---|
| #857 roadmap | All 11 observed checks PASS | Mergeable, draft | Keep open; independent review needed |
| #818 release action | All 10 observed checks PASS | Mergeable | Candidate only after independent review and dependency check |
| #820 CodeQL action | CodeQL FAIL; other observed checks PASS | Mergeable | Revise; fix CodeQL failure |
| #823 Wrangler | CI FAIL; other observed checks PASS | Mergeable | Revise CI |
| #670 trusted dependency registry | CI/DCO/Sign-off FAIL; PDR/docs/external validation/security/compliance PASS | Not mergeable | Revise CI and sign-off |
| #696 presentation brief | DCO/Sign-off FAIL; CI/docs/compliance/security PASS | Mergeable | Revise DCO/sign-off; review before merge |
| #692 JCS regression | CI FAIL; P0 evidence/E2E/DCO/sign-off/security PASS | Not mergeable | Revise CI; narrowly scoped regression fix |
| #686 DCO repair | CI CANCELLED; DCO/sign-off/PDR/docs/security PASS | Not mergeable | Rerun CI; reconcile base |
| #830 identity contract | CI/Compliance/PostgreSQL integration FAIL; supporting checks PASS | Mergeable, draft | Revise; independent contract review required |
| #831 subject/client separation | CI/Compliance/PostgreSQL integration FAIL; supporting checks PASS | Mergeable, draft | Revise; preserve identity semantics |
| #832 persisted principal binding | CI/Compliance/PostgreSQL integration/External Validation FAIL | Not mergeable, draft | Hold; depends on identity contract |
| #841 P2/runtime alignment | DCO/Sign-off FAIL; supporting docs/security/compliance PASS | Mergeable, draft | Revise sign-off; align only after runtime proof |
| #842 staging Compose | CI/Compose Validation/Compliance/security PASS; DCO/Sign-off FAIL | Mergeable, draft | Revise sign-off, then isolated staging preflight only |
| #851 brand governance | Sign-off FAIL; CI/compliance/security PASS | Mergeable, draft | Revise sign-off |
| #852 asset validation | Asset validation/docs/security PASS; CI CANCELLED | Mergeable, draft | Rerun CI and review results |
| #853 social asset audit | Render QA/asset validation PASS; DCO/Sign-off FAIL | Mergeable, draft | Revise sign-off |
| #817 Supabase subject/client | CI/Postgres/CodeQL/E2E/Compliance PASS; DCO/Sign-off FAIL | Mergeable, draft | Revise sign-off; identity dependency remains |
| #815 production API auth | Major sampled checks PASS | Not mergeable | Hold; branch/dependency alignment required |
| #794 authenticated Compose | Major sampled checks PASS | Mergeable | Hold; reconcile with #842 and identity dependencies |
| #676 preview isolation | CI/docs/DCO/CodeQL PASS; Secrets scan and Browser E2E FAIL | Not mergeable, draft | Revise; mandatory security/E2E blockers |
| #675 ambiguous outcome contract | Observed checks PASS | Not mergeable | Hold; base/supersession reconciliation |
| #648 handoff harness | Checks not refreshed | Not mergeable, draft | UNVERIFIED; assess relevance/dependencies |
| #652 candidate digest binding | Checks not refreshed | Not mergeable, draft | HOLD for execution-safety review and dependencies |

### CRM disposition snapshot
| PR | CI | Mergeability | Disposition |
|---|---|---|---|
| #69 DB integration test isolation | PASS | Mergeable, draft | Review candidate; independent review and local DB guard required |
| #68 atomic audit prototype | PASS | Mergeable, draft | Keep draft; review atomicity/rollback/fail-closed audit |
| #67 audit durability findings | PASS | Mergeable, draft | Review remediation anchor; consequential mutation coverage |
| #66 company deletion safeguard | PASS | Mergeable, draft | Review dependent-record protection |
| #65 company access validation | PASS | Mergeable, draft | Review authorization/relationship tests |
| #64 Go-Gate self-rejection | PASS | Mergeable, draft | Review separation of duties |
| #63 read-only dashboard | PASS | Mergeable, draft | HOLD until Core identity/scopes stable; scoped read-only validation |
| #62 audit hash v2 | PASS | Mergeable | Review chain compatibility/version/dependency |
| #60 legacy backfill defect reproduction | PASS | Mergeable | Maintainer self-review is COMMENTED, not independent approval |
| #37 Prisma/Outbox lifecycle | PASS | Mergeable | Older PR; verify supersession/base and independent review |

### Intelligence disposition snapshot
| PR | CI | Mergeability | Disposition |
|---|---|---|---|
| #78 ES256 verifier | PASS | Mergeable, draft | IAM closure pending; recorded assessment COMMENTED, not independent approval |
| #77 partner isolation baseline | PASS | Mergeable | Review; docs baseline does not close IAM |
| #74 DB runtime role matrix | PASS | Mergeable, draft | Review least privilege and runtime roles |
| #72 subject/ownership separation | PASS | Mergeable, draft | Review ownership binding and auth tests |
| #71 read-only Core context | FAIL | Mergeable | Revise CI; remain read-only and contract-dependent |
| #73 unified DB migration plan | FAIL | Mergeable, draft | Revise CI; documentation only, no migration/cutover authorization |
| #69 Core/GitHub resolver | FAIL | Mergeable | Revise CI; read-only and contract check |
| #67 read-only API/M1 alignment | FAIL | Mergeable | Revise CI; preserve read-only M1 scope |

### Dependency / merge decision
- **No merges performed.** Independent approvals and dependency closure are not established for the candidate set.
- Identity order: #831 semantics → #830 contract → #832 binding → scopes/ownership tests → reconcile #817/#815 auth integration.
- Staging: #842 sign-off closure → staging-only preflight/readiness; reconcile #794 with #842 to avoid duplicate Compose/auth changes.
- CRM audit: #60 defect fixture and #62 hash compatibility review → #67 findings/remediation → #68 atomic prototype → coverage across consequential mutations; order depends on review and base alignment.
- CRM integration: #69 safe DB test guard before dashboard integration; #63 stays held until Core identity/scope API stable.
- Intelligence IAM: review #72 and #78 together; #74/#77 are supporting evidence, not closure. #71/#69 read-only integrations blocked on Core contract and CI fixes. #73 migration remains unapproved.
- #675/#686/#815/#692/#676/#648/#652/#794 require branch reconciliation, reruns or supersession review before merge/close decisions.

### Phase 0 checklist refresh
- [x] Refresh open PR inventory: 41 total (Core 23, CRM 10, Intelligence 8).
- [x] Refresh mergeability metadata for all 41 PRs.
- [x] Refresh observed CI conclusions for enumerated PRs in this section.
- [x] Post remediation/review comments to sampled blockers and candidates.
- [ ] Inspect complete comment/review history and changed-file scope for all 41 PRs (not completed).
- [ ] Reconcile branch/base and stacked dependencies using explicit refs/commit ancestry (not completed; branch refs not consistently exposed in connector snapshots).
- [ ] Obtain independent technical approvals for critical/security-sensitive candidates.
- [ ] Rerun failed/cancelled workflows and refresh checks for remaining PRs.
- [ ] Close stale/superseded PRs only after history and replacement evidence confirm disposition.
- [ ] Phase 0 sign-off and final roadmap update.

**Phase 0 remains PARTIAL.** Inventory and initial disposition improved; full history/changed-file/dependency reconciliation, reruns and independent approvals remain outstanding. Do not begin broad feature development or merge critical/security PRs before closure.


### Focused changed-file review findings (2026-11-21)
- **Core #831:** patch inspection found redundant/nested role-name validation predicates in `internal/controlplane/database_target.go`; simplify and add boundary tests. CI failure is consistent with the need for correction; this is a review finding, not a confirmed root cause.
- **Core #842:** local Compose configuration does not currently pass the required `NAEOS_CONTROLPLANE_PRINCIPAL_KIND` introduced by #832. Cross-PR staging contract gap; add explicit staging-only configuration and validation before staging readiness.
- **Intelligence #72:** issuer/subject owner key uses delimiter concatenation `user:<issuer>:<subject>`; because both identifiers are opaque, encoding is not injective. Recommend canonical unambiguous tuple encoding and collision tests.
- Focused patch inspections also confirmed #832 migration is explicitly staging-only proposal, runtime binding resolution fails closed on missing/ambiguous bindings, #68 uses an optional transactional audit port with compatibility fallback (fallback needs review against production port coverage), #62 migration preserves historical hash values and maps hash lengths to metadata, and #78 restricts JWT verification to ES256 and a dedicated NAEOS scopes claim. These are design observations; they are not independent approvals.
- Review findings were posted on #831, #842 and Intelligence #72. No merge was performed.
