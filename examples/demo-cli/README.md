# NAEOS CLI Demo

This example follows the marketing demo flow with a runnable local CLI:

1. Validate a specification.
2. Generate an AI context bundle.
3. Run the NAEOS pipeline and generate artifacts.

## Run

From the repository root:

```bash
go build -o naeos ./cmd/naeos
./examples/demo-cli/run-demo.sh
```

For a one-page printable version of these instructions, see
[`PRINT.md`](PRINT.md).

The script writes an isolated run to `examples/demo-cli/.run/`. Override the
location with `NAEOS_DEMO_OUTPUT_DIR`:

```bash
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh
```

To use another binary:

```bash
NAEOS_BIN=/path/to/naeos ./examples/demo-cli/run-demo.sh
```

## What to expect

The demo validates `spec.yaml`, writes an AI context bundle, and generates a
project containing Go and TypeScript output. The script also performs a smoke
test for these files:

```text
.run/
├── context.md
├── summary.md
└── generated/
    ├── README.md
    ├── go.mod
    └── package.json
```

The exact artifact count can change as generators evolve. The demo prints the
count and fails if the expected output files are missing.

The AI compiler is intentionally not part of the default demo because it
requires an LLM API key. After the demo, compile context for a target tool
with:

```bash
naeos ai compile \
  --input-file examples/demo-cli/spec.yaml \
  --target opencode
```

## Troubleshooting

- `NAEOS CLI not found`: build the binary with `go build -o naeos ./cmd/naeos`,
  or set `NAEOS_BIN` to an existing binary.
- Permission denied when running the script: run `chmod +x
  examples/demo-cli/run-demo.sh`.
- To start from a clean output directory, remove only the demo output:
  `rm -rf examples/demo-cli/.run`.
