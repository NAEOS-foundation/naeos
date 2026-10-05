#!/usr/bin/env bash
set -euo pipefail

output_dir="${NAEOS_FLAGSHIP_OUTPUT_DIR:-/tmp/naeos-flagship-golden-path}"
rm -rf "$output_dir"

go run ./examples/flagship-golden-path

test -f "$output_dir/evidence.json"
test -f "$output_dir/positive-side-effect.json"
test ! -e "$output_dir/stale-side-effect.json"

echo "Flagship Golden Path: PASS"
echo "Evidence: $output_dir/evidence.json"
