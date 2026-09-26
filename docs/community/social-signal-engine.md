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
Authorization
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

## Authorization contract v1

Policy evaluation and authorization are separate contracts.

The authorization contract requires all of the following before `authorized: true` can be returned:

1. policy decision is `review_required`
2. policy version matches the required policy version
3. an explicit approval record exists
4. approval identity, timestamp, and approval ID are present
5. approval binds to the same proposal, action, target, and policy version

Any missing or mismatched field fails closed with `deny`. The contract does not publish content and does not execute an external action.

The contract is intentionally proposal-bound:

```
proposal
   ↓
policy decision
   ↓
explicit approval
   ↓
authorization contract
   ↓
authorized execution
```

An authorization result is not evidence that execution happened. Execution and its external receipt remain separate stages for the subsequent decision-record and publisher work.

No social network credentials, publisher calls, or automatic publication are introduced by this change.

## Human approval boundary v1

The human approval boundary is explicit and single-operator compatible. NAEOS does not assume multiple maintainers.

Before authorization can accept a social action, a human approval record must bind:

- approvalId
- approverId and approverIdentity
- explicit approve or reject decision
- approval timestamp
- proposal ID
- action
- target system/resource
- policy version
- policy decision ID

Missing or blank binding fields fail closed. A rejection is never accepted as approval.

Authorization-v1 invokes the human-approval-v1 evaluator before returning `authorized: true`, then independently rechecks proposal, action, target, policy-version, and decision-ID bindings. This makes the human boundary an authorization prerequisite rather than documentation-only metadata.

The boundary is:

```
proposal
   ↓
policy decision
   ↓
explicit human approval
   ↓
authorization contract
   ↓
execution
   ↓
receipt
```

### Identity limitation

`human-approval-v1` records approval provenance fields, but `authenticated` is explicitly `false`. The contract therefore does **not** claim cryptographic or platform-authenticated proof that a particular person performed the approval.

The current deployment may use one maintainer as the human approver. That is a real human review boundary, but it is not a multi-party approval system and it must not be represented as one.

An agent or model may propose an approval payload, but the payload itself is not proof that the human performed the approval. An authenticated approval mechanism is a later hardening step.

## Publisher dry-run and receipt v1

The publisher boundary is now represented by a dry-run contract. It accepts an action only after authorization-v1 has returned authorized: true.

The dry-run contract:

- validates the authorization contract version at runtime
- validates proposal, action, target, policy version, decision ID, receipt ID, and execution timestamp
- fails closed when authorization or required bindings are missing or unsupported
- never calls a social provider
- returns externalEffect: false for both simulated and rejected outcomes
- emits a publisher-receipt-v1 observation that distinguishes simulation from rejection

The intended boundary is now:

```
proposal
   ↓
policy decision
   ↓
human approval
   ↓
authorization
   ↓
publisher dry-run
   ↓
receipt / observation
   ↓
future external publisher
```

A dry-run receipt is evidence of the NAEOS simulation boundary, not evidence that a social network accepted or published content. An external publisher must remain a later, separately authorized integration and must return its own provider receipt.

No social credentials, network publishing calls, or automatic external publication are introduced by publisher-receipt-v1.


## External publisher boundary v1

P5 defines a provider-neutral handoff immediately before any future external publisher adapter.

The `external-publisher-v1` contract requires:

- an execution ID
- an explicit provider identifier
- `authorization-v1` with `authorized: true`
- the same proposal, action, target, policy version, and decision ID bindings
- a provider receipt that can later be verified independently of authorization

The contract exposes an `ExternalPublisherAdapter` interface but does not instantiate or call an adapter. The preparation function always reports `externalEffect: false`; it is a boundary check, not publication authorization.

A future runtime must revalidate authorization and all action bindings at the last controllable boundary immediately before invoking a provider. A provider response is observation evidence and never replaces the NAEOS policy or authorization decision.

Provider receipt validation requires the expected provider, a non-empty provider receipt ID and observation timestamp, and a status other than `unknown`. This prevents an ambiguous provider response from being treated as execution evidence.

The intended architecture is:

```
proposal
   ↓
policy decision
   ↓
human approval
   ↓
authorization-v1
   ↓
publisher-receipt-v1 dry-run
   ↓
external-publisher-v1 handoff
   ↓
future provider adapter
   ↓
provider receipt / observation
```

P5 does **not** add social credentials, provider HTTP calls, automatic publication, or a production provider adapter. Those remain a separate implementation and authorization decision.
