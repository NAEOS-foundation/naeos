# Legal Evidence Record Schema v1.0

Required top-level fields:

- `schema_version`
- `release_version`
- `repository`
- `commit_sha`
- `release_date`
- `license`
- `dco_status`
- `sbom`
- `dependency_license_status`
- `attribution_status`
- `security_status`
- `provenance_status`
- `trademark_status`
- `exceptions`
- `reviewer`
- `decision`

Allowed status values:

- `PASS`
- `HOLD`
- `REVIEWED`
- `NOT_APPLICABLE`

The schema intentionally remains small in v1.0. The executable validator is `scripts/validate-legal-evidence.sh` and is enforced by the `Legal Evidence Gate` CI job. The example release record is expected to fail validation because it contains placeholders.
