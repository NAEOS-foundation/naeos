# DANA — Design Partner Discovery Pack #001

Status: Founder preparation  
Account: DANA Indonesia  
Primary public persona: Norman Sasono, CTO  
Objective: Validate one concrete AI-assisted engineering control gap and determine whether a bounded NAEOS pilot is justified.

## Public evidence

- DANA reported in September 2026 that approximately 80–90% of its code is generated with AI support.
- DANA states that engineers remain accountable for problem formulation, architecture, validation, security, and final decisions.
- DANA previously announced GitHub Copilot adoption across backend, frontend, QA, and SRE, with nearly 300 developers using it at the time of the 2024 announcement.
- Norman Sasono currently describes his CTO remit as spanning engineering, data, AI, infrastructure, and cybersecurity.

Sources:
- DANA AI@Work Lab newsroom: https://www.dana.id/corporate/newsroom/dari-hype-ke-dampak-nyata-dana-soroti-implementasi-ai-bagi-bisnis-dan-talenta?lng=id
- DANA + GitHub Copilot newsroom: https://www.dana.id/corporate/newsroom/dana-berkolaborasi-dengan-microsoft-untuk-mempelopori-peningkatan-produktivitas-kerja-berbasis-ai-melalui-git-hub-copilot?lng=id
- Norman Sasono public profile: https://id.linkedin.com/in/normansasono

## Discovery thesis

Do not assume DANA has a control gap.

Test this question:

> As AI-assisted software development becomes a normal part of the SDLC, where does DANA enforce the boundary between what an AI agent may propose, what it may execute, and what must remain explicitly authorized by an engineer or policy?

A second question follows:

> After an AI-assisted action occurs, what durable evidence proves the authorization, execution, verification, and final outcome?

## Candidate workflow hypotheses

These are hypotheses only. Ask DANA to select the real workflow.

### H1 — Pull request change

AI agent:
- reads repository context
- proposes code changes
- creates/modifies files
- runs tests
- opens a pull request

Control questions:
- What repository scope can it access?
- Can it push directly?
- Can it modify workflow/configuration files?
- Can it create or modify CI/CD configuration?
- Which actions require human approval?

### H2 — CI/CD action

AI-assisted workflow:
- code change reaches CI
- agent/tool can trigger or influence build/test/deployment steps

Control questions:
- Which capabilities are granted to the agent?
- Can policy change while execution is in progress?
- Can a previously approved action become invalid after policy/configuration changes?
- What evidence connects the approval to the exact execution?

### H3 — Production-adjacent engineering action

AI agent or AI-assisted automation interacts with:
- infrastructure
- secrets/configuration
- deployment systems
- operational tooling

Control questions:
- What actions are explicitly prohibited?
- Which actions require step-up authorization?
- Is authorization bound to identity, repository, environment, task, and policy version?
- Can the action be independently reconstructed later?

## 20-minute conversation

### 0–3 min — Current state

1. How is AI-assisted development currently used across engineering?
2. Which workflows are considered high-value or high-risk?
3. Are you moving from code generation toward agents that can execute actions?

### 3–8 min — One real workflow

Ask:

> Could we pick one real AI-assisted engineering workflow and walk through exactly what the agent can read, propose, change, execute, and trigger?

Capture:
- repository
- agent/tool
- identity
- capabilities
- execution environment
- approval points
- CI/CD boundary
- human verification
- audit/evidence

### 8–13 min — Control model

Ask:

1. Where is authorization evaluated?
2. Is authorization capability-based, role-based, policy-based, or another mechanism?
3. Can the agent's allowed capabilities differ by repository/environment/task?
4. What happens if policy changes during execution?
5. How are tool calls/actions recorded?
6. Can you reconstruct exactly what happened after the fact?
7. Which component independently verifies the result?
8. What happens when an agent hands work to another agent/system?

### 13–17 min — Gap test

Use neutral probes:

- Is there any action you deliberately do not want an AI agent to perform even if the agent has technical credentials?
- Is there any distinction today between "the agent intended to do X" and "the system proves X actually happened"?
- Where would an authorization decision become stale?
- What evidence would security/audit need after a consequential AI-assisted change?
- Which controls exist today but are difficult to enforce consistently across tools?

Do not introduce NAEOS as the answer until the control gap is understood.

### 17–20 min — Pilot test

If a concrete gap exists, propose:

> Let's take one repository and one bounded AI-assisted workflow. We don't replace the existing coding tool. We place NAEOS at the control boundary, define the allowed capabilities, capture the policy decision and execution evidence, and independently verify the result.

## Candidate pilot

### Scope

- 1 repository
- 1 AI coding-agent workflow
- 1 engineering team
- 1–3 explicit capabilities
- 1 policy boundary
- 1 verification path
- 2–4 weeks

### Example capability boundary

Allowed:
- read selected repository paths
- create branch
- modify application code
- run tests

Require explicit authorization:
- modify CI/CD configuration
- modify security-sensitive configuration
- deploy to protected environment

Denied:
- access unrelated repositories
- access production secrets
- bypass required verification
- execute outside the declared task boundary

These are examples only. Actual DANA policy must be defined by DANA.

## NAEOS pilot acceptance criteria

A pilot is successful only if the agreed workflow demonstrates:

- [ ] Authorized action succeeds.
- [ ] Unauthorized action is blocked or escalated.
- [ ] Capability set is explicit.
- [ ] Policy decision is attributable to a policy/version.
- [ ] Execution is attributable to the authorized task/agent identity.
- [ ] Observation is separated from the agent's claim.
- [ ] Durable evidence is generated.
- [ ] Verification is independently identifiable.
- [ ] Policy/configuration change invalidates stale authorization where required.
- [ ] Existing developer workflow remains usable.
- [ ] Results are reproducible by another engineer.

## Metrics

Record baseline before intervention where possible:

| Metric | Baseline | Pilot | Evidence |
|---|---:|---:|---|
| Unauthorized-action attempts | Unknown | — | — |
| Policy violations | Unknown | — | — |
| Manual approvals | Unknown | — | — |
| Verification coverage | Unknown | — | — |
| Evidence completeness | Unknown | — | — |
| PR cycle time | Unknown | — | — |
| Developer friction | Unknown | — | — |

## Discovery outcome taxonomy

### A — Validated control gap
A concrete workflow and observable control gap exist. Proceed to bounded pilot design.

### B — Existing control is sufficient
DANA demonstrates an existing mechanism that already covers the tested boundary. Record the mechanism and do not manufacture a NAEOS gap.

### C — Problem exists but boundary is unclear
Continue discovery with engineering/security stakeholders before proposing a pilot.

### D — No relevant problem
Close the account for this hypothesis and record the learning.

## Founder rule

The goal of this conversation is not to prove NAEOS is needed.

The goal is to determine whether a real engineering-control problem exists at the boundary between AI agent intent, authorization, execution, verification, and evidence.
