#!/usr/bin/env bash
# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

DEMO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$DEMO_DIR/../.." && pwd)"
OUTPUT_DIR="${NAEOS_DEMO_OUTPUT_DIR:-$DEMO_DIR/.run}"

if [[ -n "${NAEOS_BIN:-}" ]]; then
  NAEOS="$NAEOS_BIN"
elif command -v naeos >/dev/null 2>&1; then
  NAEOS="$(command -v naeos)"
else
  NAEOS="$ROOT_DIR/naeos"
  if [[ ! -x "$NAEOS" ]]; then
    printf 'NAEOS CLI not found. Build it with: go build -o naeos ./cmd/naeos\n' >&2
    exit 1
  fi
fi

mkdir -p "$OUTPUT_DIR"
cp "$DEMO_DIR/spec.yaml" "$OUTPUT_DIR/spec.yaml"
cp "$DEMO_DIR/naeos.yaml" "$OUTPUT_DIR/naeos.yaml"

cd "$OUTPUT_DIR"

printf '\n== [1/8] Specification ==\n'
printf 'Input spec: %s\n' "spec.yaml"
"$NAEOS" validate --input-file spec.yaml --output json > "$OUTPUT_DIR/validate.json"

printf '\n== [2/8] NEIR ==\n'
"$NAEOS" inspect --input-file spec.yaml --output json > "$OUTPUT_DIR/inspect.json"

printf '\n== [3/8] Validation ==\n'
"$NAEOS" validate --input-file spec.yaml --output json

printf '\n== [4/8] Policy ==\n'
cat > "$OUTPUT_DIR/invalid-naeos.yaml" <<'EOF'
pipeline:
  name: demo-app
  mode: development
  verbose: true
  output_dir: ./generated
  language:
    - go
    - typescript
  policies:
    - rule_id: failing-rule
      condition: exists:nonexistent_key
      enabled: true
EOF

if "$NAEOS" run --config invalid-naeos.yaml --input-file spec.yaml --output json > "$OUTPUT_DIR/invalid-policy.log" 2>&1; then
  printf 'Expected invalid policy configuration to be rejected, but the run succeeded.\n' >&2
  cat "$OUTPUT_DIR/invalid-policy.log" >&2
  exit 1
fi

if ! grep -q 'policy evaluation failed' "$OUTPUT_DIR/invalid-policy.log"; then
  printf 'Expected a policy-evaluation failure message, but got:\n' >&2
  cat "$OUTPUT_DIR/invalid-policy.log" >&2
  exit 1
fi

printf 'Rejected invalid policy configuration as expected.\n'

printf '\n== [5/8] AI Context ==\n'
"$NAEOS" context --input-file spec.yaml --output markdown --output-file context.md
"$NAEOS" context --input-file spec.yaml --output json --output-file context.json
printf 'Wrote %s\n' "$OUTPUT_DIR/context.md"

printf '\n== [6/8] AI Compilation ==\n'
if [[ -n "${NAEOS_LLM_API_KEY:-}" ]]; then
  "$NAEOS" ai compile --input-file spec.yaml --target opencode --provider "${NAEOS_LLM_PROVIDER:-openai}" > "$OUTPUT_DIR/ai-compile.txt"
  if ! grep -Eq '.*' "$OUTPUT_DIR/ai-compile.txt"; then
    printf 'AI compilation returned empty output.\n' >&2
    exit 1
  fi
  printf 'AI compile step completed successfully.\n'
else
  printf 'SKIP: AI compilation requires NAEOS_LLM_API_KEY.\n'
fi

printf '\n== [7/8] Artifacts ==\n'
"$NAEOS" run --config naeos.yaml --input-file spec.yaml --output json > "$OUTPUT_DIR/run.json"
python3 - "$OUTPUT_DIR/run.json" <<'PY'
import json, sys
path = sys.argv[1]
with open(path, 'r', encoding='utf-8') as fh:
    data = json.load(fh)
required = ['run_id', 'specification_hash', 'neir_hash', 'validation', 'policy', 'context', 'audit', 'stages']
missing = [key for key in required if key not in data]
if missing:
    raise SystemExit(f'missing required run metadata: {missing}')
validation = data.get('validation', {})
if not isinstance(validation, dict):
    raise SystemExit('validation metadata must be an object')
if not data.get('run_id') or not data.get('specification_hash') or not data.get('neir_hash'):
    raise SystemExit('run_id/specification_hash/neir_hash are required')
print(f"Trace metadata verified: run_id={data['run_id']} specification_hash={data['specification_hash']} neir_hash={data['neir_hash']}")
PY

GENERATED_DIR="$OUTPUT_DIR/generated"
REQUIRED_FILES=(
  "$OUTPUT_DIR/context.md"
  "$OUTPUT_DIR/context.json"
  "$OUTPUT_DIR/run.json"
  "$GENERATED_DIR/README.md"
  "$GENERATED_DIR/go.mod"
  "$GENERATED_DIR/package.json"
)

for required_file in "${REQUIRED_FILES[@]}"; do
  if [[ ! -f "$required_file" ]]; then
    printf 'Demo verification failed: expected file was not generated: %s\n' "$required_file" >&2
    exit 1
  fi
done

ARTIFACT_COUNT="$(find "$GENERATED_DIR" -type f | wc -l | tr -d ' ')"
cat > "$OUTPUT_DIR/summary.md" <<EOF
# NAEOS CLI Demo Result

- Specification: \`spec.yaml\`
- Configuration: \`naeos.yaml\`
- Validation: \`validate.json\`
- NEIR: \`inspect.json\`
- AI context: \`context.md\`
- Run metadata: \`run.json\`
- Generated artifacts: ${ARTIFACT_COUNT}
- Generated output: \`generated/\`

The canonical demo verified the full control-plane flow: specification, NEIR, validation, policy, AI context, and generated artifacts, with traceability metadata captured in run.json.
EOF

printf '\n== [8/8] Evidence ==\n'
printf 'Artifacts generated: %s\n' "$ARTIFACT_COUNT"
printf 'Summary: %s\n' "$OUTPUT_DIR/summary.md"

printf '\nDemo complete. Output: %s\n' "$OUTPUT_DIR"
printf 'Verified %s generated artifacts and wrote %s/summary.md\n' "$ARTIFACT_COUNT" "$OUTPUT_DIR"
printf 'Inspect with: find %s -maxdepth 3 -type f | sort\n' "$OUTPUT_DIR"
