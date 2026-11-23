# Client Delivery Acceptance Gates

Status: Draft — internal review required  
Version: 1.0.0-draft

Use this gate checklist for each engagement. Status values: NOT STARTED, IN PROGRESS, PASS, FAIL, BLOCKED, NOT APPLICABLE. A gate is PASS only when its evidence, owner, date, and review are recorded.

## Pre-engagement

| Gate | Acceptance evidence | Status |
|---|---|---|
| Sponsor and technical owner identified | Names/roles and contact path | NOT STARTED |
| Scope is bounded | Workflow, repository, environment, capabilities, exclusions documented | NOT STARTED |
| Written authorization | Client authorization for assets and test activities | NOT STARTED |
| Commercial terms | Approved proposal/SOW, fees, acceptance, schedule | NOT STARTED |
| Legal/privacy review | Applicable NDA, data terms, counsel review as required | NOT STARTED |
| Data handling | Categories, access, storage, retention, deletion, approval | NOT STARTED |
| Risk and stop conditions | Threat/safety concerns and escalation path | NOT STARTED |

## Technical readiness

| Gate | Acceptance evidence | Status |
|---|---|---|
| Isolated environment | Environment and production separation evidenced | NOT STARTED |
| Least privilege | Accounts/roles and authorized capabilities reviewed | NOT STARTED |
| Secrets hygiene | No secrets in repository, logs, report, or evidence | NOT STARTED |
| Baseline reproducibility | Commit/version, toolchain, commands, exit status, outputs | NOT STARTED |
| Scenario design | Positive/negative tests and expected outcomes agreed | NOT STARTED |
| Verification | Evidence integrity/traceability and independent review plan | NOT STARTED |
| Operations | Escalation, access revocation, exit/deletion plan | NOT STARTED |

## Delivery acceptance

| Gate | Acceptance evidence | Status |
|---|---|---|
| Evidence register complete | Evidence references, provenance and classification | NOT STARTED |
| Control map complete | Control status and evidence/gap references | NOT STARTED |
| Scenario results complete | Status, actual result, evidence and limitations per test | NOT STARTED |
| Findings reviewed | Risk score, fact vs hypothesis, recommendation, owner | NOT STARTED |
| Action plan usable | Priority, owner, target, verification criterion | NOT STARTED |
| Report/readout delivered | Agreed formats and scope limitations | NOT STARTED |
| Client acceptance | Named acceptance role and recorded acceptance | NOT STARTED |
| Closure | Access revoked, data disposition and feedback recorded | NOT STARTED |

## Hard HOLD conditions

- No written authorization for requested access or tests.
- Production credentials, production data, or production actions are requested without separately approved scope and controls.
- Data/privacy handling requirements are unresolved.
- Isolation or least privilege cannot be established.
- Material scenario outcome is unsupported or evidence integrity is uncertain.
- Contracting, acceptance, or responsibilities are unclear.
- Delivery claims exceed demonstrated capability.

A HOLD means no engagement execution for the affected activity until the condition is resolved and documented. It is not a PASS or a waiver.
