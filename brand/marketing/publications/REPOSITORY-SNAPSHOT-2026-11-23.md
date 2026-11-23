# NAEOS Repository Snapshot for Publications

**Snapshot date:** 2026-11-23  
**Repository:** [NAEOS-foundation/naeos](https://github.com/NAEOS-foundation/naeos)  
**Branch reviewed:** `main`  
**Commit reviewed:** `6ff2bc40234369de17c2e34ddefdebdef97dc59e`  
**Editorial status:** Verified source snapshot; update before publishing if repository moves materially.

This snapshot anchors the publication family in the repository state reviewed on the date above. It is not a release certification or a production-readiness statement.

## 1. Repository-backed positioning

NAEOS describes itself as an open-source engineering control plane for AI coding agents. The repo-level framing connects engineering intent, policy, authorization, execution, observation, evidence, and verification. This framing is supported by the README and Verified Golden Path.

Use the following concise statement:

> NAEOS is an engineering control plane for AI coding agents that makes engineering intent, policy decisions, authorized execution, and evidence more inspectable.

Avoid broader security guarantees unless supported by scoped, reproducible evidence.

## 2. Verified public proof surfaces

| Proof surface | Repository source | What it supports | Boundary |
|---|---|---|---|
| Public control-plane demo | [naeos.dev/control-plane](https://naeos.dev/control-plane/) referenced by `START-HERE.md` and `docs/VERIFIED-GOLDEN-PATH.md` | Inspect authorization decisions through the public side-effect-free demo | Decision is not execution |
| Golden Path | `docs/GOLDEN-PATH.md`, `examples/demo-cli/run-demo.sh` | Local spec → NEIR → validation → policy rejection → context → generation → artifacts and traceable metadata | Does not establish production readiness, adoption, compliance, or universal integration safety |
| Verified Golden Path | `docs/VERIFIED-GOLDEN-PATH.md` | Connect public demo, local proof, evidence inspection, independent verification and challenge experiments | A combined onboarding path, not a second implementation |
| Reference Demo | `docs/REFERENCE-DEMO.md` | Reviewer-facing evidence story and reproducibility checklist | Use repository outputs as evidence; screenshots/claims are not substitutes |
| P1.11 Independent Verifier | `docs/control-plane/p1-11-independent-verifier-cli.md` | Read-only verification of canonical EvidenceBundle digest, signature, and trust-anchor match | Does not execute actions, contact control plane, re-run policy, or supply trust-store distribution/KMS |
| Governance Lifecycle Experiment | `experiments/governance-lifecycle/README.md` | Deterministic lifecycle scenarios separating agent claim from observation/evidence | Local deterministic experiment, not a penetration test |
| Handoff Governance Experiment | `experiments/handoff-governance/README.md` | Handoff acceptance, capability widening/escalation rejection, replay protection (as documented) | Claims must be scoped to the experiment and its test evidence |
| External Validation | `docs/EXTERNAL-VALIDATION.md`, `docs/EXTERNAL-VALIDATOR-BRIEF.md`, `docs/EXTERNAL-VALIDATOR-PROGRAM.md`, and `docs/EXTERNAL-VALIDATION-MISSIONS.md` | Reproducible external evaluation process and commit-pinned evidence | Results are commit- and evaluator-specific |
| Pilot Readiness | `docs/PILOT-READINESS.md` | Narrow evaluation flow for a design partner | Evaluation path; not a production deployment claim |

## 3. Release and implementation references

- Repository README currently advertises Go `1.26.11+` and Apache 2.0 licensing.
- Changelog lists `3.7.1` as the latest release entry at this snapshot; confirm tags/releases before citing as current version.
- `docs/READINESS-CONTRACT.md` describes gates and evidence required for repository-level readiness.
- `DOCUMENTATION-AUTHORITY.md` governs normative-document precedence. Publications are explanatory editorial surfaces, not normative specifications.

## 4. Recommended publication sequence

1. **Engineering Notes:** distinguish execution traceability (`run.json`) from canonical EvidenceBundle verification (P1.11).
2. **Founder Journey:** document the actual process of making a reproducible proof and publishing its limits; personalize and review the founder narrative before publication.
3. **Partner Brief:** invite a bounded technical evaluation through the Golden Path / External Validation / Pilot Readiness, with evidence and limitations made explicit.

## 5. Evidence hygiene

- Pin technical claims to a file, test, experiment, commit, or release.
- When discussing behavior, state whether it is documentation, implementation, deterministic experiment, or observed run.
- Keep status labels: **Implemented**, **Documented**, **Experiment**, **Planned**, **Not tested**.
- Never infer production readiness, adoption, certification, compliance, performance, or customer outcomes from repository documentation alone.
- Recheck this snapshot and source links immediately before publication.

## 6. Source links

- [README](https://github.com/NAEOS-foundation/naeos/blob/main/README.md)
- [START-HERE](https://github.com/NAEOS-foundation/naeos/blob/main/START-HERE.md)
- [Verified Golden Path](https://github.com/NAEOS-foundation/naeos/blob/main/docs/VERIFIED-GOLDEN-PATH.md)
- [Golden Path](https://github.com/NAEOS-foundation/naeos/blob/main/docs/GOLDEN-PATH.md)
- [Reference Demo](https://github.com/NAEOS-foundation/naeos/blob/main/docs/REFERENCE-DEMO.md)
- [Readiness Contract](https://github.com/NAEOS-foundation/naeos/blob/main/docs/READINESS-CONTRACT.md)
- [P1.11 Independent Verifier](https://github.com/NAEOS-foundation/naeos/blob/main/docs/control-plane/p1-11-independent-verifier-cli.md)
- [Governance Lifecycle Experiment](https://github.com/NAEOS-foundation/naeos/blob/main/experiments/governance-lifecycle/README.md)
- [External Validation](https://github.com/NAEOS-foundation/naeos/blob/main/docs/EXTERNAL-VALIDATION.md)
- [External Validator Brief](https://github.com/NAEOS-foundation/naeos/blob/main/docs/EXTERNAL-VALIDATOR-BRIEF.md)
- [External Validator Program](https://github.com/NAEOS-foundation/naeos/blob/main/docs/EXTERNAL-VALIDATOR-PROGRAM.md)
- [External Validation Missions](https://github.com/NAEOS-foundation/naeos/blob/main/docs/EXTERNAL-VALIDATION-MISSIONS.md)
- [Pilot Readiness](https://github.com/NAEOS-foundation/naeos/blob/main/docs/PILOT-READINESS.md)
- [Documentation Authority](https://github.com/NAEOS-foundation/naeos/blob/main/DOCUMENTATION-AUTHORITY.md)
