---
title: "Why AI Coding Needs a Control Plane"
description: "Agents can decide. Agents should not authorize themselves. A technical walkthrough of the NAEOS investor demo control plane and how live audit streams into SIEM and OTLP."
date: 2026-09-11
author: "NAEOS Foundation"
categories: ["concept", "tutorial"]
---

The code is the easy part now.

Frameworks, assistants, and agent runtimes can produce a working pull request faster than most teams can review it. The bottleneck has moved — from *writing code* to *knowing what the system is allowed to do, and proving it did only that*.

AI tools are good at **deciding**. The hard engineering question is: who **authorizes**?

This post walks through a reference answer we built into NAEOS — a control plane where authorization, policy enforcement, verification, and audit run **outside** the agent — and shows how it streams observability into SIEM and OTLP collectors.

## The problem: agents that authorize themselves

When you give an AI agent a task, you are handing it a decision-maker *and* letting it act on its own authority. A task like "update the payment service and run the tests" is mostly fine. But nothing stops the same agent from attempting to rotate credentials, modify policy, or pass authority to another agent — unless something outside it says no.

That "something" is a control plane. It is the difference between:

- a tool that *generates code*, and
- an engineering system that *governs what generated code is allowed to do*.

## What we actually built

NAEOS already treats software as a machine-readable engineering model (NEIR) instead of a pile of files. The investor demo applies the same discipline to agent operations: a hardened Go HTTP server where every request goes through a deterministic pipeline — policy engine, capability authority, handoff validator, execution gate, independent verifier, and an append-only audit ledger.

None of the demo's ALLOW/BLOCK results are mocked. They come from the same code paths you would run in production. That was a deliberate constraint: a demo you can't trust is a demo you can't copy.

Start it with one command:

```bash
naeos demo --addr :9091
# NAEOS investor demo listening on http://localhost:9091
```

The server exposes the whole plane: `/api/authorize`, `/api/execute`, `/api/policy`, `/api/grants`, `/api/audit`, `/api/verification`, and a scenario matrix.

## A scripted walkthrough

`POST /api/investor-demo` runs a deterministic 10-step sequence against the real engine. The interesting part is not the happy path — it is the six ways the agent tries to exceed its authority:

| # | Action | Result |
|---|--------|--------|
| 4 | Agent requests `test.execute` | ALLOW |
| 5 | Agent attempts `credential.rotate` | BLOCK |
| 6 | Agent attempts `policy.modify` | BLOCK |
| 7 | Agent A hands off to Agent B with `credential.rotate` | HANDOFF REJECTED |
| 8 | Agent replays a previously valid handoff contract | REPLAY REJECTED |
| 9 | POLICY-017 updated v17 → v18; agent reuses an old grant | STALE AUTHORIZATION → BLOCK |
| 10 | Independent verification of the whole session | VERIFICATION: FAIL |

Step 10 is the point. Because unauthorized actions were attempted, the independent verifier — running outside the agent's influence — reports **FAIL** for the session. Agents can reason. Authorization happens in the control plane.

More scenarios run individually through `/api/scenario`: IAM modification, capability escalation, replays, and stale grants all fail closed. Statefulness matters too: when a policy version changes or provenance metadata drifts, previous grants stop working instead of lingering.

## The audit trail is the backbone

Every decision becomes an audit event. Events get sequential IDs (`AUD-00001`, `AUD-00002`, ...) from the ledger itself, are append-only, and can be forwarded to any observer — that is the hook the observability bridge hangs on.

So the question "what did this agent do?" has a real answer. And with NAEOS v3.5.0, it has a *trackable* answer.

## Streaming into SIEM and OTLP

The demo control plane wires straight into the NAEOS observability stack:

```bash
naeos demo --addr :9091 \
  --siem-endpoint http://localhost:9000 \
  --otlp-endpoint http://localhost:4318
```

- **SIEM:** every audit event is forwarded to a collector as Common Event Format (CEF) or NDJSON, tagged with a tenant ID. Forwarding is asynchronous; if the collector cannot keep up, events are dropped and counted rather than blocking the control plane.
- **OTLP/HTTP:** every API request becomes a span exported to any OpenTelemetry-compatible backend, tagged with the request ID and tenant.

These are the same telemetry paths `naeos serve` and the `naeos observability` command group already use — SLO burn-rate alerting, OTLP trace export, and SIEM export are part of the release, not a separate bolt-on.

The effect is an end to "works on my machine, deployed somewhere else." A blocked capability, a rejected handoff, or a stale grant is an event with a timestamp, an agent ID, a policy version, and a trace back to the request that produced it.

## Why this matters beyond agent demos

Two years from now, most engineering teams will not ask *which AI tool to adopt*. They will ask *how to keep AI-generated change inside engineering guardrails* — consistent specs, validated models, governed actions, reproducible artifacts.

"Specify once. Build anywhere" was always about the engineering model. The control plane is the governance layer of that model: it decides what the software — and the agents building it — are allowed to do. The demo is deliberately small so the mechanism is visible, but the same pipeline is the governance backbone of the whole platform.

Try it:

```bash
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest
naeos demo --addr :9091 --siem-endpoint http://localhost:9000
curl -X POST http://localhost:9091/api/investor-demo
```

The implementation is open source (Apache-2.0). The control plane lives in [`internal/investordemo`](https://github.com/NAEOS-foundation/naeos/tree/main/internal/investordemo), the observability bridge in [`internal/demoobs`](https://github.com/NAEOS-foundation/naeos/tree/main/internal/demoobs). We build in public precisely so you can check the claims yourself — and tell us where we are wrong.