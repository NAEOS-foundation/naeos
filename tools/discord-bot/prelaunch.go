package main

// preLaunchMessages holds the paste-ready pre-launch announcements that can be
// posted with /announce. Content mirrors launch/discord-server/pre-launch.md.
var preLaunchMessages = map[string]string{
	"welcome": "## NAEOS is launching Tuesday 🚀\n\n" +
		"On **Tuesday, 18 Aug** we launch NAEOS on Product Hunt.\n\n" +
		"If you're new here: NAEOS is a declarative engineering platform. You describe\n" +
		"your system **once** in YAML/JSON — it builds an internal engineering model\n" +
		"(NEIR) and generates validated code in Go, TypeScript, Python, Java, Rust,\n" +
		"plus AI instruction sets for 6 tools and an MCP server.\n\n" +
		"**What's happening this week:**\n" +
		"- 👀 **Tonight/Monday** — community walkthrough of v3.1.0 (caching, profiling, architecture patterns)\n" +
		"- 🗳️ **Roadmap voting** — tell us what ships next (see #roadmap thread)\n" +
		"- 🏆 **Become a Launch Champion** — help us launch, get the role (see below)\n" +
		"- 🎉 **Launch day** — watch party + support thread in #product-hunt\n\n" +
		"Try NAEOS right now:\n" +
		"```bash\ncurl -fsSL https://naeos.dev/install.sh | sh\nnaeos create\ncd my-app && naeos run --input-file spec.yaml\n```\n" +
		"Links: https://naeos.dev · https://docs.naeos.dev · https://github.com/NAEOS-foundation/naeos",

	"champion": "## Become a @Launch Champion 🏆\n\n" +
		"We launch Tuesday and want 5–10 Launch Champions to help us make it loud.\n\n" +
		"**What you'd do (pick any):**\n" +
		"- Upvote + comment on the PH post within the first hour\n" +
		"- Share NAEOS with one dev community you're part of\n" +
		"- Test the quick start and report friction\n" +
		"- Hang in #launch-party and help answer questions\n\n" +
		"**What you get:**\n" +
		"- 🏅 Exclusive @Launch Champion role (permanent, not just launch week)\n" +
		"- 👀 Early access to roadmap + post-launch features\n" +
		"- 🎁 A shout-out in the launch recap\n\n" +
		"Reply to this message with **\"I'm in\"** + which part you'll help with.\n" +
		"Everyone who participates keeps the role. No brigading, no spam — just an\n" +
		"honest first-day boost from people who actually use NAEOS.",

	"walkthrough": "## Tonight: NAEOS v3.1.0 — community walkthrough 🎁\n\n" +
		"Before tomorrow's launch, here's what the community gets first:\n\n" +
		"**🚀 Pipeline caching** — `naeos run --cache-dir .naeos-cache` skips stages\n" +
		"that haven't changed. Incremental builds in seconds.\n\n" +
		"**📊 Run-level profiling** — `naeos run --profile --profile-out profile.json`\n" +
		"shows exactly where the pipeline spends time. `--pprof` for the Go-nerd view.\n\n" +
		"**🏗️ Architecture patterns** — set `architecture.pattern` in your spec\n" +
		"(monolithic, microservices, or serverless) and the NEIR model + codegen adapts to it.\n\n" +
		"**🧩 WASM plugins** — hardened plugin SDK + official examples.\n\n" +
		"Try it, then vote on what we should build next in #roadmap. The top request\n" +
		"becomes the first post-launch issue.",

	"today": "## Today's the day 🎉\n\n" +
		"NAEOS launches on Product Hunt at **14:00 WIB**.\n\n" +
		"When it's live: upvote + comment here → **https://www.producthunt.com/posts/naeos**\n" +
		"(a comment helps more than an upvote — even \"what does this do?\").\n\n" +
		"#launch-party voice opens at launch. See you there.",

	"golive": "🎉 **NAEOS is live on Product Hunt!**\n\n" +
		"We're launching today. If NAEOS has helped you — or you're curious about\n" +
		"spec-driven engineering — an upvote means the world to an open-source project.\n\n" +
		"**Upvote + comment here:** https://www.producthunt.com/posts/naeos\n\n" +
		"Every comment (even \"what does this do?\") helps. Ask us anything in\n" +
		"#product-hunt — the makers are here today.",
}
