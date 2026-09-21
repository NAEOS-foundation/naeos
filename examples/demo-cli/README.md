# NAEOS CLI Demo

This example is the canonical repository-backed developer flow for the NAEOS control plane:

1. Validate the specification.
2. Materialize the NEIR model.
3. Validate the engineering rules.
4. Evaluate policy.
5. Generate AI context.
6. Optionally compile for a target AI tool.
7. Run generation and produce artifacts.
8. Verify traceability metadata and evidence.

## Run the canonical demo

From the repository root:

```bash
go build -o naeos ./cmd/naeos
./examples/demo-cli/run-demo.sh
```

This script is the single supported demo path for the repository. It writes an isolated run to `examples/demo-cli/.run/` and performs automated verification of the pipeline stages.

Override the output location:

```bash
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh
```

Use a different binary:

```bash
NAEOS_BIN=/path/to/naeos ./examples/demo-cli/run-demo.sh
```

## Canonical workflow

```text
Specification
  ↓
Parse / Normalize / Resolve
  ↓
NEIR
  ↓
Validation
  ↓
Policy Evaluation
  ↓
AI Context
  ↓
AI Compilation (optional when NAEOS_LLM_API_KEY is set)
  ↓
Execution / Generation
  ↓
Artifacts / Evidence
```

The demo validates `spec.yaml`, inspects the resulting NEIR, demonstrates the deterministic policy rejection path, generates an AI context bundle, runs the generation pipeline, and verifies the resulting trace metadata (`run_id`, `specification_hash`, and `neir_hash`).

```text
.run/
├── inspect.json
├── validate.json
├── context.md
├── context.json
├── run.json
├── summary.md
├── generated/
│   ├── README.md
│   ├── go.mod
│   └── package.json
└── ai-compile.txt   # only when NAEOS_LLM_API_KEY is configured
```

The exact artifact count can change as generation evolves. The script fails if expected files are missing or if the trace metadata is incomplete.

### Representative `run.json`

The exact IDs and hashes change on every run. The stable metadata contract includes:

```json
{
  "run_id": "run-...",
  "specification_hash": "sha256:...",
  "neir_hash": "sha256:...",
  "validation": {},
  "policy": {},
  "context": {},
  "audit": {},
  "stages": []
}
```

Use `run.json` when an artifact needs to be traced back to the specification, derived NEIR, validation result, policy evaluation, and generated context that produced it.

## Optional AI compilation

If you want to exercise the target-tool compilation step, provide a key:

```bash
export NAEOS_LLM_API_KEY=your-key
export NAEOS_LLM_PROVIDER=openai
./examples/demo-cli/run-demo.sh
```

Then the script will run:

```bash
naeos ai compile --input-file examples/demo-cli/spec.yaml --target opencode
```

## Troubleshooting

- `NAEOS CLI not found`: run `go build -o naeos ./cmd/naeos` or set `NAEOS_BIN`.
- `Permission denied`: run `chmod +x examples/demo-cli/run-demo.sh`.
- To start fresh: `rm -rf examples/demo-cli/.run` and rerun the script.
- If AI compilation is skipped: set `NAEOS_LLM_API_KEY` to enable that stage.

For a one-page printable version, see [`PRINT.md`](PRINT.md).
