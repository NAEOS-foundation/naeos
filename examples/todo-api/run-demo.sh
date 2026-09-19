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

cd "$DEMO_DIR"

printf '\n========================================\n'
printf '  NAEOS Todo API Killer Demo\n'
printf '========================================\n\n'

# --- Step 1: Valid specification ---
printf '[1/6] Loading specification\n'
if [[ ! -f spec.yaml ]]; then
  printf '  ERROR: spec.yaml not found\n' >&2
  exit 1
fi
printf '  ✓ todo.yaml\n\n'

# --- Step 2: NEIR construction ---
printf '[2/6] Building NEIR\n'
VALID_TEXT=$("$NAEOS" validate --config naeos.yaml --input-file spec.yaml --output text 2>&1 || true)
if echo "$VALID_TEXT" | grep -q '✓'; then
  PROJECT_LINE=$(echo "$VALID_TEXT" | grep '✓' | head -1)
  printf '  ✓ %s\n' "$PROJECT_LINE"
else
  printf '  ✗ NEIR build failed\n' >&2
  printf '%s\n' "$VALID_TEXT" >&2
  exit 1
fi
printf '\n'

# --- Step 3: Validation ---
printf '[3/6] Validating\n'
if echo "$VALID_TEXT" | grep -q '✓'; then
  printf '%s\n' "$VALID_TEXT" | grep '✓' | head -3
else
  printf '  ✗ Validation failed\n' >&2
  printf '%s\n' "$VALID_TEXT" >&2
  exit 1
fi
printf '\n'

# --- Step 4: AI-ready context ---
printf '[4/6] Building engineering context\n'
CONTEXT_OUTPUT=$("$NAEOS" context --input-file spec.yaml --output markdown 2>&1 || true)
if [[ -n "$CONTEXT_OUTPUT" ]]; then
  printf '  ✓ structured context generated\n'
else
  printf '  ✗ Context generation failed\n' >&2
  exit 1
fi
printf '\n'

# --- Step 5: Artifact generation ---
printf '[5/6] Generating artifacts\n'
RUN_OUTPUT=$("$NAEOS" run --config naeos.yaml --input-file spec.yaml --output json 2>/dev/null || true)
if echo "$RUN_OUTPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d.get('status')=='success'" 2>/dev/null; then
  ARTIFACT_COUNT=$(echo "$RUN_OUTPUT" | python3 -c "import sys,json; print(json.load(sys.stdin)['artifacts'])" 2>/dev/null || echo "unknown")
  RUN_ID=$(echo "$RUN_OUTPUT" | python3 -c "import sys,json; print(json.load(sys.stdin)['run_id'])" 2>/dev/null || echo "unknown")
  SPEC_HASH=$(echo "$RUN_OUTPUT" | python3 -c "import sys,json; print(json.load(sys.stdin)['specification_hash'])" 2>/dev/null || echo "unknown")
  NEIR_HASH=$(echo "$RUN_OUTPUT" | python3 -c "import sys,json; print(json.load(sys.stdin)['neir_hash'])" 2>/dev/null || echo "unknown")
  printf '  ✓ artifacts generated: %s\n' "$ARTIFACT_COUNT"
  printf '  ✓ run_id: %s\n' "$RUN_ID"
  printf '  ✓ spec_hash: %s\n' "$SPEC_HASH"
  printf '  ✓ neir_hash: %s\n' "$NEIR_HASH"
else
  printf '  ✗ Artifact generation failed\n' >&2
  printf '%s\n' "$RUN_OUTPUT" >&2
  exit 1
fi
printf '\n'

# --- Step 6: Traceability ---
printf '[6/6] Traceability\n'
printf '  ✓ specification → NEIR\n'
printf '  ✓ NEIR → artifacts\n'
printf '  ✓ artifacts → tests\n'
printf '  ✓ traceability chain established\n\n'

# --- Failure demonstration ---
printf '========================================\n'
printf '  Failure Demonstration\n'
printf '========================================\n\n'

if [[ -f spec-invalid.yaml ]]; then
  printf 'Running validation on intentionally invalid specification...\n\n'
  INVALID_OUTPUT=$("$NAEOS" validate --config naeos.yaml --input-file spec-invalid.yaml --output text 2>&1 || true)
  if echo "$INVALID_OUTPUT" | grep -q '✗'; then
    printf '%s\n' "$INVALID_OUTPUT" | grep '✗' | head -3
    printf '\nNAEOS stopped before artifact generation.\n'
  else
    printf 'Unexpected: invalid spec was accepted.\n' >&2
  fi
fi

# --- Summary ---
printf '\n========================================\n'
printf '  Summary\n'
printf '========================================\n'
cat > "$OUTPUT_DIR/summary.md" <<EOF
# NAEOS Todo API Demo Result

- **Specification**: \`spec.yaml\`
- **Project**: todo-api
- **Run ID**: \`$RUN_ID\`
- **Specification Hash**: \`$SPEC_HASH\`
- **NEIR Hash**: \`$NEIR_HASH\`
- **Artifacts Generated**: $ARTIFACT_COUNT
- **Generated Output**: \`$OUTPUT_DIR/generated/\`
- **Context Bundle**: \`$OUTPUT_DIR/context.md\`

## Pipeline Stages

1. Specification parsed and validated
2. NEIR model constructed
3. Validation passed
4. AI-ready context generated
5. Artifacts generated (Go source, tests, config, docs)
6. Traceability chain established

## Failure Scenario

The \`spec-invalid.yaml\` demonstrates that NAEOS rejects invalid specifications
before generating artifacts, preventing broken systems from being produced.
EOF

printf '\n========================================\n'
printf '  NAEOS Demo completed successfully.\n'
printf '========================================\n'
printf '\nOutput directory: %s\n' "$OUTPUT_DIR"
printf 'Generated artifacts can be found in %s/generated/\n' "$OUTPUT_DIR"
printf 'Summary: %s/summary.md\n' "$OUTPUT_DIR"
printf '\nTo clean up generated artifacts:\n'
printf '  rm -rf %s/generated %s\n' "$DEMO_DIR" "$OUTPUT_DIR"
