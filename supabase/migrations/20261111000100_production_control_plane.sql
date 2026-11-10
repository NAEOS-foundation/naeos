-- NAEOS Production Control Plane private persistence schema.
-- Apply only to an isolated staging project during P4. Never run against production
-- until the P5 cutover gate, backup proof, and explicit operator approval are complete.
-- The runtime connects with a dedicated LOGIN role that is granted membership in
-- naeos_control_plane_runtime. Do not use the Supabase service_role API key as a
-- PostgreSQL login credential. Provision the LOGIN role password through a secret manager.
-- Example (run separately through a secure admin session; never commit the password):
--   CREATE ROLE naeos_control_plane_login LOGIN INHERIT PASSWORD '<managed-secret>';
--   GRANT naeos_control_plane_runtime TO naeos_control_plane_login;

DO $
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'naeos_control_plane_runtime') THEN
    CREATE ROLE naeos_control_plane_runtime
      NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT NOBYPASSRLS;
  END IF;
END
$;

CREATE SCHEMA IF NOT EXISTS naeos_control_plane;
REVOKE ALL ON SCHEMA naeos_control_plane FROM PUBLIC;
GRANT USAGE ON SCHEMA naeos_control_plane TO naeos_control_plane_runtime;

CREATE TABLE naeos_control_plane.policy_versions (
  policy_id TEXT NOT NULL,
  version INTEGER NOT NULL CHECK (version > 0),
  policy JSONB NOT NULL CHECK (jsonb_typeof(policy) = 'object'),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (policy_id, version)
);

COMMENT ON TABLE naeos_control_plane.policy_versions IS
  'Immutable policy snapshots; active_policies is authoritative for active-version selection.';

CREATE TABLE naeos_control_plane.active_policies (
  policy_id TEXT PRIMARY KEY,
  version INTEGER NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT active_policy_version_fk
    FOREIGN KEY (policy_id, version)
    REFERENCES naeos_control_plane.policy_versions (policy_id, version)
);

CREATE TABLE naeos_control_plane.grants (
  grant_id TEXT PRIMARY KEY,
  agent_id TEXT NOT NULL,
  policy_id TEXT NOT NULL,
  policy_version INTEGER NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('active', 'expired', 'revoked', 'inactive')),
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  grant_payload JSONB NOT NULL CHECK (jsonb_typeof(grant_payload) = 'object'),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT grant_policy_version_fk
    FOREIGN KEY (policy_id, policy_version)
    REFERENCES naeos_control_plane.policy_versions (policy_id, version)
);
CREATE INDEX grants_agent_id_idx ON naeos_control_plane.grants (agent_id);
CREATE INDEX grants_policy_version_idx ON naeos_control_plane.grants (policy_id, policy_version);

CREATE TABLE naeos_control_plane.approvals (
  approval_id TEXT PRIMARY KEY,
  decision_id TEXT NOT NULL,
  agent_id TEXT NOT NULL,
  capability TEXT NOT NULL,
  policy_id TEXT NOT NULL,
  policy_version INTEGER NOT NULL,
  approver TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'expired', 'consumed')),
  created_at TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ,
  approved_at TIMESTAMPTZ,
  consumed_at TIMESTAMPTZ,
  artifact_hash TEXT NOT NULL DEFAULT '',
  reason TEXT NOT NULL DEFAULT '',
  CONSTRAINT approval_policy_version_fk
    FOREIGN KEY (policy_id, policy_version)
    REFERENCES naeos_control_plane.policy_versions (policy_id, version)
);
CREATE INDEX approvals_decision_id_idx ON naeos_control_plane.approvals (decision_id);

CREATE TABLE naeos_control_plane.ledger_head (
  singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
  last_sequence BIGINT NOT NULL DEFAULT 0 CHECK (last_sequence >= 0),
  event_hash TEXT NOT NULL DEFAULT ''
);
INSERT INTO naeos_control_plane.ledger_head (singleton, last_sequence, event_hash)
VALUES (TRUE, 0, '')
ON CONFLICT (singleton) DO NOTHING;

CREATE TABLE naeos_control_plane.ledger_events (
  sequence BIGINT PRIMARY KEY CHECK (sequence > 0),
  event_id TEXT NOT NULL UNIQUE,
  occurred_at TIMESTAMPTZ NOT NULL,
  request_id TEXT NOT NULL DEFAULT '',
  decision_id TEXT NOT NULL DEFAULT '',
  execution_id TEXT NOT NULL DEFAULT '',
  agent_id TEXT NOT NULL DEFAULT '',
  event_type TEXT NOT NULL,
  previous_hash TEXT NOT NULL DEFAULT '',
  event_hash TEXT NOT NULL CHECK (length(event_hash) = 64),
  event_payload JSONB NOT NULL CHECK (jsonb_typeof(event_payload) = 'object')
);
CREATE INDEX ledger_events_request_id_idx ON naeos_control_plane.ledger_events (request_id);
CREATE INDEX ledger_events_decision_id_idx ON naeos_control_plane.ledger_events (decision_id);
CREATE INDEX ledger_events_agent_time_idx ON naeos_control_plane.ledger_events (agent_id, occurred_at);

-- Policy snapshots are immutable. Only active_policies may change the selected version.
CREATE OR REPLACE FUNCTION naeos_control_plane.reject_policy_version_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $$
BEGIN
  RAISE EXCEPTION 'policy versions are immutable; insert a new version instead';
END
$$;

CREATE TRIGGER policy_versions_immutable
BEFORE UPDATE OR DELETE ON naeos_control_plane.policy_versions
FOR EACH ROW EXECUTE FUNCTION naeos_control_plane.reject_policy_version_mutation();

-- Ledger events are append-only. Runtime receives SELECT, never direct INSERT/UPDATE/DELETE.
CREATE OR REPLACE FUNCTION naeos_control_plane.reject_ledger_event_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $$
BEGIN
  RAISE EXCEPTION 'control-plane ledger events are append-only';
END
$$;

CREATE TRIGGER ledger_events_immutable
BEFORE UPDATE OR DELETE ON naeos_control_plane.ledger_events
FOR EACH ROW EXECUTE FUNCTION naeos_control_plane.reject_ledger_event_mutation();

-- Serialize append operations across replicas. The application computes the canonical
-- SHA-256 event hash after locking the head; this function verifies predecessor continuity,
-- inserts the immutable event, and advances the head in the same transaction.
CREATE OR REPLACE FUNCTION naeos_control_plane.append_ledger_event(
  p_event JSONB,
  p_previous_hash TEXT,
  p_event_hash TEXT
)
RETURNS BIGINT
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
DECLARE
  v_head naeos_control_plane.ledger_head%ROWTYPE;
  v_sequence BIGINT;
  v_event_id TEXT;
BEGIN
  IF p_event IS NULL OR jsonb_typeof(p_event) <> 'object' THEN
    RAISE EXCEPTION 'ledger event must be a JSON object';
  END IF;
  IF p_event_hash IS NULL OR p_event_hash !~ '^[0-9a-f]{64}
    RAISE EXCEPTION 'event payload hash does not match supplied hash';
  END IF;

  SELECT * INTO STRICT v_head
  FROM naeos_control_plane.ledger_head
  WHERE singleton = TRUE
  FOR UPDATE;

  IF coalesce(p_event->>'previous_hash', '') <> v_head.event_hash
     OR coalesce(p_previous_hash, '') <> v_head.event_hash THEN
    RAISE EXCEPTION 'ledger predecessor mismatch';
  END IF;

  v_event_id := p_event->>'id';
  IF v_event_id IS NULL OR v_event_id = '' THEN
    RAISE EXCEPTION 'ledger event id is required';
  END IF;

  v_sequence := v_head.last_sequence + 1;

  INSERT INTO naeos_control_plane.ledger_events (
    sequence, event_id, occurred_at, request_id, decision_id, execution_id,
    agent_id, event_type, previous_hash, event_hash, event_payload
  ) VALUES (
    v_sequence,
    v_event_id,
    (p_event->>'timestamp')::TIMESTAMPTZ,
    coalesce(p_event->>'request_id', ''),
    coalesce(p_event->>'decision_id', ''),
    coalesce(p_event->>'execution_id', ''),
    coalesce(p_event->>'agent_id', ''),
    coalesce(p_event->>'event_type', ''),
    coalesce(p_event->>'previous_hash', ''),
    p_event_hash,
    p_event
  );

  UPDATE naeos_control_plane.ledger_head
  SET last_sequence = v_sequence, event_hash = p_event_hash
  WHERE singleton = TRUE;

  RETURN v_sequence;
END
$$;

-- Enable RLS as defense in depth. The private schema is not intended for browser access;
-- the dedicated backend service_role is the only runtime role granted access.
ALTER TABLE naeos_control_plane.policy_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.active_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.grants ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.approvals ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.ledger_head ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.ledger_events ENABLE ROW LEVEL SECURITY;

-- The runtime group is NOBYPASSRLS. Explicit policies and table grants jointly
-- constrain it; browser roles and Supabase service_role receive no access.
CREATE POLICY control_plane_policy_versions_runtime
  ON naeos_control_plane.policy_versions
  FOR SELECT TO naeos_control_plane_runtime USING (TRUE);
CREATE POLICY control_plane_policy_versions_insert
  ON naeos_control_plane.policy_versions
  FOR INSERT TO naeos_control_plane_runtime WITH CHECK (TRUE);
CREATE POLICY control_plane_active_policies_runtime
  ON naeos_control_plane.active_policies
  FOR ALL TO naeos_control_plane_runtime USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY control_plane_grants_runtime
  ON naeos_control_plane.grants
  FOR ALL TO naeos_control_plane_runtime USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY control_plane_approvals_runtime
  ON naeos_control_plane.approvals
  FOR ALL TO naeos_control_plane_runtime USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY control_plane_ledger_events_read
  ON naeos_control_plane.ledger_events
  FOR SELECT TO naeos_control_plane_runtime USING (TRUE);

REVOKE ALL ON SCHEMA naeos_control_plane FROM PUBLIC, anon, authenticated, service_role;
REVOKE ALL ON ALL TABLES IN SCHEMA naeos_control_plane FROM PUBLIC, anon, authenticated, service_role;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA naeos_control_plane FROM PUBLIC, anon, authenticated, service_role;

GRANT SELECT, INSERT ON naeos_control_plane.policy_versions TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.active_policies TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.grants TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.approvals TO naeos_control_plane_runtime;
GRANT SELECT ON naeos_control_plane.ledger_events TO naeos_control_plane_runtime;
GRANT EXECUTE ON FUNCTION naeos_control_plane.append_ledger_event(JSONB, TEXT, TEXT) TO naeos_control_plane_runtime;

ALTER DEFAULT PRIVILEGES IN SCHEMA naeos_control_plane
  REVOKE ALL ON TABLES FROM PUBLIC, anon, authenticated, service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA naeos_control_plane
  REVOKE ALL ON FUNCTIONS FROM PUBLIC, anon, authenticated, service_role;
 THEN
    RAISE EXCEPTION 'event hash must be a lowercase 64-character SHA-256 hex digest';
  END IF;
  IF coalesce(p_event->>'event_hash', '') <> p_event_hash THEN
    RAISE EXCEPTION 'event payload hash does not match supplied hash';
  END IF;

  SELECT * INTO STRICT v_head
  FROM naeos_control_plane.ledger_head
  WHERE singleton = TRUE
  FOR UPDATE;

  IF coalesce(p_event->>'previous_hash', '') <> v_head.event_hash
     OR coalesce(p_previous_hash, '') <> v_head.event_hash THEN
    RAISE EXCEPTION 'ledger predecessor mismatch';
  END IF;

  v_event_id := p_event->>'id';
  IF v_event_id IS NULL OR v_event_id = '' THEN
    RAISE EXCEPTION 'ledger event id is required';
  END IF;

  v_sequence := v_head.last_sequence + 1;

  INSERT INTO naeos_control_plane.ledger_events (
    sequence, event_id, occurred_at, request_id, decision_id, execution_id,
    agent_id, event_type, previous_hash, event_hash, event_payload
  ) VALUES (
    v_sequence,
    v_event_id,
    (p_event->>'timestamp')::TIMESTAMPTZ,
    coalesce(p_event->>'request_id', ''),
    coalesce(p_event->>'decision_id', ''),
    coalesce(p_event->>'execution_id', ''),
    coalesce(p_event->>'agent_id', ''),
    coalesce(p_event->>'event_type', ''),
    coalesce(p_event->>'previous_hash', ''),
    p_event_hash,
    p_event
  );

  UPDATE naeos_control_plane.ledger_head
  SET last_sequence = v_sequence, event_hash = p_event_hash
  WHERE singleton = TRUE;

  RETURN v_sequence;
END
$$;

-- Enable RLS as defense in depth. The private schema is not intended for browser access;
-- the dedicated backend service_role is the only runtime role granted access.
ALTER TABLE naeos_control_plane.policy_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.active_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.grants ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.approvals ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.ledger_head ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.ledger_events ENABLE ROW LEVEL SECURITY;

-- The runtime group is NOBYPASSRLS. Explicit policies and table grants jointly
-- constrain it; browser roles and Supabase service_role receive no access.
CREATE POLICY control_plane_policy_versions_runtime
  ON naeos_control_plane.policy_versions
  FOR SELECT TO naeos_control_plane_runtime USING (TRUE);
CREATE POLICY control_plane_policy_versions_insert
  ON naeos_control_plane.policy_versions
  FOR INSERT TO naeos_control_plane_runtime WITH CHECK (TRUE);
CREATE POLICY control_plane_active_policies_runtime
  ON naeos_control_plane.active_policies
  FOR ALL TO naeos_control_plane_runtime USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY control_plane_grants_runtime
  ON naeos_control_plane.grants
  FOR ALL TO naeos_control_plane_runtime USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY control_plane_approvals_runtime
  ON naeos_control_plane.approvals
  FOR ALL TO naeos_control_plane_runtime USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY control_plane_ledger_events_read
  ON naeos_control_plane.ledger_events
  FOR SELECT TO naeos_control_plane_runtime USING (TRUE);

REVOKE ALL ON SCHEMA naeos_control_plane FROM PUBLIC, anon, authenticated, service_role;
REVOKE ALL ON ALL TABLES IN SCHEMA naeos_control_plane FROM PUBLIC, anon, authenticated, service_role;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA naeos_control_plane FROM PUBLIC, anon, authenticated, service_role;

GRANT SELECT, INSERT ON naeos_control_plane.policy_versions TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.active_policies TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.grants TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.approvals TO naeos_control_plane_runtime;
GRANT SELECT ON naeos_control_plane.ledger_events TO naeos_control_plane_runtime;
GRANT EXECUTE ON FUNCTION naeos_control_plane.append_ledger_event(JSONB, TEXT, TEXT) TO naeos_control_plane_runtime;

ALTER DEFAULT PRIVILEGES IN SCHEMA naeos_control_plane
  REVOKE ALL ON TABLES FROM PUBLIC, anon, authenticated, service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA naeos_control_plane
  REVOKE ALL ON FUNCTIONS FROM PUBLIC, anon, authenticated, service_role;
 THEN
    RAISE EXCEPTION 'event hash must be a lowercase 64-character SHA-256 hex digest';
  END IF;
  IF coalesce(p_event->>'event_hash', '') <> p_event_hash THEN
    RAISE EXCEPTION 'event payload hash does not match supplied hash';
  END IF;

  SELECT * INTO STRICT v_head
  FROM naeos_control_plane.ledger_head
  WHERE singleton = TRUE
  FOR UPDATE;

  IF coalesce(p_event->>'previous_hash', '') <> v_head.event_hash
     OR coalesce(p_previous_hash, '') <> v_head.event_hash THEN
    RAISE EXCEPTION 'ledger predecessor mismatch';
  END IF;

  v_event_id := p_event->>'id';
  IF v_event_id IS NULL OR v_event_id = '' THEN
    RAISE EXCEPTION 'ledger event id is required';
  END IF;

  v_sequence := v_head.last_sequence + 1;

  INSERT INTO naeos_control_plane.ledger_events (
    sequence, event_id, occurred_at, request_id, decision_id, execution_id,
    agent_id, event_type, previous_hash, event_hash, event_payload
  ) VALUES (
    v_sequence,
    v_event_id,
    (p_event->>'timestamp')::TIMESTAMPTZ,
    coalesce(p_event->>'request_id', ''),
    coalesce(p_event->>'decision_id', ''),
    coalesce(p_event->>'execution_id', ''),
    coalesce(p_event->>'agent_id', ''),
    coalesce(p_event->>'event_type', ''),
    coalesce(p_event->>'previous_hash', ''),
    p_event_hash,
    p_event
  );

  UPDATE naeos_control_plane.ledger_head
  SET last_sequence = v_sequence, event_hash = p_event_hash
  WHERE singleton = TRUE;

  RETURN v_sequence;
END
$$;

-- Enable RLS as defense in depth. The private schema is not intended for browser access;
-- the dedicated backend service_role is the only runtime role granted access.
ALTER TABLE naeos_control_plane.policy_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.active_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.grants ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.approvals ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.ledger_head ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.ledger_events ENABLE ROW LEVEL SECURITY;

-- The runtime group is NOBYPASSRLS. Explicit policies and table grants jointly
-- constrain it; browser roles and Supabase service_role receive no access.
CREATE POLICY control_plane_policy_versions_runtime
  ON naeos_control_plane.policy_versions
  FOR SELECT TO naeos_control_plane_runtime USING (TRUE);
CREATE POLICY control_plane_policy_versions_insert
  ON naeos_control_plane.policy_versions
  FOR INSERT TO naeos_control_plane_runtime WITH CHECK (TRUE);
CREATE POLICY control_plane_active_policies_runtime
  ON naeos_control_plane.active_policies
  FOR ALL TO naeos_control_plane_runtime USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY control_plane_grants_runtime
  ON naeos_control_plane.grants
  FOR ALL TO naeos_control_plane_runtime USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY control_plane_approvals_runtime
  ON naeos_control_plane.approvals
  FOR ALL TO naeos_control_plane_runtime USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY control_plane_ledger_events_read
  ON naeos_control_plane.ledger_events
  FOR SELECT TO naeos_control_plane_runtime USING (TRUE);

REVOKE ALL ON SCHEMA naeos_control_plane FROM PUBLIC, anon, authenticated, service_role;
REVOKE ALL ON ALL TABLES IN SCHEMA naeos_control_plane FROM PUBLIC, anon, authenticated, service_role;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA naeos_control_plane FROM PUBLIC, anon, authenticated, service_role;

GRANT SELECT, INSERT ON naeos_control_plane.policy_versions TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.active_policies TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.grants TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.approvals TO naeos_control_plane_runtime;
GRANT SELECT ON naeos_control_plane.ledger_events TO naeos_control_plane_runtime;
GRANT EXECUTE ON FUNCTION naeos_control_plane.append_ledger_event(JSONB, TEXT, TEXT) TO naeos_control_plane_runtime;

ALTER DEFAULT PRIVILEGES IN SCHEMA naeos_control_plane
  REVOKE ALL ON TABLES FROM PUBLIC, anon, authenticated, service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA naeos_control_plane
  REVOKE ALL ON FUNCTIONS FROM PUBLIC, anon, authenticated, service_role;
 THEN
    RAISE EXCEPTION 'event hash must be a lowercase 64-character SHA-256 hex digest';
  END IF;
  IF coalesce(p_event->>'event_hash', '') <> p_event_hash THEN
    RAISE EXCEPTION 'event payload hash does not match supplied hash';
  END IF;

  SELECT * INTO STRICT v_head
  FROM naeos_control_plane.ledger_head
  WHERE singleton = TRUE
  FOR UPDATE;

  IF coalesce(p_event->>'previous_hash', '') <> v_head.event_hash
     OR coalesce(p_previous_hash, '') <> v_head.event_hash THEN
    RAISE EXCEPTION 'ledger predecessor mismatch';
  END IF;

  v_event_id := p_event->>'id';
  IF v_event_id IS NULL OR v_event_id = '' THEN
    RAISE EXCEPTION 'ledger event id is required';
  END IF;

  v_sequence := v_head.last_sequence + 1;

  INSERT INTO naeos_control_plane.ledger_events (
    sequence, event_id, occurred_at, request_id, decision_id, execution_id,
    agent_id, event_type, previous_hash, event_hash, event_payload
  ) VALUES (
    v_sequence,
    v_event_id,
    (p_event->>'timestamp')::TIMESTAMPTZ,
    coalesce(p_event->>'request_id', ''),
    coalesce(p_event->>'decision_id', ''),
    coalesce(p_event->>'execution_id', ''),
    coalesce(p_event->>'agent_id', ''),
    coalesce(p_event->>'event_type', ''),
    coalesce(p_event->>'previous_hash', ''),
    p_event_hash,
    p_event
  );

  UPDATE naeos_control_plane.ledger_head
  SET last_sequence = v_sequence, event_hash = p_event_hash
  WHERE singleton = TRUE;

  RETURN v_sequence;
END
$$;

-- Enable RLS as defense in depth. The private schema is not intended for browser access;
-- the dedicated backend service_role is the only runtime role granted access.
ALTER TABLE naeos_control_plane.policy_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.active_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.grants ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.approvals ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.ledger_head ENABLE ROW LEVEL SECURITY;
ALTER TABLE naeos_control_plane.ledger_events ENABLE ROW LEVEL SECURITY;

-- The runtime group is NOBYPASSRLS. Explicit policies and table grants jointly
-- constrain it; browser roles and Supabase service_role receive no access.
CREATE POLICY control_plane_policy_versions_runtime
  ON naeos_control_plane.policy_versions
  FOR SELECT TO naeos_control_plane_runtime USING (TRUE);
CREATE POLICY control_plane_policy_versions_insert
  ON naeos_control_plane.policy_versions
  FOR INSERT TO naeos_control_plane_runtime WITH CHECK (TRUE);
CREATE POLICY control_plane_active_policies_runtime
  ON naeos_control_plane.active_policies
  FOR ALL TO naeos_control_plane_runtime USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY control_plane_grants_runtime
  ON naeos_control_plane.grants
  FOR ALL TO naeos_control_plane_runtime USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY control_plane_approvals_runtime
  ON naeos_control_plane.approvals
  FOR ALL TO naeos_control_plane_runtime USING (TRUE) WITH CHECK (TRUE);
CREATE POLICY control_plane_ledger_events_read
  ON naeos_control_plane.ledger_events
  FOR SELECT TO naeos_control_plane_runtime USING (TRUE);

REVOKE ALL ON SCHEMA naeos_control_plane FROM PUBLIC, anon, authenticated, service_role;
REVOKE ALL ON ALL TABLES IN SCHEMA naeos_control_plane FROM PUBLIC, anon, authenticated, service_role;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA naeos_control_plane FROM PUBLIC, anon, authenticated, service_role;

GRANT SELECT, INSERT ON naeos_control_plane.policy_versions TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.active_policies TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.grants TO naeos_control_plane_runtime;
GRANT SELECT, INSERT, UPDATE ON naeos_control_plane.approvals TO naeos_control_plane_runtime;
GRANT SELECT ON naeos_control_plane.ledger_events TO naeos_control_plane_runtime;
GRANT EXECUTE ON FUNCTION naeos_control_plane.append_ledger_event(JSONB, TEXT, TEXT) TO naeos_control_plane_runtime;

ALTER DEFAULT PRIVILEGES IN SCHEMA naeos_control_plane
  REVOKE ALL ON TABLES FROM PUBLIC, anon, authenticated, service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA naeos_control_plane
  REVOKE ALL ON FUNCTIONS FROM PUBLIC, anon, authenticated, service_role;
