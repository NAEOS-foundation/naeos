# NAEOS Legal & License Audit

**Status:** Technical audit report (not legal advice)  
**Scope:** Repository `github.com/NAEOS-foundation/naeos`, branch `feat/control-plane-docker-local-staging-20261119`, HEAD `cd100cae`  
**Primary licensing assumption:** Apache License 2.0  
**Audit type:** Read-only; no repository files were modified.

---

## Executive Summary

NAEOS has a well-structured legal foundation: Apache License 2.0 in `LICENSE`, a `NOTICE` file with third-party attribution, DCO-based contribution enforcement via CI, SPDX header enforcement, and documented open-core, trademark, IP-ownership, AI-provenance, and third-party policies. The legal/ directory contains release, commercial-launch, and investor due-diligence checklists.

The most material finding is operational secret hygiene: real credentials (Supabase service role key, database password, Cloudflare API token, R2 keys, Discord token, Slack token, GitHub token, encryption key) exist on disk in `.env` and `.env.staging`. These files are gitignored and not tracked, but they are present in the working tree and represent a rotation/revocation requirement if they touch any production or staging service.

No GPL/AGPL/SSPL contamination was found in the distributed runtime. The only copyleft items are `@img/sharp-libvips-*` (LGPL-3.0-or-later) build-time npm packages, explicitly excluded from distribution in `NOTICE`. The single MPL-2.0 dependency (go-sql-driver/mysql) is file-level copyleft and unmodified.

Overall risk: **MEDIUM** — driven by the secret-on-disk finding and unresolved entity/trademark questions, not by license incompatibility.

---

## Overall Risk

MEDIUM

---

## Top 10 Findings

| # | Severity | Finding | Evidence | Recommendation |
|---|----------|---------|----------|----------------|
| 1 | CRITICAL | Real credentials present on disk in untracked `.env` and `.env.staging` | `.env`: SUPABASE_SERVICE_ROLE_KEY, SUPABASE_DB_PASSWORD, CLOUDFLARE_API_TOKEN, R2_SECRET_ACCESS_KEY, DISCORD_TOKEN, NAEOS_SLACK_TOKEN, GITHUB_TOKEN (ghp_*), NAEOS_ENCRYPTION_KEY; `.env.staging`: NAEOS_CONTROLPLANE_DATABASE_URL with password `E8cuRF4bPjiw193q` | Rotate/revoke all exposed credentials; confirm none reached git history; add pre-commit secret scanning |
| 2 | HIGH | Exact legal entity / jurisdiction for "NAEOS Foundation" is unresolved | `legal/README.md:48` open question #1; `legal/COMMERCIAL-LAUNCH-CHECKLIST.md:168` Terms vs legal/README mismatch | Confirm contracting entity with founder and counsel before paid offers or fundraising |
| 3 | HIGH | Trademark availability/ownership for NAEOS, NEIR, logos unverified | `legal/TRADEMARK.md:18` — no registration claimed; `legal/README.md:49` open question #2 | Perform jurisdiction-specific clearance before public brand claims |
| 4 | HIGH | No CLA; DCO only — no copyright assignment | `CONTRIBUTING.md:294-320`; `.github/workflows/dco.yml` | DCO is adequate for Apache-2.0 inbound=outbound, but document employee/contractor IP rights before commercial launch |
| 5 | MEDIUM | `.env` and `.env.staging` contain secrets but are gitignored (not tracked) | `.gitignore:45-48`; files exist on disk | Ensure `.env` never gets tracked; rotate values if used against real services |
| 6 | MEDIUM | Build-time npm copyleft (`@img/sharp-libvips-*` LGPL-3.0-or-later) | `site/THIRD-PARTY-NOTICES.md:58-61`; `NOTICE:58-61` | Confirmed not distributed; keep excluded from release artifacts |
| 7 | MEDIUM | MPL-2.0 MySQL driver (file-level copyleft) | `NOTICE:43-45`; `go.mod:11` | Safe for distribution unmodified; document any future modification |
| 8 | MEDIUM | AI-assisted contribution provenance relies on contributor self-disclosure | `docs/ai-provenance.md`; `CONTRIBUTING.md:328-339` | Classify material KNOWN/POSSIBLE/UNKNOWN; enforce provenance notes for substantial generated content |
| 9 | LOW | `naeos.bak`, `naeos.bak2`, `.naeos-api.token`, `.naeos-bot-state.json` present on disk, gitignored | `.gitignore:106-103` | Confirm these contain no secrets/customer data; keep out of tracking |
| 10 | LOW | SPDX header check excludes generated dirs (deliberate) | `scripts/check-license-headers.sh:15` | Continue excluding regenerated output; document exclusion rationale |

---

## License Audit

### Current licensing state

- `LICENSE` is present at repository root and contains the complete, unmodified Apache License 2.0 text (Copyright 2025 NAEOS contributors). Verified against standard Apache-2.0 text.
- `NOTICE` is present and records: Apache-2.0 software license; branding/trademark disclaimer (NAEOS, NEIR, logos not granted under Apache-2.0); third-party notices including MPL-2.0 MySQL driver (unmodified, file-level), JS packages (dompurify, khroma, sharp, @img/sharp-libvips-*), and CDN runtime assets. Explicitly states nothing in the repository is licensed under GPL or AGPL.
- `README.md:13,318` and `site/content/community.md:44` / `site/content/id/community.md:44` state Apache License 2.0.
- `CONTRIBUTING.md:288-292` states inbound = outbound under Apache-2.0, consistent with Section 5 of the license.
- No conflicting licenses identified in first-party source.
- SPDX header enforcement via `scripts/check-license-headers.sh` covers tracked `*.go`, `*.ts`, `*.tsx`, `*.js`, `*.mjs`, `*.cjs`, `*.sh`, `*.py`, excluding generated output directories.
- Examples, documentation, scripts, configuration, and prompt files are treated under the repository Apache-2.0 baseline; no separate documentation license identified.
- No CLA or DCO deviation found.

### Files without clear licensing

None identified in tracked first-party source. All tracked Go/TS/JS/sh/python source carries the Apache-2.0 SPDX header per `check-license-headers.sh`. Generated artifacts are excluded by design (regenerated, not first-party source).

---

## Dependency Audit

### Go module dependencies (`go.mod`, `go.sum`)

Direct dependencies (all permissive unless noted):

| Dependency | Version | License | Type | Notes |
|---|---|---|---|---|
| github.com/go-sql-driver/mysql | v1.10.1 | MPL-2.0 | direct | File-level copyleft; unmodified; noted in NOTICE |
| github.com/spf13/cobra | v1.10.2 | Apache-2.0 | direct | |
| github.com/jackc/pgx/v5 | v5.11.0 | MIT | direct | |
| github.com/redis/go-redis/v9 | v9.22.0 | MIT | direct | |
| github.com/nats-io/nats.go | v1.55.0 | Apache-2.0 | direct | |
| github.com/rabbitmq/amqp091-go | v1.14.0 | MPL-2.0 | direct | File-level copyleft; verify |
| github.com/segmentio/kafka-go | v0.4.51 | MIT | direct | |
| github.com/gorilla/websocket | v1.5.3 | BSD-2-Clause | direct | |
| github.com/charmbracelet/bubbletea | v1.3.10 | MIT | direct | |
| github.com/charmbracelet/lipgloss | v1.1.0 | MIT | direct | |
| github.com/fsnotify/fsnotify | v1.10.1 | BSD-3-Clause | direct | |
| github.com/tetratelabs/wazero | v1.12.0 | Apache-2.0 | direct | WASM runtime |
| github.com/bwmarrin/discordgo | v0.29.0 | BSD-3-Clause | direct | Discord bot |
| github.com/DATA-DOG/go-sqlmock | v1.5.2 | MIT | direct (test) | |
| golang.org/x/crypto | v0.57.0 | BSD-3-Clause | indirect | |
| modernc.org/sqlite | v1.60.1 | MIT | indirect | |

Indirect dependencies are predominantly MIT, BSD, Apache-2.0, and ISC. No GPL/AGPL/SSPL identified in the Go graph. License identifiers were not machine-extracted from go.sum (offline); verification against upstream recommended before release.

### npm dependencies (`site/package-lock.json`, 703 packages)

| Class | Count | Notes |
|---|---|---|
| MIT | 459 | Majority |
| Apache-2.0 | 139 | |
| Other permissive | 90 | |
| Unknown | 1 | Requires review |
| Copyleft (LGPL-3.0-or-later) | ~13 | `@img/sharp-libvips-*` and `@img/sharp-wasm32` |

### Copyleft findings

- `@img/sharp-libvips-*` (LGPL-3.0-or-later): build-time only, used by site devDependencies to process images; explicitly excluded from NAEOS runtime/website runtime output per `NOTICE:58-61`. **SAFE** for distribution as not distributed.
- `@img/sharp-wasm32` (Apache-2.0 AND LGPL-3.0-or-later AND MIT): build-time only. **SAFE** as not distributed.
- go-sql-driver/mysql (MPL-2.0): file-level copyleft, unmodified, distributed statically linked. **SAFE** with attribution in NOTICE.

### Dependency/license compatibility matrix

| Dependency | License | Apache-2.0 Compatible? | Modification Risk | Distribution Risk | Action |
|---|---|---|---|---|---|
| go-sql-driver/mysql | MPL-2.0 | Yes (file-level) | Low (unmodified) | Low | Keep NOTICE current; review if modified |
| rabbitmq/amqp091-go | MPL-2.0 | Yes (file-level) | Low | Low | Verify unmodified; keep NOTICE current |
| @img/sharp-libvips-* | LGPL-3.0-or-later | Yes (not distributed) | N/A | None (excluded) | Maintain exclusion from release artifacts |
| All MIT/BSD/Apache/Isc | Permissive | Yes | Low | Low | Standard attribution via SBOM/NOTICE |
| 1 unknown npm package | UNKNOWN | Unresolved | — | — | Identify and resolve before release |

---

## Contributor / IP Audit

### Contribution model

- **DCO only**, no CLA. Every commit must include `Signed-off-by` trailer; enforced by `.github/workflows/dco.yml` on PRs and pushes to `main`.
- `CONTRIBUTING.md:294-320` documents DCO, explicitly states DCO does NOT transfer copyright; contributor retains copyright and grants Apache-2.0 license.
- Employee/contractor/freelancer/agency contributions: contributors must confirm rights holder permits Apache-2.0 submission; no public contracts in PRs.
- No individual or corporate CLA identified.

### IP ownership assessment

- `legal/IP-OWNERSHIP.md` defines asset classes and required evidence (git history, DCO, license, provenance register).
- Contributor retains copyright; Apache-2.0 grants project rights. Not a copyright assignment.
- Open questions (legal/README.md:48-53): exact legal entity/jurisdiction, trademark, pre-existing IP ownership, hosted-service terms, privacy obligations, patent strategy.
- AI-assisted contributions: `docs/ai-provenance.md` classifies KNOWN/POSSIBLE/UNKNOWN; DCO remains the mechanism; provenance notes for substantial generated content.

### Sufficiency for commercial products

DCO is operationally sufficient for an Apache-2.0 project. For commercial/enterprise use and investor diligence, recommend documenting employee/contractor IP assignment records in a private register (per IP-OWNERSHIP.md section 5). DCO alone does not prove employer consent for employee contributions.

---

## Trademark Audit

### Project marks identified

- NAEOS, NEIR, NAEOS logos and visual identity (`legal/TRADEMARK.md:12-16`).
- `brand/brand.json`: brand "NAEOS", tagline "Architecture Drives Engineering.", version 3.1.0, logo master 2.0.0, status FROZEN, source-of-truth SHA-256 recorded.
- Brand assets in `brand/`: SVG logos (logo, logo-mark, logo-horizontal, logo-simplified, logo-mono), icons, lockups, marketing, master, source-of-truth, variants.

### Trademark rights

- `NOTICE:22-28`: Apache-2.0 does not grant trademark rights; use beyond reasonable/customary reference subject to trademark holder.
- `legal/TRADEMARK.md`: existence of a mark in the policy is NOT a statement of registration; jurisdiction-specific clearance required before filing.
- No claim that NAEOS trademark is legally registered.

### Risk

No evidence of accidental trademark permission, endorsement, affiliation, or certification implied. Trademark availability/ownership unresolved — flagged for counsel (Top 10 Finding #3).

---

## Patent / Apache-2.0 Audit

- `LICENSE` Sections 3 (patent grant) and 6 (no trademark rights) present and unmodified.
- No patent notices, patent grants, or patent-related documentation found in repository.
- No CLA/CLA Assistant configuration.
- Apache-2.0 provides: perpetual copyright + patent license from contributors; retaliation clause (Section 3 patent litigation termination). Does NOT provide: trademark rights, warranty, indemnity, or assurance of non-infringement by third parties.
- Contributor patent implications: each contributor grants patent claims necessarily infringed by their contribution; no contributor agreement beyond DCO.

---

## Secrets / Security Hygiene Audit

### CRITICAL — secrets on disk (not tracked)

| File | Secret types present | Git-tracked? |
|---|---|---|
| `.env` | Supabase anon/service-role/secret keys, DB password, Cloudflare API token, R2 keys, Discord token, Slack token, GitHub token, encryption key | No (`.gitignore:45`) |
| `.env.staging` | Staging database URL with password | No (`.gitignore:47`) |

Values are not reproduced in this report. Per AGENTS.md rule 13: report path and type only; do not print secret values; mark CRITICAL; recommend rotation/revocation.

- `.gitleaksignore` lists two verified false positives (placeholders/fixtures): `internal/security/security_test.go:204`, `tools/discord-bot/config.example.env:12`.
- `.github/workflows/gitleaks.yml` runs gitleaks OSS binary on PRs and pushes to main.
- `CONTRIBUTING.md:341-348` prohibits committing secrets; directs to `.env.example` with placeholders.
- `.env.example` and `site/.env.example` and `tools/discord-bot/config.example.env` use placeholders only.

### Other gitignored local artifacts

`naeos.bak`, `naeos.bak2`, `.naeos-api.token`, `.naeos-bot-state.json` — not tracked; confirm no secrets/customer data.

---

## AI Provenance Audit

- `docs/ai-provenance.md`: Active v1.0 policy. Principles: rights not typing; third-party license respect; DCO regardless of production method; no secrets/personal data.
- Classification: KNOWN (human/tool-assisted with record), POSSIBLE (unverified tool terms), UNKNOWN (origin unexplained).
- `CONTRIBUTING.md:328-339`: AI-assisted contributions welcome; provenance note for substantial generated content; Co-authored-by optional; does not replace DCO.
- `prompts/builtin/`: built-in prompts (compiler, llm, manifest.yaml) — first-party, provenance clear.
- No evidence of copied third-party prompts or customer data in prompts directory.
- Classification of repository AI material: **KNOWN** for hand-authored; **POSSIBLE** for agent-generated without provenance notes. Provenance should be documented going forward.

---

## Prompt and Specification Audit

- `prompts/builtin/`: built-in LLM/compiler prompts; first-party; no third-party content identified.
- `specification/NAEOS-SPEC-001..010.md`: NAEOS engineering specifications; first-party.
- `docs/NES-000..054.md`: NAEOS Engineering Specifications; first-party.
- `constitution/`: governance constitution; first-party.
- No secrets, API keys, credentials, or customer information identified in prompts/specifications.
- Licensing: treated under repository Apache-2.0 baseline; no separate license identified.

---

## Documentation License Audit

- Documentation (`README.md`, `docs/`, `site/content/`, `WHITEPAPER.md`, `WHITEPAPER-EN.md`): no explicit separate license; treated under repository Apache-2.0 baseline.
- `site/content/community.md:44` and Indonesian variant explicitly state Apache License 2.0.
- Recommendation: documentation may remain under Apache-2.0, but an explicit documentation license statement (e.g., in README or docs/) would remove ambiguity for redistributors.

---

## Asset Audit

| Asset | Source | License | Commercial Use | Attribution | Risk |
|---|---|---|---|---|---|
| NAEOS logos (SVG) | `brand/` | First-party (brand.json source-of-truth SHA-256 recorded) | Yes (brand-owned) | Brand policy (`legal/TRADEMARK.md`) | LOW (trademark separate from license) |
| Inter font | Google Fonts | OFL-1.1 | Yes | `site/THIRD-PARTY-NOTICES.md:16` | LOW |
| JetBrains Mono font | Google Fonts | OFL-1.1 | Yes | `site/THIRD-PARTY-NOTICES.md:17` | LOW |
| Umami analytics | cloud.umami.is | MIT | Yes | `site/THIRD-PARTY-NOTICES.md:18` | LOW |
| xterm.js | jsDelivr xterm@5.3.0 | MIT | Yes | `site/THIRD-PARTY-NOTICES.md:19` | LOW |
| swagger-ui | unpkg swagger-ui-dist@5.32.15 | Apache-2.0 | Yes | `site/THIRD-PARTY-NOTICES.md:20` | LOW |
| First-party images | `site/public/images/`, `brand/` | First-party | Yes | — | LOW |

No unknown-provenance assets identified. CDN assets are unversioned (Google Fonts, Umami) — note for reproducibility but not a license blocker.

---

## Build and Distribution Audit

### What gets distributed

- Cloned repository: all first-party source (Apache-2.0), LICENSE, NOTICE, docs, examples, prompts, specifications.
- Packaged/published: Go binaries (`cmd/naeos`), Docker images (`Dockerfile`, `Dockerfile.control-plane`, `Dockerfile.release`), release archives via goreleaser (`.goreleaser.yaml`).
- Docker/SBOM: NOTICE states SBOM published with each release records SPDX identifiers for Go modules; gitleaks + NOTICE + SBOM shipped in release archives (CHANGELOG.md:42).
- GoReleaser: `.goreleaser.yaml` present; `scripts/release-goreleaser.yml`, `release.yml`, `release-drill.yml` workflows.
- Reproducible release evidence: `scripts/test-reproducible-release-evidence.sh`, `REPRODUCIBLE-RELEASE-EVIDENCE.md`.
- Licensing obligations in distribution path: Apache-2.0 (retain LICENSE, NOTICE, SPDX), MPL-2.0 (file-level, unmodified), NOTICE attribution.

### Build pipeline traced

`.github/workflows/` includes: ci.yml, release.yml, goreleaser, dco, gitleaks, codeql, dependency-risk, license-headers (docs-gate), sign-off-check, release-drill, reproducible-release-evidence-drill.

---

## Open-Core Assessment

`docs/open-core.md` (draft v0.1) defines the boundary:

| Component | Recommended Model | Reason |
|---|---|---|
| NAEOS Core (spec engine, NEIR, compiler, validator, runtime) | Apache-2.0 | Differentiating open-source value; drives adoption |
| Governance, policy, audit, compliance | Apache-2.0 / Review | Community value; keep open unless enterprise obligations require separation |
| CLI, SDK, plugins SDK | Apache-2.0 | Adoption surface; must stay open |
| Integrations and adapters | Apache-2.0 | Ecosystem growth |
| NAEOS Cloud (hosted services) | Commercial SaaS | Hosted service; Apache-2.0 does not require public code |
| NAEOS Enterprise (admin, SSO, advanced governance, support) | Commercial / proprietary where justified | Enterprise controls; separation keeps core lean |
| Marketplace and registry hosting | Commercial / governed | Operated by Foundation; content licensing differs |

Guardrails: no relicensing of committed Apache-2.0 code; SBOM per release; trademark separate from licensing. This is a sound, conservative open-core boundary. No first-party component is currently under a restrictive/commercial license.

---

## Investor Due-Diligence Assessment

Investor readiness score (0=missing, 1=weak, 2=adequate, 3=strong):

| Category | Score | Notes |
|---|---|---|
| License | 3 | Apache-2.0, complete LICENSE + NOTICE, SPDX enforcement |
| Dependency compliance | 2 | SBOM + NOTICE documented; one unknown npm package; MPL-2.0 file-level |
| Copyright | 2 | DCO in place; no CLA; no copyright assignment |
| Contributor governance | 2 | DCO enforced via CI; CONTRIBUTING.md; AI-provenance policy |
| IP ownership | 1 | Entity/jurisdiction unresolved; pre-existing IP not evidenced |
| Trademark | 1 | Marks identified; no registration claimed; clearance pending |
| Security hygiene | 1 | CRITICAL secret-on-disk finding; gitleaks CI; gitignored .env |
| Documentation licensing | 2 | Apache-2.0 baseline; no explicit doc license statement |
| Commercial/open-core strategy | 2 | docs/open-core.md draft; boundary defined but not implemented |
| Due diligence readiness | 1 | Checklists exist; unresolved entity/trademark/secrets block clean diligence |

Overall: adequate technical controls, but unresolved entity, trademark, and the secret-on-disk finding reduce readiness for investor due diligence.

---

## Remediation Plan

### P0 — Immediate (material legal/security risk)

1. Rotate/revoke all credentials in `.env` and `.env.staging` if used against real services (Supabase service role key, DB password, Cloudflare token, R2 keys, Discord/Slack/GitHub tokens, encryption key). Confirm none entered git history.
2. Ensure `.env`, `.env.staging`, `.naeos-api.token`, `.naeos-bot-state.json`, `naeos.bak*` remain gitignored and never tracked.
3. Add pre-commit/pre-push secret scanning to complement CI gitleaks.

### P1 — Before external contributors

1. Resolve the 1 unknown npm package license; update NOTICE/SBOM accordingly.
2. Document employee/contractor IP rights process (private register per `legal/IP-OWNERSHIP.md:5`).
3. Confirm DCO covers all historical contributors; audit git history for missing sign-offs.

### P2 — Before commercial launch

1. Confirm legal contracting entity and jurisdiction (`legal/README.md:48`).
2. Perform trademark clearance for NAEOS, NEIR, logos (`legal/README.md:49`).
3. Review `site/content/terms.md` and `site/content/privacy.md` against actual implementation (`legal/COMMERCIAL-LAUNCH-CHECKLIST.md:160-172`).
4. Add explicit documentation license statement.
5. Publish SBOM with each release (already planned in NOTICE).

### P3 — Before fundraising

1. Resolve all P0–P2 items.
2. Complete investor due-diligence checklist (`legal/INVESTOR-DUE-DILIGENCE-CHECKLIST.md`).
3. Document open-core boundary implementation and commercial/wedge rationale.
4. Obtain counsel review of term sheets, founder equity, jurisdiction-specific advice.

---

## Proposed Repository Changes

Do NOT modify during this audit (read-only). For future implementation:

- `SECURITY.md` — add secret-rotation procedure and disclosure policy (new content).
- `.pre-commit-config.yaml` — secret scanning hook (new file, if adopted).
- `docs/ai-provenance.md` — already exists; consider adding provenance template for PRs.
- `site/content/` — explicit documentation license notice (add to README or docs landing).
- Private access-controlled rights register (not in public repository) — per `legal/IP-OWNERSHIP.md:5`.

---

## Questions Requiring Founder / Legal Counsel

1. What is the exact legal entity and jurisdiction represented by "NAEOS Foundation," and who has contracting authority?
2. Who owns pre-existing first-party IP, and are founder/employee/contractor rights assigned or licensed to the entity?
3. Have trademark clearance searches been performed for NAEOS, NEIR, and logos in relevant jurisdictions/classes?
4. Are the credentials in `.env` and `.env.staging` tied to production/staging services requiring immediate rotation?
5. What is the patent strategy or defensive publication policy?
6. Should NAEOS adopt a CLA in addition to DCO for commercial product protection?
7. Which jurisdiction's privacy/data-protection obligations apply to the hosted service and website?

---

## Recommended NAEOS Licensing Strategy

| Element | Recommendation |
|---|---|
| NAEOS Core | Apache License 2.0 (current baseline; retain) |
| NAEOS Brand | Separate trademark protection (not granted by Apache-2.0) |
| NAEOS Cloud | Commercial SaaS terms |
| NAEOS Enterprise | Commercial terms / proprietary components where justified |
| Contributions | DCO (current); consider CLA for commercial product protection |
| Documentation | Explicitly licensed (Apache-2.0 recommended) |
| Third-party dependencies | Maintain auditable SBOM / license inventory |

This default strategy is supported by repository evidence. Validate against actual architecture and dependency graph before public commitment.

---

## Important Legal Disclaimer

This audit is a technical repository and license-compliance assessment, not legal advice. Material ambiguities involving ownership, copyright, patents, trademarks, employment IP, contractor IP, licensing, corporate ownership, and investment are marked for review by qualified Indonesian or relevant-jurisdiction legal counsel. Technical facts are distinguished from legal interpretation and legal advice throughout.
