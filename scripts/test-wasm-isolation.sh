#!/usr/bin/env bash
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0
#
# Validate the NAEOS WASM containment profile from a development/Codespaces
# environment. Host access is deliberately treated as a deployment diagnostic,
# not as evidence of WASM capability exposure.

set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "=== NAEOS WASM ISOLATION HARNESS ==="
echo

echo "[1/4] WASM containment conformance"
go test ./internal/pluginsdk/wasm -run "^TestUntrustedAdapterWASMContainmentProfile$" -count=1 -v

echo
echo "[2/4] Existing filesystem boundary conformance"
go test ./internal/pluginsdk/wasm -run "^TestWASMRuntimeHasNoPreopenedFilesystem$" -count=1 -v

echo
echo "[3/4] Timeout containment conformance"
go test ./internal/pluginsdk/wasm -run "^TestWASMPluginExecuteTimeout$" -count=1 -v

echo
echo "[4/4] Development-environment diagnostics"
if [[ -S /var/run/docker.sock ]]; then
  echo "WARN: Docker socket is visible to the Codespace."
  echo "      This is a deployment-environment finding, not a WASM runtime failure."
else
  echo "PASS: Docker socket is not visible to the Codespace."
fi

credential_vars=()
while IFS="=" read -r name _; do
  case "$name" in
    AWS_ACCESS_KEY_ID|AWS_SECRET_ACCESS_KEY|AWS_SESSION_TOKEN|GITHUB_TOKEN|GH_TOKEN|OPENAI_API_KEY|ANTHROPIC_API_KEY)
      credential_vars+=("$name")
      ;;
  esac
done < <(env)

if ((${#credential_vars[@]})); then
  echo "WARN: credential-like environment variables exist in the shell:"
  printf "      %s\n" "${credential_vars[@]}"
  echo "      The harness does not treat this as WASM capability exposure."
  echo "      Verify the adapter process/container is launched without them."
else
  echo "PASS: no common credential-like environment variables detected."
fi

echo
echo "RESULT: WASM runtime containment checks passed."
echo "NOTE: OS process/container isolation requires deployment-level testing."
