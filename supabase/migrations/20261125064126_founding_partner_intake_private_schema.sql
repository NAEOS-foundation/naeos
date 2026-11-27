-- Founding Partner intake: private schema, service-role-only access.
-- Applied to the NAEOS staging project (naeos.db) on 2026-11-25.
create schema if not exists recruitment;
revoke all on schema recruitment from public, anon, authenticated;
grant usage on schema recruitment to service_role;

create table if not exists recruitment.founding_partner_applications (
  id uuid primary key default gen_random_uuid(),
  applicant_name text not null check (length(applicant_name) between 2 and 120),
  email text not null check (length(email) <= 254),
  location text,
  linkedin_url text,
  portfolio_url text,
  professional_role text,
  primary_role text not null check (primary_role in ('Technical Founding Partner / CTO','AI Systems & Engineering Partner','Product & Engineering Strategy Partner','Commercial & GTM Partner','Governance & Security Specialist')),
  alternate_role text,
  expertise_areas text[] not null default '{}',
  relevant_experience text not null,
  portfolio_summary text,
  contribution_proposal text not null,
  motivation text not null,
  weekly_hours text not null,
  collaboration_model text not null,
  earliest_availability date,
  trial_available boolean not null,
  partnership_expectations text not null,
  equity_expectations text,
  consent boolean not null check (consent = true),
  consent_at timestamptz not null default now(),
  application_status text not null default 'New Application' check (application_status in ('New Application','Screening','Shortlisted','Interview','Mutual Fit','Trial','Partnership Negotiation','Accepted','Rejected','Withdrawn','On Hold')),
  created_at timestamptz not null default now()
);

create table if not exists recruitment.intake_rate_limits (
  id uuid primary key default gen_random_uuid(),
  ip_hash text not null,
  created_at timestamptz not null default now()
);
create index if not exists founding_partner_applications_created_idx on recruitment.founding_partner_applications(created_at desc);
create index if not exists intake_rate_limits_hash_time_idx on recruitment.intake_rate_limits(ip_hash, created_at desc);

alter table recruitment.founding_partner_applications enable row level security;
alter table recruitment.intake_rate_limits enable row level security;
revoke all on all tables in schema recruitment from public, anon, authenticated;
grant insert on recruitment.founding_partner_applications to service_role;
grant select, insert, delete on recruitment.intake_rate_limits to service_role;
