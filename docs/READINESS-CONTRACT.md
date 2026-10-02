# NAEOS Readiness Contract

This document defines the repository-level readiness contract for turning the existing NAEOS control-plane proof into a consumable open-source engineering product.

It is a **readiness checklist and evidence map**, not a certification and not a claim that every deployment or external agent integration is production-safe.

## 1. Fixed baseline

The current Phase 4.5 implementation baseline is:

`daf79819bd8979c1a298e569ae84449553b50d7e`

External Validation #001 remains historical and commit-scoped to its recorded SHA. External Validation #002 is intentionally not part of this milestone.

## 2. Canonical user path

There is one canonical first-run path:

```text
START-HERE
   ↓
Verified Golden Path
   ↓
Golden Path
   ↓
Reference Demo / Evidence
   ↓
Independent verification
   ↓
Contribution or challenge
```

Use these documents for their distinct purposes:

| Surface | Purpose |
|---|---|
| `START-HERE.md` | Choose the shortest path into NAEOS |
| `docs/VERIFIED-GOLDEN-PATH.md` | Connect the public/local proof and challenge-oriented experiments |
| `docs/GOLDEN-PATH.md` | Define the runnable acceptance contract |
| `docs/REFERENCE-DEMO.md` | Explain the evidence story to an engineer/reviewer |
| `docs/EXTERNAL-EVALUATOR-QUICKSTART.md` | Pin a commit and perform an independent evaluation |
| `docs/EXTERNAL-VALIDATION.md` | Record an evaluation result |
| `docs/control-plane/p1-11-independent-verifier-cli.md` | Verify a serialized evidence bundle without executing an action |

These surfaces should not create alternate implementations of the same proof.

## 3. Readiness gates

### Gate A — Build integrity

Required:

- CLI builds from source.
- Repository tests pass for the affected change.
- Generated Go output is covered by the compile gate.
- Formatting and static-analysis requirements remain satisfied.
- CI remains green for the relevant repository gates.

Evidence:

- CI workflow results.
- Test output.
- Generated-output compile checks.

### Gate B — Control-plane proof

Required:

- Specification is accepted.
- NEIR is inspectable.
- Validation occurs before generation.
- Deliberately invalid policy input is rejected.
- Agent context is generated from the engineering model.
- Authorized generation completes.
- Run metadata connects execution to the declared specification and derived NEIR.

Evidence:

- `spec.yaml`
- `inspect.json`
- `validate.json`
- `invalid-policy.log`
- `context.md`
- `context.json`
- `run.json`
- `generated/`
- `summary.md`

### Gate C — Evidence integrity

Required:

- Evidence is inspectable after execution.
- Run identity and hashes are recorded.
- Evidence claims remain distinguishable from observations.
- Independent verification is described separately from the component that produced the claim.
- Historical validation records remain pinned to their exact commits.

Evidence:

- Golden Path output.
- Reference Demo evidence map.
- P1.11 verifier documentation.
- Commit-scoped validation records.

### Gate D — Governance boundaries

Required:

- Policy decisions remain distinct from runtime execution.
- A policy rejection is observable.
- Handoff inputs are not treated as authority merely because an agent supplied them.
- Plugin capabilities remain bounded by explicit authorization.
- Missing authorization fails closed where the governing contract requires it.

Evidence:

- Governance lifecycle experiment.
- Handoff contract documentation.
- Plugin capability-boundary documentation.
- Corresponding tests/CI gates.

### Gate E — Reproducibility

Required:

- A clean checkout can identify the exact commit to evaluate.
- The documented commands use an isolated output directory.
- Prerequisites and optional dependencies are explicit.
- Deviations are recorded rather than silently absorbed.
- No external service or LLM API key is required for the baseline Golden Path.

Evidence:

- `docs/EXTERNAL-EVALUATOR-QUICKSTART.md`
- `docs/GOLDEN-PATH.md`
- Recorded evaluator reports when available.

### Gate F — OSS contribution readiness

Required:

- A new engineer can find the contributor workflow.
- A contributor can identify a small first contribution.
- Bug, documentation, policy, evidence, plugin, and architecture challenges have a clear route.
- DCO and repository engineering requirements are visible before contribution.
- NAEOS OSS naming is used consistently.

Entry points:

- `CONTRIBUTING.md`
- `START-HERE.md`
- `.github/ISSUE_TEMPLATE/`
- `docs/community/contributor-ladder.md`

## 4. Release-readiness matrix

| Area | Minimum evidence | Status meaning |
|---|---|---|
| Build | CI + local build contract | Build path is defined and gated |
| Tests | Relevant automated tests | Behavioral regression is covered |
| Generation | Generated Go compile gate | Generated output is not merely text-checked |
| Policy | Rejection experiment + tests | Control boundary is observable |
| Evidence | `run.json` + evidence artifacts | Run is traceable |
| Verification | Independent verifier path | Evidence can be checked separately |
| Handoff | Contract + boundary semantics | Authority is explicit |
| Plugins | Capability boundary | Capability is not inferred from intent |
| Documentation | Canonical path + maps | User/evaluator path is discoverable |
| Reproducibility | Pinned-commit runbook | Evaluation can be scoped precisely |
| Contribution | CONTRIBUTING + START-HERE | External engineers can participate |

A gate is **satisfied** when the required evidence exists and is covered by the repository's applicable checks. A gate is **not satisfied** merely because a document claims that the behavior exists.

## 5. Evidence and claim boundary

NAEOS should distinguish these statements:

- **Implemented:** the repository contains the mechanism.
- **Tested:** automated checks exercise the mechanism.
- **Reproduced:** a recorded run observed the mechanism at a specific commit.
- **Independently verified:** an independent verification procedure checked the relevant evidence.
- **Production-ready:** requires deployment-specific operational, security, performance, reliability, and compliance evidence beyond this repository checklist.

The last category must not be inferred from the first four.

## 6. What this milestone intentionally does not do

Phase 4.5 does not:

- add another governance primitive;
- add a new plugin capability;
- create External Validation #002;
- convert historical validation into current-main validation;
- claim customer adoption;
- claim enterprise compliance;
- claim universal security of external AI agents;
- replace the normative architecture or technical specifications.

The objective is to make the existing proof **usable, inspectable, and contribution-ready**.

## 7. Definition of done

Phase 4.5 is ready to close when:

1. `START-HERE.md` exposes one obvious first-run path.
2. The Golden Path remains the single runnable acceptance contract.
3. The evidence story and independent verifier are linked without creating duplicate execution paths.
4. This readiness contract provides an auditable checklist for repository changes.
5. Existing Phase 4 evidence remains commit-scoped.
6. Contributors can identify how to run, inspect, challenge, report, and extend NAEOS.
7. No readiness claim depends on an unrecorded external validation run.

**Evidence first. Claims second.**

**Architecture Drives Engineering.**
