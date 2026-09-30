# P1.11 — Independent Verifier CLI

## Purpose

P1.11 adds a verifier-facing CLI that validates a serialized control-plane EvidenceBundle without contacting the control plane, evaluating policy, or executing an action.

Trust boundary:

EvidenceBundle JSON -> verifier -> PASS / FAIL

## Command

`naeos evidence verify-bundle --input-file evidence.json`

JSON output:

`naeos evidence verify-bundle --input-file evidence.json --output json`

The command reads one canonical EvidenceBundle, validates decision identity and execution bindings, recomputes the SHA-256 evidence digest, and exits successfully only when verification returns PASS.

## Verification semantics

The verifier uses the canonical `controlplane.VerifyEvidence` contract. It does not re-run policy evaluation. This keeps verification separate from authorization and execution.

A valid executed ALLOW bundle must contain matching execution evidence. A tampered artifact hash, decision identity, execution binding, or evidence digest produces FAIL and a non-zero process exit.

## Scope

P1.11 is a local verifier for exported canonical evidence. It does not provide remote evidence retrieval, distributed trust, signature verification, or enterprise compliance attestation.
