# NAEOS Client Delivery Playbook

Status: Draft for internal validation  
Version: 1.0.0-draft  
Owner: Service Delivery Lead

## 1. Operating principles

- Scope first; do not accept broad or ambiguous authorization.
- Keep model proposal, policy decision, runtime execution, observation, and verification distinct.
- Evidence before claims. Separate observed facts from client reports and assumptions.
- Fail closed where authorization, policy, scope, or evidence integrity cannot be established.
- Use minimized, client-approved data and isolated environments.
- Never claim certification, universal agent safety, production readiness, or guaranteed outcomes.

## 2. Engagement lifecycle

### Stage A — Discovery and qualification
Capture business objective, workflow, agent/tool stack, repositories, environments, identities, stakeholders, constraints, baseline, timeline, budget, and decision process. Identify risks that require a HOLD.

Exit: qualification summary, sponsor and technical owner, candidate scope, next decision.

### Stage B — Proposal and authorization
Create scope, deliverables, exclusions, schedule, fees, acceptance criteria, responsibilities, data terms, access authorization, and change-control approach. Obtain written authorization before repository access or tests.

Exit: approved scope/SOW and required legal/privacy/security gates.

### Stage C — Kickoff and baseline
Record engagement ID, scope version, contacts/roles, authorized assets, environment, data classification, commit/version/toolchain, baseline workflow, and success criteria. Confirm secure handling and communication channel.

Exit: baseline and evidence plan reviewed.

### Stage D — Evidence collection
Collect evidence through approved methods. Log evidence ID, source, date, collector, classification, scope, integrity/provenance when available, redactions, and related controls/findings. Avoid secrets and unnecessary personal data.

Exit: evidence coverage reviewed; gaps explicitly marked.

### Stage E — Control mapping and scenarios
Map controls across intent, identity, authorization, policy, runtime, observation, audit, verification, handoff, and operations. Design positive and negative scenarios, including denied actions, missing identity/approval, stale or revoked authorization, replay, tool failure, ambiguous outcome, and reconciliation where applicable.

Test outcomes:
- PASS: acceptance criteria met with evidence.
- FAIL: observed behavior violates criteria.
- BLOCKED: prerequisite or access prevents execution.
- NOT TESTED: test intentionally not run or deferred.
- INCONCLUSIVE: evidence cannot support a reliable outcome.

If an unauthorized side effect occurs, record FAIL and stop the affected scenario/workflow pending safety review.

### Stage F — Findings and remediation
Each finding must include unique ID, evidence references, observed fact, impact, likelihood, risk score/band, scope/limitations, recommendation, priority, owner, due date, and verification criteria. Keep hypotheses explicitly labeled.

### Stage G — Readout and acceptance
Deliver technical and executive summaries. Explain scope, methodology, evidence limitations, results, risks, recommended sequence, and decisions needed. Acceptance confirms delivery of agreed artifacts—not that all findings are remediated or that security/compliance is certified.

### Stage H — Closure
Confirm deliverables, outstanding actions, access removal/revocation, data return/deletion procedure (as contractually agreed), client feedback, lessons learned, and permission for any public reference. No case study or logo usage without explicit written approval.

## 3. Evidence and data handling

- Use synthetic/public/minimized data by default.
- Obtain client approval for any confidential data processing.
- Keep evidence storage access-limited; use approved encryption/storage and retention policy.
- Redact secrets, credentials, tokens, personal data, and unrelated source content.
- Record evidence provenance and distinguish verified/reported/inferred/unverified.
- Do not place client confidential evidence into public repository, demo, or sample package.
- If data-handling requirements are unresolved, pause collection.

## 4. Change control

Any expansion to repositories, environments, tools, capabilities, data types, production access, timeline, or acceptance criteria requires documented review and client approval. Reassess authorization, risk, effort, and applicable gates.

## 5. Internal review and independence

A reviewer who did not produce a finding should verify material evidence and scenario outcome when feasible. If staffing prevents independent review, disclose the limitation in the report. No test should be marked PASS solely because an expected result is described in a plan.

## 6. Minimum handoff

At closure, hand over:
- scope and authorization record;
- evidence register and evidence package (subject to approved data handling);
- control map;
- scenario results;
- findings and action plan;
- executive readout;
- limitations and residual risk;
- acceptance and closure record.

## 7. Stop conditions

Pause and escalate if scope/authorization is unclear; credentials or sensitive data are exposed; test execution risks production or unrelated systems; an unexpected side effect occurs; evidence integrity is uncertain; client requests unapproved scope; or required legal/privacy/security gates are incomplete.
