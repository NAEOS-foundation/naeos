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
| Real execution creates an independently observable side effect | `internal/runtime/gateway` + local filesystem sandbox | Level-3 Evidence v1: ALLOW | Real file write observed independently of executor return value | Level-3 Evidence job |
| DENY is tested against the side effect, not only the policy result | `internal/runtime/gateway` + control plane | Level-3 Evidence v1: DENY | Gateway denies, sandbox is not invoked, observer confirms absence | Level-3 Evidence job |
| An out-of-band side effect is not equivalent to authorization | Gateway history + evidence + independent observation | Level-3 Evidence v1: direct bypass | Side effect observed with no authorized gateway execution; verification fails | Level-3 Evidence job |
| Execution claims are not observation evidence | Gateway execution result + independent observer/verifier | Level-3 Evidence v1: execution/observation separation | Completion claim with absent effect is preserved as verification failure | Level-3 Evidence job |
| Artifact mutation after observation is detectable | `internal/evidence` + `internal/verification` | Level-3 Evidence v1: tamper | Re-read digest mismatch causes verification failure | Level-3 Evidence job |

## Interpretation

Three evidence levels should remain separate:

1. **Unit/regression evidence** — a narrow invariant is pinned by a test.
2. **Experiment evidence** — a multi-step or adversarial scenario exercises a real component.
3. **Production evidence** — deployment-level observation from an actual external side effect.

Level-3 evidence is now represented by a controlled local filesystem side effect and an independent re-read of the resulting artifact. This is stronger than simulated execution, but it remains **controlled experiment evidence**, not production security validation.

Production evidence still requires a deployment-level external side effect and an observation boundary outside the test process.
