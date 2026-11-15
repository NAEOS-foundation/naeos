\set ON_ERROR_STOP on

DO $assert_privileges$
DECLARE
  v_table TEXT;
BEGIN
  IF has_schema_privilege('anon', 'naeos_identity', 'USAGE')
     OR has_schema_privilege('authenticated', 'naeos_identity', 'USAGE')
     OR has_schema_privilege('service_role', 'naeos_identity', 'USAGE') THEN
    RAISE EXCEPTION 'browser roles or Supabase service_role must not have schema USAGE';
  END IF;

  FOREACH v_table IN ARRAY ARRAY[
    'organizations', 'workspaces', 'memberships', 'role_definitions',
    'role_assignments', 'role_capabilities', 'audit_events'
  ] LOOP
    IF has_table_privilege('anon', 'naeos_identity.' || v_table, 'SELECT')
       OR has_table_privilege('authenticated', 'naeos_identity.' || v_table, 'SELECT')
       OR has_table_privilege('service_role', 'naeos_identity.' || v_table, 'SELECT') THEN
      RAISE EXCEPTION 'browser/service_role SELECT unexpectedly granted on %', v_table;
    END IF;

    IF NOT (SELECT relrowsecurity FROM pg_catalog.pg_class
            WHERE oid = ('naeos_identity.' || v_table)::regclass) THEN
      RAISE EXCEPTION 'RLS is not enabled on %', v_table;
    END IF;
  END LOOP;
END
$assert_privileges$;

-- Dedicated connection groups are created only inside this ephemeral CI database.
CREATE ROLE naeos_identity_test_runtime_login LOGIN INHERIT;
GRANT naeos_identity_runtime TO naeos_identity_test_runtime_login;
CREATE ROLE naeos_identity_test_admin_login LOGIN INHERIT;
GRANT naeos_identity_admin TO naeos_identity_test_admin_login;

DO $assert_identity_group_role_attributes$
BEGIN
  IF NOT (SELECT rolinherit AND NOT rolcanlogin AND NOT rolsuper
                 AND NOT rolcreatedb AND NOT rolcreaterole AND NOT rolbypassrls
          FROM pg_catalog.pg_roles WHERE rolname = 'naeos_identity_runtime')
     OR NOT (SELECT rolinherit AND NOT rolcanlogin AND NOT rolsuper
                    AND NOT rolcreatedb AND NOT rolcreaterole AND NOT rolbypassrls
             FROM pg_catalog.pg_roles WHERE rolname = 'naeos_identity_admin') THEN
    RAISE EXCEPTION 'identity group roles must remain INHERIT, NOLOGIN, least-privileged and NOBYPASSRLS';
  END IF;
END
$assert_identity_group_role_attributes$;

DO $assert_role_privilege_split$
BEGIN
  IF NOT has_table_privilege('naeos_identity_test_runtime_login', 'naeos_identity.organizations', 'SELECT')
     OR has_table_privilege('naeos_identity_test_runtime_login', 'naeos_identity.organizations', 'INSERT')
     OR has_table_privilege('naeos_identity_test_runtime_login', 'naeos_identity.memberships', 'UPDATE')
     OR has_table_privilege('naeos_identity_test_runtime_login', 'naeos_identity.role_assignments', 'DELETE') THEN
    RAISE EXCEPTION 'runtime login must have read-only identity catalog privileges';
  END IF;

  IF NOT has_table_privilege('naeos_identity_test_runtime_login', 'naeos_identity.audit_events', 'INSERT')
     OR has_table_privilege('naeos_identity_test_runtime_login', 'naeos_identity.audit_events', 'UPDATE')
     OR has_table_privilege('naeos_identity_test_runtime_login', 'naeos_identity.audit_events', 'DELETE') THEN
    RAISE EXCEPTION 'runtime login audit privileges must be SELECT/INSERT only';
  END IF;

  IF NOT has_table_privilege('naeos_identity_test_admin_login', 'naeos_identity.organizations', 'INSERT')
     OR NOT has_table_privilege('naeos_identity_test_admin_login', 'naeos_identity.memberships', 'UPDATE')
     OR NOT has_table_privilege('naeos_identity_test_admin_login', 'naeos_identity.role_capabilities', 'DELETE')
     OR has_table_privilege('naeos_identity_test_admin_login', 'naeos_identity.audit_events', 'UPDATE') THEN
    RAISE EXCEPTION 'admin login must have controlled identity write privileges but no audit mutation';
  END IF;
END
$assert_role_privilege_split$;

SET ROLE naeos_identity_test_runtime_login;
SELECT count(*) FROM naeos_identity.organizations;

DO $assert_runtime_write_denial$
BEGIN
  BEGIN
    INSERT INTO naeos_identity.organizations (name, slug) VALUES ('Denied', 'denied-org');
    RAISE EXCEPTION 'expected runtime organization INSERT denial';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
END
$assert_runtime_write_denial$;


RESET ROLE;
SET ROLE naeos_identity_test_admin_login;

INSERT INTO naeos_identity.organizations (id, name, slug) VALUES
  ('22222222-2222-4222-8222-222222222222', 'Test Organization A', 'test-organization-a'),
  ('77777777-7777-4777-8777-777777777777', 'Test Organization B', 'test-organization-b');

INSERT INTO naeos_identity.workspaces (id, organization_id, name, slug) VALUES
  ('33333333-3333-4333-8333-333333333333',
   '22222222-2222-4222-8222-222222222222', 'Test Workspace A', 'test-workspace-a');

INSERT INTO naeos_identity.role_definitions (id, role_key, scope, description, is_system_role) VALUES
  ('88888888-8888-4888-8888-888888888888', 'workspace_admin', 'workspace', 'Test workspace role', true),
  ('99999999-9999-4999-8999-999999999999', 'organization_admin', 'organization', 'Test org role', true);

INSERT INTO naeos_identity.role_capabilities (role_definition_id, capability_key) VALUES
  ('88888888-8888-4888-8888-888888888888', 'crm.company.read'),
  ('88888888-8888-4888-8888-888888888888', 'crm.company.write');

INSERT INTO naeos_identity.memberships
  (id, organization_id, workspace_id, issuer, user_id, status)
VALUES
  ('44444444-4444-4444-8444-444444444444',
   '22222222-2222-4222-8222-222222222222',
   '33333333-3333-4333-8333-333333333333',
   'https://issuer.example/auth/v1',
   '11111111-1111-4111-8111-111111111111',
   'active');

INSERT INTO naeos_identity.role_assignments (membership_id, role_definition_id)
VALUES ('44444444-4444-4444-8444-444444444444',
        '88888888-8888-4888-8888-888888888888');

DO $assert_scope_guards$
DECLARE
  v_message TEXT;
BEGIN
  BEGIN
    INSERT INTO naeos_identity.memberships
      (organization_id, workspace_id, issuer, user_id, status)
    VALUES
      ('77777777-7777-4777-8777-777777777777',
       '33333333-3333-4333-8333-333333333333',
       'https://issuer.example/auth/v1',
       'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
       'active');
    RAISE EXCEPTION 'expected organization/workspace composite FK rejection';
  EXCEPTION WHEN foreign_key_violation THEN
    NULL;
  END;

  BEGIN
    INSERT INTO naeos_identity.memberships
      (organization_id, workspace_id, issuer, user_id, status)
    VALUES
      ('22222222-2222-4222-8222-222222222222',
       '33333333-3333-4333-8333-333333333333',
       'https://issuer.example/auth/v1',
       '11111111-1111-4111-8111-111111111111',
       'active');
    RAISE EXCEPTION 'expected duplicate membership rejection';
  EXCEPTION WHEN unique_violation THEN
    NULL;
  END;

  INSERT INTO naeos_identity.memberships
    (organization_id, workspace_id, issuer, user_id, status)
  VALUES
    ('22222222-2222-4222-8222-222222222222',
     NULL,
     'https://issuer.example/auth/v1',
     'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb',
     'active');

  BEGIN
    INSERT INTO naeos_identity.role_assignments (membership_id, role_definition_id)
    SELECT m.id, '88888888-8888-4888-8888-888888888888'
    FROM naeos_identity.memberships AS m
    WHERE m.user_id = 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb';
    RAISE EXCEPTION 'expected role scope guard rejection';
  EXCEPTION WHEN raise_exception THEN
    GET STACKED DIAGNOSTICS v_message = MESSAGE_TEXT;
    IF v_message = 'expected role scope guard rejection' THEN
      RAISE;
    END IF;
    IF v_message <> 'role scope does not match membership scope' THEN
      RAISE EXCEPTION 'unexpected role scope guard message: %', v_message;
    END IF;
  END;
END
$assert_scope_guards$;

INSERT INTO naeos_identity.audit_events
  (issuer, actor_subject, organization_id, workspace_id, action,
   resource_type, resource_id, request_id, outcome)
VALUES
  ('https://issuer.example/auth/v1',
   '11111111-1111-4111-8111-111111111111',
   '22222222-2222-4222-8222-222222222222',
   '33333333-3333-4333-8333-333333333333',
   'tenant.test', 'workspace', '33333333-3333-4333-8333-333333333333',
   'tenant-foundation-ci', 'success');

RESET ROLE;
SET ROLE naeos_identity_test_runtime_login;
SELECT set_config('naeos.organization_id', '22222222-2222-4222-8222-222222222222', false);
SELECT set_config('naeos.workspace_id', '33333333-3333-4333-8333-333333333333', false);
SELECT set_config('naeos.issuer', 'https://issuer.example/auth/v1', false);
SELECT set_config('naeos.subject', '11111111-1111-4111-8111-111111111111', false);

DO $assert_runtime_tenant_scope$
BEGIN
  IF (SELECT count(*) FROM naeos_identity.organizations) <> 1
     OR (SELECT count(*) FROM naeos_identity.workspaces) <> 1
     OR (SELECT count(*) FROM naeos_identity.memberships) <> 1
     OR (SELECT count(*) FROM naeos_identity.role_assignments) <> 1
     OR (SELECT count(*) FROM naeos_identity.role_definitions) <> 1
     OR (SELECT count(*) FROM naeos_identity.role_capabilities) <> 2 THEN
    RAISE EXCEPTION 'runtime tenant scope should expose only the actor assigned tenant/workspace and capabilities';
  END IF;
END
$assert_runtime_tenant_scope$;

-- A caller cannot use a tenant context for another organization to read data
-- without an active membership matching the verified issuer/subject context.
-- This simulates a trusted backend setting the context after token validation.
SELECT set_config('naeos.organization_id', '77777777-7777-4777-8777-777777777777', false);
SELECT set_config('naeos.workspace_id', '', false);
DO $assert_cross_tenant_read_denial$
BEGIN
  IF (SELECT count(*) FROM naeos_identity.organizations) <> 0
     OR (SELECT count(*) FROM naeos_identity.workspaces) <> 0
     OR (SELECT count(*) FROM naeos_identity.memberships) <> 0
     OR (SELECT count(*) FROM naeos_identity.role_assignments) <> 0
     OR (SELECT count(*) FROM naeos_identity.audit_events) <> 0 THEN
    RAISE EXCEPTION 'runtime actor without active membership must not read cross-tenant data';
  END IF;
END
$assert_cross_tenant_read_denial$;

SELECT set_config('naeos.organization_id', '22222222-2222-4222-8222-222222222222', false);
SELECT set_config('naeos.workspace_id', '33333333-3333-4333-8333-333333333333', false);

INSERT INTO naeos_identity.audit_events
  (issuer, actor_subject, organization_id, workspace_id, action,
   resource_type, resource_id, request_id, outcome)
VALUES
  ('https://issuer.example/auth/v1',
   '11111111-1111-4111-8111-111111111111',
   '22222222-2222-4222-8222-222222222222',
   '33333333-3333-4333-8333-333333333333',
   'runtime.audit.write', 'workspace', '33333333-3333-4333-8333-333333333333',
   'tenant-runtime-audit-valid', 'success');

DO $assert_forged_audit_actor_denial$
BEGIN
  BEGIN
    INSERT INTO naeos_identity.audit_events
      (issuer, actor_subject, organization_id, workspace_id, action,
       resource_type, resource_id, request_id, outcome)
    VALUES
      ('https://issuer.example/auth/v1',
       'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
       '22222222-2222-4222-8222-222222222222',
       '33333333-3333-4333-8333-333333333333',
       'runtime.audit.forged', 'workspace', '33333333-3333-4333-8333-333333333333',
       'tenant-runtime-audit-forged', 'success');
    RAISE EXCEPTION 'expected forged audit actor denial';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
END
$assert_forged_audit_actor_denial$;

DO $assert_runtime_audit_mutation_denial$
BEGIN
  BEGIN
    UPDATE naeos_identity.audit_events SET reason = 'tamper'
    WHERE request_id = 'tenant-runtime-audit-valid';
    RAISE EXCEPTION 'expected runtime audit UPDATE denial';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
  BEGIN
    DELETE FROM naeos_identity.audit_events
    WHERE request_id = 'tenant-runtime-audit-valid';
    RAISE EXCEPTION 'expected runtime audit DELETE denial';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
END
$assert_runtime_audit_mutation_denial$;

RESET ROLE;

DO $assert_audit_append_only_trigger$
DECLARE
  v_message TEXT;
BEGIN
  BEGIN
    UPDATE naeos_identity.audit_events SET reason = 'tamper'
    WHERE request_id = 'tenant-foundation-ci';
    RAISE EXCEPTION 'expected audit UPDATE trigger rejection';
  EXCEPTION WHEN raise_exception THEN
    GET STACKED DIAGNOSTICS v_message = MESSAGE_TEXT;
    IF v_message = 'expected audit UPDATE trigger rejection' THEN
      RAISE;
    END IF;
    IF v_message <> 'identity audit events are append-only' THEN
      RAISE EXCEPTION 'unexpected audit UPDATE trigger message: %', v_message;
    END IF;
  END;

  BEGIN
    DELETE FROM naeos_identity.audit_events
    WHERE request_id = 'tenant-foundation-ci';
    RAISE EXCEPTION 'expected audit DELETE trigger rejection';
  EXCEPTION WHEN raise_exception THEN
    GET STACKED DIAGNOSTICS v_message = MESSAGE_TEXT;
    IF v_message = 'expected audit DELETE trigger rejection' THEN
      RAISE;
    END IF;
    IF v_message <> 'identity audit events are append-only' THEN
      RAISE EXCEPTION 'unexpected audit DELETE trigger message: %', v_message;
    END IF;
  END;
END
$assert_audit_append_only_trigger$;

DO $assert_login_memberships$
BEGIN
  IF NOT pg_catalog.pg_has_role('naeos_identity_test_runtime_login', 'naeos_identity_runtime', 'USAGE')
     OR NOT pg_catalog.pg_has_role('naeos_identity_test_admin_login', 'naeos_identity_admin', 'USAGE')
     OR NOT pg_catalog.pg_has_role('naeos_identity_test_admin_login', 'naeos_identity_runtime', 'USAGE') THEN
    RAISE EXCEPTION 'test login role membership grants are not configured as expected';
  END IF;
END
$assert_login_memberships$;

DROP ROLE naeos_identity_test_runtime_login;
DROP ROLE naeos_identity_test_admin_login;
