# ADR-008: WASM Plugin Sandbox

> Status: Accepted
> Date: 2026-07-20

## Context

NAEOS supports third-party plugins for custom generators, validators, and deployment adapters. Plugins must run in a sandboxed environment to prevent malicious or buggy code from affecting the host process. Key requirements:

- Memory safety: plugins must not read/write arbitrary host memory
- Resource limits: CPU, memory, and execution time must be bounded
- Host API isolation: plugins must only access capabilities explicitly granted by the host
- Cross-platform: plugins must work on Linux, macOS, and Windows without recompilation
- Performance: plugin overhead should be negligible compared to the pipeline task itself

## Decision

Use **WebAssembly (WASM) via the wazero runtime** as the plugin sandbox mechanism.

Plugins compile to WASM modules and communicate with the host via JSON-over-WASI (standard input/output). The host (`pluginhost.PluginManager`) loads a WASM module, instantiates it with a pre-configured WASI environment, and communicates via JSON messages on stdin/stdout.

Resource limits are enforced via: `ModuleConfig.WithStartTimeout` (execution timeout), `RuntimeConfig.WithMemoryLimitPages` (memory cap), and a custom `wasi.ExitError` handler to prevent host exit.

Each plugin declares its required capabilities in a `plugin.yaml` manifest. The host validates capabilities against a grant policy before instantiation.

## Consequences

### Positive

- Full memory isolation: WASM modules cannot access host memory outside their linear memory
- Platform-neutral: WASM binaries run unmodified on any OS with a wazero-compatible runtime
- Fine-grained capability control: the host explicitly grants filesystem, network, and environment access
- wazero has no CGo dependency and produces tiny binaries, keeping the NAEOS distribution lean
- Hot-reload: plugins can be updated by replacing the WASM binary on disk; the host detects changes via fsnotify and reloads automatically

### Negative

- WASM overhead: JSON serialization/deserialization adds latency per call compared to native Go plugins
- Limited host API: plugins cannot call arbitrary Go functions; all interaction is via the JSON protocol
- WASM debugging tooling is less mature than native debugging (no Delve for WASM, GDB limited)

### Mitigations

- JSON overhead is amortized by batching: plugins process batches of work items per call
- The JSON protocol is designed to be replaced with a binary protocol (e.g., FlatBuffers) if profiling shows it as a bottleneck
- A `naeos plugin debug <name>` subcommand runs the plugin with verbose JSON logging for development
- SHA-256 signature verification (`internal/marketplace/`) ensures plugin integrity before loading
- Plugin execution has a configurable timeout (default 30s) to prevent runaway WASM modules

## Notes

- Implementation: `internal/pluginsdk/wasm/` — WASM runtime, protocol handler, host capability grant
- Plugin host: `internal/pluginhost/` — PluginManager, lifecycle hooks, hot-reload watcher
- Plugin marketplace: `internal/marketplace/` — search, install, verify, uninstall
- Related NES document: NES-053 (WASM Plugin Sandboxed Execution)
