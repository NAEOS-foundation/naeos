# NAEOS GitHub Discussions Guide

GitHub Discussions is the main synchronous community surface for NAEOS.
It is the place to ask questions, share ideas, show what you built, and follow
the project's thinking — before issues become formal work items.

## Guidelines

- **Search first.** Before posting, search existing discussions in the
  category you plan to use. Up-vote with reactions instead of duplicating.
- **One topic per discussion.** Keep threads focused so answers stay useful.
- **Be specific.** Prefer minimal specifications, commands, and error output
  over prose alone. Includes `naeos doctor` output when relevant.
- **Respect the Code of Conduct.** See
  [CODE_OF_CONDUCT.md](../../CODE_OF_CONDUCT.md).
- **No self-promotion spam.** Sharing a NAEOS project in Show and tell is
  welcome; repeating it everywhere is not.
- **Verifiable claims only.** Do not claim features, benchmarks, partnerships,
  or adoption that are not evidenced by the repository.

## Categories

| Category | Purpose |
|----------|---------|
| [Announcements](https://github.com/NAEOS-foundation/naeos/discussions/categories/announcements) | Maintainer releases and milestone notes. Read-only for most members. |
| [Founder Journal](https://github.com/NAEOS-foundation/naeos/discussions/categories/founder-journal) | Founder-led notes: architecture decisions, failures, lessons, open questions. Built in public. |
| [General](https://github.com/NAEOS-foundation/naeos/discussions/categories/general) | Anything that does not fit elsewhere. |
| [Ideas](https://github.com/NAEOS-foundation/naeos/discussions/categories/ideas) | Feature sketches and improvements. Use the form; the strongest ideas may become issues or RFCs. |
| [Q&A](https://github.com/NAEOS-foundation/naeos/discussions/categories/q-a) | Help and troubleshooting. Include version, environment, and what you tried. |
| [Show and tell](https://github.com/NAEOS-foundation/naeos/discussions/categories/show-and-tell) | Projects, plugins, profiles, integrations, and experiments built with NAEOS. |
| Polls | Informal community signals. |

Structured forms are defined in `.github/DISCUSSION_TEMPLATE/` and appear
automatically for Ideas, Q&A, and Show and tell when those categories are
created from the slug-matched templates.

## When to use Issues instead

Issues are for concrete, actionable work items: bugs, feature requests,
documentation gaps, plugin contributions. Discussions are for conversations.
Rule of thumb:

- **Question or idea → Discussions.**
- **Confirmed bug or agreed feature → Issue** (with a template).

## From discussion to contribution

Discussions feed the contributor ladder:

1. An **Idea** discussion that gains traction is triaged into the
   [roadmap](../../ROADMAP.md) or an RFC.
2. A **Q&A** answer can become a troubleshooting doc or a
   [Good first issue](https://github.com/NAEOS-foundation/naeos/labels/good%20first%20issue).
3. A **Show and tell** post can become a plugin, profile, or integration
   contribution.

## Moderation

Maintainers moderate with the Code of Conduct and these guidelines. Reported
content is handled privately and kindly.

## Related

- [Contributor ladder](contributor-ladder.md) — how community participation
  grows.
- [CONTRIBUTING.md](../../CONTRIBUTING.md) — engineering workflow.
- [Discord](https://discord.gg/naeos) — real-time chat (bot in
  `tools/discord-bot`).
Last updated: 2026-09-11
