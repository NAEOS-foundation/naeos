\set ON_ERROR_STOP on

BEGIN;

-- The test runs after the principal-agent binding migration in an isolated CI database.
SET ROLE naeos_control_plane_admin;
INSERT INTO naeos_control_plane.principal_agent_bindings
  (issuer, principal_kind, principal_id, agent_id)
VALUES
  ('https://issuer.example', 'human', 'test-revoke-principal', 'test-revoke-agent');

-- A valid revocation must not require the caller to provide revoked_at.
UPDATE naeos_control_plane.principal_agent_bindings
SET status = 'revoked', revoked_at = NULL
WHERE issuer = 'https://issuer.example'
  AND principal_kind = 'human'
  AND principal_id = 'test-revoke-principal'
  AND agent_id = 'test-revoke-agent';

DO $assert_server_owned_revocation_timestamp$
DECLARE
  v_revoked_at TIMESTAMPTZ;
BEGIN
  SELECT revoked_at INTO v_revoked_at
  FROM naeos_control_plane.principal_agent_bindings
  WHERE issuer = 'https://issuer.example'
    AND principal_kind = 'human'
    AND principal_id = 'test-revoke-principal'
    AND agent_id = 'test-revoke-agent';

  IF v_revoked_at IS NULL OR v_revoked_at > now() OR v_revoked_at < now() - interval '1 minute' THEN
    RAISE EXCEPTION 'revocation timestamp must be populated by the database';
  END IF;

  BEGIN
    UPDATE naeos_control_plane.principal_agent_bindings
    SET status = 'active'
    WHERE issuer = 'https://issuer.example'
      AND principal_kind = 'human'
      AND principal_id = 'test-revoke-principal'
      AND agent_id = 'test-revoke-agent';
    RAISE EXCEPTION 'expected a revoked binding to be immutable';
  EXCEPTION WHEN raise_exception THEN
    IF SQLERRM = 'expected a revoked binding to be immutable' THEN
      RAISE;
    END IF;
  END;
END
$assert_server_owned_revocation_timestamp$;

ROLLBACK;
