-- NAEOS shared identity and tenant foundation (private backend-only schema).
-- This migration is a code artifact for isolated staging rehearsal. Do not apply to
-- Production until the documented backup, rollback, security-test, and approval gates pass.
-- Provider users live in the selected CRM Auth project; no FK to auth.users is possible
-- across Supabase projects. Human principals are issuer-qualified externally.
-- Runtime connectivity is provisioned separately through a secret manager:
--   CREATE ROLE naeos_identity_runtime_login LOGIN INHERIT PASSWORD '<managed-secret>';
--   GRANT naeos_identity_runtime TO naeos_identity_runtime_login;
-- Administrative identity writes use a separate trusted backend credential:
--   CREATE ROLE naeos_identity_admin_login LOGIN INHERIT PASSWORD '<managed-secret>';
--   GRANT naeos_identity_admin TO naeos_identity_admin_login;
-- Do not use the Supabase service_role API key as a PostgreSQL login credential and
-- never commit login passwords. The migration itself does not create LOGIN roles.
-- Runtime tenant GUCs (organization_id, workspace_id, issuer, subject) must be set
-- only by a trusted backend after JWT verification; never expose the database login
-- or allow untrusted clients to set these values directly.
DO $role$
DECLARE
  v_can_login BOOLEAN;
  v_is_superuser BOOLEAN;
  v_can_create_db BOOLEAN;
  v_can_create_role BOOLEAN;
  v_bypass_rls BOOLEAN;
  v_can_inherit BOOLEAN;
  v_role_name TEXT;
BEGIN
  FOREACH v_role_name IN ARRAY ARRAY['naeos_identity_runtime', 'naeos_identity_admin'] LOOP
    IF NOT EXISTS (
      SELECT 1 FROM pg_catalog.pg_roles
      WHERE rolname = v_role_name
    ) THEN
      EXECUTE format(
        'CREATE ROLE %I NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT NOBYPASSRLS',
        v_role_name
      );
    END IF;

    SELECT rolcanlogin, rolsuper, rolcreatedb, rolcreaterole, rolbypassrls, rolinherit
      INTO v_can_login, v_is_superuser, v_can_create_db, v_can_create_role, v_bypass_rls, v_can_inherit
    FROM pg_catalog.pg_roles
    WHERE rolname = v_role_name;

    IF v_can_login OR v_is_superuser OR v_can_create_db OR v_can_create_role OR v_bypass_rls OR NOT v_can_inherit THEN
      RAISE EXCEPTION '% must remain NOLOGIN, INHERIT, non-superuser, non-creator, and NOBYPASSRLS', v_role_name;
    END IF;
  END LOOP;
END
$role$;

GRANT naeos_identity_runtime TO naeos_identity_admin;

CREATE SCHEMA IF NOT EXISTS naeos_identity;
REVOKE ALL ON SCHEMA naeos_identity FROM PUBLIC, anon, authenticated, service_role;
GRANT USAGE ON SCHEMA naeos_identity TO naeos_identity_runtime, naeos_identity_admin;

CREATE TABLE IF NOT EXISTS naeos_identity.organizations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL CHECK (length(trim(name)) > 0),
  slug TEXT NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'pending')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS naeos_identity.workspaces (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES naeos_identity.organizations(id) ON DELETE RESTRICT,
  name TEXT NOT NULL CHECK (length(trim(name)) > 0),
  slug TEXT NOT NULL CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'pending')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (organization_id, slug),
  UNIQUE (organization_id, id)
);

-- Membership is scoped either to an organization (workspace_id NULL) or to a
-- specific workspace. NULLS NOT DISTINCT prevents duplicate organization-level links.
CREATE TABLE IF NOT EXISTS naeos_identity.memberships (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES naeos_identity.organizations(id) ON DELETE RESTRICT,
  workspace_id UUID,
  issuer TEXT NOT NULL CHECK (issuer ~ '^https://[^[:space:]]+$'),
  user_id UUID NOT NULL,
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'invited', 'revoked')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (organization_id, workspace_id)
    REFERENCES naeos_identity.workspaces (organization_id, id) ON DELETE RESTRICT,
  UNIQUE NULLS NOT DISTINCT (organization_id, workspace_id, issuer, user_id),
  UNIQUE (id, workspace_id)
);

CREATE INDEX IF NOT EXISTS memberships_issuer_user_idx
  ON naeos_identity.memberships (issuer, user_id);
CREATE INDEX IF NOT EXISTS memberships_org_status_idx
  ON naeos_identity.memberships (organization_id, status);
CREATE INDEX IF NOT EXISTS memberships_workspace_status_idx
  ON naeos_identity.memberships (workspace_id, status);

CREATE TABLE IF NOT EXISTS naeos_identity.role_definitions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  role_key TEXT NOT NULL CHECK (role_key ~ '^[a-z][a-z0-9_]{1,63}$'),
  scope TEXT NOT NULL CHECK (scope IN ('organization', 'workspace')),
  description TEXT NOT NULL DEFAULT '',
  is_system_role BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (scope, role_key),
  UNIQUE (id, scope)
);

CREATE TABLE IF NOT EXISTS naeos_identity.role_capabilities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  role_definition_id UUID NOT NULL REFERENCES naeos_identity.role_definitions(id) ON DELETE CASCADE,
  capability_key TEXT NOT NULL CHECK (capability_key ~ '^[a-z][a-z0-9_.:-]{1,127}$'),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (role_definition_id, capability_key)
);

CREATE INDEX IF NOT EXISTS role_capabilities_key_idx
  ON naeos_identity.role_capabilities (capability_key);

-- Audit read is an explicit tenant capability, granted only to the built-in
-- organization/workspace administrator roles. Other roles fail closed.
INSERT INTO naeos_identity.role_definitions (role_key, scope, description, is_system_role)
VALUES
  ('organization_admin', 'organization', 'Organization administrator', TRUE),
  ('workspace_admin', 'workspace', 'Workspace administrator', TRUE)
ON CONFLICT (scope, role_key) DO NOTHING;

DO $system_role_guard$
BEGIN
  IF EXISTS (
    SELECT 1 FROM naeos_identity.role_definitions
    WHERE (scope, role_key) IN (('organization', 'organization_admin'), ('workspace', 'workspace_admin'))
      AND NOT is_system_role
  ) THEN
    RAISE EXCEPTION 'reserved NAEOS administrator role keys must remain system roles';
  END IF;
END
$system_role_guard$;

INSERT INTO naeos_identity.role_capabilities (role_definition_id, capability_key)
SELECT id, 'naeos.identity.audit.read'
FROM naeos_identity.role_definitions
WHERE (scope, role_key) IN (('organization', 'organization_admin'), ('workspace', 'workspace_admin'))
ON CONFLICT (role_definition_id, capability_key) DO NOTHING;

-- Tenant boundary identifiers are immutable after creation. Move/relink operations
-- must use explicit reviewed workflows rather than in-place scope rewrites.
CREATE OR REPLACE FUNCTION naeos_identity.reject_workspace_organization_change()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
  IF NEW.organization_id IS DISTINCT FROM OLD.organization_id THEN
    RAISE EXCEPTION 'workspace organization is immutable';
  END IF;
  RETURN NEW;
END
$function$;

DROP TRIGGER IF EXISTS workspace_organization_immutable ON naeos_identity.workspaces;
CREATE TRIGGER workspace_organization_immutable
BEFORE UPDATE OF organization_id ON naeos_identity.workspaces
FOR EACH ROW EXECUTE FUNCTION naeos_identity.reject_workspace_organization_change();

CREATE OR REPLACE FUNCTION naeos_identity.reject_membership_scope_change()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
  IF NEW.organization_id IS DISTINCT FROM OLD.organization_id
     OR NEW.workspace_id IS DISTINCT FROM OLD.workspace_id
     OR NEW.issuer IS DISTINCT FROM OLD.issuer
     OR NEW.user_id IS DISTINCT FROM OLD.user_id THEN
    RAISE EXCEPTION 'membership tenant and principal scope is immutable';
  END IF;
  RETURN NEW;
END
$function$;

DROP TRIGGER IF EXISTS membership_scope_immutable ON naeos_identity.memberships;
CREATE TRIGGER membership_scope_immutable
BEFORE UPDATE OF organization_id, workspace_id, issuer, user_id ON naeos_identity.memberships
FOR EACH ROW EXECUTE FUNCTION naeos_identity.reject_membership_scope_change();

CREATE OR REPLACE FUNCTION naeos_identity.reject_role_scope_change()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
  IF NEW.scope IS DISTINCT FROM OLD.scope THEN
    RAISE EXCEPTION 'role definition scope is immutable';
  END IF;
  RETURN NEW;
END
$function$;

DROP TRIGGER IF EXISTS role_definition_scope_immutable ON naeos_identity.role_definitions;
CREATE TRIGGER role_definition_scope_immutable
BEFORE UPDATE OF scope ON naeos_identity.role_definitions
FOR EACH ROW EXECUTE FUNCTION naeos_identity.reject_role_scope_change();

CREATE TABLE IF NOT EXISTS naeos_identity.role_assignments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  membership_id UUID NOT NULL REFERENCES naeos_identity.memberships(id) ON DELETE CASCADE,
  role_definition_id UUID NOT NULL REFERENCES naeos_identity.role_definitions(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by_issuer TEXT,
  created_by_subject UUID,
  UNIQUE (membership_id, role_definition_id)
);

-- Prevent assigning an organization-scoped role to a workspace membership, or vice versa.
CREATE OR REPLACE FUNCTION naeos_identity.validate_role_assignment_scope()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
DECLARE
  v_membership_workspace UUID;
  v_role_scope TEXT;
BEGIN
  SELECT m.workspace_id, r.scope
    INTO v_membership_workspace, v_role_scope
  FROM naeos_identity.memberships AS m
  JOIN naeos_identity.role_definitions AS r ON r.id = NEW.role_definition_id
  WHERE m.id = NEW.membership_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'membership or role definition does not exist';
  END IF;

  IF (v_role_scope = 'organization' AND v_membership_workspace IS NOT NULL)
     OR (v_role_scope = 'workspace' AND v_membership_workspace IS NULL) THEN
    RAISE EXCEPTION 'role scope does not match membership scope';
  END IF;

  RETURN NEW;
END
$function$;

DROP TRIGGER IF EXISTS role_assignment_scope_guard ON naeos_identity.role_assignments;
CREATE TRIGGER role_assignment_scope_guard
BEFORE INSERT OR UPDATE OF membership_id, role_definition_id
ON naeos_identity.role_assignments
FOR EACH ROW EXECUTE FUNCTION naeos_identity.validate_role_assignment_scope();

CREATE TABLE IF NOT EXISTS naeos_identity.audit_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  issuer TEXT NOT NULL CHECK (issuer ~ '^https://[^[:space:]]+$'),
  actor_subject UUID NOT NULL,
  organization_id UUID REFERENCES naeos_identity.organizations(id) ON DELETE RESTRICT,
  workspace_id UUID,
  action TEXT NOT NULL CHECK (length(trim(action)) > 0),
  resource_type TEXT NOT NULL CHECK (length(trim(resource_type)) > 0),
  resource_id TEXT NOT NULL CHECK (length(trim(resource_id)) > 0),
  request_id TEXT NOT NULL DEFAULT '',
  outcome TEXT NOT NULL CHECK (outcome IN ('success', 'denied', 'failure')),
  reason TEXT NOT NULL DEFAULT '',
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  schema_version TEXT NOT NULL DEFAULT '1',
  FOREIGN KEY (organization_id, workspace_id)
    REFERENCES naeos_identity.workspaces (organization_id, id) ON DELETE RESTRICT,
  CHECK (workspace_id IS NULL OR organization_id IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS identity_audit_actor_time_idx
  ON naeos_identity.audit_events (issuer, actor_subject, occurred_at);
CREATE INDEX IF NOT EXISTS identity_audit_org_time_idx
  ON naeos_identity.audit_events (organization_id, occurred_at);
CREATE INDEX IF NOT EXISTS identity_audit_request_idx
  ON naeos_identity.audit_events (request_id);

CREATE OR REPLACE FUNCTION naeos_identity.reject_audit_event_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
  RAISE EXCEPTION 'identity audit events are append-only';
END
$function$;

DROP TRIGGER IF EXISTS identity_audit_events_immutable ON naeos_identity.audit_events;
CREATE TRIGGER identity_audit_events_immutable
BEFORE UPDATE OR DELETE ON naeos_identity.audit_events
FOR EACH ROW EXECUTE FUNCTION naeos_identity.reject_audit_event_mutation();

ALTER TABLE naeos_identity.organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_identity.workspaces ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_identity.memberships ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_identity.role_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_identity.role_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_identity.role_capabilities ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_identity.audit_events ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS identity_organizations_runtime ON naeos_identity.organizations;
DROP POLICY IF EXISTS identity_organizations_runtime_read ON naeos_identity.organizations;
DROP POLICY IF EXISTS identity_organizations_admin_all ON naeos_identity.organizations;
CREATE POLICY identity_organizations_runtime_read
  ON naeos_identity.organizations FOR SELECT TO naeos_identity_runtime
  USING (
    id = NULLIF(current_setting('naeos.organization_id', true), '')::uuid
    AND status = 'active'
    AND EXISTS (
      SELECT 1 FROM naeos_identity.memberships AS m
      WHERE m.organization_id = organizations.id
        AND m.issuer = NULLIF(current_setting('naeos.issuer', true), '')
        AND m.user_id = NULLIF(current_setting('naeos.subject', true), '')::uuid
        AND m.status = 'active'
        AND (m.workspace_id IS NULL
             OR m.workspace_id = NULLIF(current_setting('naeos.workspace_id', true), '')::uuid)
    )
  );
CREATE POLICY identity_organizations_admin_all
  ON naeos_identity.organizations FOR ALL TO naeos_identity_admin
  USING (TRUE) WITH CHECK (TRUE);

DROP POLICY IF EXISTS identity_workspaces_runtime ON naeos_identity.workspaces;
DROP POLICY IF EXISTS identity_workspaces_runtime_read ON naeos_identity.workspaces;
DROP POLICY IF EXISTS identity_workspaces_admin_all ON naeos_identity.workspaces;
CREATE POLICY identity_workspaces_runtime_read
  ON naeos_identity.workspaces FOR SELECT TO naeos_identity_runtime
  USING (organization_id = NULLIF(current_setting('naeos.organization_id', true), '')::uuid
    AND status = 'active'
    AND EXISTS (
      SELECT 1 FROM naeos_identity.organizations AS o
      WHERE o.id = workspaces.organization_id AND o.status = 'active'
    )
    AND (NULLIF(current_setting('naeos.workspace_id', true), '') IS NULL
         OR id = NULLIF(current_setting('naeos.workspace_id', true), '')::uuid)
    AND EXISTS (
      SELECT 1 FROM naeos_identity.memberships AS m
      WHERE m.organization_id = workspaces.organization_id
        AND m.issuer = NULLIF(current_setting('naeos.issuer', true), '')
        AND m.user_id = NULLIF(current_setting('naeos.subject', true), '')::uuid
        AND m.status = 'active'
        AND (m.workspace_id IS NULL OR m.workspace_id = workspaces.id)
    ));
CREATE POLICY identity_workspaces_admin_all
  ON naeos_identity.workspaces FOR ALL TO naeos_identity_admin
  USING (TRUE) WITH CHECK (TRUE);

DROP POLICY IF EXISTS identity_memberships_runtime ON naeos_identity.memberships;
DROP POLICY IF EXISTS identity_memberships_runtime_read ON naeos_identity.memberships;
DROP POLICY IF EXISTS identity_memberships_admin_all ON naeos_identity.memberships;
CREATE POLICY identity_memberships_runtime_read
  ON naeos_identity.memberships FOR SELECT TO naeos_identity_runtime
  USING (organization_id = NULLIF(current_setting('naeos.organization_id', true), '')::uuid
    AND EXISTS (
      SELECT 1 FROM naeos_identity.organizations AS o
      WHERE o.id = memberships.organization_id AND o.status = 'active'
    )
    AND (workspace_id IS NULL OR EXISTS (
      SELECT 1 FROM naeos_identity.workspaces AS w
      WHERE w.id = memberships.workspace_id
        AND w.organization_id = memberships.organization_id
        AND w.status = 'active'
    ))
    AND issuer = NULLIF(current_setting('naeos.issuer', true), '')
    AND user_id = NULLIF(current_setting('naeos.subject', true), '')::uuid
    AND status = 'active'
    AND (NULLIF(current_setting('naeos.workspace_id', true), '') IS NULL
         OR workspace_id IS NULL
         OR workspace_id = NULLIF(current_setting('naeos.workspace_id', true), '')::uuid));
CREATE POLICY identity_memberships_admin_all
  ON naeos_identity.memberships FOR ALL TO naeos_identity_admin
  USING (TRUE) WITH CHECK (TRUE);

DROP POLICY IF EXISTS identity_role_definitions_runtime ON naeos_identity.role_definitions;
DROP POLICY IF EXISTS identity_role_definitions_runtime_read ON naeos_identity.role_definitions;
DROP POLICY IF EXISTS identity_role_definitions_admin_all ON naeos_identity.role_definitions;
CREATE POLICY identity_role_definitions_runtime_read
  ON naeos_identity.role_definitions FOR SELECT TO naeos_identity_runtime
  USING (EXISTS (
      SELECT 1 FROM naeos_identity.role_assignments AS ra
      JOIN naeos_identity.memberships AS m ON m.id = ra.membership_id
      WHERE ra.role_definition_id = role_definitions.id
        AND m.organization_id = NULLIF(current_setting('naeos.organization_id', true), '')::uuid
        AND m.issuer = NULLIF(current_setting('naeos.issuer', true), '')
        AND m.user_id = NULLIF(current_setting('naeos.subject', true), '')::uuid
        AND m.status = 'active'
        AND EXISTS (
          SELECT 1 FROM naeos_identity.organizations AS o
          WHERE o.id = m.organization_id AND o.status = 'active'
        )
        AND (m.workspace_id IS NULL OR EXISTS (
          SELECT 1 FROM naeos_identity.workspaces AS w
          WHERE w.id = m.workspace_id AND w.organization_id = m.organization_id
            AND w.status = 'active'
        ))
        AND (NULLIF(current_setting('naeos.workspace_id', true), '') IS NULL
             OR m.workspace_id = NULLIF(current_setting('naeos.workspace_id', true), '')::uuid)
    ));
CREATE POLICY identity_role_definitions_admin_all
  ON naeos_identity.role_definitions FOR ALL TO naeos_identity_admin
  USING (TRUE) WITH CHECK (TRUE);

DROP POLICY IF EXISTS identity_role_assignments_runtime ON naeos_identity.role_assignments;
DROP POLICY IF EXISTS identity_role_assignments_runtime_read ON naeos_identity.role_assignments;
DROP POLICY IF EXISTS identity_role_assignments_admin_all ON naeos_identity.role_assignments;
CREATE POLICY identity_role_assignments_runtime_read
  ON naeos_identity.role_assignments FOR SELECT TO naeos_identity_runtime
  USING (EXISTS (
      SELECT 1 FROM naeos_identity.memberships AS m
      WHERE m.id = role_assignments.membership_id
        AND m.organization_id = NULLIF(current_setting('naeos.organization_id', true), '')::uuid
        AND m.issuer = NULLIF(current_setting('naeos.issuer', true), '')
        AND m.user_id = NULLIF(current_setting('naeos.subject', true), '')::uuid
        AND m.status = 'active'
        AND EXISTS (
          SELECT 1 FROM naeos_identity.organizations AS o
          WHERE o.id = m.organization_id AND o.status = 'active'
        )
        AND (m.workspace_id IS NULL OR EXISTS (
          SELECT 1 FROM naeos_identity.workspaces AS w
          WHERE w.id = m.workspace_id AND w.organization_id = m.organization_id
            AND w.status = 'active'
        ))
        AND (NULLIF(current_setting('naeos.workspace_id', true), '') IS NULL
             OR m.workspace_id = NULLIF(current_setting('naeos.workspace_id', true), '')::uuid)
    ));
CREATE POLICY identity_role_assignments_admin_all
  ON naeos_identity.role_assignments FOR ALL TO naeos_identity_admin
  USING (TRUE) WITH CHECK (TRUE);

DROP POLICY IF EXISTS identity_role_capabilities_runtime_read ON naeos_identity.role_capabilities;
CREATE POLICY identity_role_capabilities_runtime_read
  ON naeos_identity.role_capabilities FOR SELECT TO naeos_identity_runtime
  USING (EXISTS (
      SELECT 1 FROM naeos_identity.role_assignments AS ra
      JOIN naeos_identity.memberships AS m ON m.id = ra.membership_id
      WHERE ra.role_definition_id = role_capabilities.role_definition_id
        AND m.organization_id = NULLIF(current_setting('naeos.organization_id', true), '')::uuid
        AND m.issuer = NULLIF(current_setting('naeos.issuer', true), '')
        AND m.user_id = NULLIF(current_setting('naeos.subject', true), '')::uuid
        AND m.status = 'active'
        AND (NULLIF(current_setting('naeos.workspace_id', true), '') IS NULL
             OR m.workspace_id = NULLIF(current_setting('naeos.workspace_id', true), '')::uuid)
    ));
DROP POLICY IF EXISTS identity_role_capabilities_admin_all ON naeos_identity.role_capabilities;
CREATE POLICY identity_role_capabilities_admin_all
  ON naeos_identity.role_capabilities FOR ALL TO naeos_identity_admin
  USING (TRUE) WITH CHECK (TRUE);

DROP POLICY IF EXISTS identity_audit_events_runtime_read ON naeos_identity.audit_events;
-- Audit visibility requires both active tenant membership and the explicit
-- naeos.identity.audit.read capability on a role assigned to that membership.
CREATE POLICY identity_audit_events_runtime_read
  ON naeos_identity.audit_events FOR SELECT TO naeos_identity_runtime
  USING (
    organization_id = NULLIF(current_setting('naeos.organization_id', true), '')::uuid
    AND EXISTS (
      SELECT 1 FROM naeos_identity.organizations AS o
      WHERE o.id = audit_events.organization_id AND o.status = 'active'
    )
    AND (workspace_id IS NULL OR EXISTS (
      SELECT 1 FROM naeos_identity.workspaces AS w
      WHERE w.id = audit_events.workspace_id
        AND w.organization_id = audit_events.organization_id
        AND w.status = 'active'
    ))
    AND (NULLIF(current_setting('naeos.workspace_id', true), '') IS NULL
         OR workspace_id = NULLIF(current_setting('naeos.workspace_id', true), '')::uuid)
    AND EXISTS (
      SELECT 1
      FROM naeos_identity.memberships AS m
      JOIN naeos_identity.role_assignments AS ra ON ra.membership_id = m.id
      JOIN naeos_identity.role_capabilities AS rc ON rc.role_definition_id = ra.role_definition_id
      WHERE m.organization_id = audit_events.organization_id
        AND m.issuer = NULLIF(current_setting('naeos.issuer', true), '')
        AND m.user_id = NULLIF(current_setting('naeos.subject', true), '')::uuid
        AND m.status = 'active'
        AND (
          (m.workspace_id IS NULL AND NULLIF(current_setting('naeos.workspace_id', true), '') IS NULL)
          OR m.workspace_id = NULLIF(current_setting('naeos.workspace_id', true), '')::uuid
        )
        AND rc.capability_key = 'naeos.identity.audit.read'
    )
  );
DROP POLICY IF EXISTS identity_audit_events_runtime_insert ON naeos_identity.audit_events;
CREATE POLICY identity_audit_events_runtime_insert
  ON naeos_identity.audit_events FOR INSERT TO naeos_identity_runtime
  WITH CHECK (
    issuer = NULLIF(current_setting('naeos.issuer', true), '')
    AND actor_subject = NULLIF(current_setting('naeos.subject', true), '')::uuid
    AND organization_id = NULLIF(current_setting('naeos.organization_id', true), '')::uuid
    AND workspace_id IS NOT DISTINCT FROM NULLIF(current_setting('naeos.workspace_id', true), '')::uuid
    AND EXISTS (
      SELECT 1 FROM naeos_identity.memberships AS m
      WHERE m.organization_id = audit_events.organization_id
        AND m.workspace_id IS NOT DISTINCT FROM audit_events.workspace_id
        AND m.issuer = audit_events.issuer
        AND m.user_id = audit_events.actor_subject
        AND m.status = 'active'
        AND EXISTS (
          SELECT 1 FROM naeos_identity.organizations AS o
          WHERE o.id = m.organization_id AND o.status = 'active'
        )
        AND (m.workspace_id IS NULL OR EXISTS (
          SELECT 1 FROM naeos_identity.workspaces AS w
          WHERE w.id = m.workspace_id AND w.organization_id = m.organization_id
            AND w.status = 'active'
        ))
    )
  );

-- Trusted identity administration may append/read audit events independently
-- of an end-user tenant context. UPDATE/DELETE remain unavailable by grants
-- and are additionally rejected by the append-only trigger.
DROP POLICY IF EXISTS identity_audit_events_admin_read ON naeos_identity.audit_events;
CREATE POLICY identity_audit_events_admin_read
  ON naeos_identity.audit_events FOR SELECT TO naeos_identity_admin
  USING (TRUE);
DROP POLICY IF EXISTS identity_audit_events_admin_insert ON naeos_identity.audit_events;
CREATE POLICY identity_audit_events_admin_insert
  ON naeos_identity.audit_events FOR INSERT TO naeos_identity_admin
  WITH CHECK (TRUE);

REVOKE ALL ON ALL TABLES IN SCHEMA naeos_identity FROM PUBLIC, anon, authenticated, service_role;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA naeos_identity FROM PUBLIC, anon, authenticated, service_role;

GRANT SELECT ON naeos_identity.organizations TO naeos_identity_runtime;
GRANT SELECT ON naeos_identity.workspaces TO naeos_identity_runtime;
GRANT SELECT ON naeos_identity.memberships TO naeos_identity_runtime;
GRANT SELECT ON naeos_identity.role_definitions TO naeos_identity_runtime;
GRANT SELECT ON naeos_identity.role_assignments TO naeos_identity_runtime;
GRANT SELECT ON naeos_identity.role_capabilities TO naeos_identity_runtime;
GRANT SELECT, INSERT ON naeos_identity.audit_events TO naeos_identity_runtime;

GRANT SELECT, INSERT, UPDATE, DELETE ON naeos_identity.organizations TO naeos_identity_admin;
GRANT SELECT, INSERT, UPDATE, DELETE ON naeos_identity.workspaces TO naeos_identity_admin;
GRANT SELECT, INSERT, UPDATE, DELETE ON naeos_identity.memberships TO naeos_identity_admin;
GRANT SELECT, INSERT, UPDATE, DELETE ON naeos_identity.role_definitions TO naeos_identity_admin;
GRANT SELECT, INSERT, UPDATE, DELETE ON naeos_identity.role_assignments TO naeos_identity_admin;
GRANT SELECT, INSERT, UPDATE, DELETE ON naeos_identity.role_capabilities TO naeos_identity_admin;
GRANT SELECT, INSERT ON naeos_identity.audit_events TO naeos_identity_admin;

GRANT EXECUTE ON FUNCTION naeos_identity.validate_role_assignment_scope() TO naeos_identity_admin;
GRANT EXECUTE ON FUNCTION naeos_identity.reject_audit_event_mutation() TO naeos_identity_runtime, naeos_identity_admin;

ALTER DEFAULT PRIVILEGES IN SCHEMA naeos_identity
  REVOKE ALL ON TABLES FROM PUBLIC, anon, authenticated, service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA naeos_identity
  REVOKE ALL ON FUNCTIONS FROM PUBLIC, anon, authenticated, service_role;
