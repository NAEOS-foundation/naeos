# NAEOS Experiment Evidence Matrix v1

Purpose: connect NAEOS architectural claims to implemented control points, executable experiments, regression tests, and CI evidence.

This matrix is deliberately evidence-first. A document or diagram is not counted as proof unless an executable test or experiment exercises the relevant implementation.

| NAEOS claim | Implementation | Experiment / test | Evidence | CI |
|---|---|---|---|---|
| Model intent is not authorization | `internal/governance/control` | Governance Lifecycle v1; policy tests | Policy decision record independent of agent intent | Governance Lifecycle job |
| Governed execution fails closed when no policy is effective | Pipeline governance requirement + policy-bypass oracle | Policy Bypass: empty governance configuration | Explicit DENY / governance-unconfigured state | Policy Bypass Experiment job |
| Policy evaluation must receive the data it claims to govern | Versioned policy context in pipeline | Policy Bypass H4 | Context version/digest and NOT_EVALUATED distinction | Policy Bypass Experiment job |
| Mutable instructions are not the policy authority | Prompt/compiler boundary + adversarial harness | Policy Bypass H1 | Override scenario and governance outcome | Policy Bypass Experiment job |
| Non-finite numeric operands cannot silently satisfy finite constraints | Policy evaluator numeric validation | Policy Bypass H3 | NaN / +/-Inf regression outcomes | Policy Bypass Experiment job |
| Approval binds to an exact artifact | `internal/evidence` + `internal/verification` | Governance Lifecycle: artifact mutation | Artifact hash mismatch causes verification failure | Governance Lifecycle job |
| Agent claims are not execution evidence | Evidence + observation separation | Governance Lifecycle: claim vs observation | Agent claim contradicted by independent observation | Governance Lifecycle job |
| Policy freshness matters at execution/verification | Policy registry + freshness verifier | Governance Lifecycle: stale replay | Old policy version fails independent verification | Governance Lifecycle job |
| Handoff cannot widen capability | `internal/investordemo.HandoffValidator` | Agent Handoff Governance v1 | Capability widening / downstream escalation are rejected | Handoff Governance job |
| Handoff payload is integrity-bound | Handoff validator payload digest | Agent Handoff Governance v1 | Tampered payload is rejected | Handoff Governance job |
| Handoff provenance is explicit | Handoff validator provenance check | Agent Handoff Governance v1 | Missing/mismatched source fails closed | Handoff Governance job |
| Handoff protocol/canonicalization versions are trust boundaries | Handoff validator version checks | Agent Handoff Governance v1 | Unsupported versions are rejected | Handoff Governance job |
| Handoff replay is detectable | Handoff validator nonce ledger | Agent Handoff Governance v1 | Second use of same nonce is rejected | Handoff Governance job |

## Interpretation

Three evidence levels should remain separate:

1. **Unit/regression evidence** — a narrow invariant is pinned by a test.
2. **Experiment evidence** — a multi-step or adversarial scenario exercises a real component.
3. **Production evidence** — deployment-level observation from an actual external side effect.

NAEOS currently has strong deterministic evidence at levels 1 and 2 for the claims above. That should not be described as equivalent to production security validation.

## Next evidence gap

The largest remaining gap is level-3 evidence: a real adapter/tool execution where NAEOS records the requested action, authorization, actual side effect, and independent observation without trusting the agent's self-report.
