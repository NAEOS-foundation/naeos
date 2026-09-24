#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASE_SHA="${BASE_SHA:-${GITHUB_BASE_SHA:-}}"
REQUEST="${NAEOS_DEPENDENCY_RISK_REQUEST:-}"
OUTPUT="${NAEOS_DEPENDENCY_RISK_OUTPUT:-${ROOT}/dependency-risk-evidence.json}"
cd "${ROOT}"
if [[ -z "${BASE_SHA}" ]]; then echo "BASE_SHA is required; dependency risk gate fails closed."; exit 1; fi
mapfile -t changed < <(git diff --name-only "${BASE_SHA}"...HEAD)
dependency_changed=()
for path in "${changed[@]}"; do
  case "${path}" in
    go.mod|go.sum|*/go.mod|*/go.sum|package.json|package-lock.json|npm-shrinkwrap.json|*/package.json|*/package-lock.json|*/npm-shrinkwrap.json|pnpm-lock.yaml|*/pnpm-lock.yaml|yarn.lock|*/yarn.lock|requirements*.txt|*/requirements*.txt|pyproject.toml|*/pyproject.toml|Cargo.toml|Cargo.lock|*/Cargo.toml|*/Cargo.lock) dependency_changed+=("${path}") ;;
  esac
done
if ((${#dependency_changed[@]} == 0)); then
  mkdir -p "$(dirname "${OUTPUT}")"
  cat > "${OUTPUT}" <<EOF
{
  "policy_id": "dependency-risk-governance",
  "policy_version": "1.0.0",
  "decision": "allow",
  "risk": "low",
  "reason": "no dependency manifest changed",
  "changed_dependency_manifests": []
}
EOF
  echo "No dependency manifest changed; dependency risk gate passed."; exit 0
fi
echo "Dependency manifests changed:"; printf " - %s\n" "${dependency_changed[@]}"
if [[ -z "${REQUEST}" ]]; then echo "NAEOS_DEPENDENCY_RISK_REQUEST is required for dependency changes; failing closed."; exit 1; fi
export NAEOS_DEPENDENCY_RISK_REQUEST="${REQUEST}"
export NAEOS_DEPENDENCY_RISK_OUTPUT="${OUTPUT}"
go run ./cmd/naeos-dependency-risk