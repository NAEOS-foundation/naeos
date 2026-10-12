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
