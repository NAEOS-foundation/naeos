#!/usr/bin/env bash
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

command -v jq >/dev/null 2>&1 || { echo "jq is required" >&2; exit 1; }
command -v sha256sum >/dev/null 2>&1 || { echo "sha256sum is required" >&2; exit 1; }
command -v openssl >/dev/null 2>&1 || { echo "openssl is required" >&2; exit 1; }

fixture="$tmp/release"
mkdir -p "$fixture"

git init -q "$fixture/repo"
cd "$fixture/repo"
git config user.name "NAEOS Reproducibility Drill"
git config user.email "reproducibility-drill@naeos.dev"
git remote add origin https://github.com/NAEOS-foundation/naeos.git
printf 'NAEOS reproducibility fixture\n' > source.txt
git add source.txt
git commit -q -m "test: reproducible release fixture" -m "Signed-off-by: NAEOS Reproducibility Drill <reproducibility-drill@naeos.dev>"

release_sha="$(git rev-parse HEAD)"
git cat-file -e "$release_sha^{commit}"

printf 'NAEOS release artifact fixture\n' > naeos-fixture.bin
cat > naeos-fixture.bom.json <<'EOF'
{"bomFormat":"CycloneDX","specVersion":"1.6","components":[]}
EOF
sha256sum naeos-fixture.bin > checksums.txt

export GITHUB_REPOSITORY="NAEOS-foundation/naeos"
export GITHUB_SHA="$release_sha"
export RELEASE_VERSION="0.0.3-repro-drill"
export RELEASE_DATE="2026-01-01T00:00:00Z"
export LEGAL_EVIDENCE_REVIEWER="reproducibility-drill"
export DCO_STATUS="PASS"
export DEPENDENCY_LICENSE_STATUS="PASS"
export ATTRIBUTION_STATUS="PASS"
export SECURITY_STATUS="PASS"
export PROVENANCE_STATUS="PASS"
export TRADEMARK_STATUS="PASS"

bash "$repo_root/scripts/check-release-dco.sh" "$release_sha" "$release_sha"

bash "$repo_root/scripts/generate-legal-evidence.sh" \
  --sbom naeos-fixture.bom.json \
  --output release-legal-evidence.json
bash "$repo_root/scripts/validate-legal-evidence.sh" release-legal-evidence.json
bash "$repo_root/scripts/release-legal-gate.sh" release-legal-evidence.json

[[ "$(jq -r '.repository' release-legal-evidence.json)" == "$GITHUB_REPOSITORY" ]]
[[ "$(jq -r '.commit_sha' release-legal-evidence.json)" == "$release_sha" ]]
[[ "$(jq -r '.release_version' release-legal-evidence.json)" == "$RELEASE_VERSION" ]]

verify_sbom_binding() {
  local expected actual
  expected="$(jq -r '.sbom.sha256' release-legal-evidence.json)"
  actual="$(sha256sum naeos-fixture.bom.json | awk '{print $1}')"
  [[ "$expected" == "$actual" ]]
}

verify_sbom_binding

sha256sum -c checksums.txt

openssl genpkey -algorithm ED25519 -out release-key.pem >/dev/null 2>&1
openssl pkey -in release-key.pem -pubout -out release-public.pem >/dev/null 2>&1
openssl pkeyutl -sign -inkey release-key.pem -rawin -in checksums.txt -out checksums.sig >/dev/null 2>&1
openssl pkeyutl -verify -pubin -inkey release-public.pem -rawin -in checksums.txt -sigfile checksums.sig >/dev/null 2>&1

echo "Mutation test: artifact checksum must fail"
printf 'tampered\n' >> naeos-fixture.bin
if sha256sum -c checksums.txt >/dev/null 2>&1; then
  echo "Reproducibility drill: mutated artifact unexpectedly passed checksum verification" >&2
  exit 1
fi
printf 'NAEOS release artifact fixture\n' > naeos-fixture.bin

echo "Mutation test: SBOM binding must fail"
printf 'tampered\n' >> naeos-fixture.bom.json
if verify_sbom_binding; then
  echo "Reproducibility drill: mutated SBOM unexpectedly matched evidence digest" >&2
  exit 1
fi
cat > naeos-fixture.bom.json <<'EOF'
{"bomFormat":"CycloneDX","specVersion":"1.6","components":[]}
EOF

echo "Mutation test: signature must fail"
printf 'tampered\n' >> checksums.txt
if openssl pkeyutl -verify -pubin -inkey release-public.pem -rawin -in checksums.txt -sigfile checksums.sig >/dev/null 2>&1; then
  echo "Reproducibility drill: mutated checksum manifest unexpectedly passed signature verification" >&2
  exit 1
fi

echo "Reproducibility drill: PASS"
echo "Verified: release SHA, artifact checksum, SBOM digest binding, evidence registry, legal gate, and signature fixture."
echo "Verified mutation failures: artifact, SBOM metadata, and signed checksum manifest."
