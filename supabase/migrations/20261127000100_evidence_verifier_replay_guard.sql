-- Dedicated durable replay protection for the standalone Evidence Verifier.
-- Apply only to an isolated verifier staging database. Never run against production.
DO $role$
BEGIN
 IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname='naeos_evidence_verifier_runtime') THEN
  CREATE ROLE naeos_evidence_verifier_runtime NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
 END IF;
END
$role$;
CREATE SCHEMA IF NOT EXISTS naeos_evidence_verifier;
REVOKE ALL ON SCHEMA naeos_evidence_verifier FROM PUBLIC;
GRANT USAGE ON SCHEMA naeos_evidence_verifier TO naeos_evidence_verifier_runtime;
CREATE TABLE IF NOT EXISTS naeos_evidence_verifier.replay_claims (
 verifier_scope TEXT NOT NULL CHECK (length(verifier_scope) BETWEEN 1 AND 128),
 request_id TEXT NOT NULL CHECK (request_id ~ '^[0-9a-f]{32}$'),
 claimed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 PRIMARY KEY (verifier_scope,request_id)
);
REVOKE ALL ON TABLE naeos_evidence_verifier.replay_claims FROM PUBLIC;
GRANT INSERT ON TABLE naeos_evidence_verifier.replay_claims TO naeos_evidence_verifier_runtime;
