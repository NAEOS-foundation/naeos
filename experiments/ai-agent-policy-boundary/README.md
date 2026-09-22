# AI Agent Policy Boundary Experiment v1

> An AI agent may propose an action. It must not become the authority that authorizes its own side effect.

This flagship experiment composes existing NAEOS components into one end-to-end boundary:

AGENT INTENT -> POLICY DECISION -> EXECUTION GATEWAY -> SIDE EFFECT -> INDEPENDENT OBSERVATION -> EVIDENCE -> INDEPENDENT VERIFICATION

It does not introduce a second policy engine, runtime gateway, evidence store, or verifier.

## Run

    go run ./experiments/ai-agent-policy-boundary

Exit code 0 means all expected scenario assertions passed.

## Scenarios

| Scenario | Expected proof |
|---|---|
| ALLOW | A policy-authorized write crosses the gateway and is independently observed and verified. |
| DENY | A denied write never reaches the sandbox and the observer confirms no side effect. |
| REQUIRE_APPROVAL | The gateway refuses execution until an approval path exists. |
| DIRECT BYPASS | A side effect performed outside the gateway is observable but is not treated as authorized; independent verification fails. |

The fourth scenario is the key boundary test:

> A DENY decision is not proof that an out-of-band side effect was prevented.

NAEOS must distinguish authorization state from what actually happened.

## What is real

- internal/governance/control — deterministic policy decision.
- internal/runtime/gateway — execution enforcement boundary.
- internal/evidence — append-only, hash-chained evidence.
- internal/verification — independent verification chain.
- A local filesystem sandbox — real side effect target for the experiment.
- An independent filesystem observer — reads the resulting artifact separately from the executor.

## What is not claimed

This is a controlled local experiment. It does not prove that every deployment can prevent every possible out-of-band side effect. Production enforcement requires an external deployment boundary capable of constraining the agent/runtime itself.

## Why this is the flagship experiment

The question is not:

> Did the model follow the instruction?

The question is:

> Can NAEOS independently establish what the agent requested, what policy authorized, what the gateway executed, what actually happened, and whether the evidence still verifies?

That is the AI Agent Policy Boundary.