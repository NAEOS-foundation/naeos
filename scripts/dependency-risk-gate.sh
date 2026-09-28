#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASE_SHA="${BASE_SHA:-${GITHUB_BASE_SHA:-}}"
REQUEST="${NAEOS_DEPENDENCY_RISK_REQUEST:-}"
OUTPUT="${NAEOS_DEPENDENCY_RISK_OUTPUT:-${ROOT}/dependency-risk-evidence.json}"
cd "${ROOT}"
if [[ -z "${BASE_SHA}" ]]; then
  # workflow_dispatch does not provide github.event.before. For a manual
  # validation run, compare the checked-out commit with its first parent.
  if BASE_SHA="$(git rev-parse HEAD^ 2>/dev/null)"; then
    echo "BASE_SHA was not supplied; using first parent ${BASE_SHA} for manual validation."
  else
    echo "BASE_SHA is required and HEAD has no parent; dependency risk gate fails closed."
    exit 1
  fi
fi
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
export NAEOS_DEPENDENCY_RISK_BASE_SHA="${BASE_SHA}"
export NAEOS_DEPENDENCY_RISK_OUTPUT="${OUTPUT}"
if [[ -n "${REQUEST}" ]]; then
  export NAEOS_DEPENDENCY_RISK_REQUEST="${REQUEST}"
elif printf '%s\n' "${dependency_changed[@]}" | grep -qx 'go.mod'; then
  echo "No manual request supplied; deriving Go dependency risk from BASE_SHA."
else
  echo "Automatic dependency-risk derivation currently supports go.mod only; manual request is required for other ecosystems/manifests."
  exit 1
fi
go run ./cmd/naeos-dependency-risk