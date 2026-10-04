# Phase 4.4 — Plugin Capability Boundary

## Status

Implementation baseline for the plugin capability boundary.

The governing rule is:

> **Plugin capability declaration is not plugin authority.**

A plugin may declare that an action requires a capability. The host must explicitly grant that capability before the action can execute.

## Boundary Contract

The plugin execution path is:

```
Plugin action
    |
    v
Declared required capabilities
    |
    v
CapabilityBoundary
    |
    +-- all capabilities explicitly granted --> continue
    |
    +-- any capability missing ----------------> DENY
    |
    v
Sandbox timeout / rate limit
    |
    v
Plugin Execute()
```

The capability boundary is deliberately separate from the sandbox. Sandbox controls execution isolation and resource limits; the capability boundary controls **what privileged access the plugin is authorized to request**.

## Capability Semantics

Capabilities use exact identifiers.

Examples:

- `fs.read`
- `fs.write`
- `network.egress`
- `env.read`
- `secret.read`

Wildcard grants such as `fs.*` are rejected. This prevents a broad namespace grant from silently widening authority.

An action with no declared capabilities is capability-free. An action with one or more required capabilities is denied unless every required capability is explicitly granted to that plugin.

## Non-Escalation

The effective capability set is:

```
effective_capabilities(action)
    ⊆
explicit_plugin_grant
```

The plugin cannot enlarge its authority by:

- changing its requested action;
- adding a capability at execution time;
- relying on a wildcard grant;
- declaring a capability without a corresponding host grant.

The boundary is therefore fail-closed for privileged actions.

## Implementation

- `internal/pluginhost/capability.go` implements exact-match capability authorization.
- `internal/pluginhost/pluginhost.go` adds action-level capability declarations.
- `internal/pluginhost/manager.go` invokes the capability boundary immediately before plugin execution.
- `internal/pluginhost/capability_test.go` verifies granted, denied, and wildcard cases.

Persisted host grants are represented by:

```json
{
  "capability_grants": {
    "example-plugin": ["fs.read"]
  }
}
```

Action requirements are represented by `PluginInfo.ActionCapabilities`.

## Security Invariant

A plugin having code execution inside the plugin sandbox does **not** imply authority to access a protected capability.

The invariant is:

```
technical execution capability ≠ authorized capability
```

This preserves the NAEOS architectural distinction between capability and authority and aligns plugin execution with the Control Plane / Runtime authorization model.

## Acceptance

- [x] Explicit capability grants exist.
- [x] Privileged actions are denied without matching grants.
- [x] Exact capability matching is enforced.
- [x] Wildcard capability grants are rejected.
- [x] Capability checks occur before plugin execution.
- [x] Existing sandbox controls remain separate from authorization.
- [ ] Fresh-checkout reproducible plugin capability experiment recorded.
- [ ] External evaluator evidence recorded.

The final two items remain evidence/verification work and are intentionally not marked complete by this implementation alone.
