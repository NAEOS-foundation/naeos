# NAEOS Product Hunt Launch Kit

Everything you need to launch NAEOS on Product Hunt — listing content, maker bio, first comment, checklist, and social copy. All content is in English (Product Hunt's standard), with placeholders where personal details are needed.

## Contents

| File | What it's for |
|------|---------------|
| `listing.md` | Core submission: name, tagline options (≤60 chars), publish-ready description, topics, gallery image plan, links |
| `first-comment.md` | Maker's kickoff comment to anchor the discussion at 00:01 PT |
| `maker-bio.md` | PH profile: name, links, bio options, about section (fill in placeholders) |
| `launch-checklist.md` | Full timeline: D-7 pre-launch, launch-day hour-by-hour, D+1–D+7 post-launch, metrics |
| `social-posts.md` | Ready-to-post copy: X thread, LinkedIn, Discord/Slack, HN Show HN, Reddit, supporter email |
| `reddit-posts.md` | Per-subreddit launch posts (r/golang, r/devops, r/programming) written as developer stories, not ads |
| `support-emails.md` | D-1 support-list emails: teaser, co-maker invite, launch-day brief (personalizable) |
| `assets/` | 5 gallery images (1600×900 PNG) + 240×240 logo — brand-accurate, ready to upload |
| `assets/generate.js` | Script that renders all images (uses `sharp`); regenerate after brand/CLI changes |
| `../discord-server/pre-launch.md` | Pre-launch content to grow the Discord community: social teasers, Discord announcements, Launch Champion call, sneeak-peek series (D-2 → D-0) |

## Quick start

1. Fill in the placeholders in `maker-bio.md` and the `[PH LINK]` entries in `social-posts.md`.
2. Copy the listing from `listing.md` into Product Hunt's saved-draft form.
3. Upload `assets/logo-240.png` as the product logo and the five `assets/0*.png` as gallery images (see `listing.md` §3).
4. Follow `launch-checklist.md` starting at D-7.

## Supporting material (already in the repo)

- **Brand assets** — `brand/` (logo variants, `brand.json` colors: dark `#05050a`, gradient `#08d6ff → #9333ea`)
- **Website** — `site/` (Hugo site at https://naeos.dev)
- **Launch-adjacent blog posts** — `site/content/blog/`:
  - `v3.1.0-release.md` — the release being launched
  - `why-declarative-engineering.md` — explainer for the core idea
  - `ai-driven-development.md` — the AI integration story
- **Discord server** — `launch/discord-server/` (`blueprint.md` server setup, `templates.md` paste-ready channel templates, `pre-launch.md` community growth content)
- **Slack workspace** — `launch/slack-server/blueprint.md` (channel structure, roles, onboarding, paste-ready messages)
- **Whitepaper** — `WHITEPAPER-EN.md` (English) / `WHITEPAPER.md` (Bahasa Indonesia)
- **Screenshots source** — gallery images are pre-generated from real CLI output in `assets/`; regenerate with `node launch/producthunt/assets/generate.js`

## Notes

- Launch day must be **Tue–Thu** for best visibility; publish between **00:01–07:00 PT**.
- One Product Hunt account per product — launch from a personal hunter account, not a company account.
- Never put links inside gallery images (PH strips them).
- After launch, consider writing a recap post in `site/content/blog/` and linking it from the repo README.