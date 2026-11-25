-- Atomic intake rate-limit enforcement (service-role only).
create or replace function recruitment.enforce_intake_rate_limit(
  p_ip_hash text,
  p_max_submissions integer default 5,
  p_window_seconds integer default 3600
)
returns boolean
language plpgsql
security definer
set search_path = pg_catalog, recruitment
as $$
declare
  recent_count integer;
begin
  if p_ip_hash is null or p_ip_hash = '' or p_max_submissions < 1 or p_window_seconds < 1 then
    return false;
  end if;

  perform pg_advisory_xact_lock(hashtextextended(p_ip_hash, 0));
  select count(*) into recent_count
    from recruitment.intake_rate_limits
   where ip_hash = p_ip_hash
     and created_at >= now() - make_interval(secs => p_window_seconds);

  if recent_count >= p_max_submissions then
    return false;
  end if;

  insert into recruitment.intake_rate_limits(ip_hash) values (p_ip_hash);
  return true;
end;
$$;

revoke all on function recruitment.enforce_intake_rate_limit(text, integer, integer) from public, anon, authenticated;
grant execute on function recruitment.enforce_intake_rate_limit(text, integer, integer) to service_role;
