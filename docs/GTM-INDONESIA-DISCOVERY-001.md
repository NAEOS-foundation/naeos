# Indonesia Design Partner — Discovery Scorecard #001

Status: Operating artifact  
Purpose: Convert a 20-minute technical exchange into a reproducible pilot decision.

## 1. Call structure

### 0–3 min — Context

- What AI-assisted engineering workflow are you using today?
- Which AI coding agent or internal agent is involved?
- What changed recently that made governance/control relevant?

### 3–8 min — Workflow

Map one concrete workflow:

`Intent → Agent → Proposed Action → Authorization → Execution → Observation → Verification → Evidence`

Capture the actual systems involved, not a hypothetical architecture.

### 8–13 min — Control boundary

Ask:

1. What can the agent currently read?
2. What can it write or execute?
3. Which actions require explicit approval?
4. Which actions are prohibited?
5. Where is authorization evaluated?
6. Can authorization change while a task is running?
7. What happens when policy and agent context disagree?
8. What evidence proves what the agent actually did?
9. Who independently verifies the result?
10. What happens when a handoff crosses an authority boundary?

### 13–17 min — Impact

Capture measurable baseline where available:

- PR cycle time
- Review/rework time
- Manual approval count
- Agent intervention count
- Policy violations
- Unauthorized-action attempts
- Verification coverage
- Evidence completeness
- Developer context-preparation time
- Incident or near-miss examples

Do not invent baselines. Mark unknown values as unknown.

### 17–20 min — Pilot test

Confirm:

- one repository or bounded engineering surface
- one agent/workflow
- one policy boundary
- one verification path
- one evidence format
- technical owner
- 2–4 week evaluation window
- explicit acceptance criteria

## 2. Discovery record

### Account

- Account:
- Date:
- Participant:
- Role:
- Channel:
- NAEOS participant:

### Workflow

- Workflow:
- Repository/system:
- Agent:
- Agent capabilities:
- Consequential actions:
- Current authorization mechanism:
- Current verification:
- Current evidence/audit:
- Existing controls:

### Control-gap statement

Write one falsifiable sentence:

> Today, [agent/system] can [action] under [control], but [observable gap] makes [risk/operational problem] difficult to control or verify.

If no concrete gap exists, record:

> No validated NAEOS control gap identified during this discovery.

### Pilot hypothesis

> If NAEOS is placed at [boundary], then [specific control/evidence behavior] should become observable, while [existing workflow] remains unchanged.

### Acceptance criteria

- [ ] Allowed action succeeds.
- [ ] Disallowed action is blocked or requires the defined approval.
- [ ] Policy decision is attributable to a policy/version.
- [ ] Authorized capability set is observable.
- [ ] Execution produces durable evidence.
- [ ] Verification result is independently identifiable.
- [ ] Failure path is reproducible.
- [ ] Existing developer workflow remains usable.

### Outcome

- Discovery status: Not started / Scheduled / Completed
- Pilot candidate: Yes / No / Unclear
- Partner owner identified: Yes / No
- Next action:
- Next-action date:
- Evidence/source:

## 3. Do not overclaim

Discovery notes must distinguish:

- **Observed:** directly demonstrated during the call.
- **Reported:** stated by the participant.
- **Documented:** supported by public/internal documentation.
- **Hypothesis:** proposed explanation or pilot idea.
- **Unknown:** not established.

A positive conversation is not a pilot. A pilot is not a customer. A successful pilot is not automatically a case study or production deployment.
