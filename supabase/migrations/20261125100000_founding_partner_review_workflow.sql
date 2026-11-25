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

-- Authenticated must resolve objects in this private schema; table privileges and RLS still limit access.
grant usage on schema recruitment to authenticated;

grant select on recruitment.founding_partner_reviewers to authenticated;
grant select on recruitment.founding_partner_review_events to authenticated;
grant select on recruitment.founding_partner_applications to authenticated;

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

create or replace function public.founding_partner_review_transition(
  p_application_id uuid,
  p_new_status text,
  p_note text default null
)
returns boolean
language plpgsql
security definer
set search_path = pg_catalog, recruitment
as $$
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
$$;

revoke all on function public.founding_partner_review_transition(uuid,text,text) from public, anon;
grant execute on function public.founding_partner_review_transition(uuid,text,text) to authenticated;

-- Reviewer bootstrap must be performed by a trusted project owner through SQL editor:
-- insert into recruitment.founding_partner_reviewers(user_id, display_name)
-- values ('<AUTH_USER_UUID>', 'Founder Reviewer');
-- Never expose reviewer bootstrap via a public endpoint.
