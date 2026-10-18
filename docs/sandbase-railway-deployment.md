# SandBase Authorization on Railway

This runbook describes a **separate** Railway service for the NAEOS `naeos serve` daemon. Do not repoint the existing `naeos-control-plane` service: it uses `Dockerfile.control-plane` and serves the demo.

## Service setup

1. Create a new Railway service from `NAEOS-foundation/naeos`, branch `main`.
2. Set its Dockerfile path to `Dockerfile.sandbase-authz`.
3. Set the Railway health check path to `/api/health` and expose port `8080` over Railway HTTPS.
4. Add `NAEOS_API_JWT_SECRET` as a strong random secret. This protects the general API; it is separate from the SandBase hook token.
5. Until the policy/grant is approved, leave both SandBase-specific variables unset. The endpoint will not be registered yet.

## Enable the hook only after policy approval

Set these Railway variables:

- `NAEOS_SANDBASE_AUTHZ_CONFIG_JSON`: the approved JSON policy/grant document documented in [the adapter README](../internal/integrations/sandbase/README.md). Keep capability scope least-privilege and use the actual approved agent identity, policy version, grant expiry, and capability list.
- `NAEOS_SANDBASE_AUTHZ_TOKEN`: a generated shared secret. Configure the same value as `MANAGED_AGENTS_EXTERNAL_AUTHZ_TOKEN` on the SandBase side. Never commit or paste it into an issue or log.

The entrypoint writes the policy/grant JSON to a private temporary file (mode restricted by `umask 077`) at startup and exports `NAEOS_SANDBASE_AUTHZ_CONFIG` to the daemon. Both SandBase-specific variables must be present together. The handler rereads the file for each request; changing the JSON in Railway requires a restart/redeploy.

The bearer-token route is `POST /api/v1/integrations/sandbase/authorize`. Keep it on HTTPS and do not disable authentication on the general API. Railway terminates public HTTPS; the daemon listens on the container HTTP port.

## Acceptance gates

Do not call this integration live or close NAEOS Issue #501 until the real SandBase runtime is configured and Evidence Run #001 passes:

1. Authenticated valid request gets the expected decision; missing/wrong bearer token is rejected.
2. Policy/grant configuration is least-privilege and matches the real SandBase agent identity and capability.
3. Make one real governed tool call.
4. Change policy version during the same live session while keeping the grant stale; the next consequential call must be refused before tool execution.
5. Re-authorize with a grant bound to the new policy version and verify expected behavior.
6. Export and independently verify invocation-bound `agent.external_authorization` and tool-result evidence; confirm secrets are absent.

This is deployment scaffolding, not evidence of a successful live pilot.
