# NAEOS Social Signal Engine

The social signal endpoint turns meaningful repository activity into reviewable social-post drafts.

## Endpoint

`GET /api/social/draft`

The endpoint reads current GitHub state and returns:

- open issues and bug-labelled issues
- open pull requests
- recently merged pull requests
- recent workflow state
- latest commit
- latest release
- candidate signal classification
- LinkedIn and X draft copy

It deliberately does **not** publish content or mutate the repository.

## Signal policy

Not every repository event becomes a social post.

| Event | Candidate |
| --- | --- |
| Small commit | No |
| Comment / review | No |
| Release | Yes |
| Meaningful PR merged | Yes |
| Bug opened/visible | Candidate |
| CI failure | Candidate |
| Routine CI retry | No |
| Weekly engineering summary | Future |

## Publishing boundary

The intended flow is:

```
GitHub event/state
      ↓
Social Signal Engine
      ↓
Draft
      ↓
Policy / human review
      ↓
Social publisher
      ↓
External receipt
```

The current implementation stops at the draft boundary. A publisher integration must be connected separately.

## Why this is separate from Status

Status answers: "What is the current engineering state?"

Social answers: "Which meaningful change is worth communicating?"

Both can consume the same GitHub source of truth without creating commits on protected `main`.

## Future webhook mode

A future GitHub webhook can move this from request-time polling to event-driven processing. GitHub supports repository webhooks for issues, pull requests, releases, workflow-related events, and other repository activity.

Webhook delivery should be authenticated, deduplicated using the GitHub delivery ID, and processed asynchronously before a social publication is considered.


## Policy gate v1

The draft endpoint evaluates a separate social policy before returning a candidate:

- `release`, `merged-pr`, `bug`, and `ci-failure` can produce a reviewable candidate.
- A valid GitHub source is required.
- The response uses `candidate` to describe a reviewable proposal; it is not publication authorization.
- The policy decision is `review_required`, not publish authorization.
- `authorized` is always `false` in v1.
- Unsupported or unavailable source state fails closed with `deny`.
- The response records the policy version and decision reasons so proposal, authorization, execution, and evidence remain distinct.
- Error responses also include the fail-closed policy decision.

The effective boundary is:

```
GitHub state
    ↓
Signal classification
    ↓
Draft proposal
    ↓
Policy evaluation
    ↓
Human approval
    ↓
Publisher
    ↓
External receipt
```

No social network credentials, publisher calls, or automatic publication are introduced by this change.
