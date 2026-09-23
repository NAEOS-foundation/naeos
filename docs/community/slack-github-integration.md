# NAEOS Slack + GitHub Integration

## Purpose

NAEOS uses Slack as the real-time coordination layer and GitHub as the durable engineering source of truth.

> Slack coordinates. GitHub records. CI verifies. Evidence closes the loop.

## Workspace setup

A Slack workspace administrator should install the official GitHub integration for Slack:

- https://slack.com/apps/A01BP7R4KNY-github

After installation:

1. Connect the administrator's GitHub account with `/github signin`.
2. Subscribe the relevant channels to `NAEOS-foundation/naeos` with:
   `/github subscribe NAEOS-foundation/naeos`
3. Add GitHub to private channels explicitly when required.
4. Tune notifications so high-signal engineering events are visible without flooding community channels.

The integration supports repository notifications, GitHub issue and pull-request collaboration, rich link previews, and GitHub actions from Slack.

## Recommended channel routing

| Slack channel | GitHub activity |
|---|---|
| `#community` | Selected discussions, issue links, contributor questions |
| `#engineering` | Issues, PRs, CI, implementation work |
| `#architecture` | Architecture discussions and design decisions |
| `#policy-governance` | Policy and authorization issues/PRs |
| `#audit-verification` | Evidence and verification work |
| `#runtime` | Runtime and execution-boundary work |
| `#ai-agents` | Agent integration and workflow work |
| `#handoffs` | Handoff contracts and boundary validation |

Do not mirror every GitHub event into every channel. Route only the activity that helps the channel perform its purpose.

## Community funnel

The public website should point visitors to a Slack shared-invite URL:

`Website → Slack invite → #community → technical discussion → GitHub issue/PR → CI → evidence`

Use a Slack shared-invite URL rather than the workspace home URL for public acquisition. Shared invite links can be rotated independently of the NAEOS repository.

## Source of truth

- Conversation: Slack
- Durable decision: GitHub issue/discussion
- Implementation: GitHub PR
- Verification: GitHub Actions
- Evidence: repository artifacts and experiment output

The Slack conversation should not be treated as the authoritative record of an engineering decision when the decision affects the repository.
