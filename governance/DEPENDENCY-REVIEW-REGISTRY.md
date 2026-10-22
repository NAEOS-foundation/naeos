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

The schema and validation primitives are being introduced first. The CI integration must next load the registry from the trusted base revision, remove any PR-controlled approval path, and add integration tests for self-approval attempts and mixed-risk dependency changes. Until that integration and its independent review are complete, this work must not be used to unblock PR #648 or #652.
