# Control Plane Principal-to-Agent Binding

Status: implementation proposal; migration artifact is not applied.

## Purpose

A verified JWT identifies an issuer-qualified principal. It does not, by itself, prove authority over a NAEOS agent. The production Control Plane resolves the verified tuple `(iss, principal_kind, sub)` through `naeos_control_plane.principal_agent_bindings` and uses only the resulting active `agent_id`.

`client_id` remains a distinct JWT claim. It is not used as a principal ID or agent ID.

## Principal kind

The runtime requires `NAEOS_CONTROLPLANE_PRINCIPAL_KIND` to be explicitly set to `human` or `service`. This is a trusted deployment profile, not a value inferred from JWT claim presence. A deployment must use the value appropriate to its configured issuer/token profile.

## Resolution rules

- Issuer and subject must be present in the verified JWT.
- Binding lookup requires exact issuer, configured principal kind, subject, and active status.
- If the request supplies an agent ID, it only narrows the lookup to that exact active binding.
- If the request omits an agent ID, resolution succeeds only when exactly one active binding exists.
- Zero matches, revoked bindings, and an unbound requested agent return 403.
- Multiple active bindings without an explicit selection return 403; the service never guesses.
- Database or integrity failures fail closed with 503.
- Caller-supplied policy, grant, authorization context, timestamp, and client ID are never authority sources.

## Database controls

The migration creates a private binding table with a composite identity key, RLS, and separate roles:
- `naeos_control_plane_runtime`: SELECT only.
- `naeos_control_plane_admin`: SELECT/INSERT/UPDATE, no DELETE grant.
- Identity tuple columns are immutable.
- An active binding may transition once to revoked; revoked bindings are terminal.
- The migration does not create LOGIN roles or passwords. Provision the administrative login out of band through the approved secret manager.

## Staging rollout gates

1. Review and merge the code/migration PRs in dependency order.
2. Create an isolated Supabase staging branch or use the approved staging project `naeos.db`; do not target production.
3. Review the migration SQL and role/RLS behavior independently.
4. Apply the migration only after explicit staging-application approval.
5. Provision the admin login securely and create only explicitly approved bindings through the administrative workflow.
6. Set `NAEOS_CONTROLPLANE_PRINCIPAL_KIND` in staging deployment configuration only after confirming the token profile.
7. Run preflight, authorization tests, cross-agent denial tests, revoked-binding tests, and readiness checks.
8. Keep production untouched until a separate production cutover decision, backup proof, and explicit authorization.

## Current limitations

This change implements decision-route principal-to-agent resolution only. Route-specific scope enforcement, tenant-aware binding administration/audit workflow, and agent-scoped runs/evidence queries remain separate gates. The migration file is an artifact only; no SQL has been applied to Supabase by this change.
