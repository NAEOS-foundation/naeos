\set ON_ERROR_STOP on

BEGIN;

-- The workflow provisions a minimal Supabase-compatible auth schema and roles.
-- All data below is synthetic and the transaction is rolled back.
INSERT INTO auth.users (id) VALUES
  ('11111111-1111-4111-8111-111111111111'),
  ('22222222-2222-4222-8222-222222222222'),
  ('33333333-3333-4333-8333-333333333333');

INSERT INTO recruitment.founding_partner_applications (
  id, applicant_name, email, primary_role, relevant_experience,
  contribution_proposal, motivation, weekly_hours, collaboration_model,
  trial_available, partnership_expectations, consent
) VALUES (
  'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
  'Synthetic Applicant', 'synthetic@example.invalid',
  'Technical Founding Partner / CTO',
  'Synthetic experience description for isolated database regression tests.',
  'Synthetic contribution proposal for isolated database regression tests.',
  'Synthetic motivation statement for isolated database regression tests.',
  '5–10', 'Part-time', true,
  'Synthetic partnership expectations for isolated database regression tests.',
  true
);

INSERT INTO recruitment.founding_partner_reviewers (user_id, display_name, enabled)
VALUES
  ('11111111-1111-4111-8111-111111111111', 'Active Test Reviewer', true),
  ('33333333-3333-4333-8333-333333333333', 'Disabled Test Reviewer', false);

DO $assert_privileges_and_rls$
BEGIN
  IF has_schema_privilege('anon', 'recruitment', 'USAGE') THEN
    RAISE EXCEPTION 'anon must not have recruitment schema USAGE';
  END IF;
  IF has_schema_privilege('authenticated', 'recruitment', 'USAGE') THEN
    RAISE EXCEPTION 'authenticated must not receive USAGE on the private intake schema';
  END IF;
  IF has_schema_privilege('anon', 'recruitment_private', 'USAGE')
     OR NOT has_schema_privilege('authenticated', 'recruitment_private', 'USAGE') THEN
    RAISE EXCEPTION 'only authenticated sessions may resolve the private reviewer RPC schema';
  END IF;
  IF has_table_privilege('anon', 'recruitment.founding_partner_applications', 'SELECT')
     OR has_table_privilege('anon', 'recruitment.founding_partner_reviewers', 'SELECT')
     OR has_table_privilege('anon', 'recruitment.founding_partner_review_events', 'SELECT') THEN
    RAISE EXCEPTION 'anon must not have SELECT on founding partner review tables';
  END IF;
  IF has_table_privilege('authenticated', 'recruitment.founding_partner_applications', 'SELECT')
     OR has_table_privilege('authenticated', 'recruitment.founding_partner_reviewers', 'SELECT')
     OR has_table_privilege('authenticated', 'recruitment.founding_partner_review_events', 'SELECT') THEN
    RAISE EXCEPTION 'authenticated must not receive direct SELECT on private recruitment tables';
  END IF;
  IF has_table_privilege('authenticated', 'recruitment.founding_partner_applications', 'INSERT')
     OR has_table_privilege('authenticated', 'recruitment.founding_partner_applications', 'UPDATE')
     OR has_table_privilege('authenticated', 'recruitment.founding_partner_review_events', 'INSERT')
     OR has_table_privilege('authenticated', 'recruitment.founding_partner_review_events', 'UPDATE')
     OR has_table_privilege('authenticated', 'recruitment.founding_partner_review_events', 'DELETE') THEN
    RAISE EXCEPTION 'authenticated must not directly mutate applications or review audit events';
  END IF;
  IF NOT (SELECT relrowsecurity FROM pg_catalog.pg_class
          WHERE oid = 'recruitment.founding_partner_applications'::regclass)
     OR NOT (SELECT relrowsecurity FROM pg_catalog.pg_class
             WHERE oid = 'recruitment.founding_partner_reviewers'::regclass)
     OR NOT (SELECT relrowsecurity FROM pg_catalog.pg_class
             WHERE oid = 'recruitment.founding_partner_review_events'::regclass) THEN
    RAISE EXCEPTION 'RLS must be enabled on all review workflow tables';
  END IF;
  IF (SELECT prosecdef FROM pg_catalog.pg_proc WHERE oid = 'public.founding_partner_review_transition(uuid,text,text)'::regprocedure)
     OR (SELECT prosecdef FROM pg_catalog.pg_proc WHERE oid = 'public.founding_partner_review_context()'::regprocedure)
     OR (SELECT prosecdef FROM pg_catalog.pg_proc WHERE oid = 'public.founding_partner_review_list(integer,integer,text)'::regprocedure)
     OR (SELECT prosecdef FROM pg_catalog.pg_proc WHERE oid = 'public.founding_partner_review_detail(uuid)'::regprocedure) THEN
    RAISE EXCEPTION 'public PostgREST wrappers must remain SECURITY INVOKER';
  END IF;
  IF NOT (SELECT prosecdef FROM pg_catalog.pg_proc WHERE oid = 'recruitment_private.founding_partner_review_transition_internal(uuid,text,text)'::regprocedure)
     OR NOT (SELECT prosecdef FROM pg_catalog.pg_proc WHERE oid = 'recruitment_private.founding_partner_review_context_internal()'::regprocedure)
     OR NOT (SELECT prosecdef FROM pg_catalog.pg_proc WHERE oid = 'recruitment_private.founding_partner_review_list_internal(integer,integer,text)'::regprocedure)
     OR NOT (SELECT prosecdef FROM pg_catalog.pg_proc WHERE oid = 'recruitment_private.founding_partner_review_detail_internal(uuid)'::regprocedure) THEN
    RAISE EXCEPTION 'privileged implementations must remain in the private recruitment schema';
  END IF;
  IF has_function_privilege('anon', 'recruitment_private.founding_partner_review_transition_internal(uuid,text,text)', 'EXECUTE')
     OR has_function_privilege('anon', 'recruitment_private.founding_partner_review_context_internal()', 'EXECUTE')
     OR has_function_privilege('anon', 'recruitment_private.founding_partner_review_list_internal(integer,integer,text)', 'EXECUTE')
     OR has_function_privilege('anon', 'recruitment_private.founding_partner_review_detail_internal(uuid)', 'EXECUTE')
     OR NOT has_function_privilege('authenticated', 'recruitment_private.founding_partner_review_transition_internal(uuid,text,text)', 'EXECUTE')
     OR NOT has_function_privilege('authenticated', 'recruitment_private.founding_partner_review_context_internal()', 'EXECUTE')
     OR NOT has_function_privilege('authenticated', 'recruitment_private.founding_partner_review_list_internal(integer,integer,text)', 'EXECUTE')
     OR NOT has_function_privilege('authenticated', 'recruitment_private.founding_partner_review_detail_internal(uuid)', 'EXECUTE') THEN
    RAISE EXCEPTION 'private review implementations must be executable only by authenticated sessions';
  END IF;
  IF NOT has_function_privilege('authenticated',
       'public.founding_partner_review_transition(uuid,text,text)', 'EXECUTE')
     OR has_function_privilege('anon',
       'public.founding_partner_review_transition(uuid,text,text)', 'EXECUTE')
     OR NOT has_function_privilege('authenticated', 'public.founding_partner_review_context()', 'EXECUTE')
     OR has_function_privilege('anon', 'public.founding_partner_review_context()', 'EXECUTE')
     OR NOT has_function_privilege('authenticated', 'public.founding_partner_review_list(integer,integer,text)', 'EXECUTE')
     OR has_function_privilege('anon', 'public.founding_partner_review_list(integer,integer,text)', 'EXECUTE')
     OR NOT has_function_privilege('authenticated', 'public.founding_partner_review_detail(uuid)', 'EXECUTE')
     OR has_function_privilege('anon', 'public.founding_partner_review_detail(uuid)', 'EXECUTE') THEN
    RAISE EXCEPTION 'review RPCs must be executable only by authenticated';
  END IF;
END
$assert_privileges_and_rls$;

-- Inject an audit-sink failure to prove status mutation and audit append roll back together.
CREATE FUNCTION public.reject_review_event_insert() RETURNS trigger
LANGUAGE plpgsql
AS $trigger$
BEGIN
  IF NEW.new_status = 'Shortlisted' THEN
    RAISE EXCEPTION 'synthetic audit sink failure';
  END IF;
  RETURN NEW;
END
$trigger$;
CREATE TRIGGER reject_review_event_insert
BEFORE INSERT ON recruitment.founding_partner_review_events
FOR EACH ROW EXECUTE FUNCTION public.reject_review_event_insert();

-- A signed-in user who is not allowlisted sees no applications and cannot transition.
SET LOCAL ROLE authenticated;
SELECT set_config('request.jwt.claim.sub', '22222222-2222-4222-8222-222222222222', true);
DO $assert_dashboard_rpcs_deny_non_reviewers$
BEGIN
  BEGIN
    PERFORM public.founding_partner_review_context();
    RAISE EXCEPTION 'non-reviewer unexpectedly obtained reviewer context';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
  BEGIN
    PERFORM public.founding_partner_review_list(50, 0, NULL);
    RAISE EXCEPTION 'non-reviewer unexpectedly listed applications';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
  BEGIN
    PERFORM public.founding_partner_review_detail('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa');
    RAISE EXCEPTION 'non-reviewer unexpectedly opened application detail';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
END
$assert_dashboard_rpcs_deny_non_reviewers$;

DO $assert_non_reviewer_denied$
BEGIN
  BEGIN
    PERFORM 1 FROM recruitment.founding_partner_applications;
    RAISE EXCEPTION 'authenticated non-reviewer unexpectedly received direct table access';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
  BEGIN
    PERFORM public.founding_partner_review_transition(
      'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'Screening', 'must be denied'
    );
    RAISE EXCEPTION 'non-reviewer unexpectedly executed review transition';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
END
$assert_non_reviewer_denied$;

-- An enabled allowlisted reviewer can read and transition; audit is atomic.
SELECT set_config('request.jwt.claim.sub', '11111111-1111-4111-8111-111111111111', true);
DO $assert_active_reviewer_flow$
DECLARE
  v_detail jsonb;
BEGIN
  IF public.founding_partner_review_context()->>'user_id' <> '11111111-1111-4111-8111-111111111111' THEN
    RAISE EXCEPTION 'enabled reviewer could not obtain reviewer context';
  END IF;
  IF (public.founding_partner_review_list(10, 0, NULL)->>'total')::integer <> 1 THEN
    RAISE EXCEPTION 'enabled reviewer could not list applications through dashboard RPC';
  END IF;
  IF public.founding_partner_review_detail('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa')->'application'->>'applicant_name' <> 'Synthetic Applicant' THEN
    RAISE EXCEPTION 'enabled reviewer could not retrieve application detail through dashboard RPC';
  END IF;
  IF public.founding_partner_review_transition(
       'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'Screening', 'synthetic review'
     ) IS DISTINCT FROM true THEN
    RAISE EXCEPTION 'valid reviewer transition did not return true';
  END IF;
  v_detail := public.founding_partner_review_detail('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa');
  IF v_detail->'application'->>'application_status' <> 'Screening' THEN
    RAISE EXCEPTION 'valid transition did not update application status';
  END IF;
  IF jsonb_array_length(v_detail->'events') <> 1
     OR v_detail->'events'->0->>'reviewer_id' <> '11111111-1111-4111-8111-111111111111'
     OR v_detail->'events'->0->>'previous_status' <> 'New Application'
     OR v_detail->'events'->0->>'new_status' <> 'Screening'
     OR v_detail->'events'->0->>'note' <> 'synthetic review' THEN
    RAISE EXCEPTION 'valid transition must append exactly one matching audit event';
  END IF;
  BEGIN
    PERFORM public.founding_partner_review_transition(
      'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'New Application', 'invalid backward transition'
    );
    RAISE EXCEPTION 'invalid transition unexpectedly succeeded';
  EXCEPTION WHEN SQLSTATE '22023' THEN
    NULL;
  END;
  v_detail := public.founding_partner_review_detail('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa');
  IF v_detail->'application'->>'application_status' <> 'Screening'
     OR jsonb_array_length(v_detail->'events') <> 1 THEN
    RAISE EXCEPTION 'invalid transition changed application or audit history';
  END IF;
  DECLARE
    v_failure_state text;
    v_failure_message text;
  BEGIN
    PERFORM public.founding_partner_review_transition(
      'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'Shortlisted', 'audit failure rollback'
    );
    RAISE EXCEPTION 'transition unexpectedly succeeded when audit insert failed';
  EXCEPTION WHEN OTHERS THEN
    GET STACKED DIAGNOSTICS
      v_failure_state = RETURNED_SQLSTATE,
      v_failure_message = MESSAGE_TEXT;
    IF v_failure_state <> 'P0001' OR v_failure_message <> 'synthetic audit sink failure' THEN
      RAISE EXCEPTION 'expected audit sink failure P0001, got %: %',
        v_failure_state, v_failure_message;
    END IF;
  END;
  v_detail := public.founding_partner_review_detail('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa');
  IF v_detail->'application'->>'application_status' <> 'Screening'
     OR jsonb_array_length(v_detail->'events') <> 1 THEN
    RAISE EXCEPTION 'audit failure did not roll back the application status mutation';
  END IF;
END
$assert_active_reviewer_flow$;

-- Disabling a reviewer immediately removes read and transition access.
SELECT set_config('request.jwt.claim.sub', '33333333-3333-4333-8333-333333333333', true);
DO $assert_disabled_reviewer_denied$
BEGIN
  BEGIN
    PERFORM 1 FROM recruitment.founding_partner_applications;
    RAISE EXCEPTION 'disabled reviewer unexpectedly received direct table access';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
  BEGIN
    PERFORM public.founding_partner_review_list(10, 0, NULL);
    RAISE EXCEPTION 'disabled reviewer unexpectedly listed applications through dashboard RPC';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
  BEGIN
    PERFORM public.founding_partner_review_transition(
      'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'Interview', 'must be denied'
    );
    RAISE EXCEPTION 'disabled reviewer unexpectedly executed review transition';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
END
$assert_disabled_reviewer_denied$;

RESET ROLE;

DO $assert_audit_retention$
BEGIN
  BEGIN
    DELETE FROM recruitment.founding_partner_applications
    WHERE id = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa';
    RAISE EXCEPTION 'application deletion unexpectedly removed referenced review history';
  EXCEPTION WHEN foreign_key_violation THEN
    NULL;
  END;
  IF NOT EXISTS (
    SELECT 1 FROM recruitment.founding_partner_review_events
    WHERE application_id = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa'
      AND new_status = 'Screening'
  ) THEN
    RAISE EXCEPTION 'review audit history was not retained';
  END IF;
END
$assert_audit_retention$;

DROP TRIGGER reject_review_event_insert ON recruitment.founding_partner_review_events;
DROP FUNCTION public.reject_review_event_insert();
ROLLBACK;
