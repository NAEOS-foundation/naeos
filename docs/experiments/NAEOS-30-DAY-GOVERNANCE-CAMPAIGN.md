# NAEOS — 30 Days of AI Agent Governance

Campaign ID: NAEOS-CAM-001
Status: Active
Campaign thesis: AI coding agents need explicit authority, runtime enforcement, evidence, and independent verification.
Primary CTA: Run the experiment.
Secondary CTA: Review the architecture / Request a design-partner session.

## Operating rule
Every public claim must point to reproducible repository evidence. An experiment is not blanket proof of production security, compliance, customer adoption, or safety.

## 30-day schedule

| Day | Theme | Repository evidence | Public angle | CTA |
|---|---|---|---|---|
| 01 | Governance lifecycle | experiments/governance-lifecycle | Can an agent action be traced from intent to verification? | Run it |
| 02 | Agent claim vs evidence | experiments/governance-lifecycle | The agent's claim is not the evidence | Run it |
| 03 | Policy change mid-run | docs/experiments/EXP-001-policy-change-mid-run.md | What happens when permission changes halfway through a task? | Inspect evidence |
| 04 | Stale authorization | EXP-001 + control-plane tests | Old approval must not silently survive a new policy | Review behavior |
| 05 | Atomic execution boundary | docs/experiments/EXP-002-atomic-execution-commit-boundary.md | Policy changes and consequential execution need one boundary | Read experiment |
| 06 | Real local side effect | experiments/level3-evidence | Move from simulated lifecycle to an observable filesystem effect | Run it |
| 07 | Direct bypass | experiments/level3-evidence | What if execution bypasses the gateway? | Inspect result |
| 08 | Evidence tampering | experiments/level3-evidence | Can post-observation mutation remain verified? | Run it |
| 09 | Handoff authority | experiments/handoff-governance | Can one agent silently widen another agent's authority? | Run it |
| 10 | Capability widening | experiments/handoff-governance | Handoff should transfer explicit authority, not create new authority | Review contract |
| 11 | Replay resistance | experiments/handoff-governance | Why replay state matters for agent-to-agent handoffs | Run it |
| 12 | Payload integrity | experiments/handoff-governance | A valid contract cannot authorize a changed payload | Inspect evidence |
| 13 | Version negotiation | experiments/handoff-governance | Protocol mismatch should not become implicit authority | Read contract |
| 14 | Evidence completeness | experiments/evidence-completion-enforcement-v3 | Missing evidence must fail the completion boundary | Run it |
| 15 | Run binding | experiments/evidence-integrity-run-binding-v5 | Evidence from another run must not be accepted | Run it |
| 16 | Provenance | experiments/evidence-runtime-provenance-v5-1 | Bind evidence to runtime provenance | Inspect digest |
| 17 | Runtime event binding | experiments/evidence-runtime-event-binding-v5-2 | Evidence must reference the event that actually occurred | Run it |
| 18 | Event ledger | experiments/evidence-runtime-event-ledger-v5-3 | Separate observation from evidence construction | Review design |
| 19 | Independent observer | experiments/evidence-runtime-observer-v5-4 | The evidence builder should not invent the observation | Run it |
| 20 | Durable ledger | experiments/evidence-durable-runtime-ledger-v5-5 | Runtime evidence needs durable integrity | Inspect result |
| 21 | Durable receipt | experiments/evidence-durable-receipt-v5-6 | Recover and verify a sealed runtime receipt | Run it |
| 22 | Adversarial governance | experiments/policy-bypass | Test the control boundary instead of trusting the prompt | Run it |
| 23 | Attack matrix | experiments/attack-matrix | What boundaries are actually covered? | Review matrix |
| 24 | Verification independence | internal/verification + governance lifecycle | Verification should not depend solely on the claimant | Review architecture |
| 25 | Specification-first flow | README + Golden Path | Engineering intent should precede agent execution | Run Golden Path |
| 26 | Vendor neutrality | README / AI compiler | Changing agents should not require rebuilding engineering rules | Review model |
| 27 | Evidence story | docs/REFERENCE-DEMO.md | Turn one run into a reviewer-readable evidence narrative | Review demo |
| 28 | External validation | docs/EXTERNAL-VALIDATION.md | Let a third party reproduce the result | Validate independently |
| 29 | Limitations | experiment READMEs | Characterization is not a blanket production-security claim | Read limitations |
| 30 | Campaign synthesis | All above | 30 days of concrete boundaries, failures, evidence, and lessons | Request design-partner session |

## Publication format
Each day produces: one primary technical post, one short-form derivative, one repository evidence link, one screenshot or terminal capture where useful, one discussion question, and exactly one primary CTA.

Recommended structure:
Hook → concrete problem → experiment setup → observed result → why it matters → limitation → evidence → single CTA.

## Measurement
Track repository visits, experiment/demo executions where measurable, documentation engagement, GitHub discussions/issues, qualified engineering conversations, and design-partner requests.

Do not optimize around follower count alone.

## Claim discipline
Prefer: demonstrates this specific invariant; reproduces this failure mode; records this observable result; blocks this tested path; does not by itself prove X.

Never turn a characterization result into a blanket production-security, compliance, adoption, or safety claim.

## Week objectives
Week 1: make governance visible.
Week 2: make authority boundaries visible.
Week 3: make evidence durable.
Week 4: make the story externally testable.

## Campaign success criterion
The campaign succeeds when technical audiences can understand one concrete governance problem, reproduce an experiment, inspect the evidence, identify its limitation, and start a technical discussion or design-partner conversation.
