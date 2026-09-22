#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

# Benchmark regression gate.
# Runs the core pipeline benchmarks repeatedly and compares the per-benchmark median
# ns/op against the stored baseline. Fails if any benchmark regresses by more than
# the configured relative threshold, if a benchmark disappears, or if a new
# benchmark appears without an explicit baseline update.
#
# A second sample is taken only after a regression is detected. This avoids
# failing the gate on a transient hosted-runner performance spike while still
# failing persistent regressions.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BENCH_DIR="${ROOT}/bench"
BASELINE="${BENCH_DIR}/baseline.txt"
OUT="${BENCH_DIR}/current.txt"
RECHECK_OUT="${BENCH_DIR}/recheck.txt"
BENCH_RE="BenchmarkPipeline(Run|Validate|New)\\$"
THRESHOLD="${THRESHOLD:-0.35}"
COUNT="${COUNT:-5}"

mkdir -p "${BENCH_DIR}"

run_benchmarks() {
  go test -run='^$' -bench="${BENCH_RE}" -benchtime=200x -count="${COUNT}" -benchmem ./pkg/pipeline/
}

evaluate() {
  python3 - "${BASELINE}" "${1}" "${THRESHOLD}" <<'PYEOF'
import re
import statistics
import sys

baseline_path, current_path, threshold = sys.argv[1], sys.argv[2], float(sys.argv[3])
ns_re = re.compile(r"^(Benchmark\\S+)\\s+\\d+\\s+([\\d.]+)\\s+ns/op")


def medians(path):
    vals = {}
    for line in open(path):
        m = ns_re.match(line)
        if m:
            vals.setdefault(m.group(1), []).append(float(m.group(2)))
    return {k: statistics.median(v) for k, v in vals.items()}


base = medians(baseline_path)
cur = medians(current_path)

print(f"{'Benchmark':<40}{'baseline':>14}{'current':>14}{'delta':>10}  verdict")
failed = False
for name in sorted(set(base) | set(cur)):
    if name not in base:
        print(f"{name:<40}{'-':>14}{cur[name]:>14.2f}{'-':>10}  new benchmark")
        failed = True
        continue
    if name not in cur:
        print(f"{name:<40}{base[name]:>14.2f}{'-':>14}{'-':>10}  missing run")
        failed = True
        continue

    delta = (cur[name] - base[name]) / base[name]
    verdict = "FAIL" if delta > threshold else "ok"
    if delta > threshold:
        failed = True
    print(f"{name:<40}{base[name]:>14.2f}{cur[name]:>14.2f}{delta:>+10.1%}  {verdict}")

if failed:
    print(f"\\nRegression detected: delta above threshold ({threshold:.0%})")
    sys.exit(1)

print("\\nNo regression detected.")
PYEOF
}

echo "Running benchmarks..."
run_benchmarks > "${OUT}"

if evaluate "${OUT}"; then
  exit 0
fi

echo
echo "Initial regression detected; re-running benchmark gate to distinguish a transient runner spike from a persistent regression..."
run_benchmarks > "${RECHECK_OUT}"

if evaluate "${RECHECK_OUT}"; then
  echo
  echo "Initial regression did not reproduce on the recheck; accepting the benchmark gate."
  exit 0
fi

echo
echo "Persistent benchmark regression detected."
exit 1
