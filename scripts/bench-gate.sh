#!/usr/bin/env bash
set -euo pipefail

# Benchmark regression gate.
# Runs the core pipeline benchmarks and compares the per-benchmark median
# ns/op against the stored baseline. Fails (exit 1) if any benchmark
# regresses by more than the configured relative threshold.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BENCH_DIR="${ROOT}/bench"
BASELINE="${BENCH_DIR}/baseline.txt"
OUT="${BENCH_DIR}/current.txt"
BENCH_RE="BenchmarkPipeline(Run|Validate|New)\$"
THRESHOLD="${THRESHOLD:-0.35}"

mkdir -p "${BENCH_DIR}"

if [ ! -f "${BASELINE}" ]; then
  echo "No baseline found at ${BASELINE}. Generating baseline from this run..."
  go test -run='^$' -bench="${BENCH_RE}" -benchtime=200x -count=3 -benchmem \
    ./pkg/pipeline/ | tee "${BASELINE}"
  echo "Baseline written to ${BASELINE}. Commit it to the repository."
  exit 0
fi

echo "Running benchmarks..."
go test -run='^$' -bench="${BENCH_RE}" -benchtime=200x -count=3 -benchmem \
  ./pkg/pipeline/ > "${OUT}"

python3 - "${BASELINE}" "${OUT}" "${THRESHOLD}" <<'PYEOF'
import re
import statistics
import sys

baseline_path, current_path, threshold = sys.argv[1], sys.argv[2], float(sys.argv[3])
ns_re = re.compile(r"^(Benchmark\S+)\s+\d+\s+([\d.]+)\s+ns/op")


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
    print(f"\nRegression detected: delta above threshold ({threshold:.0%})")
    sys.exit(1)
print("\nNo regression detected.")
PYEOF
