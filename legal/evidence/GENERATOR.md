# Release Legal Evidence Generator

The release legal evidence generator creates a machine-readable Legal Evidence Registry record from
the actual release commit, release version, release timestamp, and generated SBOM.

## Design

The generator is intentionally **fail-closed**.

Legal or compliance statuses are supplied explicitly through environment variables:

- `DCO_STATUS`
- `DEPENDENCY_LICENSE_STATUS`
- `ATTRIBUTION_STATUS`
- `SECURITY_STATUS`
- `PROVENANCE_STATUS`
- `TRADEMARK_STATUS`

Each status must be one of:

- `PASS`
- `HOLD`
- `REVIEWED`
- `NOT_APPLICABLE`

If a status is omitted, it becomes `HOLD`. Therefore the generated decision is `HOLD` until the required evidence has been explicitly established.

The generator calculates the SBOM SHA-256 itself and binds the record to `GITHUB_SHA` and `GITHUB_REF_NAME` when run in GitHub Actions.

## Usage

```bash
RELEASE_VERSION=3.6.0 \
GITHUB_SHA="$(git rev-parse HEAD)" \
bash scripts/generate-legal-evidence.sh \
  --sbom naeos_3.6.0.bom.json \
  --output legal-evidence.json
```

For a CI release, the workflow supplies the actual GitHub commit and release tag automatically.

This generator creates evidence metadata; it does not establish legal compliance, replace legal review, or infer a PASS decision from the existence of CI artifacts.
