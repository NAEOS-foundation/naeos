#!/bin/bash
# Run benchmarks and compare against baseline
set -euo pipefail

BASELINE_FILE=".benchmark-baseline.json"
PKG="./pkg/pipeline/..."

echo "=== Running Benchmarks ==="
go test -bench=. -benchmem -count=5 -run='^$' $PKG 2>&1 | tee /tmp/bench-results.txt

# Extract results and compare (using the Go tool if available)
echo ""
echo "=== Benchmark Summary ==="
grep "Benchmark" /tmp/bench-results.txt | head -20

echo ""
echo "To update baseline: cp /tmp/bench-results.txt .benchmark-baseline.txt"
