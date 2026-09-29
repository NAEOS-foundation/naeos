# P1.9 — Evidence & Verification Plane

## Purpose

P1.9 turns the existing append-only control-plane ledger into a verifier-facing evidence contract.

The objective is to make one agent action independently reconstructable:

    request → authorization decision → execution outcome → verification → evidence digest

## Evidence contract

EvidenceBundle binds request and decision identity, agent and capability, artifact hash, policy ID/version, grant ID, canonical ledger events, deterministic verification, and a SHA-256 evidence digest.

## Verification

VerifyEvidence checks decision identity, execution identity, ALLOW/EXECUTION_ALLOWED consistency, DENY versus allowed execution consistency, and evidence digest integrity.

Ledger reconstruction also verifies the event hash chain and previous-hash links.

## Acceptance criteria

| Criterion | Evidence |
|---|---|
| Reconstruct one action lifecycle | BuildEvidence(decisionID) |
| Bind policy/version | policy_id + policy_version |
| Bind grant | grant_id |
| Bind artifact | artifact_hash |
| Bind execution | execution_event + execution_id |
| Independent verification | VerifyEvidence(bundle) |
| Tamper detection | digest + ledger hash-chain tests |

## Scope

P1.9 does not claim distributed consensus, external-system transaction semantics, or enterprise compliance. It defines a deterministic evidence contract over the existing in-process control-plane ledger.
