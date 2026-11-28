# NAEOS Founder Journey — Draft 01

**Publication:** NAEOS Founder Journey  
**Document ID:** NAEOS-PUB-2026-FJ-001  
**Date:** 2026-11-23  
**Version:** 0.1  
**Status:** Draft for founder review — do not publish without personal voice/experience review  
**Audience:** Builders, founders, engineers, open-source community

## Building a proof before making a claim

The current NAEOS repository is organized around a practical question: can an engineer reproduce a workflow and inspect what happened without relying only on an AI agent's statement?

The project now documents a public control-plane demo, a local Golden Path, evidence inspection, independent verification, and challenge-oriented experiments. The point of connecting these surfaces is not to make the story look bigger. It is to make the evaluation path easier to follow.

## What the repository shows today

- `START-HERE.md` directs newcomers through the Verified Golden Path.
- `docs/GOLDEN-PATH.md` defines a local runnable baseline.
- `docs/REFERENCE-DEMO.md` explains how a reviewer inspects outputs.
- `docs/control-plane/p1-11-independent-verifier-cli.md` documents a read-only verifier for canonical EvidenceBundle records.
- `experiments/governance-lifecycle` explores the difference between an agent's claim and observed evidence.
- `docs/EXTERNAL-VALIDATION.md` invites commit-pinned independent reproduction.

These are repository facts. The personal lessons and decisions below must be completed by the founder before release.

## Founder reflection — complete before publication

**Why I chose to focus on this engineering problem:**  
[Founder: add your own account. Do not invent a motivation.]

**A decision or trade-off that changed the project:**  
[Describe a real decision and evidence behind it.]

**What did not work or remains incomplete:**  
[Share an actual setback, limitation, or unresolved gap.]

**What I learned from making the proof reproducible:**  
[Personal lesson based on your experience, not generic marketing language.]

## The next challenge

The repository invites engineers to challenge its assumptions and reproduce the workflow. The next meaningful milestone is not a bigger claim; it is useful independent feedback, recorded with the exact commit and evidence.

## Invitation

If you are an engineer, try the documented path from a clean checkout, record where the steps are clear or unclear, and report reproducible deviations. A failed reproduction is useful feedback when it is accompanied by the command, environment, commit, and evidence.

## Sources

- `START-HERE.md`
- `docs/VERIFIED-GOLDEN-PATH.md`
- `docs/GOLDEN-PATH.md`
- `docs/REFERENCE-DEMO.md`
- `docs/EXTERNAL-VALIDATION.md`
- `docs/PILOT-READINESS.md`

**Editorial reminder:** Replace bracketed founder prompts with Bayu's own real story and approve the tone before publication.
