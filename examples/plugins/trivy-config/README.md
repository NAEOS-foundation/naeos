# NAEOS Plugin - trivy-config

Runs the external Trivy configuration scanner and returns a structured summary
of misconfiguration findings. This is a native Go plugin because it invokes the
`trivy` executable; it is not a WASM plugin.

## Requirements

- Go 1.25+
- NAEOS plugin runtime
- Trivy installed and available on `PATH`

## Build and test

```bash
go test -race ./examples/plugins/trivy-config
go build -o trivy-config ./examples/plugins/trivy-config
```

## Run

```bash
printf '%s\n' '{"method":"scan","params":{"target":"."}}' | ./trivy-config
```

The plugin runs:

```text
trivy config --format json --quiet <target>
```

For a local or CI artifact without committing the report, run:

```bash
./scripts/trivy-config-pilot.sh
TRIVY_TARGET=Dockerfile TRIVY_OUTPUT="$RUNNER_TEMP/trivy-report.json" \
	./scripts/trivy-config-pilot.sh
```

It returns `finding_count`, `blocking_count`, `severity_counts`, the target, and
the parsed Trivy report. `CRITICAL` and `HIGH` findings make `ok` false;
`MEDIUM`, `LOW`, and other severities remain visible but do not block the
result. The plugin does not store credentials or scan outside the target passed
by the caller.

The intentionally small fixture in `testdata/Dockerfile` can be used for a
local pilot run:

```bash
printf '%s\n' '{"method":"scan","params":{"target":"testdata/Dockerfile"}}' \
	| ./trivy-config > trivy-report.json
```
