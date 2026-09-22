# NAEOS Agent Handoff Governance Experiment v1

## Thesis

> A handoff transfers context, not authority.

This experiment tests whether a downstream agent/component can receive an authorization handoff without silently widening the authority granted by the upstream boundary.

It exercises the real `internal/investordemo.HandoffValidator` rather than a parallel mock implementation.

## Run

```bash
go run ./experiments/handoff-governance
```

Exit code `0` means every scenario passed.

## Scenarios

| Scenario | Invariant |
|---|---|
| Valid signed handoff | A correctly bound contract is accepted. |
| Capability widening | Requested capability must be inside the authorized set. |
| Downstream escalation | A downstream handoff cannot request authority absent from the parent. |
| Replay | A nonce cannot be accepted twice. |
| Payload tampering | The payload must match its declared digest. |
| Provenance mismatch | The claimed source must match the initiator. |
| Contract version mismatch | Unsupported protocol versions fail closed. |
| Canonicalization mismatch | Unsupported canonicalization versions fail closed. |
| Expired handoff | An expired authorization cannot be accepted. |
| Signature tampering | A modified signature invalidates the contract. |

## Evidence

The experiment emits JSON containing each scenario, expected invariant, observed result, and validator evidence. It is intentionally deterministic at the scenario level: time-sensitive cases use fixed past/future offsets relative to the process clock and do not depend on network services or an LLM.

## Claim discipline

This experiment demonstrates behavior of the current NAEOS handoff validator under these deterministic inputs. It is not a production penetration test and does not establish that all handoff implementations are secure.

CI validation: this experiment is exercised by the repository test/lint pipeline and is expected to remain deterministic and network-free.
