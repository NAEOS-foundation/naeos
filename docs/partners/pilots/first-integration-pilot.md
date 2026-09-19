# Recommended First Integration Pilot

This is a partner-neutral starting proposal. It can be adapted for a developer
tool, testing service, observability tool, security tool, or API without naming
an organization before a technical conversation.

## Pilot Objective

Validate one external tool integration through the NAEOS plugin contract and
produce a reproducible example that does not modify the NAEOS core kernel.

## Problem

Teams need to connect an engineering tool to a specification-driven workflow
without creating a one-off integration that bypasses validation, lifecycle
controls, or reproducibility.

## Proposed Integration

Build a minimal plugin or adapter that accepts one defined NAEOS input, invokes
the partner tool in an approved test environment, and returns a structured
result. The initial scope should use one operation and one fixture.

Possible first scopes:

- validate a specification or generated artifact;
- collect a tool result as pipeline evidence;
- run a bounded security, testing, or observability check;
- expose a partner API through a documented plugin action.

The implementation should follow the existing plugin lifecycle and contract
validation model. A WASM build may be used when the partner tool supports a
suitable isolated interface.

## Partner Contribution

Potential contribution: API or sandbox access, interface documentation, a test
account or fixture, rate-limit guidance, and one technical contact. Access is
optional, subject to eligibility and approval, and mutually agreed before work
starts. No payment, credits, or production commitment is assumed.

## NAEOS Contribution

NAEOS can provide:

- a reference plugin based on the official examples;
- integration and test work accepted under repository review;
- documentation and a reproducible command sequence;
- review of lifecycle, failure, licensing, and secret-handling behavior;
- an open-source example if the contribution is accepted.

## Repository Baseline

- [Official example plugins](../../../examples/plugins/README.md) demonstrate
  Go plugins, structured reports, JSON-over-stdio, and WASM builds.
- The [trivy-config example](../../../examples/plugins/trivy-config/README.md)
  provides the first concrete native integration reference for this pilot.
- [NES-009 Plugin](../../NES-009-Plugin.md) defines discovery, registration,
  validation, activation, isolation, and deactivation expectations.
- [Partner Policy](../PARTNER-POLICY.md) governs security, branding, claims,
  licensing, and publicity.

## Verified Starting Point

The repository baseline was checked before proposing this pilot:

```text
go test -race ./examples/plugins/...
GOOS=wasip1 GOARCH=wasm go build -o hello.wasm ./examples/plugins/hello
```

The example plugin tests pass, and the `hello` example produces a non-empty
WASM module. These checks validate the starting plugin path only; they do not
validate any partner tool or future integration.

## Deliverables

1. One plugin or adapter with a clearly documented input and output contract.
2. One non-sensitive fixture and a local test command.
3. Tests for the success path and at least one partner-tool failure path.
4. Documentation showing installation, execution, limitations, and cleanup.
5. A short pilot record stating what was verified and what remains unverified.
6. A JSON report stored as a local or CI artifact, not committed to Git.

## Technical Acceptance Criteria

- The plugin passes contract validation before activation.
- The core kernel remains unchanged by the integration.
- The fixture produces a structured, inspectable result.
- Partner-tool failures return a documented error without leaking credentials.
- `CRITICAL` and `HIGH` findings are blocking; lower severities remain visible
  as non-blocking findings.
- The example can be executed from a clean checkout with approved prerequisites.
- Tests pass for the agreed scope.

## Timeline

Two to four weeks:

- Week 1: confirm interface, fixture, access, security boundary, and acceptance
  criteria.
- Week 2: implement the smallest working integration and tests.
- Week 3: review failure handling, documentation, and reproducibility.
- Week 4: decide whether to publish, revise, or stop.

## Publicity / Case Study

No partner name, logo, quote, metric, performance claim, or endorsement is
published without explicit approval. A declined publication does not prevent a
private technical conclusion.

## Exit Criteria

The pilot is complete when the acceptance criteria are reviewed, limitations
are recorded, credentials and temporary access are removed or returned, and
both parties agree whether to stop, revise, or extend the work.
