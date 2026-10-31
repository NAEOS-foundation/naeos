# Repository Structure Hygiene Plan

**Status:** Active  
**Scope:** Incremental cleanup of the NAEOS monorepo without changing runtime behavior.

## Goals

- Make the repository easier to navigate for contributors and AI coding agents.
- Keep canonical specifications, implementation, experiments, website code, and operational material distinguishable.
- Prevent local caches and generated run output from becoming source-controlled artifacts.
- Preserve existing paths until references, build configuration, CI, and published URLs have been checked.

## Current baseline

The repository already has a structural map in [REPOSITORY-ARCHITECTURE.md](REPOSITORY-ARCHITECTURE.md) and a navigation index in [DOCUMENTATION-INDEX.md](DOCUMENTATION-INDEX.md). The first cleanup pass should enforce those boundaries rather than introduce a competing architecture.

The current root contains approximately 49 first-level files/directories, while the repository has several thousand tracked paths. Root-level documents include onboarding, governance, technical specifications, project operations, partner materials, and marketing content. These have different lifecycles, so moving them in one bulk change would create unnecessary broken-link and contributor-workflow risk.

## Target top-level responsibilities

| Area | Canonical location | Keep here |
|---|---|---|
| Repository entry points | Root | README, START-HERE, GETTING-STARTED, CONTRIBUTING, SECURITY, LICENSE, core build/config files |
| Normative architecture and contracts | Existing Reference Architecture, specification, and canonical technical specification files | Normative documents; do not duplicate |
| Engineering guides and ADRs | docs/ | Guides, ADRs, domain reference docs, contributor docs |
| Governance | constitution/, governance/, policy/, profile/ | Constitutional and policy artifacts |
| Runtime implementation | cmd/, internal/, pkg/, kernel/ | Executable code and reusable packages |
| Reproducible examples | examples/ | Minimal, documented examples and templates |
| Experiments | experiments/, benchmarks/, bench/ | Clearly labeled experimental work and benchmark inputs |
| Website | site/ | Website implementation and publication content |
| Infrastructure | infra/, supabase/, root deployment files where tooling requires them | Deployment and service configuration |
| Project operations | .github/, brand/, selected project records | CI, contribution workflows, brand source assets |
| Partner and marketing material | Existing docs/partners/ and a future clearly defined documentation domain | Public-facing material, with one canonical source per topic |

## Safe migration rules

1. Do not move files merely to reduce the root item count.
2. Keep build-critical files at their current paths until Docker, Go, CI, hosting, and release references are verified.
3. Before moving a document, search for inbound links and update every canonical index and relative link in the same change.
4. Do not delete generated or experiment files until their generation command, reproducibility, and consumer references are known.
5. Keep normative documents authoritative; summaries must link to them instead of copying their requirements.
6. Make structural changes in small pull requests with explicit path maps and a validation checklist.
7. Never claim CI, build, or deployment passed unless the corresponding checks were observed.

## Phases

### Phase 1 — Hygiene and navigation

- Ignore local caches and generated run output.
- Establish this migration plan as the review checklist.
- Do not change runtime code or move public-facing paths.

### Phase 2 — Root-level document classification

- Classify root Markdown files as entry point, normative, guide, project record, partner, or marketing.
- Move only low-risk noncanonical documents after link and workflow checks.
- Keep compatibility links or redirects where paths are externally referenced.

### Phase 3 — Generated artifacts and examples

- Identify generated output currently tracked under example run directories.
- Verify whether each artifact is required for tests or documentation before removing it from version control.
- Document the exact command to regenerate retained artifacts.

### Phase 4 — Navigation and automation

- Update DOCUMENTATION-INDEX.md, README links, contribution instructions, and any site navigation.
- Add lightweight checks for broken local Markdown links and accidental cache/output files where practical.

## Acceptance checklist

- [ ] No runtime behavior changes in structure-only PRs.
- [ ] No canonical document duplicated or silently superseded.
- [ ] All moved paths have their references updated.
- [ ] Go tests and relevant website checks run, or explicitly marked not run.
- [ ] CI results are inspected before merge.
- [ ] Main remains the stable integration branch.

## Root Markdown inventory (Phase 2 audit)

This classification is a review aid, not an instruction to move every item. Keep canonical and high-traffic entry points in place. Any path move requires a separate inbound-reference audit and a coordinated update to links, CI, release workflows, and site navigation.

| Root file | Class | Initial disposition |
|---|---|---|
| `README.md` | Entry point | Keep at root |
| `START-HERE.md`, `GETTING-STARTED.md` | Onboarding | Keep at root for discoverability; link to deeper guides |
| `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md` | Community and security policy | Keep at root (standard GitHub discovery paths) |
| `AGENTS.md` | Agent instructions | Keep at root; preserve tooling discovery |
| `LICENSE`, `NOTICE` | Legal / licensing | Keep at root (not Markdown inventory, but root-critical) |
| `DOCUMENTATION-AUTHORITY.md`, `DOCUMENTATION-INDEX.md` | Documentation governance/navigation | Keep at root; canonical cross-repository entry points |
| `REPOSITORY-ARCHITECTURE.md` | Repository map | Keep at root; reference this plan rather than duplicating its rules |
| `NAEOS-MTS-001.md` | Normative technical specification | Keep at root until authority and external references are checked |
| `NAEOS-PER-001.md` | Project evolution record | Candidate for a project-record location after link audit |
| `NAEOS-RDP-001.md` | Roadmap/development plan record | Candidate for a project-record location after checking roadmap authority and links |
| `ARCHITECTURE-OVERVIEW.md`, `TRUST-BOUNDARY-MAP.md` | Architecture/reference | Check authority overlap and references before moving or consolidating |
| `WHITEPAPER.md`, `WHITEPAPER-EN.md` | Public-facing whitepaper | Keep published paths stable until external links and site publishing are audited |
| `MARKETING-STRATEGY.md` | Marketing/operations | Candidate for `brand/marketing/` after inbound-reference audit |
| `NAEOS-PARTNER-TARGET-LIST-v1.0.md`, `PARTNER-PROGRAM.md`, `PARTNER-REQUESTS.md`, `PARTNERS.md` | Partner operations | Candidate for `docs/partners/`; inspect website, forms, and relative links first |
| `DEVELOPMENT_PLAN.md`, `ROADMAP.md`, `CHANGELOG.md` | Project planning/history | Keep canonical authority explicit; do not merge or relocate until overlapping roadmap roles are reconciled |
| `GLOSSARY.md` | Terminology reference | Candidate for `docs/` only after auditing inbound links and generated docs |

### Phase 2 reference audit findings

- `.run/generated` is referenced in `docs/EXTERNAL-VALIDATION-CASE-STUDY-001.md` and two `brand/marketing/` assets. These references describe generated output, so ignoring it does not remove the documented workflow; however, whether its tracked contents are needed for reproducibility must be checked before deletion.
- `Reference Architecture/NAEOS-NRA-001.md` and `Reference Architecture/NAEOS-NRA-002.md` both contain NAEOS-NRA-001 identity text. Treat this as a separate document-authority review; do not rename or consolidate either file as part of a general cleanup without resolving the apparent identifier overlap.
- The root has multiple roadmap/planning artifacts with explicit authority distinctions. Preserve those distinctions and avoid creating a second canonical roadmap.
- No root Markdown files are moved in this phase. The next structural PR should move only low-risk partner/marketing records after the full inbound-link and workflow search is complete.

### Phase 2 reference audit — migration blockers and generated output

The inbound-reference search found these concrete dependencies for the candidate documents:

- `PARTNER-PROGRAM.md`: referenced by `README.md`, `PARTNERS.md`, and `DOCUMENTATION-INDEX.md`; the partner directory links back to it.
- `PARTNER-REQUESTS.md`: referenced by `PARTNER-PROGRAM.md`, `docs/partners/30-DAY-PLAN.md`, and `DOCUMENTATION-INDEX.md`. The 30-day plan uses a relative path (`../../PARTNER-REQUESTS.md`).
- `PARTNERS.md` and `NAEOS-PARTNER-TARGET-LIST-v1.0.md`: indexed from `DOCUMENTATION-INDEX.md`.
- `MARKETING-STRATEGY.md`: referenced by `DOCUMENTATION-INDEX.md` and `.github/copilot/prompts/investor-page.md`.
- `GLOSSARY.md`: linked from `GETTING-STARTED.md` and `DOCUMENTATION-INDEX.md`.
- `ARCHITECTURE-OVERVIEW.md`: linked from `START-HERE.md`, `GETTING-STARTED.md`, and `DOCUMENTATION-AUTHORITY.md`; it is also linked by the published investor page using an absolute GitHub URL.
- `NAEOS-PER-001.md` and `NAEOS-RDP-001.md`: indexed from `DOCUMENTATION-INDEX.md`; roadmap and authority semantics must be settled before either moves.

**Decision:** do not move these files in the current hygiene PR. A move must update the known relative links and index entries in the same change, then run a repository-wide search and link/build checks. Absolute GitHub URLs and external consumers may continue to depend on the old paths, so a compatibility file or redirect may be required.

The checked-in `examples/demo-cli/.run/generated/` tree contains **76 tracked tree/blob entries**, including generated Go and TypeScript projects. The canonical demo script writes output under `.run/`, and the external validation case study cites specific generated Go files as evidence; the marketing HN draft and video storyboard also describe this output. The ignore rule added in Phase 1 only prevents new untracked output from being added—it does not remove files already tracked by Git.

**Decision:** preserve the existing generated snapshot for now. Before untracking or deleting it, determine whether the demo/test contract intentionally uses the snapshot, whether it can be regenerated deterministically, and whether the case study should point to an archived fixture or the output from its pinned validation commit. Do not claim the current snapshot proves the behavior of current `main`.

**Status (2026-10-31):** reproducibility verified. Executing `examples/demo-cli/run-demo.sh` with the freshly built CLI into a clean output directory reproduced 53 of the 54 tracked snapshot files byte-for-byte (54 materialized files from 60 logical artifacts, 6 path collisions). The only difference is `generated/demoapp/package.go`, stale output from an older generator version. No Go test references the snapshot; the demo contract regenerates output under `.run/` at run time. The tracked snapshot was therefore untracked with `git rm -r --cached` (local copy kept and already covered by the `examples/demo-cli/.run/generated/` ignore rule), so the case-study references remain valid as archived evidence of the pinned run while new output is no longer source-controlled.

### Next safe migration sequence

1. Let the current PR checks finish; do not merge while CI is pending.
2. In a dedicated follow-up PR, migrate at most one low-risk document group and update all known links together. Partner docs are not yet low-risk because they have reciprocal and relative references.
3. Treat generated-output cleanup as a separate change from document moves.
4. Add broken-local-Markdown-link validation only after checking the existing repository's conventions and exceptions; do not introduce a broad gate that fails on known intentional external or generated links.

## Phase 3 reconnaissance — structural boundaries and tracked artifacts

The current main tree contains 2,408 entries (recursive Git tree; not a count of files alone). The largest first-level domains by entry count include internal/ (761), docs/ (472), site/ (315), cmd/ (193), and examples/ (165). These sizes alone are not evidence that a directory should be split or moved.

### Findings that need follow-up

- **Tracked SQLite sidecars:** cmd/naeos/test.db-shm (32 KiB) and cmd/naeos/test.db-wal (empty) are tracked in main. These are SQLite shared-memory/write-ahead-log sidecars, not source files. The existing .gitignore already contains *.db-shm and *.db-wal, but those rules do not untrack files already committed. Before removing them, verify no test or fixture intentionally relies on their exact bytes; then delete them in a dedicated cleanup commit and run relevant Go tests.

**Status (2026-10-31):** verification done. No Go test or fixture references the tracked sidecar bytes; `cmd/naeos/db_cmd_test.go` and `api_db_test.go` create their own `test.db` under command-managed or `t.TempDir()` paths. Both sidecars were untracked with `git rm --cached` (local file copies kept; the ignore rules already cover them), and `go test ./cmd/naeos/...` passes. A dedicated commit is still required to land this change.
- **Benchmark locations:** bench/baseline.txt is a small Go benchmark result, while benchmarks/governance-boundary-v1/README.md defines a reproducible governance-boundary benchmark backed by scripts/governance-boundary-benchmark.sh. They are related but not equivalent. Do not blindly merge the directories; first define whether bench/ is historical measured output and benchmarks/ is the reproducible benchmark interface.
- **Built-in prompt assets:** prompts/builtin/ contains YAML compiler and LLM prompts plus a manifest. Treat these as product/runtime assets, not miscellaneous documentation. Keep their paths stable until loader/embed configuration and tests are audited.
- **Runtime concentration:** internal/ is the largest domain and contains many separately named packages. Do not perform a mass package move. Start with a package/dependency map and identify actual coupling or naming overlap before proposing a split.
- **Separate kernel domain:** kernel/ is small compared with internal/, but it has an explicit architectural role in REPOSITORY-ARCHITECTURE.md. Directory size is not a reason to merge it into internal/.

### Recommended order

1. Finish and review the current hygiene PR.
2. Remove only confirmed accidental runtime artifacts such as tracked SQLite sidecars in a separate PR, with Go tests.
3. Document the intended roles of bench/ versus benchmarks/ and check scripts/workflows before any rename.
4. Build an internal-package dependency map before considering source-code reorganization.

## Phase 4 reconnaissance — Go package boundaries (first pass)

This is a source-level reconnaissance pass, not a complete static dependency graph. The recursive Git tree for `main` contains **612 Go files under `internal/`**, including tests, spread across **78 immediate subdirectories**. That confirms the area is large enough to justify an explicit boundary map, but not that it should be split into multiple modules or moved wholesale.

### Observed responsibility seams

| Area | Evidence observed in current source | Restructuring implication |
|---|---|---|
| `internal/controlplane/` | Defines policy, grant, action, decision, execution-lock, and ledger/evidence-oriented types in package `controlplane`. The sampled `model.go`, `evaluator.go`, `ledger.go`, `gateway.go`, and `evidence.go` are primarily standard-library imports. | Treat as a core domain boundary; do not relocate it before tracing its callers and public contracts. |
| `internal/evidence/` | Evidence records import `internal/governance/control`; runtime evidence builder also imports that decision package. | There is an explicit evidence-to-governance dependency. Preserve direction and inspect types/contracts before attempting to unify evidence packages. |
| `internal/runtime/gateway/` | Imports both `internal/errors` and `internal/governance/control`. | Runtime gateway currently depends on shared error and governance-control contracts; those contracts are potential boundary anchors. |
| `internal/runtime/engine/` | The sampled engine imports `internal/errors` plus standard-library packages. | Runtime engine and runtime gateway are separate Go packages; keep their roles distinct until call paths prove otherwise. |
| `internal/audit/` | Sampled audit implementation imports `internal/errors`; encrypted audit implementation also imports `internal/securityext`. | Audit storage/formatting and evidence lifecycle should not be assumed to be the same domain merely because both retain records. |
| `internal/investordemo/` | `controlplane_bridge.go` imports `internal/controlplane`. | Demo/experiment code is an explicit consumer of control-plane types; preserve the demo contract while mapping dependencies. |

### Boundary risks to investigate before any source moves

- **Potentially overlapping concepts, not proven duplicates:** `internal/controlplane`, `internal/governance/control`, `internal/evidence`, and `internal/audit` all touch authorization decisions, policy context, or records. They are not interchangeable based on names alone. Compare their types, persistence guarantees, callers, and tests before consolidation.
- **Large surface area:** `cmd/naeos/main.go` registers a broad CLI command set. Moving packages can affect many command handlers even when a package compiles in isolation; the dependency map must include CLI entry points and integration tests.
- **No module split proposed:** the repository currently declares one Go module, `github.com/NAEOS-foundation/naeos`. Do not introduce nested modules or public package moves as a cleanup shortcut.
- **Tests count toward the 612-file inventory:** the count is a path inventory, not a count of production source files or independent packages. Some immediate directories contain nested packages, so it is not a full package count.

### Next validation steps

1. Generate a reproducible import graph from the repository checkout (for example, `go list -deps -json ./...` or a dedicated import-graph script) and retain package-level edges, not only directory sizes.
2. Identify high-fan-in packages and cycles, if any, from the actual graph; do not infer cycles from folder names.
3. Trace `controlplane`, `governance/control`, `evidence`, `audit`, and `runtime` from CLI/API entry points through tests and persistence adapters.
4. Record a proposed boundary change only when it has a specific coupling problem, a path/import migration plan, and targeted regression tests.
5. Keep source moves out of the current structure-hygiene PR. A source refactor must be a separate, behavior-preserving PR with full Go and relevant integration checks.

## Phase 4 implementation — reproducible import graph

Added `scripts/internal-import-graph.sh` to emit sorted, deduplicated TSV edges in the form `importing_package<TAB>imported_package`. It scopes output to first-party imports in the NAEOS Go module and includes packages across `./...` so command-line and library consumers are represented, not just packages below `internal/`.

Run it from any directory inside a checkout with:

```bash
bash scripts/internal-import-graph.sh > /tmp/naeos-import-graph.tsv
```

This is an audit primitive, not a claim that the graph has already been generated or reviewed. **Validation status:** not run in this environment; CI and local Go tests must confirm the script works before merge. Next, derive fan-in/fan-out and inspect actual edges for the control-plane, governance, evidence, audit, and runtime packages. Do not move packages based only on the output volume or directory size.

## Phase 5 reconnaissance — target map for the current checkout

**Observed checkout date:** 2026-10-25. This section records the working tree examined for the next restructuring step; older counts and branch observations above describe earlier snapshots and must not be treated as current.

The checkout has 1,988 tracked paths. It has a staged CLI extraction under `cmd/naeos/commands/` and helper package under `cmd/naeos/internal/clihelpers/`. The tracked `cmd/naeos/main.go` imports both trees, and CLI tests import command packages from `cmd/naeos/commands/`. The current `HEAD` tree contains neither source tree, while its parent contained the prior flat command implementations. The staged change now contains the extracted source needed by the entry point; a clean checkout will receive those packages once this complete change is committed.

`go list` inspection resolved package metadata before the helper move. The 40 CLI import sites were updated when `internal/clihelpers/` was moved to `cmd/naeos/internal/clihelpers/`, after confirming every consumer was under `cmd/naeos/`. A CLI build to `/tmp/naeos-cli-restructure-check` succeeded both before and after the move, and `go test ./cmd/naeos/...` now passes. This validates the complete staged working tree; `HEAD` itself does not include the staged additions until they are committed. The helper imports Cobra, `internal/pipelinecache`, `internal/governance/policy`, `internal/errors`, and `pkg/pipeline`. Keep the dependency direction toward product/domain code; implementation packages must not import CLI command packages.

The current `go list` graph for `./internal/... ./pkg/...` resolves **131 packages** and **257 direct in-repository import edges** with no package import cycle reported. Highest fan-in packages are `internal/errors` (74 importing packages), `internal/neir/model` (17), `internal/specification/parser` (10), and `internal/securityext` (10). This suggests stable shared contracts, not automatic candidates for relocation. `pkg/pipeline` coordinates several internal domains, including NEIR, generation, specification, governance, planning, profiling, and evidence; inspect its public API/type leakage before treating it as a reusable standalone module. No module split is justified by this graph alone.

The repository's normative SDK spec defines the intended boundary: [NES-019](docs/NES-019-SDK.md) says packages under `pkg/` are public, stable, and usable without knowledge of `internal/` (FR-001, NFR-001, NFR-002). An additive `pkg/sdk` façade now exposes public config, policy, result, artifact, task/review/policy summaries, governance provenance, and graph counts; its focused tests and a temporary external-module integration test passed. NEIR builders use the documented version mapping (`2.0.0` model, `0.5.0` input spec), and builder output is checked against the model schema with full draft-07 validation. The generator publishes separate model and input-spec schema contracts; the CLI and pipeline input validation use the input schema, and representative repository specs pass. Next, add typed public DTO round-trip tests before exposing NEIR in the SDK. A v3.7.0 GitHub release exists, while this checkout's `VERSION`/`CHANGELOG.md` say 3.7.1 and the local tag set stops at v3.7.0; preserve v3 compatibility and reconcile release metadata before publishing. This does not remove the existing leak: `pkg/pipeline.Config`, `pkg/pipeline.Result`, adapter methods, and `pkg/config.Pipeline.Policies` still expose `internal/` types. The façade checks cancellation at stage boundaries, though it cannot interrupt an active stage; both Run and Validate results include hashes after successful completion. See [the SDK boundary design](docs/architecture/SDK-PUBLIC-API-DESIGN.md).

Static comparison of `HEAD^:cmd/naeos/main.go` with the current `cmd/naeos/main.go` found 78 root-command registrations in each, with equal normalized constructor-name sets. A source-level declaration scan also matched all 248 command constructor names from the parent flat source to the extracted packages, with no unmatched names in either set. This is strong evidence that root registrations and constructor inventory were carried over, but it is not proof that command behavior, flags, output, error handling, exit codes, or generated documentation are unchanged. The parent had 88 non-test Go source files directly under `cmd/naeos/`; the extracted tree currently has 84 non-test Go files, so helper/file consolidation still needs review.

A static comparison of literal Cobra metadata found 193 shared `Use` command names (the old source also had the root `naeos` name) and 152 shared flag-name candidates, with no old-only or new-only flag names detected by the scan. This is a source-text inventory only; it does not compare flag types/defaults/required status, aliases, command output, or runtime behavior.

The extraction exposed that several old `package main` tests depended on unexported command helpers. Those unit tests were moved into their owning CLI packages; root-command integration tests remain under `cmd/naeos/`. The CLI suite now passes with `go test ./cmd/naeos/...`. Cobra reference docs were regenerated to 298 Markdown pages. Generation used `HOME='~'` to keep the plugin-directory default portable, and `CONTRIBUTING.md` now documents that reproducible invocation. No unrelated runtime package was moved.

### Proposed target map

| Component | Current path | Target path | Allowed dependency direction | Migration risk |
|---|---|---|---|---|
| Main NAEOS CLI entry and command registration | `cmd/naeos/main.go` | Keep `cmd/naeos/`; keep entry small and registration explicit | CLI entry → command constructors; no domain package imports back into `cmd/` | High until extracted source is tracked and its ownership/status is resolved |
| CLI command adapters | `cmd/naeos/commands/*` (staged); predecessor versions were flat files under `cmd/naeos/` in `HEAD^` | Keep `cmd/naeos/commands/<area>/`; include the extraction and entry-point change together | command adapters → `internal/` and stable `pkg/` APIs; never `internal/` → `cmd/` | Medium until committed; behavioral coverage now passes in the staged working tree |
| CLI-only helpers | `cmd/naeos/internal/clihelpers/` (staged; moved in this working tree) | Keep under `cmd/naeos/internal/` | CLI commands → CLI helpers → domain/API packages; no runtime package should depend on CLI helpers | Low: scoped to CLI consumers and tests pass; helper still imports product packages |
| Other Go executables and generators | `cmd/gentest/`, `cmd/neir-schema-gen/`, `cmd/naeos-change-risk/`, `cmd/naeos-dependency-risk/`, `cmd/naeos-demo/` | Keep separate under `cmd/`; consolidate only if they share a stable library/API | executable → `internal/` or `pkg/`; no executable-to-executable dependency | Low to medium |
| Public Go SDK | `pkg/config/`, `pkg/kernel/`, `pkg/pipeline/`, initial façade `pkg/sdk/` | Keep paths stable; use `pkg/sdk/` as the additive external entry point while legacy signatures remain | public DTOs/interfaces → internal implementation adapters; downstream consumers must not need `internal/` imports | High: legacy exported structures still expose internal types; release compatibility and NEIR/plugin contracts need a migration plan |
| Runtime and domain implementation | `internal/` (many domain packages) | Keep one Go module and current domain paths initially; group only after import graph identifies a concrete boundary | domain packages → narrow domain contracts/adapters; avoid cycles and cross-domain reach-through | High: broad coupling and many tests/consumers |
| Kernel architecture artifacts | `kernel/` and `pkg/kernel/` | Keep distinct: normative artifacts remain in `kernel/`, implementation API in `pkg/kernel/` | docs/specification define contracts; implementation realizes them | Medium: naming overlap can confuse navigation but paths serve different roles |
| Normative specifications and architecture | `Reference Architecture/`, `specification/`, root `NAEOS-MTS-001.md`, `constitution/` | Keep canonical paths stable; improve index links before considering moves | normative authority → implementation; guides/site may summarize and link, not supersede | High: authority and external-link risk |
| Engineering docs and generated references | `docs/`, root `DOCUMENTATION-*`, `GLOSSARY.md` | Keep current locations for now; `docs/` remains home for guides, ADRs, references | docs derive from normative sources and implementation; generated docs must be marked | Medium: many internal and published links |
| Partner material | root partner files and existing `docs/partners/` | Candidate consolidation under `docs/partners/` | navigation/index → one canonical partner record | Medium/high: reciprocal links and external URLs already exist |
| Marketing material | root `MARKETING-STRATEGY.md` and `brand/marketing/` | Candidate canonical strategy under `brand/marketing/`; root may retain a short compatibility link if needed | marketing claims → verified product/docs/release evidence | Medium: inbound links and public paths must be audited |
| Website | `site/` | Keep as isolated publication app | site content → canonical docs/specs; no site dependency from Go runtime | Low for internal moves; higher for public URLs and deployment |
| Worker and deployment configuration | `infra/ws-worker/`, `supabase/`, root `Dockerfile*`, `.github/workflows/` | Keep at current paths; document each build/deploy boundary | deployment artifacts consume release outputs; infrastructure does not become runtime authority | Medium: CI/CD paths and secrets context |
| Prompts and plugin mechanisms | `prompts/`, `internal/promptlib/`, `internal/pluginhost/`, `internal/pluginsdk/`, `examples/plugins/` | Keep product prompt assets and implementation distinct; do not create a top-level `plugins/` until a separate lifecycle exists | prompt/template assets are loaded by implementation; plugins use documented SDK/contracts | Medium: runtime loading, generated artifacts, and provenance |
| Experiments and benchmarks | `experiments/`, `bench/`, `benchmarks/` | Keep separate until reproducibility and ownership roles are explicit | experiments observe named implementation; benchmark harness produces traceable evidence | Medium: some docs/marketing refer to checked-in generated results |
| Brand, assets, and legal evidence | `brand/`, `legal/` | Keep distinct from runtime source; fill provenance register before redistribution | product/site may consume assets; legal evidence governs release claims | Medium: source and ownership evidence incomplete |

### Recommended migration order

1. **Stabilize the CLI extraction.** Root registrations, constructor names, literal `Use` names, detected flag names, and CLI package tests now match/pass. The extraction, moved tests, entry point, generated CLI reference, and contributor generation instructions are staged together so they can be reviewed as one change.
2. **Constrain CLI helper scope.** The helper is under `cmd/naeos/internal/clihelpers/`; all 40 CLI imports were updated, and the CLI build and package tests passed. Keep it outside the general product runtime API.
3. **Continue the public SDK boundary work.** The additive `pkg/sdk` façade is in place and has passed focused and external-module checks. NEIR metadata and separate model/input schemas now have focused validation. Next define a typed public NEIR DTO and round-trip contract, reconcile v3.7.1 release metadata, evaluate cancellation inside long-running stages, and design the plugin public contract. Keep legacy signatures unchanged. Do not split Go modules or rename domain packages based on directory size alone.
4. **Consolidate documentation in small groups.** Partner material and the root marketing strategy are candidates only after repository-wide link, workflow, and published-URL searches. Keep legal, normative, onboarding, and GitHub-discovery files in their canonical locations.
5. **Handle generated artifacts separately.** Determine which `.run/` snapshots are reproducibility evidence before untracking or deleting them; do not mix this with source or document moves.

### Decision boundary

This map is a proposal, not approval to move or delete files. The current CLI extraction is the only source-boundary change already visible in the working tree. Before accepting it as the target, maintainers should review it as a complete change (including all new files), then verify the build, CLI behavior, generated command docs, and CI. No Go module split, top-level domain creation, or bulk move is recommended by this reconnaissance.

