# Untrusted Adapter Deployment Profile

Status: Normative deployment profile for external/untrusted adapters

## Security objective

An adapter that is not trusted application code MUST NOT become a side-effect authority merely because it can submit a valid NAEOS adoption envelope.

The enforced deployment topology is:

```
Untrusted adapter
    |
    | protocol-only intent
    v
Isolation boundary (WASM or separate OS process/container)
    |
    | normalized request
    v
NAEOS Execution Gateway
    |
    +--> policy / authorization / replay
    |
    +--> NAEOS-controlled runtime
          |
          +--> credentials
          +--> filesystem write authority
          +--> network side effects
          +--> external receipts
```

The adapter is never the owner of consequential credentials or capability handles.

## Required deployment modes

### Mode A — WASM adapter

Use the NAEOS WASM sandbox when the adapter can be compiled to WASM.

The sandbox MUST retain:

- bounded execution time;
- bounded linear memory;
- no preopened filesystem unless explicitly required;
- no implicit environment/credential access;
- explicit host capability grants.

NAEOS's WASM runtime is the reference implementation. WASM conformance is not itself authorization; the gateway remains authoritative.

### Mode B — Out-of-process adapter

Use a separate OS process/container when the provider adapter cannot run as WASM.

The adapter process MUST:

- run as a non-root identity;
- receive no consequential credentials;
- have no host socket or Docker socket access;
- use a read-only root filesystem where practical;
- expose only the protocol transport required to submit intents;
- have no direct write access to the target repository or deployment system;
- have network egress disabled by default, or explicitly allowlisted only for provider communication;
- communicate with NAEOS through an authenticated, integrity-protected transport;
- never receive NAEOS capability handles or runtime credentials.

The NAEOS runtime process/container owns all consequential credentials and side-effect access.

## Prohibited topology

This topology MUST NOT be used for untrusted adapters:

```
untrusted adapter
      |
      +--> filesystem credentials
      +--> deployment credentials
      +--> repository write access
      +--> unrestricted network
      +--> NAEOS process memory
```

An adapter interface, JSON schema, signature, or successful conformance test does not make this topology safe.

## Operator verification checklist

Before enabling an untrusted adapter, verify:

- [ ] adapter is WASM-sandboxed or a separate OS process/container;
- [ ] adapter runs without consequential credentials;
- [ ] adapter cannot write the target repository directly;
- [ ] adapter cannot invoke deployment APIs directly;
- [ ] adapter cannot access the NAEOS process memory;
- [ ] adapter has no Docker/container runtime socket;
- [ ] adapter filesystem is read-only except for an explicitly scoped temporary area;
- [ ] network egress is disabled or explicitly allowlisted;
- [ ] NAEOS gateway remains the only consequential execution path;
- [ ] replay protection uses a durable store for restart/shared-replica deployments;
- [ ] external side effects have independent observation/evidence where required.

## Verification evidence

Record the deployment facts with the external adoption evidence:

- isolation mode: `wasm` or `process`;
- image/module digest;
- runtime identity/UID;
- filesystem policy;
- network policy;
- credential bindings;
- capability grants;
- NAEOS commit;
- conformance test result.

The deployment profile closes the architectural ambiguity: **adapter protocol conformance is separate from adapter containment, and containment is enforced by the deployment boundary.**

## Claim discipline

This profile does not claim that every arbitrary host deployment is automatically isolated. It defines the required posture for an operator to claim an adapter is treated as untrusted.

> Untrusted adapter code submits intent; NAEOS-owned runtime owns authority and side effects.
