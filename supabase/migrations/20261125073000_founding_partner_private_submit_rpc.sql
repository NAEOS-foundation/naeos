-- Private Founding Partner submission RPC. Tables remain in the non-exposed recruitment schema.
create or replace function public.founding_partner_intake_submit(
  p_ip_hash text,
  p_application jsonb
)
returns boolean
language plpgsql
security definer
set search_path = pg_catalog, recruitment
as $$
declare
  recent_count integer;
begin
  if p_ip_hash is null or p_ip_hash = '' or p_application is null or jsonb_typeof(p_application) <> 'object' then
    return false;
  end if;
  perform pg_advisory_xact_lock(hashtextextended(p_ip_hash, 0));
  select count(*) into recent_count
    from recruitment.intake_rate_limits
   where ip_hash = p_ip_hash
     and created_at >= now() - interval '1 hour';
  if recent_count >= 5 then
    return false;
  end if;
  insert into recruitment.intake_rate_limits(ip_hash) values (p_ip_hash);
  insert into recruitment.founding_partner_applications(
    applicant_name,email,location,linkedin_url,portfolio_url,professional_role,primary_role,alternate_role,
    expertise_areas,relevant_experience,portfolio_summary,contribution_proposal,motivation,weekly_hours,
    collaboration_model,earliest_availability,trial_available,partnership_expectations,equity_expectations,
    consent,consent_at
  ) values (
    p_application->>'applicant_name',p_application->>'email',p_application->>'location',
    p_application->>'linkedin_url',p_application->>'portfolio_url',p_application->>'professional_role',
    p_application->>'primary_role',p_application->>'alternate_role',
    coalesce(array(select jsonb_array_elements_text(p_application->'expertise_areas')), '{}'),
    p_application->>'relevant_experience',p_application->>'portfolio_summary',
    p_application->>'contribution_proposal',p_application->>'motivation',p_application->>'weekly_hours',
    p_application->>'collaboration_model',nullif(p_application->>'earliest_availability','')::date,
    true,p_application->>'partnership_expectations',p_application->>'equity_expectations',true,now()
  );
  return true;
end;
$$;
revoke all on function public.founding_partner_intake_submit(text,jsonb) from public, anon, authenticated;
grant execute on function public.founding_partner_intake_submit(text,jsonb) to service_role;
