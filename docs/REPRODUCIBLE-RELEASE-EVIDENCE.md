# Reproducible Release Evidence

## Purpose

The reproducibility drill validates the release-evidence chain as an isolated, deterministic fixture without publishing a real release.

It verifies that one release commit SHA can be bound to:

1. release version and repository identity;
2. release artifact checksums;
3. SBOM digest;
4. legal evidence registry record;
5. fail-closed release legal gate;
6. a cryptographic signature fixture.

It also proves that post-generation mutations are detected.

## Drill contract

The fixture must pass these checks:

- the release SHA resolves to a Git commit;
- release DCO verification passes for the fixture commit;
- the legal evidence `commit_sha`, `repository`, and `release_version` match the fixture;
- the artifact checksum matches the artifact on disk;
- the SBOM SHA-256 matches the digest recorded in legal evidence;
- the legal evidence validator accepts the registry record;
- the release legal gate accepts a complete `PASS` record;
- an Ed25519 signature over the checksum manifest verifies.

The drill then mutates:

- the release artifact;
- the SBOM;
- the signed checksum manifest.

Each mutation must fail its corresponding verification.

## Scope

This is a **fixture-level reproducibility drill**, not a real tagged-release verification.

The drill does not create a Git tag, publish a GitHub Release, attest production artifacts through GitHub's attestation service, or exercise the production signing secret. Those remain separate release-environment operations.

The signature is an isolated Ed25519 fixture generated during the drill. It proves mutation detection for the cryptographic verification boundary without asserting equivalence to the production NAEOS signing-key workflow.

## Evidence chain

`Release SHA → Artifact Checksums → SBOM Digest → Legal Evidence Registry → Legal Gate → Signature Verification`

A future production release audit can use the same chain against actual release outputs.
