#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

for cmd in git jq sha256sum openssl; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "missing required command: $cmd" >&2; exit 1; }
done

repo="$tmp/repo"
git init -q "$repo"
cd "$repo"
git config user.name "NAEOS Release Reproducibility Drill"
git config user.email "release-drill@naeos.dev"
git remote add origin https://github.com/NAEOS-foundation/naeos.git
git commit --allow-empty -q -m "test: previous reproducible release" -m "Signed-off-by: NAEOS Release Drill <release-drill@naeos.dev>"
git tag v0.0.1-drill
git commit --allow-empty -q -m "test: reproducible release candidate" -m "Signed-off-by: NAEOS Release Drill <release-drill@naeos.dev>"

source_commit="$(git rev-parse HEAD)"
base_tag="$(git describe --tags --abbrev=0 HEAD^)"
[[ "$base_tag" == "v0.0.1-drill" ]] || { echo "previous release tag mismatch" >&2; exit 1; }
bash "$repo_root/scripts/check-release-dco.sh" "$base_tag" "$source_commit"

version="0.0.2-drill"
release_date="2026-01-01T00:00:00Z"
artifact="$repo/naeos-linux-amd64"
sbom="$repo/naeos_$version.bom.json"
manifest="$repo/release-manifest.json"
attestation="$repo/attestation-fixture.json"
signature="$repo/naeos-linux-amd64.sig"
public_key="$repo/release-public.pem"
evidence="$repo/release-legal-evidence.json"
release_record="$repo/github-release-fixture.json"

printf 'NAEOS reproducible release fixture %s\nsource=%s\n' "$version" "$source_commit" > "$artifact"
printf '%s\n' '{"bomFormat":"CycloneDX","specVersion":"1.6","components":[]}' > "$sbom"

openssl genpkey -algorithm ED25519 -out "$tmp/release-key.pem" >/dev/null 2>&1
openssl pkey -in "$tmp/release-key.pem" -pubout -out "$public_key" >/dev/null 2>&1
openssl pkeyutl -sign -rawin -inkey "$tmp/release-key.pem" -in "$artifact" -out "$signature" >/dev/null 2>&1
openssl pkeyutl -verify -rawin -pubin -inkey "$public_key" -in "$artifact" -sigfile "$signature" >/dev/null 2>&1

artifact_sha256="$(sha256sum "$artifact" | awk '{print $1}')"
sbom_sha256="$(sha256sum "$sbom" | awk '{print $1}')"
signature_sha256="$(sha256sum "$signature" | awk '{print $1}')"
public_key_sha256="$(sha256sum "$public_key" | awk '{print $1}')"

jq -n --arg version "$version" --arg repo "NAEOS-foundation/naeos" --arg commit "$source_commit" --arg date "$release_date" --arg artifact "naeos-linux-amd64" --arg artifact_sha "$artifact_sha256" --arg sbom "naeos_$version.bom.json" --arg sbom_sha "$sbom_sha256" --arg signature "naeos-linux-amd64.sig" --arg signature_sha "$signature_sha256" --arg public_key "release-public.pem" --arg public_key_sha "$public_key_sha256" '
{schema_version:"1.0",release_version:$version,repository:$repo,commit_sha:$commit,release_date:$date,
 artifacts:[{path:$artifact,sha256:$artifact_sha}],
 sbom:{path:$sbom,sha256:$sbom_sha},
 signature:{path:$signature,sha256:$signature_sha,public_key:$public_key,public_key_sha256:$public_key_sha}}' > "$manifest"

jq -n --arg repo "NAEOS-foundation/naeos" --arg commit "$source_commit" --arg digest "sha256:$artifact_sha256" '
{fixture:true,note:"Synthetic provenance fixture; not a GitHub/Sigstore attestation.",repository:$repo,commit_sha:$commit,subject:"naeos-linux-amd64",digest:$digest}' > "$attestation"

export GITHUB_REPOSITORY="NAEOS-foundation/naeos"
export GITHUB_SHA="$source_commit"
export RELEASE_VERSION="$version"
export RELEASE_DATE="$release_date"
export LEGAL_EVIDENCE_REVIEWER="release-reproducibility-drill"
export DCO_STATUS=PASS DEPENDENCY_LICENSE_STATUS=PASS ATTRIBUTION_STATUS=PASS SECURITY_STATUS=PASS PROVENANCE_STATUS=PASS TRADEMARK_STATUS=PASS

bash "$repo_root/scripts/generate-legal-evidence.sh" --sbom "$sbom" --output "$evidence"
bash "$repo_root/scripts/validate-legal-evidence.sh" "$evidence"
bash "$repo_root/scripts/release-legal-gate.sh" "$evidence"

jq -n --arg tag "v$version" --arg version "$version" --arg repo "NAEOS-foundation/naeos" --arg commit "$source_commit" '
{fixture:true,note:"Synthetic GitHub Release metadata; not a published GitHub Release.",tag:$tag,version:$version,repository:$repo,commit_sha:$commit}' > "$release_record"

verify_manifest() {
  jq -e --arg commit "$source_commit" --arg version "$version" --arg repo "NAEOS-foundation/naeos" --arg artifact_sha "$artifact_sha256" --arg sbom_sha "$sbom_sha256" --arg signature_sha "$signature_sha256" --arg public_key_sha "$public_key_sha256" '
  (.commit_sha==$commit) and (.release_version==$version) and (.repository==$repo)
  and (.artifacts[0].sha256==$artifact_sha) and (.sbom.sha256==$sbom_sha)
  and (.signature.sha256==$signature_sha) and (.signature.public_key_sha256==$public_key_sha)' "$manifest" >/dev/null
  [[ "$(sha256sum "$artifact" | awk '{print $1}')" == "$artifact_sha256" ]]
  [[ "$(sha256sum "$sbom" | awk '{print $1}')" == "$sbom_sha256" ]]
  [[ "$(sha256sum "$signature" | awk '{print $1}')" == "$signature_sha256" ]]
  [[ "$(sha256sum "$public_key" | awk '{print $1}')" == "$public_key_sha256" ]]
  openssl pkeyutl -verify -rawin -pubin -inkey "$public_key" -in "$artifact" -sigfile "$signature" >/dev/null 2>&1
}

verify_legal_binding() {
  jq -e --arg commit "$source_commit" --arg version "$version" --arg sbom "naeos_$version.bom.json" --arg sbom_sha "$sbom_sha256" '
  (.repository=="NAEOS-foundation/naeos") and (.commit_sha==$commit) and (.release_version==$version)
  and (.sbom.artifact==$sbom) and (.sbom.sha256==$sbom_sha) and (.decision=="PASS")' "$1" >/dev/null
}

jq -e --arg commit "$source_commit" --arg digest "sha256:$artifact_sha256" '
(.fixture==true) and (.repository=="NAEOS-foundation/naeos") and (.commit_sha==$commit) and (.digest==$digest)' "$attestation" >/dev/null

verify_legal_binding "$evidence"

jq -e --arg commit "$source_commit" --arg version "$version" '
(.fixture==true) and (.tag==("v"+$version)) and (.version==$version)
and (.repository=="NAEOS-foundation/naeos") and (.commit_sha==$commit)' "$release_record" >/dev/null

verify_manifest

cp "$artifact" "$tmp/artifact.original"
printf 'tampered\n' >> "$artifact"
if verify_manifest; then
  echo "artifact mutation unexpectedly passed" >&2
  exit 1
fi
mv "$tmp/artifact.original" "$artifact"

jq '.commit_sha = ("0" * 40)' "$evidence" > "$tmp/evidence-tampered.json"
if verify_legal_binding "$tmp/evidence-tampered.json"; then
  echo "evidence binding mutation unexpectedly passed" >&2
  exit 1
fi

echo "Release reproducibility drill: PASS"
echo "Bound source commit: $source_commit"
echo "Artifact SHA256: $artifact_sha256"
echo "SBOM SHA256: $sbom_sha256"
echo "Signature fixture verified: PASS"
echo "Synthetic attestation binding verified: PASS"
echo "Legal evidence binding verified: PASS"
echo "Synthetic GitHub Release binding verified: PASS"
echo "Mutation tests: PASS"
