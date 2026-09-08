#!/usr/bin/env bash
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

printf '\n== 1/3 Validate specification ==\n'
"$NAEOS" validate --input-file spec.yaml --output json

printf '\n== 2/3 Generate AI context bundle ==\n'
"$NAEOS" context --input-file spec.yaml --output markdown --output-file context.md
printf 'Wrote %s\n' "$OUTPUT_DIR/context.md"

printf '\n== 3/3 Run generation pipeline ==\n'
"$NAEOS" run --config naeos.yaml --input-file spec.yaml --output json

GENERATED_DIR="$OUTPUT_DIR/generated"
REQUIRED_FILES=(
  "$OUTPUT_DIR/context.md"
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
- AI context: \`context.md\`
- Generated artifacts: ${ARTIFACT_COUNT}
- Generated output: \`generated/\`

The smoke test verified that the context bundle, Go module, TypeScript package,
and generated project README were created.
EOF

printf '\nDemo complete. Output: %s\n' "$OUTPUT_DIR"
printf 'Verified %s generated artifacts and wrote %s/summary.md\n' "$ARTIFACT_COUNT" "$OUTPUT_DIR"
printf 'Inspect with: find %s -maxdepth 3 -type f | sort\n' "$OUTPUT_DIR"
