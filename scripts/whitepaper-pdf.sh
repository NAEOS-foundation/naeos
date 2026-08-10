#!/usr/bin/env bash
# Generate whitepaper PDF with mermaid diagrams rendered via mermaid-cli.
#
# Usage: scripts/whitepaper-pdf.sh <input.md> <output.pdf> <title>
#
# Requires: pandoc (with xelatex), node, and @mermaid-js/mermaid-cli (mmdc).

set -euo pipefail

INPUT="${1:?usage: whitepaper-pdf.sh <input.md> <output.pdf> <title>}"
OUTPUT="${2:?usage: whitepaper-pdf.sh <input.md> <output.pdf> <title>}"
TITLE="${3:?usage: whitepaper-pdf.sh <input.md> <output.pdf> <title>}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

if ! command -v mmdc >/dev/null 2>&1; then
    echo "error: mermaid-cli (mmdc) not found. Install with: npm install -g @mermaid-js/mermaid-cli" >&2
    exit 1
fi

mkdir -p "$WORKDIR/diagrams"

# Split mermaid code fences from the source into numbered .mmd files.
python3 - "$INPUT" "$WORKDIR" <<'PYEOF'
import re
import sys

source, workdir = sys.argv[1], sys.argv[2]
text = open(source, encoding="utf-8").read()

pattern = re.compile(r"```mermaid\s*\n(.*?)```", re.DOTALL)
index = 0
output = []
last = 0

for match in pattern.finditer(text):
    output.append(text[last : match.start()])
    index += 1
    mmd_path = f"{workdir}/diagrams/{index}.mmd"
    with open(mmd_path, "w", encoding="utf-8") as fh:
        fh.write(match.group(1))
    output.append(f"![mermaid diagram {index}]({workdir}/diagrams/{index}.png)\n")
    last = match.end()

output.append(text[last:])
with open(f"{workdir}/source.md", "w", encoding="utf-8") as fh:
    fh.write("".join(output))

print(f"extracted {index} mermaid diagrams")
PYEOF

# Render each diagram to SVG.
for mmd in "$WORKDIR"/diagrams/*.mmd; do
    name="$(basename "$mmd" .mmd)"
    echo "rendering diagram $name"
    mmdc --input "$mmd" --output "$WORKDIR/diagrams/$name.png" \
        --puppeteerConfigFile "$REPO_ROOT/scripts/puppeteer-config.json"
done

# Produce the PDF.
mkdir -p "$(dirname "$OUTPUT")"
pandoc "$WORKDIR/source.md" \
    --pdf-engine=xelatex \
    --resource-path="$WORKDIR" \
    -V geometry:margin=1in \
    -V title="$TITLE" \
    -V subtitle="Version $(cat "$REPO_ROOT/VERSION")" \
    -V author="NAEOS Foundation" \
    -V date="$(date +%Y-%m-%d)" \
    -o "$OUTPUT"

echo "PDF written to $OUTPUT"
