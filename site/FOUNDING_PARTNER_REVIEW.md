# Founding Partner Review Dashboard

The reviewer dashboard is available at `/{lang}/founding-partners/review` (for example,
`/id/founding-partners/review`). It is a restricted workspace, not a public application form.

## Staging-only setup

1. Configure `NEXT_PUBLIC_SUPABASE_URL` and `NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY` with the
   **NAEOS staging Supabase project** values before building the site.
2. Do not use a service-role key or any secret in a `NEXT_PUBLIC_*` variable.
3. Create the founder reviewer account through the trusted Supabase staging Auth
   administration path. Public sign-up and reviewer self-enrollment are not implemented.
4. Obtain that Auth user's UUID from the staging Auth user record and insert it into
   `recruitment.founding_partner_reviewers` through the trusted staging SQL editor:

   ```sql
   insert into recruitment.founding_partner_reviewers (user_id, display_name, enabled)
   values ('<STAGING_AUTH_USER_UUID>', 'Founder Reviewer', true);
   ```

5. Never copy a production user UUID into staging or apply this migration to production.

## Security boundary

- The `recruitment` intake schema remains unexposed and retains its restricted ACL. The separate `recruitment_private` schema is also not exposed to PostgREST; authenticated sessions receive schema USAGE only to invoke its narrowly scoped reviewer functions.
- Browser requests use the public anon/publishable key plus the signed-in user's access
  token. Browser requests use narrowly scoped `public.founding_partner_review_*` SECURITY INVOKER RPC wrappers. Those wrappers delegate to SECURITY DEFINER implementations in the non-exposed `recruitment_private` schema.
- Authenticated clients receive no direct SELECT/INSERT/UPDATE/DELETE grants on recruitment tables. Each private implementation checks `auth.uid()` against the enabled reviewer allowlist. The public wrappers do not execute with elevated privileges; an authenticated account alone is insufficient.
- The dashboard does not persist access tokens; refresh requires signing in again.
- The status transition RPC remains the only write path and appends the audit event
  atomically. Invalid transitions and audit-write failures roll back.
- This page is not a substitute for staging verification, browser E2E, or a production
  security review. Keep PR #886 Draft until those gates and dependency PR #885 are complete.
