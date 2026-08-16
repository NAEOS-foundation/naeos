# Reddit Posts — Launch Day (D-0, 12:00–16:00 PT)

Post **manually** from your personal account (not a bot) — Reddit's self-promo
rules require human interaction. Post **after** the PH post is live. Before
posting, verify each sub's self-promo policy (sidebar/wiki). Some subs require
approval via modmail first.

**Golden rules:**
- Never post the same text twice — tailor to each sub.
- Engage in comments for the first few hours; don't post-and-run.
- If a sub has a 10:1 ratio rule, use an account that's genuinely been active
  in that community.
- Use the discussion question as the hook, not the product.

---

## r/golang — "I built a Go pipeline that treats your spec as the source of truth"

> I kept seeing the same drift: docs say one thing, code says another, and
> every new engineer reverse-engineers the system.
>
> So I built a pipeline in Go that treats a YAML spec as the single source of
> truth. It parses → normalizes → resolves → validates → schedules a DAG →
> generates code. The internal model (NEIR) captures project, modules, services,
> APIs, storage, security — so generation isn't string-splitting, it's
> model-driven.
>
> Highlights:
> - Validated output in Go/TS/Python/Java/Rust from one spec
> - `naeos run --cache-dir` skips unchanged stages (incremental builds)
> - `--profile` shows where the pipeline spends time
> - WASM plugin SDK to add your own adapters
> - Stage caching + distributed execution for larger projects
>
> Apache 2.0: https://github.com/NAEOS-foundation/naeos
> One-page overview: https://naeos.dev
>
> What's the worst spec/code drift you've dealt with? I collected enough horror
> stories to build a whole pipeline around it — curious what yours look like.

---

## r/devops — "Spec-driven engineering, but the spec stays the source of truth (not scaffolding)"

> Most "spec-driven" tools scaffold once and you're on your own. I wanted the
> spec to stay authoritative for the whole lifecycle, not just day one.
>
> I built NAEOS (Go, Apache 2.0): one YAML spec → validated code in
> Go/TS/Python/Java/Rust + governance checks + docs + AI context. The pipeline
> is auditable end to end: parse → normalize → resolve → validate → schedule →
> generate. Errors are explicit (`naeos validate --output json` gives codes +
> field locations), so CI failures are fixable, not mysteries.
>
> For ops folks specifically:
> - Reproducible builds: same spec → same output (artifact store, audit trail)
> - Watch mode re-runs on change; diff engine shows what drifted
> - Schema migrations across spec versions
> - Docker-ready, single static binary
>
> https://github.com/NAEOS-foundation/naeos · https://naeos.dev
>
> Is "spec as source of truth" actually practical at your scale, or does it
> break down once teams and monorepos get big? Genuinely asking — that's the
> boundary I want to understand.

---

## r/programming — "The spec/code drift problem, inverted"

> I work on a tool that flips the drift problem around: instead of docs chasing
> code, the spec is the source of truth and the code is generated from it.
>
> It's a Go pipeline — parse, validate against a schema, build an engineering
> model (NEIR), generate code in 5 languages, and emit AI instruction sets so
> assistants work from the model instead of guessing from your repo.
>
> It's not a magic "no more coding" tool — it's for systems where the spec IS
> the artifact: APIs, services, schemas, governance. When the spec changes, the
> diff engine shows exactly what drifted, and the pipeline re-runs only what
> changed (stage caching).
>
> Open source (Apache 2.0): https://github.com/NAEOS-foundation/naeos
>
> Design decisions I'd love feedback on: using `$ref`/`$include` for spec
> composition, `$fn` for custom functions, and keeping the whole pipeline
> deterministic + auditable.

---

## Posting checklist

- [ ] PH post is live before posting to Reddit
- [ ] Check each sub's self-promo rule; modmail first if required
- [ ] Use your personal account with genuine history in the sub
- [ ] Personalize the discussion question per sub
- [ ] Reply to every comment for 2–3 hours after posting
- [ ] Don't post to more than 2–3 subs; don't cross-post the identical text
