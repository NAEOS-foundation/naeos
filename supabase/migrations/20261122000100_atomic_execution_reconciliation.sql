-- Index the event stream lookup performed by the reconciliation guard and
-- enforce at most one durable reconciliation event per execution ID.
CREATE INDEX IF NOT EXISTS ledger_events_execution_id_type_idx
  ON naeos_control_plane.ledger_events (execution_id, event_type)
  WHERE execution_id <> '';
CREATE UNIQUE INDEX IF NOT EXISTS ledger_events_single_execution_start_idx
  ON naeos_control_plane.ledger_events (execution_id)
  WHERE execution_id <> '' AND event_type = 'EXECUTION_STARTED';
CREATE UNIQUE INDEX IF NOT EXISTS ledger_events_single_reconciliation_idx
  ON naeos_control_plane.ledger_events (execution_id)
  WHERE execution_id <> '' AND event_type = 'EXECUTION_RECONCILED';

-- Atomic, append-only reconciliation for ambiguous execution outcomes.
-- Apply only to isolated staging until the explicit production cutover gate.
-- The function serializes on ledger_head and revalidates the execution binding
-- while holding that lock, so concurrent reconcilers cannot both succeed.

-- Keep the generic append API from bypassing reconciliation-specific validation.
CREATE OR REPLACE FUNCTION naeos_control_plane.append_ledger_event(
  p_event JSONB,
  p_previous_hash TEXT,
  p_event_hash TEXT
)
RETURNS BIGINT
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
  v_head naeos_control_plane.ledger_head%ROWTYPE;
  v_sequence BIGINT;
  v_event_id TEXT;
BEGIN
  IF p_event IS NULL OR jsonb_typeof(p_event) <> 'object' THEN
    RAISE EXCEPTION 'ledger event must be a JSON object';
  END IF;
  IF p_event_hash IS NULL OR p_event_hash !~ '^[0-9a-f]{64}-- observation may contradict that terminal operator decision.
CREATE OR REPLACE FUNCTION naeos_control_plane.reject_execution_outcome_after_reconciliation()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
  IF NEW.event_type IN ('EXECUTION_STARTED', 'EXECUTION_ALLOWED', 'EXECUTION_UNKNOWN', 'EXECUTION_RECONCILED')
     AND NEW.execution_id <> ''
     AND EXISTS (
       SELECT 1
       FROM naeos_control_plane.ledger_events existing
       WHERE existing.execution_id = NEW.execution_id
         AND existing.event_type = 'EXECUTION_RECONCILED'
     ) THEN
    RAISE EXCEPTION 'execution already reconciled; later outcome events are forbidden';
  END IF;
  RETURN NEW;
END
$function$;

CREATE TRIGGER execution_outcome_after_reconciliation_guard
BEFORE INSERT ON naeos_control_plane.ledger_events
FOR EACH ROW EXECUTE FUNCTION naeos_control_plane.reject_execution_outcome_after_reconciliation();

REVOKE ALL ON FUNCTION naeos_control_plane.reject_execution_outcome_after_reconciliation()
  FROM PUBLIC, anon, authenticated, service_role;

CREATE OR REPLACE FUNCTION naeos_control_plane.reconcile_execution(
  p_execution_id TEXT,
  p_event JSONB
)
RETURNS BIGINT
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
  v_head naeos_control_plane.ledger_head%ROWTYPE;
  v_start JSONB;
  v_sequence BIGINT;
  v_event_type TEXT;
  v_metadata JSONB;
BEGIN
  IF p_execution_id IS NULL OR p_execution_id = ''
     OR p_event IS NULL OR jsonb_typeof(p_event) <> 'object' THEN
    RAISE EXCEPTION 'execution id and reconciliation event are required';
  END IF;

  v_event_type := p_event->>'event_type';
  v_metadata := p_event->'metadata';
  IF v_event_type IS DISTINCT FROM 'EXECUTION_RECONCILED'
     OR p_event->>'execution_id' IS DISTINCT FROM p_execution_id
     OR coalesce(v_metadata->>'evidence_ref', '') = ''
     OR coalesce(v_metadata->>'reconciler_id', '') = ''
     OR coalesce(v_metadata->>'reconciliation_outcome', '') NOT IN ('COMMITTED', 'NOT_COMMITTED') THEN
    RAISE EXCEPTION 'invalid reconciliation event';
  END IF;

  -- Global ledger serialization also fences concurrent appenders and reconcilers.
  SELECT * INTO STRICT v_head
  FROM naeos_control_plane.ledger_head
  WHERE singleton = TRUE
  FOR UPDATE;

  SELECT event_payload INTO v_start
  FROM naeos_control_plane.ledger_events
  WHERE execution_id = p_execution_id
    AND event_type = 'EXECUTION_STARTED'
  ORDER BY sequence ASC
  LIMIT 1;

  IF v_start IS NULL THEN
    RAISE EXCEPTION 'execution start not found';
  END IF;

  IF p_event->>'request_id' IS DISTINCT FROM v_start->>'request_id'
     OR p_event->>'decision_id' IS DISTINCT FROM v_start->>'decision_id'
     OR p_event->>'agent_id' IS DISTINCT FROM v_start->>'agent_id'
     OR p_event->>'capability' IS DISTINCT FROM v_start->>'capability'
     OR p_event->>'artifact_hash' IS DISTINCT FROM v_start->>'artifact_hash' THEN
    RAISE EXCEPTION 'reconciliation binding mismatch';
  END IF;

  IF EXISTS (
    SELECT 1 FROM naeos_control_plane.ledger_events
    WHERE execution_id = p_execution_id
      AND event_type IN ('EXECUTION_ALLOWED', 'EXECUTION_RECONCILED')
  ) THEN
    RAISE EXCEPTION 'execution already has a terminal outcome';
  END IF;

  -- A durable start with no terminal event is unresolved even if the process
  -- crashed before it could append EXECUTION_UNKNOWN.

  -- This dedicated path writes directly while holding the ledger-head lock.
  -- Generic append_ledger_event explicitly rejects EXECUTION_RECONCILED.
  IF coalesce(p_event->>'id', '') = ''
     OR coalesce(p_event->>'timestamp', '') = ''
     OR coalesce(p_event->>'event_hash', '') !~ '^[0-9a-f]{64}
END
$function$;

REVOKE ALL ON FUNCTION naeos_control_plane.reconcile_execution(TEXT, JSONB)
  FROM PUBLIC, anon, authenticated, service_role;
GRANT EXECUTE ON FUNCTION naeos_control_plane.reconcile_execution(TEXT, JSONB)
  TO naeos_control_plane_runtime;
 THEN
    RAISE EXCEPTION 'event hash must be a lowercase 64-character SHA-256 hex digest';
  END IF;
  IF coalesce(p_event->>'event_hash', '') <> p_event_hash THEN
    RAISE EXCEPTION 'event payload hash does not match supplied hash';
  END IF;
  IF coalesce(p_event->>'event_type', '') = '' THEN
    RAISE EXCEPTION 'ledger event type is required';
  END IF;
  IF p_event->>'event_type' = 'EXECUTION_RECONCILED' THEN
    RAISE EXCEPTION 'reconciliation events must use the atomic reconciliation function';
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

  IF p_event->>'event_type' = 'EXECUTION_STARTED'
     AND coalesce(p_event->>'decision_id', '') <> ''
     AND EXISTS (
       SELECT 1
       FROM naeos_control_plane.ledger_events AS existing
       WHERE existing.decision_id = p_event->>'decision_id'
         AND existing.event_type IN ('EXECUTION_STARTED', 'EXECUTION_ALLOWED', 'EXECUTION_UNKNOWN')
     ) THEN
    RAISE EXCEPTION 'execution already claimed for decision';
  END IF;

  v_sequence := v_head.last_sequence + 1;
  INSERT INTO naeos_control_plane.ledger_events (
    sequence, event_id, occurred_at, request_id, decision_id, execution_id,
    agent_id, event_type, previous_hash, event_hash, event_payload
  ) VALUES (
    v_sequence, v_event_id, (p_event->>'timestamp')::TIMESTAMPTZ,
    coalesce(p_event->>'request_id', ''), coalesce(p_event->>'decision_id', ''),
    coalesce(p_event->>'execution_id', ''), coalesce(p_event->>'agent_id', ''),
    p_event->>'event_type', coalesce(p_event->>'previous_hash', ''),
    p_event_hash, p_event
  );

  UPDATE naeos_control_plane.ledger_head
  SET last_sequence = v_sequence, event_hash = p_event_hash
  WHERE singleton = TRUE;
  RETURN v_sequence;
END
$function$;

-- Once reconciliation is committed, no later provider receipt or UNKNOWN
-- observation may contradict that terminal operator decision.
CREATE OR REPLACE FUNCTION naeos_control_plane.reject_execution_outcome_after_reconciliation()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
  IF NEW.event_type IN ('EXECUTION_STARTED', 'EXECUTION_ALLOWED', 'EXECUTION_UNKNOWN', 'EXECUTION_RECONCILED')
     AND NEW.execution_id <> ''
     AND EXISTS (
       SELECT 1
       FROM naeos_control_plane.ledger_events existing
       WHERE existing.execution_id = NEW.execution_id
         AND existing.event_type = 'EXECUTION_RECONCILED'
     ) THEN
    RAISE EXCEPTION 'execution already reconciled; later outcome events are forbidden';
  END IF;
  RETURN NEW;
END
$function$;

CREATE TRIGGER execution_outcome_after_reconciliation_guard
BEFORE INSERT ON naeos_control_plane.ledger_events
FOR EACH ROW EXECUTE FUNCTION naeos_control_plane.reject_execution_outcome_after_reconciliation();

REVOKE ALL ON FUNCTION naeos_control_plane.reject_execution_outcome_after_reconciliation()
  FROM PUBLIC, anon, authenticated, service_role;

CREATE OR REPLACE FUNCTION naeos_control_plane.reconcile_execution(
  p_execution_id TEXT,
  p_event JSONB
)
RETURNS BIGINT
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
  v_head naeos_control_plane.ledger_head%ROWTYPE;
  v_start JSONB;
  v_sequence BIGINT;
  v_event_type TEXT;
  v_metadata JSONB;
BEGIN
  IF p_execution_id IS NULL OR p_execution_id = ''
     OR p_event IS NULL OR jsonb_typeof(p_event) <> 'object' THEN
    RAISE EXCEPTION 'execution id and reconciliation event are required';
  END IF;

  v_event_type := p_event->>'event_type';
  v_metadata := p_event->'metadata';
  IF v_event_type IS DISTINCT FROM 'EXECUTION_RECONCILED'
     OR p_event->>'execution_id' IS DISTINCT FROM p_execution_id
     OR coalesce(v_metadata->>'evidence_ref', '') = ''
     OR coalesce(v_metadata->>'reconciler_id', '') = ''
     OR coalesce(v_metadata->>'reconciliation_outcome', '') NOT IN ('COMMITTED', 'NOT_COMMITTED') THEN
    RAISE EXCEPTION 'invalid reconciliation event';
  END IF;

  -- Global ledger serialization also fences concurrent appenders and reconcilers.
  SELECT * INTO STRICT v_head
  FROM naeos_control_plane.ledger_head
  WHERE singleton = TRUE
  FOR UPDATE;

  SELECT event_payload INTO v_start
  FROM naeos_control_plane.ledger_events
  WHERE execution_id = p_execution_id
    AND event_type = 'EXECUTION_STARTED'
  ORDER BY sequence ASC
  LIMIT 1;

  IF v_start IS NULL THEN
    RAISE EXCEPTION 'execution start not found';
  END IF;

  IF p_event->>'request_id' IS DISTINCT FROM v_start->>'request_id'
     OR p_event->>'decision_id' IS DISTINCT FROM v_start->>'decision_id'
     OR p_event->>'agent_id' IS DISTINCT FROM v_start->>'agent_id'
     OR p_event->>'capability' IS DISTINCT FROM v_start->>'capability'
     OR p_event->>'artifact_hash' IS DISTINCT FROM v_start->>'artifact_hash' THEN
    RAISE EXCEPTION 'reconciliation binding mismatch';
  END IF;

  IF EXISTS (
    SELECT 1 FROM naeos_control_plane.ledger_events
    WHERE execution_id = p_execution_id
      AND event_type IN ('EXECUTION_ALLOWED', 'EXECUTION_RECONCILED')
  ) THEN
    RAISE EXCEPTION 'execution already has a terminal outcome';
  END IF;

  -- A durable start with no terminal event is unresolved even if the process
  -- crashed before it could append EXECUTION_UNKNOWN.

  -- Reuse the canonical append function. It takes the same row lock
  -- re-entrantly and verifies the candidate predecessor/hash.
  SELECT naeos_control_plane.append_ledger_event(
    p_event,
    coalesce(p_event->>'previous_hash', ''),
    coalesce(p_event->>'event_hash', '')
  ) INTO v_sequence;

  RETURN v_sequence;
END
$function$;

REVOKE ALL ON FUNCTION naeos_control_plane.reconcile_execution(TEXT, JSONB)
  FROM PUBLIC, anon, authenticated, service_role;
GRANT EXECUTE ON FUNCTION naeos_control_plane.reconcile_execution(TEXT, JSONB)
  TO naeos_control_plane_runtime;

     OR coalesce(p_event->>'event_hash', '') <> coalesce(p_event->>'event_hash', '') THEN
    RAISE EXCEPTION 'invalid reconciliation event envelope';
  END IF;
  IF coalesce(p_event->>'previous_hash', '') <> v_head.event_hash THEN
    RAISE EXCEPTION 'ledger predecessor mismatch';
  END IF;

  v_sequence := v_head.last_sequence + 1;
  INSERT INTO naeos_control_plane.ledger_events (
    sequence, event_id, occurred_at, request_id, decision_id, execution_id,
    agent_id, event_type, previous_hash, event_hash, event_payload
  ) VALUES (
    v_sequence,
    p_event->>'id',
    (p_event->>'timestamp')::TIMESTAMPTZ,
    coalesce(p_event->>'request_id', ''),
    coalesce(p_event->>'decision_id', ''),
    p_execution_id,
    coalesce(p_event->>'agent_id', ''),
    'EXECUTION_RECONCILED',
    coalesce(p_event->>'previous_hash', ''),
    p_event->>'event_hash',
    p_event
  );

  UPDATE naeos_control_plane.ledger_head
  SET last_sequence = v_sequence, event_hash = p_event->>'event_hash'
  WHERE singleton = TRUE;

  RETURN v_sequence;
END
$function$;

REVOKE ALL ON FUNCTION naeos_control_plane.reconcile_execution(TEXT, JSONB)
  FROM PUBLIC, anon, authenticated, service_role;
GRANT EXECUTE ON FUNCTION naeos_control_plane.reconcile_execution(TEXT, JSONB)
  TO naeos_control_plane_runtime;
 THEN
    RAISE EXCEPTION 'event hash must be a lowercase 64-character SHA-256 hex digest';
  END IF;
  IF coalesce(p_event->>'event_hash', '') <> p_event_hash THEN
    RAISE EXCEPTION 'event payload hash does not match supplied hash';
  END IF;
  IF coalesce(p_event->>'event_type', '') = '' THEN
    RAISE EXCEPTION 'ledger event type is required';
  END IF;
  IF p_event->>'event_type' = 'EXECUTION_RECONCILED' THEN
    RAISE EXCEPTION 'reconciliation events must use the atomic reconciliation function';
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

  IF p_event->>'event_type' = 'EXECUTION_STARTED'
     AND coalesce(p_event->>'decision_id', '') <> ''
     AND EXISTS (
       SELECT 1
       FROM naeos_control_plane.ledger_events AS existing
       WHERE existing.decision_id = p_event->>'decision_id'
         AND existing.event_type IN ('EXECUTION_STARTED', 'EXECUTION_ALLOWED', 'EXECUTION_UNKNOWN')
     ) THEN
    RAISE EXCEPTION 'execution already claimed for decision';
  END IF;

  v_sequence := v_head.last_sequence + 1;
  INSERT INTO naeos_control_plane.ledger_events (
    sequence, event_id, occurred_at, request_id, decision_id, execution_id,
    agent_id, event_type, previous_hash, event_hash, event_payload
  ) VALUES (
    v_sequence, v_event_id, (p_event->>'timestamp')::TIMESTAMPTZ,
    coalesce(p_event->>'request_id', ''), coalesce(p_event->>'decision_id', ''),
    coalesce(p_event->>'execution_id', ''), coalesce(p_event->>'agent_id', ''),
    p_event->>'event_type', coalesce(p_event->>'previous_hash', ''),
    p_event_hash, p_event
  );

  UPDATE naeos_control_plane.ledger_head
  SET last_sequence = v_sequence, event_hash = p_event_hash
  WHERE singleton = TRUE;
  RETURN v_sequence;
END
$function$;

-- Once reconciliation is committed, no later provider receipt or UNKNOWN
-- observation may contradict that terminal operator decision.
CREATE OR REPLACE FUNCTION naeos_control_plane.reject_execution_outcome_after_reconciliation()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
  IF NEW.event_type IN ('EXECUTION_STARTED', 'EXECUTION_ALLOWED', 'EXECUTION_UNKNOWN', 'EXECUTION_RECONCILED')
     AND NEW.execution_id <> ''
     AND EXISTS (
       SELECT 1
       FROM naeos_control_plane.ledger_events existing
       WHERE existing.execution_id = NEW.execution_id
         AND existing.event_type = 'EXECUTION_RECONCILED'
     ) THEN
    RAISE EXCEPTION 'execution already reconciled; later outcome events are forbidden';
  END IF;
  RETURN NEW;
END
$function$;

CREATE TRIGGER execution_outcome_after_reconciliation_guard
BEFORE INSERT ON naeos_control_plane.ledger_events
FOR EACH ROW EXECUTE FUNCTION naeos_control_plane.reject_execution_outcome_after_reconciliation();

REVOKE ALL ON FUNCTION naeos_control_plane.reject_execution_outcome_after_reconciliation()
  FROM PUBLIC, anon, authenticated, service_role;

CREATE OR REPLACE FUNCTION naeos_control_plane.reconcile_execution(
  p_execution_id TEXT,
  p_event JSONB
)
RETURNS BIGINT
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
  v_head naeos_control_plane.ledger_head%ROWTYPE;
  v_start JSONB;
  v_sequence BIGINT;
  v_event_type TEXT;
  v_metadata JSONB;
BEGIN
  IF p_execution_id IS NULL OR p_execution_id = ''
     OR p_event IS NULL OR jsonb_typeof(p_event) <> 'object' THEN
    RAISE EXCEPTION 'execution id and reconciliation event are required';
  END IF;

  v_event_type := p_event->>'event_type';
  v_metadata := p_event->'metadata';
  IF v_event_type IS DISTINCT FROM 'EXECUTION_RECONCILED'
     OR p_event->>'execution_id' IS DISTINCT FROM p_execution_id
     OR coalesce(v_metadata->>'evidence_ref', '') = ''
     OR coalesce(v_metadata->>'reconciler_id', '') = ''
     OR coalesce(v_metadata->>'reconciliation_outcome', '') NOT IN ('COMMITTED', 'NOT_COMMITTED') THEN
    RAISE EXCEPTION 'invalid reconciliation event';
  END IF;

  -- Global ledger serialization also fences concurrent appenders and reconcilers.
  SELECT * INTO STRICT v_head
  FROM naeos_control_plane.ledger_head
  WHERE singleton = TRUE
  FOR UPDATE;

  SELECT event_payload INTO v_start
  FROM naeos_control_plane.ledger_events
  WHERE execution_id = p_execution_id
    AND event_type = 'EXECUTION_STARTED'
  ORDER BY sequence ASC
  LIMIT 1;

  IF v_start IS NULL THEN
    RAISE EXCEPTION 'execution start not found';
  END IF;

  IF p_event->>'request_id' IS DISTINCT FROM v_start->>'request_id'
     OR p_event->>'decision_id' IS DISTINCT FROM v_start->>'decision_id'
     OR p_event->>'agent_id' IS DISTINCT FROM v_start->>'agent_id'
     OR p_event->>'capability' IS DISTINCT FROM v_start->>'capability'
     OR p_event->>'artifact_hash' IS DISTINCT FROM v_start->>'artifact_hash' THEN
    RAISE EXCEPTION 'reconciliation binding mismatch';
  END IF;

  IF EXISTS (
    SELECT 1 FROM naeos_control_plane.ledger_events
    WHERE execution_id = p_execution_id
      AND event_type IN ('EXECUTION_ALLOWED', 'EXECUTION_RECONCILED')
  ) THEN
    RAISE EXCEPTION 'execution already has a terminal outcome';
  END IF;

  -- A durable start with no terminal event is unresolved even if the process
  -- crashed before it could append EXECUTION_UNKNOWN.

  -- Reuse the canonical append function. It takes the same row lock
  -- re-entrantly and verifies the candidate predecessor/hash.
  SELECT naeos_control_plane.append_ledger_event(
    p_event,
    coalesce(p_event->>'previous_hash', ''),
    coalesce(p_event->>'event_hash', '')
  ) INTO v_sequence;

  RETURN v_sequence;
END
$function$;

REVOKE ALL ON FUNCTION naeos_control_plane.reconcile_execution(TEXT, JSONB)
  FROM PUBLIC, anon, authenticated, service_role;
GRANT EXECUTE ON FUNCTION naeos_control_plane.reconcile_execution(TEXT, JSONB)
  TO naeos_control_plane_runtime;
