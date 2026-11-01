# NAEOS Public Plugin Contract Design

Status: Design proposal; steps 1–3 partially implemented (see "Integration status"); not a compatibility promise
Date: 2026-10-25

## Integration status (2026-10-31)

- Step 1 (v1 schema, finite error-code set): implemented in `pkg/pluginapi/v1`
  (`ProtocolVersion`, documented error codes, `ValidErrorCode`). Response
  validation rejects undocumented codes.
- Step 2 (wire DTO integration): implemented for the WASM adapter. The host
  now sends v1 `Request` envelopes and requires a single v1 response per
  invocation; `internal/pluginsdk/scaffold` and `examples/plugins/hello`
  speak the same public DTOs. The generic process sandbox in
  `internal/pluginsdk/sandbox` still exchanges its legacy `{"method","params"}`
  representation with subprocess plugins.
- Step 3 (strict host validation): implemented host-side in
  `internal/pluginsdk/wasm`: one-response stdout rule, request-ID echo,
  protocol-version match, documented error codes, provisional response size
  limit (`MaxResponseBytes`, default still an open decision), and no raw
  untrusted output in host error text. Host-side manifest loading and
  protocol-range selection from `Manifest.ProtocolVersions` are not
  implemented yet.
- Native `.so` loading and the internal `pluginhost.Plugin` interface remain
  internal and trusted-only, as the migration sequence requires.

End-to-end host tests exercise the v1 envelope against a real wasip1 build of
`examples/plugins/hello` (success, error-code mapping, cancellation, timeout).

## Goal and boundary

Define a portable plugin contract for third-party authors without exposing
`internal/pluginhost` or making its Go interface a public compatibility
commitment. The primary external contract should be a versioned JSON protocol
used by WASM plugins. Keep native Go `.so` loading an internal,
platform-specific mechanism.

The current plugin documentation labels the implementation experimental. This
proposal does not change that status: the host and protocol must implement and
test the contract before it can be advertised as supported.

## Evidence from the current implementation

| Area | Current behavior | Contract gap |
|---|---|---|
| Request | `internal/pluginsdk/wasm/wasm.go` sends `{method, params}` as stdin to one WASM invocation. | No protocol version or request ID. |
| Response | Host parses `{ok, result, error, elapsed_ms}` from stdout and replaces `elapsed_ms`. | No stable error code, version echo, or request correlation. |
| Process model | `examples/plugins/hello/main.go` decodes one request, initializes the plugin with an empty `PluginContext`, executes once, and exits. | No host-supplied plugin configuration or capability grant is delivered to WASM. |
| Manifest | `internal/pluginhost/pluginhost.go` defines plugin metadata, actions, and declared capabilities. | Manifest has no protocol compatibility range and is an internal Go type. |
| Authorization | `CapabilityBoundary` uses exact matching and requires explicit host grants for declared privileged actions. `Manager.Execute` checks the configured action capabilities before calling a plugin. | The WASM wire request contains no grant context; authorization remains host-side and must stay there. |
| Timeout | `Manager.Execute` now passes a combined caller/deadline context to context-aware plugins. The WASM adapter uses `WASMPlugin.ExecuteContext`; wazero closes the module when that context ends. The generic sandbox wrapper races a goroutine against the context; it cannot stop a native plugin function that ignores cancellation. | Timeout/cancellation behavior differs by runtime. Native code may continue after the caller receives timeout. |
| Resource limits | WASM runtime sets a memory page limit. Generic sandbox also has call-count and path checks. | These controls are not uniformly implemented for every runtime; avoid a cross-runtime guarantee. |
| Stdout | WASM response is decoded from the captured stdout buffer. | Extra stdout text makes the response invalid; this must be explicit and tested. |

## Proposed v1 envelope

Use JSON objects with these fields. The initial public DTOs are implemented
under `pkg/pluginapi/v1`, but host integration and protocol compatibility
remain incomplete.

```json
{
  "protocol_version": "1.0",
  "request_id": "8c50f92c",
  "method": "execute",
  "action": "ping",
  "params": {}
}
```

Successful response:

```json
{
  "protocol_version": "1.0",
  "request_id": "8c50f92c",
  "ok": true,
  "result": {"status": "ok"}
}
```

Error response:

```json
{
  "protocol_version": "1.0",
  "request_id": "8c50f92c",
  "ok": false,
  "error": {"code": "invalid_params", "message": "name is required"}
}
```

Rules proposed for v1:

- One request and one response per module invocation; both are one JSON object.
- The host generates a request ID and requires the response to echo it.
- Protocol version uses `major.minor`. A host accepts a plugin only when the
  manifest's declared supported range includes a host-supported version.
  Different major versions are incompatible; additive optional fields may be
  added in a minor version.
- The host owns elapsed-time measurement and records it in host-side
  execution metadata, not plugin-controlled protocol fields.
- stdout is reserved for exactly one response object. Diagnostics go to
  stderr and must not include credentials or request secrets.
- Reject malformed JSON, multiple responses, an unexpected response ID,
  unsupported protocol versions, and oversized input/output. Return a
  structured host error without echoing raw untrusted output to end users.
- Error codes should be a documented finite set (for example
  `invalid_request`, `unsupported_version`, `invalid_params`, `denied`,
  `timeout`, `resource_limit`, `plugin_error`, `invalid_response`). The host
  must not trust plugin error text as authorization or policy evidence.

## Version negotiation

The public manifest declares an exact `protocol_versions` list (for example
`["1.0"]`). The host must choose an exact common version and validate it
before execution. A future explicit handshake may advertise host capabilities,
but must not grant privileges. Do not infer compatibility from plugin package
version or from an unversioned `describe` action.

The WASM host adapter, scaffold, and hello example now use the public DTO
package; the generic subprocess sandbox still uses the legacy wire shape, and
host-side manifest negotiation remains to do. Compatibility fixtures and host
integration tests must keep evolving together.

## Capabilities and authorization

- A plugin manifest may declare exact required capability names per action.
- An administrator or caller grants capabilities separately in host policy.
- The host compares exact names and denies missing or unknown privileged
  requirements before invoking the plugin. Wildcards are not supported.
- A plugin declaration is a request, never an authority. Do not transmit
  host grants as trusted claims to the plugin or let a plugin self-authorize.
- Keep any host service API capability-scoped. Never give a plugin ambient
  filesystem, network, process, credential, or event-bus access.
- Define a reviewed capability registry and grant UX before declaring
  capability names stable.

The current `CapabilityBoundary` already applies exact-match checks to
configured action requirements, but the WASM request does not carry a
capability context. That is desirable for authorization authority; any future
host services must enforce grants on every call.

## Timeout, cancellation, and resource limits

Proposed external behavior:

- The host enforces a configured wall-clock deadline and returns `timeout`.
- WASM execution is interrupted by runtime context cancellation and module
  closure. The host must test that execution actually stops and releases
  resources before promising this guarantee.
- Native plugins are trusted in-process code and cannot be safely force-stopped
  by a Go context. Do not advertise native `.so` timeout as a hard execution
  boundary. If a hard deadline is required, isolate the plugin in a process.
- Specify independent input/output size limits, memory limits, and invocation
  quotas, and test each in the runtime where it is enforced.
- Cancellation from a caller must flow to the WASM runtime context rather than
  being replaced with `context.Background()`.

The manager now propagates caller cancellation and its timeout to the WASM
runtime. The generic wrapper can still return while a native function that
ignores context continues in the background. Keep native plugins outside any
hard timeout guarantee; verify WASM module closure and resource release before
publishing the WASM guarantee.

## Migration sequence

1. Finalize the v1 schema, finite error-code set, and resource-size limits.
   The DTO currently uses exact protocol-version strings to avoid range-syntax
   ambiguity.
2. Integrate the public manifest and wire DTOs under `pkg/pluginapi/v1` into
   the host and scaffold; the DTO package exists, but host dispatch still uses
   the legacy unversioned wire representation.
3. Add JSON Schema and strict request/response validation in the host.
4. Caller context and the manager deadline now reach WASM execution through
   `ExecuteContext`; tests cover cancellation. Keep native timeout behavior
   documented as best-effort unless native execution moves to a process.
5. Update scaffold and examples to use only public protocol types. A public
   helper library may encode/decode messages, but must not expose the host
   manager or internal context.
6. Add host-level compatibility fixtures and integration tests for manifest
   version selection, malformed and oversized messages, error mapping, denial,
   timeout, and cancellation. DTO validation and WASM cancellation tests
   already cover the package/runtime boundaries, not host protocol integration.
7. Update plugin docs only after host integration and tests pass. Keep
   native `.so` documented as an internal/trusted extension mechanism.

## Decisions still required

- Which exact semantic-version/range syntax should the protocol use?
- Should a manifest be embedded in the package, shipped as a sidecar, or both?
- Which capabilities are needed by initial external use cases, and what
  host-mediated APIs provide them?
- What input/output, memory, and invocation limits should be defaults?
- Is process isolation required for native plugins, or should native plugins
  remain explicitly trusted-only?
- Who owns protocol compatibility and its release policy?

## Non-goals

- This proposal does not make the current JSON-over-stdio protocol stable.
- It does not promise that WASM or native plugins are safe for arbitrary
  untrusted code.
- It does not grant plugin authors NAEOS trademarks or imply endorsement.
- It does not change plugin code, runtime behavior, or repository manifests.
