-- Explicit issuer-qualified principal-to-agent bindings for the production Control Plane.
-- Apply only to isolated staging after PR review and an approved migration rehearsal.
-- This migration does not create LOGIN roles or provision credentials.
-- Provision the administrator login out-of-band through the approved secret manager:
--   CREATE ROLE naeos_control_plane_admin_login LOGIN INHERIT PASSWORD '<managed-secret>';
--   GRANT naeos_control_plane_admin TO naeos_control_plane_admin_login;
-- Never use service_role API keys as PostgreSQL login credentials.

DO $roles$
DECLARE
  v_name TEXT;
  v_can_login BOOLEAN;
  v_super BOOLEAN;
  v_createdb BOOLEAN;
  v_createrole BOOLEAN;
  v_bypass BOOLEAN;
  v_inherit BOOLEAN;
BEGIN
  FOREACH v_name IN ARRAY ARRAY['naeos_control_plane_runtime', 'naeos_control_plane_admin'] LOOP
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = v_name) THEN
      EXECUTE format(
        'CREATE ROLE %I NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT NOBYPASSRLS',
        v_name
      );
    END IF;
    SELECT rolcanlogin, rolsuper, rolcreatedb, rolcreaterole, rolbypassrls, rolinherit
      INTO v_can_login, v_super, v_createdb, v_createrole, v_bypass, v_inherit
    FROM pg_catalog.pg_roles WHERE rolname = v_name;
    IF v_can_login OR v_super OR v_createdb OR v_createrole OR v_bypass OR NOT v_inherit THEN
      RAISE EXCEPTION '% must remain NOLOGIN, INHERIT, non-privileged, and NOBYPASSRLS', v_name;
    END IF;
  END LOOP;
END
$roles$;

GRANT USAGE ON SCHEMA naeos_control_plane TO naeos_control_plane_runtime, naeos_control_plane_admin;

CREATE TABLE naeos_control_plane.principal_agent_bindings (
  issuer TEXT NOT NULL CHECK (issuer ~ '^https://[^[:space:]]+$'),
  principal_kind TEXT NOT NULL CHECK (principal_kind IN ('human', 'service')),
  principal_id TEXT NOT NULL CHECK (length(btrim(principal_id)) > 0),
  agent_id TEXT NOT NULL CHECK (length(btrim(agent_id)) > 0),
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  revoked_at TIMESTAMPTZ,
  PRIMARY KEY (issuer, principal_kind, principal_id, agent_id),
  CHECK ((status = 'active' AND revoked_at IS NULL) OR (status = 'revoked' AND revoked_at IS NOT NULL))
);
COMMENT ON TABLE naeos_control_plane.principal_agent_bindings IS
  'Trusted issuer-qualified principal-to-agent authority. Runtime can read only; provisioning is administrative and audited out of band.';
CREATE INDEX principal_agent_bindings_active_lookup_idx
  ON naeos_control_plane.principal_agent_bindings (issuer, principal_kind, principal_id, agent_id)
  WHERE status = 'active';

CREATE OR REPLACE FUNCTION naeos_control_plane.guard_principal_agent_binding_update()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
  IF NEW.issuer IS DISTINCT FROM OLD.issuer
     OR NEW.principal_kind IS DISTINCT FROM OLD.principal_kind
     OR NEW.principal_id IS DISTINCT FROM OLD.principal_id
     OR NEW.agent_id IS DISTINCT FROM OLD.agent_id
     OR NEW.created_at IS DISTINCT FROM OLD.created_at THEN
    RAISE EXCEPTION 'principal-agent binding identity is immutable';
  END IF;
  IF OLD.status <> 'active' OR NEW.status <> 'revoked' THEN
    RAISE EXCEPTION 'principal-agent bindings may only transition once from active to revoked';
  END IF;
  -- The database owns revocation time; callers must not supply it.
  NEW.revoked_at := now();
  NEW.updated_at := now();
  RETURN NEW;
END
$function$;
CREATE TRIGGER principal_agent_binding_update_guard
BEFORE UPDATE ON naeos_control_plane.principal_agent_bindings
FOR EACH ROW EXECUTE FUNCTION naeos_control_plane.guard_principal_agent_binding_update();

ALTER TABLE naeos_control_plane.principal_agent_bindings ENABLE ROW LEVEL SECURITY;
CREATE POLICY principal_agent_bindings_runtime_read
  ON naeos_control_plane.principal_agent_bindings
  FOR SELECT TO naeos_control_plane_runtime USING (TRUE);
CREATE POLICY principal_agent_bindings_admin_all
  ON naeos_control_plane.principal_agent_bindings
  FOR ALL TO naeos_control_plane_admin USING (TRUE) WITH CHECK (TRUE);

REVOKE ALL ON naeos_control_plane.principal_agent_bindings
  FROM PUBLIC, anon, authenticated, service_role;
REVOKE ALL ON FUNCTION naeos_control_plane.guard_principal_agent_binding_update()
  FROM PUBLIC, anon, authenticated, service_role;
GRANT SELECT ON naeos_control_plane.principal_agent_bindings TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.principal_agent_bindings TO naeos_control_plane_admin;
GRANT EXECUTE ON FUNCTION naeos_control_plane.guard_principal_agent_binding_update()
  TO naeos_control_plane_admin;
