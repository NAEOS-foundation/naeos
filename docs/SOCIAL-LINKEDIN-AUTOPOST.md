# LinkedIn Organization Auto-Posting

Status: P0 implementation baseline

## Flow

`content-calendar.json` → `social-post.yml` → `scripts/social-post.sh` → `scripts/linkdin-post.sh` → LinkedIn Posts API.

The workflow is scheduled daily and can also be started manually. LinkedIn publication is an externally visible action, so the runtime fails closed when the required credentials are missing or LinkedIn rejects the request.

## GitHub Actions secrets

Configure these repository secrets:

- `LINKEDIN_ACCESS_TOKEN`: OAuth access token authorized for organization posting.
- `LINKEDIN_ORGANIZATION_URN`: NAEOS organization URN, for example `urn:li:organization:123456`.

The token must have `w_organization_social` access and the authenticated member must have a LinkedIn Page role that is authorized to create organization posts.

Do not commit access tokens or put them in repository files.

## Manual test

Use **Actions → social-post → Run workflow**.

For the first verification, set:

- `platforms`: `linkedin`
- `date`: a calendar date that has a LinkedIn entry
- `dry_run`: `true`

After the dry run is verified, repeat with `dry_run=false`.

## Governance boundary

The content calendar is the proposal/source content. The GitHub Actions job is the runtime executor. A successful LinkedIn HTTP response is the execution observation; it is not inferred from the presence of a calendar entry.

Do not enable LinkedIn for every commit. Publish only calendar-approved content or another explicitly authorized event source.

## Failure behavior

The publisher exits non-zero when:

- the access token is missing;
- the organization URN is missing;
- the LinkedIn API returns a non-success response.

This prevents a failed or unauthorized external publication from being represented as successful.

## External authorization

LinkedIn organization posting requires the LinkedIn Community Management API access and the `w_organization_social` permission. Authorization is granted to the member through LinkedIn OAuth; the repository stores only the resulting secret, not the authorization flow itself.

See the official LinkedIn documentation for current access and organization authorization requirements.
