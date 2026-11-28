# NAEOS Engineering Note #002 (Draft)

## Run Traceability Is Not Independent Evidence Verification

**Publication:** NAEOS Engineering Notes  
**Document ID:** NAEOS-PUB-2026-002  
**Date:** 2026-11-23  
**Version:** 0.2  
**Status:** Draft — repository-grounded; execution evidence and editorial review required  
**Audience:** AI engineering, platform, security, and technical evaluators

## Summary

A control-plane run can leave traceability metadata, but traceability and independent evidence verification answer different questions.

NAEOS's Golden Path produces run metadata connecting an execution to its declared specification and derived NEIR. The separate P1.11 Independent Verifier checks a serialized canonical EvidenceBundle in read-only mode, using an operator-supplied trust anchor obtained independently.

These are complementary proof surfaces—not interchangeable labels.

## Context

The NAEOS Golden Path documented in `docs/GOLDEN-PATH.md` is a local reproducible workflow: specification validation, NEIR inspection, deliberate policy rejection, AI context generation, authorized generation, artifact inspection, and run metadata. Its acceptance contract also requires expected generated files, a summary, and zero generated-artifact path collisions.

The Reference Demo asks reviewers to inspect repository outputs rather than rely on screenshots or narrative claims.

## Two evidence boundaries

### 1. Run traceability

The canonical demo's `run.json` contract requires:

- `run_id`
- `specification_hash`
- `neir_hash`
- validation, policy, context, audit, and stage metadata
- `artifact_summary`, including logical-artifact count, materialized-file count, and path-collision count

These anchors help connect the recorded run to its declared input and derived representation, and account for generated artifacts. They do not, by themselves, establish the correctness or trustworthiness of every external dependency or action.

The demo acceptance criteria require the path-collision count to be zero. A successful script also checks expected output files and writes `summary.md`.

### 2. Independent EvidenceBundle verification

P1.11 documents the command:

```bash
naeos evidence verify-bundle \
  --input-file evidence.json \
  --trusted-public-key /secure/trust/naeos-evidence.pub
```

The verifier separately reports digest integrity, signature validity, and whether the signer matches the operator-supplied trust anchor. Decision and execution bindings are checked independently. The verifier is read-only: it does not contact the control plane, re-run policy, or execute an action. Missing or malformed trust configuration fails closed.

The trusted public key must come from an operator-controlled trust channel, not merely from the bundle being verified.

## Why the distinction matters

A record that describes a run is not automatically a cryptographically verifiable canonical evidence bundle. Independent verification checks properties defined by the EvidenceBundle contract; it does not prove that the external world is safe, that an action was desirable, or that every claim about an event is true. Its conclusions are bounded by the evidence contract, trust anchor, and inputs observed.

## Reproduce the documented baseline

From a clean checkout, build the CLI and run the demo with an isolated output directory:

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
go build -o naeos ./cmd/naeos
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh
find /tmp/naeos-demo -maxdepth 3 -type f | sort
```

The script uses the repository-built `./naeos` binary when no installed CLI or explicit `NAEOS_BIN` is supplied. It validates the required run metadata and artifact accounting, verifies representative generated files, and fails on path collisions.

Then follow `docs/GOLDEN-PATH.md` and `docs/VERIFIED-GOLDEN-PATH.md` to inspect the output and the separate verification capability. The command above is a reproducibility procedure—not execution evidence from this editorial review. Record the exact commit, toolchain, exit status, run identifiers/hashes, artifact counts, and deviations when actually run.

## Evidence and sources

- `docs/GOLDEN-PATH.md` — runnable acceptance criteria and evidence map.
- `examples/demo-cli/run-demo.sh` — executable demo contract, required metadata, generated-file checks, and collision gate.
- `docs/REFERENCE-DEMO.md` — reviewer checklist and traceability contract.
- `docs/VERIFIED-GOLDEN-PATH.md` — canonical onboarding spine and proof boundaries.
- `docs/control-plane/p1-11-independent-verifier-cli.md` — verifier semantics, trust anchor, and scope.
- `docs/READINESS-CONTRACT.md` — repository readiness gates and evidence map.

## Takeaway

**Traceability connects a run to its declared engineering inputs and artifact accounting. Independent verification checks a canonical evidence record against integrity, signature, trust, and binding requirements.**

Keep these proof boundaries explicit in engineering docs, demos, and partner conversations.

## Editorial gate

Before publication, run the Golden Path on a recorded commit and attach observed output/evidence; confirm the current verifier command and semantics; and update date/version metadata if the source snapshot changes. Do not claim security certification, production readiness, customer adoption, or universal agent safety.
