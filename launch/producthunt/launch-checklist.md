# Launch Checklist — Pre, Day-of, Post

A step-by-step timeline for a successful Product Hunt launch. PH launches are most effective on **Tuesdays–Thursdays**; publish your post between **00:01–07:00 PT** so it rides the full US day. Work backwards from your chosen launch date.

---

## Pre-launch — D-7 to D-1

### D-7
- [ ] Choose launch date (Tue/Wed/Thu recommended) and time (00:01–07:00 PT)
- [ ] Register the Product Hunt account that will post (one product per hunter account — do not post from a company account)
- [ ] Complete profile per `maker-bio.md` (photo, headline, website, links)
- [ ] Assemble gallery images per `listing.md` §3 (hero, pipeline diagram, CLI screenshots, AI compiler output, LSP/VS Code shot)
- [ ] Export logo as 240×240px PNG (transparent or brand background) from `brand/logo.svg`
- [ ] Draft the listing in Product Hunt's saved-draft feature: name, tagline, description (`listing.md`), topics, links, gallery

### D-6 to D-5
- [ ] Verify all links resolve: `https://naeos.dev`, `https://docs.naeos.dev`, GitHub repo, releases page
- [ ] Re-test the quick start end-to-end on a clean machine:
  ```bash
  curl -fsSL https://naeos.dev/install.sh | sh
  naeos create my-app
  naeos run --input-file spec.yaml
  ```
- [ ] Confirm `naeos version` reports v3.0.0 and the latest release is tagged on GitHub
- [ ] Have 1–2 community members (not you) dry-run the quick start and note friction — fix anything they trip on

### D-4 to D-2 — rally your audience
- [ ] Create a private list of supporters (Discord, X, LinkedIn, email) who will upvote on launch day
- [ ] Send teaser messages ("launching on Product Hunt this Thursday — would love your support") — no public posts yet
- [ ] Prepare the X thread and LinkedIn post from `social-posts.md`, ready to publish at 00:01 PT
- [ ] Prepare community copies (Discord, HN, Reddit, Indie Hackers) from `social-posts.md`

### D-1
- [ ] Final review of the saved draft: tagline ≤60 chars, description hook intact, 3+ gallery images, links correct
- [ ] Pick a launch date/time with your team — note launch day is Pacific Time, not local
- [ ] Prepare the first comment (`first-comment.md`) in a text file for instant paste
- [ ] Brief your support list: exact link, exact time, "upvote + comment if you can"
- [ ] Schedule the X thread / LinkedIn post for 00:01–00:30 PT (or post manually)
- [ ] Get a good night's sleep — launch day is all-day engagement

---

## Launch Day — D-day

### 00:01–00:30 PT (submit window)
- [ ] Publish the PH post (Product Hunt lets you set a launch date — set it to today, hit publish)
- [ ] Paste the first comment immediately (`first-comment.md`)
- [ ] Post the X thread and LinkedIn announcement
- [ ] Send the "we're live" message to your support list with the direct link

### 00:30–09:00 PT (US East Coast wake-up)
- [ ] Check in on the post — reply to every single comment within minutes
- [ ] Update the first comment if there are obvious questions (link to docs, FAQ)
- [ ] Ask early upvoters to share if they genuinely like it — never ask for spammy upvotes

### 09:00–17:00 PT (peak traffic)
- [ ] Stay present: reply to every comment, thank every supporter, answer technical questions in depth
- [ ] Post community copies (Discord, HN, Reddit, Indie Hackers) — see `social-posts.md`
- [ ] Share the post in relevant Discord/Slack communities you're active in (respect each community's self-promo rules)
- [ ] Pin a helpful link in comments (docs, quick start, GitHub) — comment, not gallery image
- [ ] Recruit 1–2 team members to help respond while you rotate

### 17:00–23:59 PT (wrap-up)
- [ ] Do a final pass on unanswered comments
- [ ] Save the top questions/feedback — this is product research, log it
- [ ] Post a "day one wrap-up" comment with thanks and a teaser of what's next

### Launch-day rules to remember
- No links inside gallery images (PH strips them)
- No paid promotion for votes (PH bans it — account risk)
- Don't edit the tagline/description after the first few hours — momentum beats polish
- Track rank and upvotes in the first 24h; the first 6 hours matter most

---

## Post-launch — D+1 to D+7

### D+1
- [ ] Post a thank-you comment / thread: numbers (honest), top feedback themes, what's next
- [ ] Send a thank-you message to everyone who commented or upvoted
- [ ] Screenshot the final stats (upvotes, comments, rank) for the recap

### D+2 to D+7
- [ ] Publish a launch recap blog post on `site/content/blog/` (e.g., "We launched NAEOS on Product Hunt") — link it in the repo and socials
- [ ] Turn the top 3–5 feedback items into GitHub issues with labels, link them in a follow-up comment
- [ ] Answer late comments; PH notifications keep coming for a week
- [ ] Cross-post the recap on X, LinkedIn, and the community channels
- [ ] Update `ROADMAP.md` / `DEVELOPMENT_PLAN.md` with any new commitments made during the launch

## Metrics to track

| Metric | Target |
|--------|--------|
| Upvotes (24h) | 100+ (strong), 300+ (very strong for dev tools) |
| Comments (24h) | 20+ |
| Website visits from PH | 500+ |
| GitHub stars added | 2–5× the upvote count is not realistic — set 50–150 |
| Installs / downloads | 100+ |
| Demo/quick-start completions | 50% of installs |