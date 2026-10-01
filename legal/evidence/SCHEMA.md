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

The schema intentionally remains small in v1.0. Release automation may later emit a machine-readable JSON Schema once the field semantics stabilize.
