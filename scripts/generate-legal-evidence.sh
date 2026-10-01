#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

usage() {
  echo "usage: $0 --sbom FILE --output FILE" >&2
  exit 2
}

SBOM=""
OUTPUT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --sbom)
      [[ $# -ge 2 ]] || usage
      SBOM="$2"
      shift 2
      ;;
    --output)
      [[ $# -ge 2 ]] || usage
      OUTPUT="$2"
      shift 2
      ;;
    *)
      usage
      ;;
  esac
done

[[ -n "$SBOM" && -n "$OUTPUT" ]] || usage
[[ -f "$SBOM" ]] || { echo "SBOM not found: $SBOM" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "jq is required" >&2; exit 1; }
command -v sha256sum >/dev/null 2>&1 || { echo "sha256sum is required" >&2; exit 1; }

release_version="${RELEASE_VERSION:-${GITHUB_REF_NAME#v}}"
commit_sha="${GITHUB_SHA:-}"
release_date="${RELEASE_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
repository="${GITHUB_REPOSITORY:-NAEOS-foundation/naeos}"
reviewer="${LEGAL_EVIDENCE_REVIEWER:-github-actions}"

[[ "$release_version" =~ ^[^[:space:]]+$ ]] || { echo "invalid release version" >&2; exit 1; }
[[ "$commit_sha" =~ ^[0-9a-f]{40}$ ]] || { echo "GITHUB_SHA must be a 40-character lowercase SHA-1" >&2; exit 1; }
[[ "$repository" == "NAEOS-foundation/naeos" ]] || { echo "unexpected repository: $repository" >&2; exit 1; }

sha256="$(sha256sum "$SBOM" | awk '{print $1}')"

status_or_hold() {
  local name="$1"
  local value
  value="${!name:-HOLD}"
  case "$value" in
    PASS|HOLD|REVIEWED|NOT_APPLICABLE) printf '%s' "$value" ;;
    *) echo "invalid status for $name: $value" >&2; exit 1 ;;
  esac
}

dco_status="$(status_or_hold DCO_STATUS)"
dependency_license_status="$(status_or_hold DEPENDENCY_LICENSE_STATUS)"
attribution_status="$(status_or_hold ATTRIBUTION_STATUS)"
security_status="$(status_or_hold SECURITY_STATUS)"
provenance_status="$(status_or_hold PROVENANCE_STATUS)"
trademark_status="$(status_or_hold TRADEMARK_STATUS)"

decision="PASS"
for status in "$dco_status" "$dependency_license_status" "$attribution_status" "$security_status" "$provenance_status" "$trademark_status"; do
  if [[ "$status" == "HOLD" ]]; then
    decision="HOLD"
    break
  fi
done

mkdir -p "$(dirname "$OUTPUT")"
jq -n   --arg schema_version "1.0"   --arg release_version "$release_version"   --arg repository "$repository"   --arg commit_sha "$commit_sha"   --arg release_date "$release_date"   --arg license "Apache-2.0"   --arg dco_status "$dco_status"   --arg sbom_artifact "$(basename "$SBOM")"   --arg sbom_sha256 "$sha256"   --arg dependency_license_status "$dependency_license_status"   --arg attribution_status "$attribution_status"   --arg security_status "$security_status"   --arg provenance_status "$provenance_status"   --arg trademark_status "$trademark_status"   --arg reviewer "$reviewer"   --arg decision "$decision"   '{
    schema_version: $schema_version,
    release_version: $release_version,
    repository: $repository,
    commit_sha: $commit_sha,
    release_date: $release_date,
    license: $license,
    dco_status: $dco_status,
    sbom: {artifact: $sbom_artifact, sha256: $sbom_sha256},
    dependency_license_status: $dependency_license_status,
    attribution_status: $attribution_status,
    security_status: $security_status,
    provenance_status: $provenance_status,
    trademark_status: $trademark_status,
    exceptions: [],
    reviewer: $reviewer,
    decision: $decision
  }' > "$OUTPUT"

echo "Generated legal evidence: $OUTPUT"
echo "Decision: $decision"
