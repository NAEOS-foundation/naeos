# Trusted Dependency Review Registry

Status: implementation in progress. This contract does not approve any dependency by itself.

## Security boundary

The registry is trusted only when loaded from the protected merge-base revision selected by the CI workflow. A record added or modified in the same pull request as a dependency change must not approve that dependency. Callers must record the exact trusted source revision and matching record in generated decision evidence.

Do not use pull-request-controlled request JSON, environment variables, or registry changes as proof of approval. If the trusted source cannot be fetched or verified, the gate must deny.

## Record requirements

Every record identifies an exact ecosystem, package/module name, and version, and provides:
- upstream source URL;
- SPDX license identifier and URL to license evidence;
- checksum and provenance URL;
- purpose and affected architectural domains;
- criticality;
- reviewer identity, review reference, RFC3339 review timestamp, and review expiry;
- policy and schema versions at the registry level.

A record is valid only when every required field is present, URLs use HTTPS, the checksum is well-formed, domains and criticality are recognized, the review timestamp is not in the future, the review has not expired, and the exact version matches the dependency under evaluation. Duplicate exact-version entries are invalid.

## Decision rules

1. Match ecosystem, package name, and version exactly; no wildcard or implicit version range approval.
2. Missing, malformed, stale-schema, policy-mismatched, or version-mismatched records deny.
3. A registry record supplies review evidence, not an automatic allow; architectural domain and criticality policy still determines the required gates.
4. Keep the highest-risk result across all changed dependencies.
5. Emit the trusted registry revision and matched review reference in evidence.
6. Keep dependency-risk governance fail-closed if trusted provenance or verification evidence is unavailable.

## Implementation status

The CLI now reads `governance/dependency-review-registry.json` from the trusted base commit with `git show`, rejects missing/malformed/stale/policy-mismatched registries, requires an exact version match for every changed dependency, rejects reviewer/PR-author identity matches, and records the trusted source revision plus matched review references in evidence. The checked-in registry is intentionally empty: no dependency is approved by default. Go module versions are extracted from `go.mod`; npm package versions must be exact semantic versions in `package.json` at present, so ranges and unsupported lockfile-only changes fail closed rather than being guessed. The registry remains evidence, not an automatic allow, and the existing risk classifier still decides the required gates.

PR #670 remains draft pending CI and independent security review. Do not use it to unblock PR #648 or #652 until CI passes and the trusted-base integration has been independently reviewed.
