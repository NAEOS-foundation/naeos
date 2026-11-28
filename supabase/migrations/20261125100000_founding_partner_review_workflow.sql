-- Copyright 2025 NAEOS contributors
-- SPDX-License-Identifier: Apache-2.0
-- Founding Partner review workflow: authenticated allowlisted reviewer access only.
-- Apply to staging after PR review; production is intentionally untouched.

create table if not exists recruitment.founding_partner_reviewers (
  user_id uuid primary key references auth.users(id) on delete cascade,
  display_name text,
  enabled boolean not null default true,
  created_at timestamptz not null default now()
);

create table if not exists recruitment.founding_partner_review_events (
  id uuid primary key default gen_random_uuid(),
  application_id uuid not null references recruitment.founding_partner_applications(id) on delete restrict,
  reviewer_id uuid not null references auth.users(id),
  previous_status text not null,
  new_status text not null,
  note text,
  created_at timestamptz not null default now(),
  constraint founding_partner_review_event_status_check check (
    previous_status in ('New Application','Screening','Shortlisted','Interview','Mutual Fit','Trial','Partnership Negotiation','Accepted','Rejected','Withdrawn','On Hold')
    and new_status in ('New Application','Screening','Shortlisted','Interview','Mutual Fit','Trial','Partnership Negotiation','Accepted','Rejected','Withdrawn','On Hold')
  )
);

create index if not exists founding_partner_review_events_application_idx
  on recruitment.founding_partner_review_events(application_id, created_at desc);

alter table recruitment.founding_partner_reviewers enable row level security;
alter table recruitment.founding_partner_review_events enable row level security;

revoke all on recruitment.founding_partner_reviewers from public, anon, authenticated;
revoke all on recruitment.founding_partner_review_events from public, anon, authenticated;
revoke all on recruitment.founding_partner_applications from public, anon, authenticated;

-- Keep the intake schema ACL restricted. A separate non-exposed schema hosts reviewer RPC implementations.
create schema if not exists recruitment_private;
revoke all on schema recruitment_private from public, anon;
grant usage on schema recruitment_private to authenticated;

-- No direct table SELECT is granted to authenticated; reads are returned only by allowlisted private RPCs.

create policy "reviewer can read own allowlist record"
  on recruitment.founding_partner_reviewers
  for select to authenticated
  using (user_id = (select auth.uid()) and enabled = true);

create policy "enabled reviewers can read applications"
  on recruitment.founding_partner_applications
  for select to authenticated
  using (exists (
    select 1 from recruitment.founding_partner_reviewers r
    where r.user_id = (select auth.uid()) and r.enabled = true
  ));

create policy "enabled reviewers can read review audit"
  on recruitment.founding_partner_review_events
  for select to authenticated
  using (exists (
    select 1 from recruitment.founding_partner_reviewers r
    where r.user_id = (select auth.uid()) and r.enabled = true
  ));


-- Privileged implementations live in the private, non-exposed recruitment_private schema.
-- Public PostgREST entrypoints below are SECURITY INVOKER wrappers only.
create or replace function recruitment_private.founding_partner_review_transition_internal(
  p_application_id uuid,
  p_new_status text,
  p_note text default null
)
returns boolean
language plpgsql
security definer
set search_path = pg_catalog, recruitment, auth
as $body$
declare
  v_reviewer uuid := auth.uid();
  v_old_status text;
  v_allowed boolean := false;
begin
  if v_reviewer is null or not exists (
    select 1 from recruitment.founding_partner_reviewers
    where user_id = v_reviewer and enabled = true
  ) then
    raise exception 'reviewer authorization required' using errcode = '42501';
  end if;

  select application_status into v_old_status
    from recruitment.founding_partner_applications
   where id = p_application_id
   for update;

  if v_old_status is null then
    raise exception 'application not found' using errcode = 'P0002';
  end if;

  v_allowed := case v_old_status
    when 'New Application' then p_new_status in ('Screening','Rejected','Withdrawn','On Hold')
    when 'Screening' then p_new_status in ('Shortlisted','Interview','Rejected','Withdrawn','On Hold')
    when 'Shortlisted' then p_new_status in ('Interview','Mutual Fit','Rejected','Withdrawn','On Hold')
    when 'Interview' then p_new_status in ('Mutual Fit','Trial','Rejected','Withdrawn','On Hold')
    when 'Mutual Fit' then p_new_status in ('Trial','Partnership Negotiation','Rejected','Withdrawn','On Hold')
    when 'Trial' then p_new_status in ('Partnership Negotiation','Accepted','Rejected','Withdrawn','On Hold')
    when 'Partnership Negotiation' then p_new_status in ('Accepted','Rejected','Withdrawn','On Hold')
    when 'On Hold' then p_new_status in ('Screening','Shortlisted','Interview','Mutual Fit','Trial','Partnership Negotiation','Rejected','Withdrawn')
    else false
  end;

  if not v_allowed then
    raise exception 'invalid application status transition' using errcode = '22023';
  end if;

  update recruitment.founding_partner_applications
     set application_status = p_new_status
   where id = p_application_id;

  insert into recruitment.founding_partner_review_events(
    application_id, reviewer_id, previous_status, new_status, note
  ) values (
    p_application_id, v_reviewer, v_old_status, p_new_status, nullif(left(trim(coalesce(p_note,'')),1000),'')
  );
  return true;
end;
$body$;

create or replace function public.founding_partner_review_transition(
  p_application_id uuid,
  p_new_status text,
  p_note text default null
)
returns boolean
language sql
security invoker
set search_path = pg_catalog, recruitment_private
as $wrapper$
  select recruitment_private.founding_partner_review_transition_internal(
    p_application_id, p_new_status, p_note
  );
$wrapper$;

create or replace function recruitment_private.founding_partner_review_context_internal()
returns jsonb
language plpgsql
security definer
set search_path = pg_catalog, recruitment, auth
as $body$
declare
  v_reviewer uuid := auth.uid();
  v_display_name text;
begin
  if v_reviewer is null then
    raise exception 'reviewer authorization required' using errcode = '42501';
  end if;

  select display_name into v_display_name
    from recruitment.founding_partner_reviewers
   where user_id = v_reviewer and enabled = true;

  if not found then
    raise exception 'reviewer authorization required' using errcode = '42501';
  end if;

  return jsonb_build_object('user_id', v_reviewer, 'display_name', v_display_name);
end;
$body$;

create or replace function public.founding_partner_review_context()
returns jsonb
language sql
security invoker
set search_path = pg_catalog, recruitment_private
as $wrapper$
  select recruitment_private.founding_partner_review_context_internal();
$wrapper$;

create or replace function recruitment_private.founding_partner_review_list_internal(
  p_limit integer default 50,
  p_offset integer default 0,
  p_status text default null
)
returns jsonb
language plpgsql
security definer
set search_path = pg_catalog, recruitment, auth
as $body$
declare
  v_reviewer uuid := auth.uid();
  v_items jsonb;
  v_total bigint;
begin
  if v_reviewer is null or not exists (
    select 1 from recruitment.founding_partner_reviewers
    where user_id = v_reviewer and enabled = true
  ) then
    raise exception 'reviewer authorization required' using errcode = '42501';
  end if;

  if p_limit is null or p_limit < 1 or p_limit > 100
     or p_offset is null or p_offset < 0 then
    raise exception 'invalid pagination parameters' using errcode = '22023';
  end if;

  if p_status is not null and p_status not in (
    'New Application','Screening','Shortlisted','Interview','Mutual Fit','Trial',
    'Partnership Negotiation','Accepted','Rejected','Withdrawn','On Hold'
  ) then
    raise exception 'invalid application status filter' using errcode = '22023';
  end if;

  select coalesce(jsonb_agg(
      jsonb_build_object(
        'id', page_row.id,
        'applicant_name', page_row.applicant_name,
        'primary_role', page_row.primary_role,
        'application_status', page_row.application_status,
        'created_at', page_row.created_at
      ) order by page_row.created_at desc
    ), '[]'::jsonb)
    into v_items
    from (
      select a.id, a.applicant_name, a.primary_role, a.application_status, a.created_at
        from recruitment.founding_partner_applications a
       where p_status is null or a.application_status = p_status
       order by a.created_at desc, a.id
       limit p_limit offset p_offset
    ) as page_row;

  select count(*) into v_total
    from recruitment.founding_partner_applications a
   where p_status is null or a.application_status = p_status;

  return jsonb_build_object('items', v_items, 'total', v_total, 'limit', p_limit, 'offset', p_offset);
end;
$body$;

create or replace function public.founding_partner_review_list(
  p_limit integer default 50,
  p_offset integer default 0,
  p_status text default null
)
returns jsonb
language sql
security invoker
set search_path = pg_catalog, recruitment_private
as $wrapper$
  select recruitment_private.founding_partner_review_list_internal(p_limit, p_offset, p_status);
$wrapper$;

create or replace function recruitment_private.founding_partner_review_detail_internal(p_application_id uuid)
returns jsonb
language plpgsql
security definer
set search_path = pg_catalog, recruitment, auth
as $body$
declare
  v_reviewer uuid := auth.uid();
  v_application jsonb;
  v_events jsonb;
begin
  if v_reviewer is null or not exists (
    select 1 from recruitment.founding_partner_reviewers
    where user_id = v_reviewer and enabled = true
  ) then
    raise exception 'reviewer authorization required' using errcode = '42501';
  end if;

  select to_jsonb(a) into v_application
    from recruitment.founding_partner_applications a
   where a.id = p_application_id;

  if v_application is null then
    raise exception 'application not found' using errcode = 'P0002';
  end if;

  select coalesce(jsonb_agg(to_jsonb(e) order by e.created_at asc), '[]'::jsonb)
    into v_events
    from recruitment.founding_partner_review_events e
   where e.application_id = p_application_id;

  return jsonb_build_object('application', v_application, 'events', v_events);
end;
$body$;

create or replace function public.founding_partner_review_detail(p_application_id uuid)
returns jsonb
language sql
security invoker
set search_path = pg_catalog, recruitment_private
as $wrapper$
  select recruitment_private.founding_partner_review_detail_internal(p_application_id);
$wrapper$;

-- Private implementations are callable only by authenticated sessions through
-- the public SECURITY INVOKER wrappers; the recruitment schema is not exposed.
revoke all on function recruitment_private.founding_partner_review_transition_internal(uuid,text,text) from public, anon;
revoke all on function recruitment_private.founding_partner_review_context_internal() from public, anon;
revoke all on function recruitment_private.founding_partner_review_list_internal(integer,integer,text) from public, anon;
revoke all on function recruitment_private.founding_partner_review_detail_internal(uuid) from public, anon;
grant execute on function recruitment_private.founding_partner_review_transition_internal(uuid,text,text) to authenticated;
grant execute on function recruitment_private.founding_partner_review_context_internal() to authenticated;
grant execute on function recruitment_private.founding_partner_review_list_internal(integer,integer,text) to authenticated;
grant execute on function recruitment_private.founding_partner_review_detail_internal(uuid) to authenticated;

revoke all on function public.founding_partner_review_transition(uuid,text,text) from public, anon;
revoke all on function public.founding_partner_review_context() from public, anon;
revoke all on function public.founding_partner_review_list(integer,integer,text) from public, anon;
revoke all on function public.founding_partner_review_detail(uuid) from public, anon;
grant execute on function public.founding_partner_review_transition(uuid,text,text) to authenticated;
grant execute on function public.founding_partner_review_context() to authenticated;
grant execute on function public.founding_partner_review_list(integer,integer,text) to authenticated;
grant execute on function public.founding_partner_review_detail(uuid) to authenticated;
