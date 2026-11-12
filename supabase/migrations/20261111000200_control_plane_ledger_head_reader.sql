-- Add a narrow, read-only ledger-head accessor for the Control Plane runtime.
-- Runtime intentionally has no direct SELECT/UPDATE privileges on ledger_head;
-- writes and locking remain inside append_ledger_event (SECURITY DEFINER).

CREATE OR REPLACE FUNCTION naeos_control_plane.read_ledger_head()
RETURNS TABLE (last_sequence BIGINT, event_hash TEXT)
LANGUAGE sql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
  SELECT head.last_sequence, head.event_hash
  FROM naeos_control_plane.ledger_head AS head
  WHERE head.singleton = TRUE
$function$;

REVOKE ALL ON FUNCTION naeos_control_plane.read_ledger_head() FROM PUBLIC, anon, authenticated, service_role;
GRANT EXECUTE ON FUNCTION naeos_control_plane.read_ledger_head() TO naeos_control_plane_runtime;
