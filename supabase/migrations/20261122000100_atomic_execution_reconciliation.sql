-- Atomic, append-only reconciliation for ambiguous execution outcomes.
-- Apply only to isolated staging until the explicit production cutover gate.
-- The function serializes on ledger_head and revalidates the execution binding
-- while holding that lock, so concurrent reconcilers cannot both succeed.

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
  IF v_event_type <> 'EXECUTION_RECONCILED'
     OR p_event->>'execution_id' <> p_execution_id
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

  IF NOT EXISTS (
    SELECT 1 FROM naeos_control_plane.ledger_events
    WHERE execution_id = p_execution_id
      AND event_type = 'EXECUTION_UNKNOWN'
  ) AND NOT EXISTS (
    SELECT 1 FROM naeos_control_plane.ledger_events
    WHERE execution_id = p_execution_id
      AND event_type = 'EXECUTION_STARTED'
  ) THEN
    RAISE EXCEPTION 'execution has no unresolved outcome';
  END IF;

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
