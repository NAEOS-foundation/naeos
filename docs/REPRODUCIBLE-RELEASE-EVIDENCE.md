# Reproducible Release Evidence Drill

Status: Phase 4.5 experiment  
Scope: isolated CI fixture; no production release behavior is changed.

## Purpose

The drill verifies that one release source commit can be bound to the evidence required to reconstruct and inspect a release:

`Release SHA → artifact digest → SBOM digest → signature → provenance binding → legal evidence → release metadata`

It also proves that tampering with an artifact or evidence record is detected.

## What the drill verifies

1. A synthetic previous release tag resolves from the release candidate.
2. Release DCO verification passes for the candidate commit.
3. A release artifact has a recorded SHA-256 digest.
4. A CycloneDX SBOM has a recorded SHA-256 digest.
5. An Ed25519 signature fixture verifies against the artifact.
6. The release manifest binds repository, version, source commit, artifact digest, SBOM digest, signature digest, and public-key digest.
7. The provenance fixture binds the artifact digest to the source commit.
8. Existing Legal Evidence generation, validation, and fail-closed gate pass.
9. Legal evidence binds the same source commit, release version, repository, and SBOM digest.
10. Synthetic GitHub Release metadata binds the same source commit and release tag.
11. Mutating the artifact causes manifest verification to fail.
12. Mutating the legal evidence commit causes legal evidence validation to fail.

## Explicit proof boundary

The drill intentionally uses synthetic fixtures for provenance attestation and GitHub Release metadata.

It does **not** claim that the fixture is a GitHub/Sigstore attestation, nor that a real GitHub Release has been reconstructed. Production releases already generate GitHub artifact attestations through `actions/attest`; GitHub documents those attestations as cryptographically signed provenance tied to the repository and commit. The real tagged-release path remains a separate operational verification step.

## Why this matters

A checksum proves artifact integrity. A signature proves possession of the signing key for the signed content. Provenance connects the artifact to its source/build context. Legal evidence records the release-control decision. The release record identifies the distribution boundary.

The reproducibility contract is therefore not just "the hash matches"; it is the ability to follow the same immutable release identity across each evidence class.

## Production follow-up

After this isolated drill is merged, the next production hardening step is to bind the real release manifest and attestation identity into the release evidence bundle, then execute one real tagged release and verify it externally.

