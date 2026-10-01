#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

git init -q "$tmp/repo"
cd "$tmp/repo"
git config user.name "NAEOS Release Drill"
git config user.email "release-drill@naeos.dev"
git remote add origin https://github.com/NAEOS-foundation/naeos.git

git commit --allow-empty -q -m "test: previous release" -m "Signed-off-by: NAEOS Release Drill <release-drill@naeos.dev>"
git tag v0.0.1

git commit --allow-empty -q -m "test: release candidate" -m "Signed-off-by: NAEOS Release Drill <release-drill@naeos.dev>"

base="$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || true)"
[[ "$base" == "v0.0.1" ]] || {
  echo "Release drill: failed to resolve previous release tag (got: ${base:-<empty>})" >&2
  exit 1
}

bash "$repo_root/scripts/check-release-dco.sh" "$base" HEAD

cat > "$tmp/repo/sbom.json" <<'EOF'
{"bomFormat":"CycloneDX","specVersion":"1.6","components":[]}
EOF

export GITHUB_REPOSITORY="NAEOS-foundation/naeos"
export GITHUB_SHA="$(git rev-parse HEAD)"
export RELEASE_VERSION="0.0.2-drill"
export RELEASE_DATE="2026-01-01T00:00:00Z"
export LEGAL_EVIDENCE_REVIEWER="release-drill"
export DCO_STATUS="PASS"
export DEPENDENCY_LICENSE_STATUS="PASS"
export ATTRIBUTION_STATUS="PASS"
export SECURITY_STATUS="PASS"
export PROVENANCE_STATUS="PASS"
export TRADEMARK_STATUS="PASS"

bash "$repo_root/scripts/generate-legal-evidence.sh" --sbom "$tmp/repo/sbom.json" --output "$tmp/repo/pass.json"
bash "$repo_root/scripts/validate-legal-evidence.sh" "$tmp/repo/pass.json"
bash "$repo_root/scripts/release-legal-gate.sh" "$tmp/repo/pass.json"

export SECURITY_STATUS="HOLD"
bash "$repo_root/scripts/generate-legal-evidence.sh" --sbom "$tmp/repo/sbom.json" --output "$tmp/repo/hold.json"
bash "$repo_root/scripts/validate-legal-evidence.sh" "$tmp/repo/hold.json"

if bash "$repo_root/scripts/release-legal-gate.sh" "$tmp/repo/hold.json"; then
  echo "Release drill: HOLD fixture unexpectedly passed the release gate" >&2
  exit 1
fi

echo "Release evidence drill: PASS"
