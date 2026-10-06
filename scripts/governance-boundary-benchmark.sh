#!/usr/bin/env bash
# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

output_dir="${NAEOS_GOVERNANCE_BENCHMARK_DIR:-/tmp/naeos-governance-benchmark}"
rm -rf "$output_dir"
mkdir -p "$output_dir"

commit_sha="$(git rev-parse HEAD)"
go_version="$(go version)"
runner_version="governance-boundary-benchmark-v1"

printf '%s\n'   "benchmark=NAEOS Governance Boundary Benchmark v1"   "runner_version=$runner_version"   "repository=NAEOS-foundation/naeos"   "commit_sha=$commit_sha"   "go_version=$go_version"   "started_at_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"   > "$output_dir/manifest.txt"

go run ./experiments/ai-agent-policy-boundary --json > "$output_dir/benchmark.json"
go run ./experiments/ai-agent-policy-boundary > "$output_dir/benchmark.txt"

printf '%s\n' "completed_at_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$output_dir/manifest.txt"

echo "NAEOS Governance Boundary Benchmark: PASS"
echo "Evidence: $output_dir"
echo "Commit: $commit_sha"
echo "Scenarios: ALLOW, DENY, REQUIRE_APPROVAL, DIRECT_BYPASS"
