# NAEOS Development Plan — v2.2.0 → v3.0.0

## Fase 1: Kualitas & Keandalan

| Item | Area | Detail |
|------|------|--------|
| Test coverage minimum 80% | Backend | ✅ `supabase` 84.1%, `messagequeue` 93.5%, `marketplace` 88.7%, `mcp` 85.1%, `migration` 97.9%. Target: `watch`, `rollback`, `cicd`, `distributed`, `gateway`, `websocket`, `eventsourcing` |
| CI lint zero-failure ✅ | Backend | Semua pelanggaran `bodyclose`, `noctx`, `gofmt`, `unconvert`, `errcheck` diperbaiki. Lint lulus 100% |
| Integration test suite | Backend | ✅ Supabase integration tests (Auth, Storage, SQL, Admin) di CI tiap commit dengan secrets |
| Fuzz testing | Backend | ✅ 5 fuzz targets jalan di CI tiap commit |
| Error message audit ✅ | UX | `internal/errors/` — 15 error codes (`PARSE_ERROR`, `VALIDATION_ERROR`, `AUTH_ERROR`, dll) + sentinel errors + `ErrorGroup`. Tinggal audit penggunaan di seluruh package |
| Logging standardization | Backend | ✅ 0 `log.Print` sisa. ❌ 7 `fmt.Println` di `internal/create/create.go` (CLI wizard) — perlu migrasi ke `slog` |

## Fase 2: Website & Dokumentasi

| Item | Area | Detail |
|------|------|--------|
| Wiki → Hugo migration ✅ | Site | Semua halaman wiki sudah dimigrasi ke Hugo site. Wiki/ dihapus. |
| CLI docs auto-generate ✅ | Site | `naeos docsgen` — regenerate 150+ file CLI docs (termasuk `naeos_supabase*.md`) |
| API docs auto-generate ✅ | Site | `.github/workflows/website.yml` — copy `docs/openapi.yaml` ke `site/static/` tiap build Hugo (raw YAML, bukan Swagger UI) |
| Blog content pipeline ✅ | Site | `.github/workflows/release-blog.yml` — triggered on `release: [published]`, auto-create blog post EN + ID, open PR |
| Interactive playground ✅ | Site | xterm.js + WebSocket ke server demo di homepage. Hero terminal interaktif, fallback ke animasi statis. Demo server di `cmd/naeos-demo/` |
| PDF generation ✅ | Site | CLI reference + getting-started sebagai PDF download via GitHub Action (`pdf-docs.yml`). Tersedia di `/downloads/` |
| Dark mode OG image ✅ | Site | SVG OG image dengan `prefers-color-scheme` CSS + PNG fallback (dark & light) via sharp. Tersedia di `/images/og-default.svg` |
| Swagger UI page | Site | Generate halaman `/docs/api/` dari `docs/openapi.yaml` (saat ini masih raw YAML, belum ada UI renderer) |

## Fase 3: Platform & Ekosistem

| Item | Area | Detail |
|------|------|--------|
| Supabase backend integration ✅ | Backend/CLI | Database adapter (pgx), Auth (GoTrue), Storage, Edge Functions, Admin API. CLI: `naeos supabase init/auth/storage/sql/status` |
| Plugin registry publik ✅ | Backend/Site | `registry.json` di `/plugins/`, halaman browse di `/plugins/`, `RemoteRegistry` client di `internal/marketplace/`, GitHub workflow auto-discover `naeos-plugin` topic |
| Plugin template generator ✅ | CLI | `naeos plugin init` — scaffolding dengan SDK boilerplate, tests, Makefile, GitHub Actions CI, WASM entry point |
| NEIR schema registry | Backend | Host NEIR JSON Schema di `schemaregistry/` dengan versioning, validasi spec terhadap schema terbaru |
| Template marketplace ✅ | CLI | `naeos template publish [path]` — publikasi starter project template ke registry, `naeos template search`, `naeos template init` |
| GoReleaser release pipeline ✅ | CI | Auto-build binary untuk linux/darwin/windows × amd64/arm64 tiap tag. Checksum + Docker image ke ghcr.io |

## Fase 4: Performa & Skalabilitas

| Item | Area | Detail |
|------|------|--------|
| Pipeline caching v2 | Backend | ❌ Saat ini hanya full-spec hash cache. Target: incremental/partial — skip stage jika input tidak berubah |
| Parallel generation | Backend | ❌ Saat ini sequential `for range` di `CompileAll()`. Target: concurrent multi-adapter via `errgroup` |
| Lazy NEIR loading | Backend | ❌ Load NEIR model on-demand untuk proyek besar (1000+ module) |
| Benchmark suite | Backend | ❌ Belum ada benchmark terstandarisasi untuk 3 skala (small/medium/large) |
| Memory profiling | QA | ❌ Belum ada profiling untuk leak detection |

## Fase 5: AI & Developer Experience

| Item | Area | Detail |
|------|------|--------|
| AI recommendation engine ✅ | Backend | `naeos ai suggest` — analisa spec via LLM, rekomendasi arsitektur & best practices. Juga `ai explain`, `ai enrich`, `ai compile` |
| NEIR-aware LSP | Backend | ❌ Language Server Protocol untuk spec YAML: autocomplete, diagnostics, hover info, go-to-definition |
| VS Code extension | Plugin | ❌ Extension dengan syntax highlighting, LSP integration, inline validation, playground |
| NEIR diff visualization | CLI/TUI | ❌ `naeos diff --format unified` (text). Target: `--visual` side-by-side tree view |

## Fase 6: Rilis v3.0.0

| Item | Area | Detail |
|------|------|--------|
| NEIR v2.0 specification | Core | ❌ Saat ini NEIR v0.1.0. Target: conditional modules, environment profiles, inheritance, multi-file |
| GUI Dashboard ✅ | Site | `naeos dashboard` — web dashboard dengan stats, activity log, component health, WebSocket live updates |
| RBAC ✅ | Backend | `internal/auth/rbac.go` — admin/developer/viewer roles, resource-based permissions |
| OAuth2/OIDC ✅ | Backend | Google OAuth2 provider + `/.well-known/openid-configuration` + `/.well-known/jwks.json` |
| Enterprise features | Backend | ✅ RBAC + OAuth2/OIDC dasar. ❌ SSO framework, SAML, LDAP, audit log export (JSON/Splunk), compliance reports (SOC2, HIPAA) |
| v3.0.0 release | All | Changelog, migration guide v2→v3, release party blog post, deprecation notices |

## Metrik Progress

| Metrik | Saat Ini | Target Q1 2027 | Target Q3 2027 |
|--------|----------|----------------|----------------|
| Test coverage (overall) | ~65% | ≥80% | ≥85% |
| Test coverage (target packages) | ✅ 84–98% (5 packages) | — | — |
| CLI commands test coverage | ~55% | 100% | 100% |
| Website pages (EN) | 24 | 35+ (wiki migrated) | 40+ |
| Blog posts | 2 | 6+ | 12+ |
| Plugin ecosystem | 0 | 5+ community plugins | 20+ |
| Build time (pipeline) | ~2s (small) | <1s (small) | <5s (medium) |
| CI lint pass rate | 100% ✅ | 100% | 100% |
| `fmt.Println`/`log.Print` sisa | 7 (`create.go`) | 0 | 0 |

## Completed (v2.2.0)

- **Supabase backend integration** — database adapter, Auth, Storage, Edge Functions, Admin API, CLI, CI
- **Lint zero-failure** — 28 lint issues fixed (`bodyclose`, `noctx`, `gofmt`, `unconvert`, `errcheck`)
- **Unit tests supabase** — 44 tests, coverage 84.1%
- **Test flakiness** — `TestQueueFull` race condition fixed, `TestRealMySQLConnectNoOptionalConfig` timeout fixed
- **Dead code removal** — entire `realtime.go` (151 lines) + `DeployFunctionFromFile` removed
- **CI hardening** — codecov-action `file:` → `files:`, coverage fail-safe
- **CLI docs regenerated** — 21 `naeos_supabase*.md` + auto-generated via `docsgen`
- **v2.2.0 release** — GoReleaser binary builds for linux/darwin/windows × amd64/arm64 + Docker image
- **Coverage audit** — ditemukan 5 package sudah ≥80%: `supabase` (84.1%), `messagequeue` (93.5%), `marketplace` (88.7%), `mcp` (85.1%), `migration` (97.9%)
- **Feature inventory** — dikonfirmasi 12 fitur sudah implement: errors package, ai suggest, template publish, OpenAPI CI, blog pipeline, dashboard, RBAC, OAuth2/OIDC

## Notes

- **Prioritas**: Fase 1 dulu — kualitas sebelum fitur baru
- **Website**: Setiap fase include update konten website sesuai fitur yang dirilis
- **CI**: Tiap PR wajib lint + test + coverage check; coverage drop → block merge
- **Dokumentasi**: Tiap API/fitur baru harus include doc PR sebelum code merge