# NAEOS Development Plan — v2.2.0 → v4.0.0 (Enterprise)

## Fase 1: Kualitas & Keandalan

| Item | Area | Detail |
|------|------|--------|
| Test coverage minimum 80% | Backend | ✅ `supabase` 84.1%, `messagequeue` 93.5%, `marketplace` 88.7%, `mcp` 85.1%, `migration` 97.9%. ✅ Target tercapai: `watch` 84.8%, `rollback` 85.7%, `cicd` 88.3%, `distributed` 84.4%, `gateway` 83.9%, `websocket` 80.1%, `eventsourcing` 83.7% |
| CI lint zero-failure ✅ | Backend | Semua pelanggaran `bodyclose`, `noctx`, `gofmt`, `unconvert`, `errcheck` diperbaiki. Lint lulus 100% |
| Integration test suite | Backend | ✅ Supabase integration tests (Auth, Storage, SQL, Admin) di CI tiap commit dengan secrets |
| Fuzz testing | Backend | ✅ 5 fuzz targets jalan di CI tiap commit |
| Error message audit ✅ | UX | `internal/errors/` — 15 error codes (`PARSE_ERROR`, `VALIDATION_ERROR`, `AUTH_ERROR`, dll) + sentinel errors + `ErrorGroup`. ✅ Semua `fmt.Errorf` di production code sudah migrasi ke `naeoserr.New`/`naeoserr.Wrapf` |
| Logging standardization | Backend | ✅ 0 `log.Print` sisa. ✅ 7 `fmt.Println` di `create.go` → helper methods `output`/`outputLine`/`errorLine` (stdout/stderr terpusat) |

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
| Swagger UI page ✅ | Site | Generate halaman `/docs/api/` dari `docs/openapi.yaml` dengan Swagger UI interaktif (CDN v5) |

## Fase 3: Platform & Ekosistem

| Item | Area | Detail |
|------|------|--------|
| Supabase backend integration ✅ | Backend/CLI | Database adapter (pgx), Auth (GoTrue), Storage, Edge Functions, Admin API. CLI: `naeos supabase init/auth/storage/sql/status` |
| Plugin registry publik ✅ | Backend/Site | `registry.json` di `/plugins/`, halaman browse di `/plugins/`, `RemoteRegistry` client di `internal/marketplace/`, GitHub workflow auto-discover `naeos-plugin` topic |
| Plugin template generator ✅ | CLI | `naeos plugin init` — scaffolding dengan SDK boilerplate, tests, Makefile, GitHub Actions CI, WASM entry point |
| NEIR schema registry ✅ | Backend | `Registry` (in-memory semver), `NEIRClient` (fetch from URL/file), `ValidateNEIRSpec`/`ValidateSpec`, CLI `naeos schema validate/info`. Integrasi pipeline via `--schema-source` flag + `Config.SchemaSource` |
| Template marketplace ✅ | CLI | `naeos template publish [path]` — publikasi starter project template ke registry, `naeos template search`, `naeos template init` |
| GoReleaser release pipeline ✅ | CI | Auto-build binary untuk linux/darwin/windows × amd64/arm64 tiap tag. Checksum + Docker image ke ghcr.io |

## Fase 4: Performa & Skalabilitas

| Item | Area | Detail |
|------|------|--------|
| Pipeline caching v2 ✅ | Backend | Stage-level cache keyed by NEIR hash (JSON+SHA256) untuk schedule & generate. StageCache via `Config.StageCache` / `WithStageCache()` option. `--profile` CLI flag menampilkan hit rate |
| Parallel generation ✅ | Backend | Concurrent multi-adapter via `sync.WaitGroup` + goroutine di `CompileAll()`. Benchmark: 3 adapters @ ~1.4ms (1ms sleep per adapter) vs ~3ms sequential |
| Lazy NEIR loading ✅ | Backend | `LazyNEIR` struct dengan per-section lazy accessors (`Modules()`, `Services()`, `Architecture()`, dll). `LazyBuilder` interface + `BuildLazy()` pada `DefaultBuilder`. 16 test |
| Benchmark suite ✅ | Backend | Benchmark terstandarisasi untuk 3 skala (small/medium/large) + single-adapter + 1ms-sleep tiap adapter di `internal/compiler/bench_test.go`. `BenchmarkResult` dengan percentile/histogram/stddev di `internal/profiling` |
| Pipeline profiling ✅ | Backend | `PipelineProfile` mencatat timing & memory per stage (validate, build_graph, policy_eval, schedule, generate, review, write_artifacts). `--profile` flag di `naeos build`. `naeos profile run` subcommand standalone |
| Memory profiling ✅ | QA | `MemProfiler` — heap snapshot tiap stage boundary, heap diffing, GC pressure analysis, leak detection (`DetectLeaks()`, `Analyze()`). `--memprofile` flag di `naeos build` & `naeos profile run` |
| Distributed build real ✅ | Backend | `naeos build --distributed --workers N` — workers jalankan pipeline real, spec di-split per module, hasil di-aggregate (bukan stub fake duration) |

## Fase 5: AI & Developer Experience

| Item | Area | Detail |
|------|------|--------|
| AI recommendation engine ✅ | Backend | `naeos ai suggest` — analisa spec via LLM, rekomendasi arsitektur & best practices. Juga `ai explain`, `ai enrich`, `ai compile` |
| NEIR-aware LSP ✅ | Backend | Language Server Protocol untuk spec YAML: autocomplete, diagnostics, hover info, go-to-definition, document symbols, code actions. Server stdio JSON-RPC 2.0, context-aware completions, real parser integration. |
| VS Code extension ✅ | Plugin | Extension generator via `naeos dx vscode-gen` — TextMate grammar, LSP client integration, commands (compile/validate/dashboard), keybindings, menus, config. Output: package.json, extension.js, syntaxes/, README. |
| NEIR diff visualization ✅ | CLI/TUI | `naeos diff --format unified` (unified diff), `--visual` side-by-side tree view, `FormatSpecDiff` untuk perbandingan spec-level. Semua format terintegrasi di CLI |

## Fase 6: Rilis v3.0.0

| Item | Area | Detail |
|------|------|--------|
| NEIR v2.0 specification ✅ | Core | Conditional modules (`Condition` field + `ConditionalResolver`), environment profiles (`ActiveProfile`/`Inherits` pada model + `ProfileResolver`), `$if{}/$endif` di parser + resolver level |
| GUI Dashboard ✅ | Site | `naeos dashboard` — web dashboard dengan stats, activity log, component health, WebSocket live updates |
| RBAC ✅ | Backend | `internal/auth/rbac.go` — admin/developer/viewer roles, **hierarchy (`Parents`) + deny rules (`Deny`) override**, `SetupRoleTemplate()` untuk 4 compliance templates |
| OAuth2/OIDC ✅ | Backend | Google OAuth2 provider + `/.well-known/openid-configuration` + `/.well-known/jwks.json`. **OIDC Provider** (discovery, JWKS RSA sig verify, auth code flow) |
| Enterprise features ✅ | Backend | **SSO**: OIDC, SAML 2.0 (XML parsing, NameID/attribute extraction), LDAP (TCP/TLS bind, ASN.1 BER search). **Audit**: hashed chain (`HashedAuditor`), encrypted (`EncryptedAuditor` AES-256-GCM), cloud export (AWS SigV4, GCS HMAC, Azure SharedKey). **Compliance**: SOC2 (8 controls), HIPAA (11), GDPR (8), `GenerateReport()`, CLI `naeos compliance report/list-frameworks/verify/cloud-export` |
| v3.0.0 release ✅ | All | Changelog, migration guide v2→v3, release party blog post, deprecation notices. **Unreleased section di CHANGELOG.md sudah berisi semua 20+ item fitur v3.0.0.** |

## Kebijakan Versi (Semver Ketat)

Mengikuti semantic versioning secara disiplin — **tidak ada lompatan nomor versi**:

| Jenis rilis | Format | Isi yang diperbolehkan |
|-------------|--------|------------------------|
| Patch | `x.y.z` → `x.y.(z+1)` | Bugfix, security fix, dependensi update. **Tanpa** fitur baru & perubahan API |
| Minor | `x.y.z` → `x.(y+1).0` | Fitur baru backward-compatible |
| Major | `x.y.z` → `(x+1).0.0` | Hanya jika ada breaking change yang tak terhindarkan |

Aturan wajib:
- `VERSION` di repo harus selalu sama dengan tag rilis (verifikasi di CI: tag == `v$(cat VERSION)`).
- CHANGELOG diisi untuk setiap versi **sebelum** merge, bukan setelah rilis.
- Setiap PR wajib mencantumkan target versi; CI memastikan kenaikan tepat satu angka (no-skip).
- Setiap Major/Minor diikuti patch recovery bila perlu (mis. `v3.2.0` → `v3.2.1`).
- Fitur baru = Minor. Bugfix = Patch. Tidak ada fitur baru dalam Patch.

## Release Train — Arah Enterprise (v3.1.1 → v4.0.0)

Roadmap enterprise self-hosted dibagi menjadi rilis kecil berurutan. Setiap rilis = kenaikan tepat satu versi dari rilis sebelumnya.

| Rilis | Jenis | Fase | Isi utama |
|-------|-------|------|-----------|
| v3.1.1 | Patch | Fase 7 | Backlog bugfix + security fix + dependensi update |
| v3.1.2 | Patch | Fase 7 | Perbaikan race condition, queue, LSP, parser edge cases |
| v3.1.3 | Patch | Fase 8 | `govulncheck` zero-critical + secret-scan CI + perbaikan penanganan secret |
| v3.2.0 | Minor | Fase 7 | `naeos serve` daemon produksi: TLS, graceful shutdown, multi-listener, systemd unit, server config file |
| v3.2.1 | Patch | Fase 7 | Follow-up hardening `naeos serve` |
| v3.3.0 | Minor | Fase 8 | SBOM (Syft/CycloneDX) + Cosign signing + `naeos verify` + aset rilis tersign |
| v3.4.0 | Minor | Fase 7 | Helm chart + Kustomize, bundle air-gapped, config provider (env/file/K8s Secret/Vault), `naeos backup`/`restore`, API server stateless (state store Postgres) |
| v3.5.0 | Minor | Fase 10 | Ekspor tracing OpenTelemetry (OTLP) + korelasi request_id → trace_id → tenant + SLO & alerting Prometheus + ekspor audit ke SIEM |
| v3.6.0 | Minor | Fase 11 | Durable job queue (Postgres outbox) + worker pipeline jaringan (NATS/Kafka) + idempotency |
| v3.7.0 | Minor | Fase 9 | REST API v2: pagination cursor, idempotency key, error envelope RFC 7807, rate-limit berjenjang, kebijakan versioning API |
| v3.8.0 | Minor | Fase 9 | Webhooks/events outbound (HMAC-signed, retry + dead-letter, katalog event) |
| v3.9.0 | Minor | Fase 9 | SDK resmi Go + TypeScript via OpenAPI codegen + drift-check CI |
| v3.10.0 | Minor | Fase 8 | MFA TOTP/WebAuthn + kebijakan password + rotasi session/refresh token + API key terpetakan |
| v3.11.0 | Minor | Fase 8 | SCIM 2.0 provisioning + JIT provisioning + break-glass & access review + FIPS-ready |
| v3.12.0 | Minor | Fase 11 | Kuota/limit per-tenant (concurrency, artefak) + tier cache Redis + circuit breaker & retry policy |
| v3.13.0 | Minor | Fase 12 | Policy-as-code antar-tenant + drift detection (`naeos plan`) + approval workflow & change management |
| v3.14.0 | Minor | Fase 12 | Audit trail immutable/WORM/object storage + kebijakan retensi + bukti compliance live (SOC2/HIPAA/GDPR) |
| v3.15.0 | Minor | Fase 12 | Edition/license gating (community vs enterprise), RBAC delegation + access reviews lanjutan |
| v4.0.0 | Major | Semua | Hanya bila ada breaking change riil; jika tidak, nama rilis tetap v3.16.0 dst |

Catatan:
- **Prioritas**: stabil (Patch) → ops (v3.2.0–v3.4.0) → observability (v3.5.0) → skala (v3.6.0) → API/integrasi (v3.7.0–v3.9.0) → security lanjutan (v3.10.0–v3.11.0) → governance (v3.12.0–v3.15.0).
- Setiap rilis Minor wajib menyertakan: CHANGELOG entry, CLI docs `naeos docsgen` regenerate, update website/site konten, dan doc PR sebelum code merge.
- Jika sebuah item belum siap pada slotnya, rilis tersebut tetap rilis (berisi item yang siap) dan item dipindah ke slot berikutnya — **nomor versi tidak pernah di-skip**.

## Metrik Progress

| Metrik | Saat Ini | Target Q1 2027 | Target Q3 2027 |
|--------|----------|----------------|----------------|
| Test coverage (overall) | ✅ 81.5%+ | ≥80% | ≥85% |
| Test coverage (target packages) | ✅ ≥80% (13 packages: watch 84.8%, rollback 85.7%, cicd 88.3%, distributed 84.4%, gateway 83.9%, eventsourcing 83.7%, **websocket 95.8%**, **configschema 94.7%**, **monitoring 93.4%**, **configreload 91.9%**, **database 87.1%**, **auth 92.5%**, supabase 84.1%) | — | — |
| CLI commands test coverage | ✅ 80.8% (693 test, 39 command) | 100% | 100% |
| Website pages (EN) | ~59 | 35+ (wiki migrated) | 40+ |
| Blog posts | 8 | 6+ | 12+ |
| Plugin ecosystem | 0 | 5+ community plugins | 20+ |
| Build time (pipeline) | ~2s (small) | <1s (small) | <5s (medium) |
| CI lint pass rate | 100% ✅ | 100% | 100% |
| `fmt.Println`/`log.Print` sisa | 0 | 0 | 0 |
| Versi rilis — no-skip (urutan rilis == urutan versi) | — | ✅ terbukti di v3.1.1–v3.1.3 | ✅ terbukti hingga v3.10.0 |
| `naeos serve` daemon (TLS, graceful shutdown) | ❌ belum ada | ✅ v3.2.0 | ✅ + systemd unit |
| SBOM + artifact signing | ❌ belum ada | ✅ v3.3.0 (SBOM + Cosign + `naeos verify`) | ✅ di setiap rilis |
| Dependency vuln scan (govulncheck) | ❌ belum ada di CI | ✅ v3.1.3 zero-critical | ✅ zero-critical terus |
| Helm chart & air-gap bundle | ❌ belum ada | ✅ v3.4.0 | ✅ v3.4.0 + Kustomize |
| Observability (OTLP tracing + SLO) | ❌ metrics Prometheus saja | ✅ v3.4.0 metrics | ✅ v3.5.0 OTLP + SLO |
| Durable job queue + worker jaringan | ❌ async in-memory saja | ✅ v3.6.0 (Postgres outbox) | ✅ NATS/Kafka workers |
| REST API v2 (cursor, idempotency, RFC 7807) | ❌ v1 | ✅ v3.7.0 | ✅ + webhooks v3.8.0 + SDK v3.9.0 |
| MFA + SCIM provisioning | ❌ belum ada | review (v3.10.0–v3.11.0) | ✅ v3.11.0 SCIM + MFA |
| Governance (policy-as-code, drift, approval) | ❌ belum ada | design (v3.12.0–v3.13.0) | ✅ v3.14.0 sebagian besar |
| Kehilangan data audit | 0 | 0 | 0 (immutable/WORM v3.14.0) |

## Completed (v2.2.0 → v3.0.0)

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
- **NEIR v2.0 spec** — conditional modules (`Condition` field, `ConditionalResolver`), environment profiles (`ActiveProfile`/`Inherits`, `ProfileResolver`), builtin function `env()`/`eq()`/`has_prefix()`, validator `validateConditionExpr()` + 8 resolver tests
- **RBAC hierarchy + deny** — `Parents []string` parent chain, `Deny` override (deny wins over allow), `hasPermissionRecursive()`, 4 compliance role templates (auditor, soc2_auditor, gdpr_admin, hipaa_admin), CLI `naeos auth create-role --parents --deny`
- **Audit hashed chain** — `HashedAuditor` dengan SHA256 previous-hash chain, `VerifyChain()`/`VerifyChainFile()` tamper detection
- **Audit encrypted** — `EncryptedAuditor` AES-256-GCM, `DecryptedReader`, `NewEncryptedFileAuditor()`
- **Audit cloud export** — `CloudExporter` interface, `S3Exporter` (AWS SigV4), `GCSExporter` (GOOG1 HMAC), `AzureBlobExporter` (SharedKey), CLI `naeos compliance cloud-export`
- **Compliance frameworks** — SOC2 (8 controls: CC1.1–CC8.1), HIPAA (11 controls: 164.308–164.312), GDPR (8 articles: 5, 7, 17, 25, 30, 32, 33, 35), `GenerateReport()` evaluasi audit events + evidence, CLI `naeos compliance report/list-frameworks/verify`
- **SSO framework** — OIDC (discovery `.well-known/openid-configuration`, JWKS RSA signature verification, auth code + token exchange), SAML 2.0 (XML Response parsing, NameID/attribute extraction), LDAP (TCP/TLS bind, ASN.1 BER search encode/decode), `SSORegistry` di `Manager`, CLI `naeos auth sso configure/list/remove`
- **Pipeline caching v2** — stage-level cache (`schedule`, `generate`) keyed by NEIR hash, `StageCache` dengan file-backed store, `WithStageCache()` option
- **Lazy NEIR loading** — `LazyNEIR` per-section lazy accessor, `LazyBuilder` interface, backward-compatible dengan `Builder` existing
- **Memory profiling** — `MemProfiler` (heap snapshot, GC tracking, leak detection), terintegrasi ke pipeline via `--memprofile`
- **Pipeline profiling** — `PipelineProfile` stage timing + memory stats, `--profile` flag, `naeos profile run` subcommand
- **Schema-based spec validation** — `ValidateSpec()` in-memory, integrasi pipeline via `--schema-source`, `naeos build --schema-source file://schema.json`
- **Distributed build real** — `naeos build --distributed` sekarang jalankan pipeline real per-module (bukan stub fake duration)
- **OAuth2 refactor** — type aliases (`GoogleOAuth2 = GenericOAuth2`) diganti embedded structs, `ScopeStr` derived dari `Config.Scopes` via `strings.Join`
- **Database factory refactor** — `NewFromConfig` shared antara `factory.go` dan `factory_nosql.go`, `"mariadb"` alias untuk MySQL
- **Workflow error logging** — 14 `_ = w.Machine.Trigger(...)` → `slog.Warn(...)`, semua `warn()` helper konsisten
- **BenchmarkResult bugfix** — method `durations()` selalu return nil → fixed dengan field `allDurations` + public `Durations()`
- **NEIR diff visualization** — `naeos diff --visual` side-by-side tree, `--format unified` unified diff, `FormatSpecDiff` spec-level comparison
- **VS Code extension generator** — `naeos dx vscode-gen` dengan output directory, TextMate grammar, LSP client, commands, keybindings, menus, config
- **NEIR-aware LSP** — 30 test, 7 file server stdio JSON-RPC 2.0, code actions, fix duplicate module detection bug
- **CLI tests added** — 9 test baru untuk diff, build, lsp, dx, vscode-gen commands
- **Rollback coverage boost** — 73.9% → 85.7% (merge-restore, walk errors, import traversal/symlink, export failures, List/Latest edge cases)
- **CLI coverage expanded** — 31 test baru untuk version, config, init, status, health, template, plugin, schema, compliance, search, ai, security, docgen, auth, dashboard, observability, events, mcp, validate, workflow, supabase, marketplace, watch, rollback, lock, migrate, deploy, run, create, lint, cicd, gateway, sso, profile, artifacts
- **Plugin search fix** — `naeos plugin search` terhubung ke `RemoteRegistry` (dari hardcoded stub); default URL kini menunjuk GitHub Pages static registry (`naeos-foundation.github.io/naeos/plugins/registry.json`), URL `.json` dieksekusi langsung, `platforms` array dari workflow discovery di-decode, dokuments & contoh diperbarui dari `registry.naeos.dev` fiktif
- **Database coverage boost** — `internal/database` 59.7% → 87.1%: 25 test baru untuk connected paths MySQL/PostgreSQL/Supabase via SQLite in-memory + `go-sqlmock` (Exec/Query/QueryRow, transaksi, Migrate/Rollback success+error, Supabase REST *Context, query cache, RealSupabaseTx)
- **Auth coverage boost** — `internal/auth` 66.8% → 92.5%: 16 test baru (LDAP Authenticate via mock BER TCP server, bind success/failure/user-bind/TLS, OIDC Discover/ExchangeCode/GetUser dengan RSA-signed JWT + JWKS, verifyIDToken, SAML loadCertFile)
- **DB adapter `Begin()` bugfix** — 4 adapter real memanggil `defer cancel()` pada ctx `BeginTx` sehingga semua transaksi langsung invalid; sekarang `context.Background()`
- **LDAP bind parsing bugfix** — pemeriksaan `resp[n-2] == 0x0a` tidak pernah cocok dengan BindResponse nyata (tag ENUMERATED di `n-3`); sekarang memindai `0x0a 0x01` result code
- **LDAP search parsing rewrite** — `parseLDAPResult` pakai BER decoder minimal (`parseBER`), attribute values diekstrak dari `SearchResultEntry` standar; fuzz targets di-konsolidasi ke `FuzzParseBER`
- **Websocket coverage boost** — 80.1% → 95.8%: HTTP upgrade success/failure, readPump (ping→pong, invalid JSON, unknown type, interceptor block), writePump broadcast + close frame, Stop dengan connected client, disconnect unregister (13 test HTTP end-to-end via gorilla dialer, sebelumnya hanya di build-tag integration)
- **Configschema coverage boost** — 80.2% → 94.7%: AddValidator/AddCustomType, ValidateFile (YAML/JSON/unknown-ext/missing), ValidateData (yaml/yml/json + invalid), LoadSchemaFromFile (yaml/json/missing/bad parse)
- **Monitoring coverage boost** — 81.7% → 93.4%: GaugeSet edge cases (unknown family/existing/new/cardinality), statusResponseWriter (WriteHeader/Write/Unwrap), RecordRuntime, StartRuntimeCollector, ObservePipelineDuration, IncSpecValidations, IncArtifacts, SetWebSocketConnections, MetricsObserver lifecycle, middleware status capture
- **Configreload coverage boost** — 81.9% → 91.9%: GetString/GetInt/GetBool type-mismatch, LastModified, HotReloader Start twice/invalid dir/Stop not-running, loop error channel, non-matching event, matchesConfigFile, closed watcher channels terminate loop
- **WASM coverage boost** — 27% → 98% (getter tests, marshal error path)
- **LSP coverage boost** — 74% → 86% (CodeAction, Hover, Definition, handleMessage)
- **Audit coverage boost** — 78% → 81% (ExportCSV, escapeCSV, edge cases)
- **v3.0.0 changelog** — Unreleased section dengan 20+ item fitur baru untuk rilis v3.0.0

## Notes

- **Prioritas**: Fase 1 dulu — kualitas sebelum fitur baru
- **Versi**: Semver ketat, no-skip — lihat [Kebijakan Versi] dan [Release Train] di atas
- **Website**: Setiap fase include update konten website sesuai fitur yang dirilis
- **CI**: Tiap PR wajib lint + test + coverage check; coverage drop → block merge
- **Dokumentasi**: Tiap API/fitur baru harus include doc PR sebelum code merge