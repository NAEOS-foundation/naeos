# NAEOS Development Plan — v2.1.0 → v3.0.0

## Fase 1: Kualitas & Keandalan

| Item | Area | Detail |
|------|------|--------|
| Test coverage minimum 80% | Backend | Packages dengan coverage <50%: `marketplace`, `mcp`, `migration`, `watch`, `rollback`, `cicd`, `distributed`, `gateway`, `graphql`, `messagequeue`, `websocket`, `eventsourcing` — tambah table-driven tests |
| Integration test suite | Backend | Pipeline E2E dari spec → generate → compile dengan fixture real (bukan stub). Jalankan di CI tiap commit |
| Fuzz testing | Backend | Fuzz parser, resolver, compiler input untuk boundary cases |
| Error message audit | UX | Semua error dari `errors/` package → human-readable + actionable messages. Tambah error codes |
| Logging standardization | Backend | Ganti semua `fmt.Println`/`log.Print` sisa → `log/slog` structured logging via kernel |

## Fase 2: Website & Dokumentasi

| Item | Area | Detail |
|------|------|--------|
| Wiki → Hugo migration ✅ | Site | Semua halaman wiki sudah dimigrasi ke Hugo site. Wiki/ dihapus. |
| API docs auto-generate | Site | CI job baca `docs/openapi.yaml` → generate Swagger UI page di `/docs/api/` |
| Blog content pipeline | Site | GitHub Action: detect release tag → auto-create blog post dari changelog |
| Interactive playground | Site | Integrasi xterm.js + WebSocket ke server demo di homepage |
| PDF generation ✅ | Site | CLI reference + getting-started sebagai PDF download via GitHub Action (`pdf-docs.yml`). Tersedia di `/downloads/` |
| Dark mode OG image | Site | Generate PNG OG image yang sesuai dark theme + light theme |

## Fase 3: Platform & Ekosistem

| Item | Area | Detail |
|------|------|--------|
| Plugin registry publik | Backend/Site | Bikin `marketplace.naeos.dev` API + halaman browse plugin dari GitHub topic `naeos-plugin` |
| Plugin template generator | CLI | `naeos plugin init` — scaffolding plugin project dengan SDK boilerplate, testing, CI template |
| NEIR schema registry | Backend | Host NEIR JSON Schema di `schemaregistry/` dengan versioning, validasi spec terhadap schema terbaru |
| Template marketplace | CLI/Site | Publikasi template starter project (microservices-go, serverless-ts, dll) via `naeos template publish` |

## Fase 4: Performa & Skalabilitas

| Item | Area | Detail |
|------|------|--------|
| Pipeline caching v2 | Backend | Cache partial: skip stage jika input spec tidak berubah (incremental build nyata) |
| Parallel generation | Backend | Generate multi-module secara concurrent (saat ini sequential) |
| Lazy NEIR loading | Backend | Load NEIR model on-demand untuk proyek besar (1000+ module) |
| Benchmark suite | QA | Benchmark pipeline untuk 3 skala: small (5 modul), medium (50), large (500). Target <5s untuk small |
| Memory profiling | QA | Profiling leak di parser + compiler untuk spec besar. Target <100MB untuk medium |

## Fase 5: AI & Developer Experience

| Item | Area | Detail |
|------|------|--------|
| NEIR-aware LSP | Backend | Language Server Protocol untuk spec YAML: autocomplete, diagnostics, hover info, go-to-definition |
| VS Code extension | Plugin | Extension dengan syntax highlighting, LSP integration, inline validation, playground |
| AI recommendation engine | Backend | `naeos ai suggest` — analisa spec dan rekomendasi arsitektur, pola, best practices berdasarkan knowledge graph |
| NEIR diff visualization | CLI/TUI | `naeos diff --visual` — side-by-side tree view spec changes |

## Fase 6: Rilis v3.0.0

| Item | Area | Detail |
|------|------|--------|
| NEIR v2.0 specification | Core | Conditional modules, environment profiles, inheritance, multi-file inheritance |
| GUI Dashboard | Site | Visual project management — drag-and-drop module graph, real-time pipeline status |
| Enterprise features | Backend | SSO (OIDC), audit log export (JSON/Splunk), team RBAC, compliance reports (SOC2, HIPAA) |
| v3.0.0 release | All | Changelog, migration guide v2→v3, release party blog post, deprecation notices |

## Metrik Progress

| Metrik | Saat Ini | Target Q1 2027 | Target Q3 2027 |
|--------|----------|----------------|----------------|
| Test coverage | ~60% | ≥80% | ≥85% |
| CLI commands test coverage | ~50% | 100% | 100% |
| Website pages (EN) | 24 | 35+ (wiki migrated) | 40+ |
| Blog posts | 2 | 6+ | 12+ |
| Plugin ecosystem | 0 | 5+ community plugins | 20+ |
| Build time (pipeline) | ~2s (small) | <1s (small) | <5s (medium) |

## Notes

- **Prioritas**: Fase 1 dulu — kualitas sebelum fitur baru
- **Website**: Setiap fase include update konten website sesuai fitur yang dirilis
- **CI**: Tiap PR wajib lint + test + coverage check; coverage drop → block merge
- **Dokumentasi**: Tiap API/fitur baru harus include doc PR sebelum code merge
