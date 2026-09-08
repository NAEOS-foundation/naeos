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

The script writes an isolated run to `examples/demo-cli/.run/`. Override the
location with `NAEOS_DEMO_OUTPUT_DIR`:

```bash
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh
```

To use another binary:

```bash
NAEOS_BIN=/path/to/naeos ./examples/demo-cli/run-demo.sh
```

The AI compiler is intentionally not part of the default demo because it
requires an LLM API key. After the demo, compile context for a target tool
with:

```bash
naeos ai compile \
  --input-file examples/demo-cli/spec.yaml \
  --target opencode
```
