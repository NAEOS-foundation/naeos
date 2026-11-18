# NAEOS Pilot Readiness Pack #001

Status: Proposed — internal working document  
Version: 1.0  
Owner: NAEOS pilot lead  
Applies to: Service-led AI Engineering Governance & Verification pilot using NAEOS Core  
Decision boundary: This pack is not a production-readiness certification, contract, privacy notice, or legal opinion.

## 1. Purpose and decision

This pack defines the minimum operational controls and evidence needed to initiate a bounded, service-led pilot with one design partner. It supplements, and does not replace:

- [Commercial Launch Evidence Checklist](../legal/COMMERCIAL-LAUNCH-CHECKLIST.md)
- [Reference Demo & Evidence Story](REFERENCE-DEMO.md)
- [Golden Path](GOLDEN-PATH.md)
- [Indonesia Design Partner Discovery Pack #001](GTM-INDONESIA-DANA-DISCOVERY-001.md)

Initial decision recommendation: **GO for discovery and a controlled, non-sensitive technical evaluation; HOLD for processing client confidential source code, production credentials, personal data, or production-adjacent actions until the applicable gates in this document have evidence and approval.**

A pilot is service-led. SaaS multi-tenancy, hosted production service, autonomous production deployment, and broad agent execution are out of scope unless separately reviewed and accepted.

## 2. Proposed pilot profile

| Field | Default proposal | Must be confirmed with client |
|---|---|---|
| Offer | AI Engineering Governance & Verification Assessment/Pilot | Yes |
| Duration | 2–4 weeks | Yes |
| Client team | One engineering team | Yes |
| Repository | One explicitly authorized repository | Yes |
| Workflow | One bounded AI-assisted engineering workflow | Yes |
| Capabilities | 1–3 declared capabilities | Yes |
| Deployment | Client-controlled or isolated evaluation environment | Yes |
| Data | Synthetic, public, or client-approved minimized data | Yes |
| Production access | Excluded by default | Explicit separate approval required |
| Deliverables | Assessment report, policy/control map, evidence pack, findings/remediation plan, acceptance report | Yes |
| Commercial terms | Separate SOW and counsel-reviewed terms | Yes |

## 3. Scope boundaries

### In scope (candidate)

- Discovery of one real AI-assisted engineering workflow.
- Mapping intent, identity, repository, capability, policy, execution boundary, observation, evidence, and verification.
- Reproducible NAEOS Core Golden Path / Reference Demo evaluation.
- Configuration of a narrowly scoped evaluation policy where supported.
- Evaluation of authorization allow/deny behavior and evidence traceability.
- Findings and recommendations.

### Out of scope unless separately accepted

- Production deployment or direct production changes.
- Access to production secrets or unrelated repositories.
- Unrestricted agent/tool execution.
- Claiming that NAEOS secures every AI provider, agent, integration, or workflow.
- Enterprise compliance certification, penetration-test certification, SLA, or availability guarantee.
- Multi-tenant SaaS deployment.
- Processing client data before data-flow, retention, access, deletion, and contractual requirements are approved.

## 4. Readiness gates (P0)

A gate is satisfied only when its evidence is recorded, reviewed, and has an owner/date. A checkbox or narrative assertion alone is not evidence.

| Gate | Required evidence | Owner | Status |
|---|---|---|---|
| G0 — Client & scope authorization | Named sponsor/technical owner; written repository and workflow authorization; scope and exclusions | Pilot lead | NOT STARTED |
| G1 — Legal/commercial | Contracting entity confirmed; NDA/SOW and applicable service/data terms reviewed by qualified counsel; fees and acceptance defined | Founder/legal | NOT STARTED |
| G2 — Data & privacy | Data-flow map for actual deployment; data categories, destinations/providers, access, region, retention, deletion; client approval | Security/privacy owner | NOT STARTED |
| G3 — Environment isolation | Environment boundary documented; least-privilege identities; secrets excluded from repo and evidence; production separated | Engineering owner | NOT STARTED |
| G4 — Security controls | Threat model; authentication/authorization checks; policy enforcement tests; dependency/vulnerability review appropriate to scope | Security owner | NOT STARTED |
| G5 — Core reproducibility | Clean-checkout Golden Path/Reference Demo run with commit SHA, toolchain, command, exit status and outputs recorded | Pilot engineer | NOT STARTED |
| G6 — Evidence integrity/verification | Evidence bundle/artifacts inspected; verifier behavior tested where applicable; reviewer independent of producing claim | Verification owner | NOT STARTED |
| G7 — Operations | Service owner, incident/escalation contact, backup/restore if stateful, access revocation and client exit/deletion procedure tested | Operations owner | NOT STARTED |
| G8 — Client acceptance | Baseline, success criteria, report format, review cadence, sign-off role and exit process agreed | Client sponsor + pilot lead | NOT STARTED |

**Gate policy:** No sensitive client data or consequential external action is allowed while a relevant P0 gate is open. If scope changes, re-evaluate the gates before proceeding. Fail closed when authorization, policy, schema, or version compatibility is uncertain.

## 5. Data-handling minimum

Before any client repository or data is accessed, document:

- Data owner and client authorization.
- Repository URLs/identifiers and permitted paths.
- Data types: source code, specifications, issue text, logs, generated artifacts, evidence, account data, secrets, personal data.
- Data processing location and each provider/service destination.
- Whether inputs leave the client's environment and which model/provider receives them.
- Purpose and necessity for each category.
- Access roles, credential mechanism, audit records.
- Retention period and deletion trigger, including backup behavior.
- Incident escalation and client notification responsibilities.
- Export/return/deletion procedure and evidence of completion.

Never place client confidential data, personal data, secrets, private contracts, or raw customer evidence in this public repository. Store sensitive evidence in an approved, access-controlled private location. In public or broadly shared records, use a safe reference, digest, status, and owner only.

## 6. Pilot workflow and decision boundary

1. **Discovery:** choose a real workflow and confirm the client problem without assuming a control gap.
2. **Baseline:** record current controls and relevant metrics; distinguish unknowns from observed values.
3. **Design:** declare task, repository, actor/agent identity, capabilities, policy/version, environment, prohibited actions, and verification path.
4. **Authorize:** client-designated authority approves the scope and allowed actions. Model output is not authorization.
5. **Execute:** only the approved, scoped evaluation workflow may run. Revalidate authorization at the last controllable boundary before consequential side effects.
6. **Observe:** capture actual execution result separately from the agent's statement.
7. **Evidence:** capture run ID, commit/artifact digests, specification/NEIR hashes where applicable, policy decision/version, timestamps, result, and evidence references.
8. **Verify:** a reviewer who is not the sole producer checks artifacts against acceptance criteria.
9. **Report:** classify findings, limitations, deviations, and remediation recommendations.
10. **Close:** client acceptance, access revocation, data return/deletion, and evidence-retention decision.

## 7. Acceptance criteria (proposed; client must approve)

| ID | Criterion | Verification method | Result |
|---|---|---|---|
| AC-01 | Authorized, in-scope action succeeds | Reproducible test and execution record | NOT RUN |
| AC-02 | Unauthorized/out-of-scope action is denied or escalated | Negative test; inspect policy result and runtime behavior | NOT RUN |
| AC-03 | Capability boundary is explicit and task-bound | Inspect declared policy/scope | NOT RUN |
| AC-04 | Decision is attributable to policy/version and actor/task | Inspect evidence record | NOT RUN |
| AC-05 | Observation is distinct from agent claim | Compare runtime result with model statement | NOT RUN |
| AC-06 | Required evidence is durable and traceable | Inspect artifacts, identifiers, hashes, references | NOT RUN |
| AC-07 | Verification is independently identifiable | Reviewer record and reproducible checks | NOT RUN |
| AC-08 | Policy/configuration changes invalidate stale authorization where required | Negative/staleness test, if applicable to integration | NOT RUN |
| AC-09 | Existing developer workflow remains usable within pilot scope | Client engineer walkthrough and feedback | NOT RUN |
| AC-10 | Another engineer can reproduce or review the result | Independent reviewer checklist | NOT RUN |
| AC-11 | Scope, deviations, limitations, and residual risks are documented | Final report review | NOT RUN |
| AC-12 | Client exit and access revocation are completed | Checklist and evidence record | NOT RUN |

Do not report a criterion as passed without observed evidence. Criteria not applicable to the chosen workflow require a written rationale and client approval.

## 8. Measurement plan

Capture baseline and pilot values only where measurement is reliable. Do not invent baselines or imply causal improvement from a small pilot.

| Metric | Baseline | Pilot result | Evidence/source |
|---|---|---|---|
| Unauthorized action test outcomes | TBD | TBD | Test record |
| Policy enforcement coverage for agreed scenarios | TBD | TBD | Scenario matrix |
| Evidence completeness | TBD | TBD | Evidence checklist |
| Independent verification success | TBD | TBD | Reviewer record |
| Manual approval count / latency | TBD | TBD | Workflow records |
| PR cycle time (if available) | TBD | TBD | Client-approved source |
| Developer friction / feedback | TBD | TBD | Structured feedback |
| Findings by severity | TBD | TBD | Assessment report |

## 9. Final report outline

1. Executive summary and pilot decision.
2. Client-approved scope, environment, and exclusions.
3. Workflow and architecture/control-boundary diagram.
4. Baseline and methodology.
5. Test scenarios and observed outcomes.
6. Evidence index (safe references/digests; no sensitive payloads).
7. Findings: severity, impact, evidence, recommendation, owner.
8. Acceptance criteria status (PASS / FAIL / NOT RUN / NOT APPLICABLE with rationale).
9. Limitations, assumptions, and residual risks.
10. Client feedback and next-step recommendation.
11. Data return/deletion and access-revocation confirmation.

## 10. Go / hold / stop decision

- **GO — controlled pilot:** relevant P0 gates are complete; scope and data are authorized; acceptance criteria agreed; no unresolved material blocker.
- **HOLD:** missing evidence, unclear authority, unapproved data flow, inadequate isolation, unresolved security finding, or contractual/privacy review incomplete.
- **STOP:** client authorization withdrawn; scope cannot be safely bounded; material incident occurs; prohibited access/action is attempted; integrity of evidence cannot be established.

The pilot lead records decision, rationale, approver(s), date, scope version, and supporting evidence reference. This is an internal delivery decision, not a certification or legal conclusion.

## 11. Client discovery record (copy for each prospect)

- Account / client:
- Sponsor / technical owner:
- Date and participants:
- Engineering team and size:
- AI coding agent(s) in use:
- Workflow selected:
- Repository and environment boundary (safe identifier only):
- Current control/approval model:
- Observable pain point (client stated):
- Existing controls (client evidenced):
- Hypothesis to test:
- Data categories and sensitivity:
- Candidate capabilities:
- Prohibited actions:
- Baseline metrics / unknowns:
- Client acceptance authority:
- Discovery outcome: A — validated gap / B — sufficient existing control / C — boundary unclear / D — no relevant problem
- Follow-up and owner:

## 12. Pilot decision record (copy for each pilot)

- Pilot ID:
- Client/account (private identifier if necessary):
- Scope version:
- Repository commit SHA:
- NAEOS release/toolchain:
- Environment identifier:
- Gate status G0–G8:
- Acceptance criteria version:
- Run IDs / evidence references:
- Deviations and incidents:
- Final decision: GO / HOLD / STOP
- Client acceptance: pending / accepted / rejected
- Data return/deletion status:
- Access revocation status:
- Approver(s) and date:

## 13. Evidence principle

**Claims must be bounded by observed evidence.** A successful local demo supports only the behavior demonstrated by that run. It does not, by itself, prove production readiness, security of arbitrary integrations, enterprise compliance, scalability, customer adoption, or SLA performance.
