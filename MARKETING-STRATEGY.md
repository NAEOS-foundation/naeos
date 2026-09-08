# NAEOS Marketing Strategy

**Status:** Draft operasional 1.0  
**Tanggal:** 2026-09-08  
**Pemilik:** NAEOS Foundation  
**Sumber utama:** [README.md](README.md), [ARCHITECTURE-OVERVIEW.md](ARCHITECTURE-OVERVIEW.md), [GETTING-STARTED.md](GETTING-STARTED.md), [ROADMAP.md](ROADMAP.md), [CONTRIBUTING.md](CONTRIBUTING.md), [site/content/community.md](site/content/community.md)

Dokumen ini menerjemahkan kondisi repository menjadi strategi marketing yang dapat dijalankan. Setiap klaim produk harus diverifikasi kembali terhadap repository sebelum dipublikasikan. Angka adopsi, pelanggan, revenue, benchmark, partnership, dan compliance **belum terverifikasi dari repository** dan tidak boleh digunakan sebagai proof point.

## 1. Executive Summary

NAEOS diposisikan sebagai **declarative engineering platform** dengan pesan utama **“Specify Once. Build Anywhere.”**. Produk ini menghubungkan spesifikasi YAML/JSON, validasi, NEIR, generation, AI context, governance, dan artifacts dalam satu pipeline.

Strategi 90 hari berfokus pada:

1. Membuat repository dan quick start menjadi jalur aktivasi utama.
2. Menjelaskan masalah engineering di sekitar AI-generated code tanpa menjual hype.
3. Mengubah fitur teknis yang sudah ada menjadi tutorial, demo, diagram, dan issue yang bisa diikuti.
4. Mengembangkan contributor ladder dari pembaca menjadi pengguna, issue reporter, contributor, dan plugin/profile builder.
5. Mengukur aktivasi nyata: kunjungan GitHub, penyelesaian quick start, run pertama, issue/discussion, contributor, profile, dan plugin.

## 2. Current NAEOS Positioning

### What

NAEOS adalah Nusantara Engineering & Architecture Operating System, sebuah platform engineering deklaratif yang mengubah spesifikasi menjadi sistem software melalui pipeline yang konsisten, tervalidasi, dan dapat diperluas.

### Why

AI coding tools dapat mempercepat pembuatan kode, tetapi repository menunjukkan bahwa NAEOS berfokus pada lapisan yang lebih luas: spesifikasi, model engineering, validasi, generation, konteks AI, governance, traceability, dan lifecycle alignment.

### How

`Specification -> Normalization -> Resolution -> NEIR -> Validation -> Scheduling -> Generation -> AI Context -> Governance -> Artifacts -> Verification`

### Category

Kategori utama yang diuji: **Specification-Driven Engineering**. Kategori pendukung: **Declarative Software Engineering** dan **AI-Native Engineering Infrastructure**. Hindari menyebut NAEOS sekadar code generator atau AI coding assistant.

### Verified proof points

- README mendokumentasikan parser, normalizer, resolver, NEIR builder, validator, scheduler, dan generator.
- README mendokumentasikan adapter untuk GitHub Copilot, Claude Code, Cursor, Gemini CLI, Codex, OpenCode, dan Windsurf.
- README mendokumentasikan profile/plugin marketplace serta governance dan audit trail.
- Quick start menyediakan alur `init`, `run`, `context`, dan `ai compile`.
- Website memiliki jalur GitHub, Discussions, contributing, docs, roadmap, dan community.

## 3. Target Audiences

| Audiens | Masalah | Pesan | CTA pertama |
|---|---|---|---|
| AI engineers | Kode AI cepat berubah dan konteks engineering terpecah | NAEOS menyusun spesifikasi dan konteks untuk workflow AI | Jalankan quick start dan compile context |
| Software architects | Keputusan arsitektur tersebar di kode, prompt, tiket, dan dokumen | NEIR memberi model engineering terpusat | Baca architecture overview |
| Platform engineers | Standar lintas tim sulit dipakai konsisten | Profile, policy, plugin, dan pipeline dapat dipakai ulang | Coba profile/policy example |
| Engineering leaders | Kecepatan AI dapat memperbesar drift dan technical debt | NAEOS menambah control layer untuk specification-driven development | Diskusikan use case dan trade-off |
| Open-source developers | Ingin tool yang bisa diperiksa dan diperluas | NAEOS terbuka untuk code, docs, issues, profiles, dan plugins | Star, issue, atau PR |
| AI agent builders | Agent memerlukan konteks engineering terstruktur | NAEOS dapat menyediakan context bundle dan adapter | Eksplorasi AI compiler dan MCP |
| Regulated/enterprise teams | Perlu traceability dan governance | NAEOS menyediakan mekanisme governance yang terdokumentasi | Review policy dan audit docs |

Enterprise compliance, production readiness, customer adoption, dan security certification **belum terverifikasi dari repository**.

## 4. Core Narrative

**Model lama:** intent -> ticket -> prompt -> generated code -> validasi terfragmentasi -> architecture drift.

**Model NAEOS:** engineering specification -> normalized model -> validation -> pipeline -> generated artifacts -> AI context -> governance -> verification -> lifecycle evolution.

Kalimat inti:

> AI dapat menghasilkan kode. NAEOS membantu menstrukturkan sistem engineering di sekitar kode tersebut.

## 5. Messaging Framework

Gunakan urutan berikut untuk setiap pesan:

**Problem -> Consequence -> New approach -> NAEOS -> Proof -> CTA**

Contoh pesan:

- **Problem:** Tim mendeskripsikan system intent di banyak tempat.
- **Consequence:** Konteks AI, arsitektur, dan implementasi dapat drift.
- **New approach:** Jadikan intent engineering machine-readable.
- **NAEOS:** Spesifikasi diproses menjadi model NEIR dan pipeline tervalidasi.
- **Proof:** Repository, architecture docs, quick start, tests, dan CLI tersedia.
- **CTA:** Clone repository dan jalankan spesifikasi pertama.

Terjemahan teknis ke manfaat:

| Konsep | Penjelasan sederhana | Contoh | Manfaat |
|---|---|---|---|
| NEIR | Satu model terstruktur untuk sistem yang dibangun | Project, module, service, API, deployment | Tool downstream memakai konteks yang konsisten |
| Validator | Pemeriksaan sebelum output dibuat | Dependency cycle atau port conflict | Masalah ditemukan lebih awal |
| AI compiler | Mengubah model engineering menjadi instruksi tool AI | `naeos ai compile --target opencode` | Konteks tidak perlu ditulis ulang per tool |
| Profile/plugin marketplace | Komponen reusable untuk workflow | Profile SaaS atau plugin | Eksperimen lebih mudah dibagikan |

## 6. Competitive Landscape

NAEOS berada di sekitar beberapa kategori, bukan klaim pengganti langsung:

| Kategori | Fokus umum | Peran NAEOS |
|---|---|---|
| AI coding assistant | Membantu menghasilkan atau mengubah kode | Menstrukturkan engineering context di sekitarnya |
| Code generator/scaffolder | Membuat struktur awal | Menghubungkan generation dengan specification dan validation |
| IaC | Mendefinisikan infrastruktur | Menempatkan infrastructure dalam model engineering yang lebih luas |
| Internal developer platform | Golden paths dan workflow tim | Menyediakan specification, profiles, policies, dan artifacts |
| Architecture tool | Dokumentasi dan model arsitektur | Menghubungkan model dengan pipeline dan generation |
| CI/CD | Build, test, deploy | Menjadi titik integrasi untuk validation, policy, dan artifacts |
| AI agent framework | Orkestrasi agent | Menyediakan context engineering yang lebih terstruktur |

Jangan menyatakan NAEOS lebih baik atau menggantikan kategori tersebut tanpa bukti repository dan data penggunaan.

## 7. Content Pillars

1. Specification-Driven Engineering: mengapa intent harus machine-readable.
2. AI-Native Engineering: mengapa generated code memerlukan struktur engineering.
3. NEIR: model pusat dan alasan arsitekturalnya.
4. AI Tool Integration: tujuh adapter yang tercantum di README.
5. Governance: policy, artifact review, audit trail, dan verification.
6. Developer Workflow: specification -> validate -> run -> generate -> test -> compile context.
7. Architecture: parser, normalizer, resolver, NEIR, validator, scheduler, generator, compiler.
8. Open Source Journey: release, keputusan, eksperimen, kegagalan, dan lessons learned.
9. Ecosystem: plugins, profiles, marketplace, MCP, dan integrasi yang benar-benar tersedia.
10. Founder Journey: pembangunan terbuka dengan batas klaim yang jelas.

## 8. Channel Strategy

| Prioritas | Channel | Peran | Format |
|---|---|---|---|
| 1 | GitHub | Aktivasi, proof, kontribusi | README, examples, issues, Discussions, releases |
| 1 | LinkedIn | Architecture, leadership, founder voice | Technical post, diagram, lesson |
| 1 | Reddit | Kritik teknis dan problem discovery | Founder experiment, question, failure |
| 1 | Hacker News | Technical credibility | Deep dive, architecture, demo |
| 2 | Dev.to/Medium | Searchable education | Tutorial, comparison, architecture |
| 2 | YouTube | Demonstrasi workflow | 30 detik, 3 menit, architecture explainer |
| 3 | Discord/Slack | Retention dan contributor support | Office hour, Q&A, build log |

Setiap kanal mendapat versi native. Jangan menyalin post yang sama persis ke semua kanal.

## 9. GitHub Growth Strategy

Jalur utama:

`README -> Quick Start -> First successful run -> Experiment -> Star -> Issue/Discussion -> Contribution -> Integration`

Prioritas repository:

1. Pertahankan README sebagai landing page teknis dengan satu quick win.
2. Tambahkan minimal examples untuk `run`, `validate`, `context`, dan `ai compile`.
3. Buat issue berlabel `good first issue`, `documentation`, `profile`, dan `plugin`.
4. Gunakan Discussions untuk pertanyaan arsitektur dan feedback workflow.
5. Setiap release menyertakan why, example, technical detail, migration note, dan contribution path.
6. Tampilkan link strategy, roadmap, contributing, dan architecture secara konsisten.

## 10. Community Strategy

Contributor ladder:

`Observer -> User -> Experimenter -> Issue reporter -> Documentation contributor -> Code contributor -> Plugin/profile contributor -> Maintainer -> Ecosystem partner`

Aktivasi komunitas:

- Mingguan: satu architecture/workflow post dan satu prompt untuk Discussion.
- Dua mingguan: office hour atau walkthrough quick start.
- Setiap milestone: ajakan kontribusi yang spesifik, bukan “help wanted” yang umum.
- Setiap issue: sertakan konteks, expected behavior, evidence, dan acceptance criteria.

## 11. Partnership Strategy

Target riset awal: AI tooling, developer platforms, cloud/DevOps, security tooling, open source, universitas, dan komunitas engineering.

Kriteria: strategic fit, technical fit, audience overlap, integration opportunity, distribution opportunity, credibility benefit, mutual value, dan first collaboration proposal.

Proposal pertama yang realistis:

- review arsitektur terbuka;
- plugin/profile experiment;
- tutorial integrasi;
- benchmark yang metodologinya dipublikasikan;
- research discussion tentang specification-driven engineering.

Tidak ada partnership yang dianggap aktif sebelum tercatat dan diverifikasi.

## 12. SEO Strategy

| Keyword cluster | Intent | Content awal | CTA |
|---|---|---|---|
| specification driven engineering | Informational | What is specification-driven engineering? | Read architecture |
| AI software engineering | Problem discovery | AI-generated code needs engineering context | Try quick start |
| declarative software engineering | Informational | From specification to artifacts | Read README |
| engineering intermediate representation | Technical | NEIR as a central model | Read NEIR docs |
| AI coding governance | Problem discovery | Validation and governance around AI workflows | Review policy docs |
| AI development workflow | Practical | Build, validate, compile AI context | Run example |

Internal links harus mengarah ke README, architecture, getting started, NEIR docs, policy docs, roadmap, dan community. Hindari keyword stuffing.

## 13. Launch Strategy

Gunakan milestone nyata: release, adapter, governance capability, marketplace capability, architecture document, benchmark, atau integration yang sudah ada.

**Pra-launch, 7-14 hari:** problem statement, technical preview, architecture note, example, dan changelog preview.  
**Launch day:** GitHub release, technical article, demo, diagram, dan native channel posts.  
**Pasca-launch:** tutorial, deep dive, feedback thread, known limitations, dan next experiment.

Jangan membuat urgency buatan atau klaim production-ready tanpa bukti.

## 14. 30-Day Calendar

| Hari | Objective | Content / hook | Channel | CTA | Signal |
|---|---|---|---|---|---|
| 1 | Awareness | AI dapat menghasilkan kode, tetapi siapa yang menjaga intent? | LinkedIn | Baca README | Click-through |
| 2 | Problem | Architecture drift setelah AI-assisted coding | Reddit | Beri kritik | Comments |
| 3 | Category | Definisi specification-driven engineering | Blog/GitHub | Read architecture | Page views |
| 4 | Workflow | YAML spec pertama dalam 5 menit | GitHub | Run quick start | Successful run |
| 5 | NEIR | Satu model untuk project dan service | LinkedIn | Read NEIR docs | Saves |
| 6 | Validation | Contoh error yang ditemukan sebelum generation | Blog | Try validate | CLI runs |
| 7 | Recap | Ringkasan minggu pertama + pertanyaan terbuka | Discussions | Join thread | Replies |
| 8 | AI context | Mengapa context bundle penting | LinkedIn | Explore context | Link clicks |
| 9 | Adapter | Satu spec, tujuh target AI | Hacker News | Inspect adapters | Qualified visits |
| 10 | Demo | `naeos ai compile --target opencode` | YouTube/GitHub | Reproduce demo | Reproductions |
| 11 | Architecture | Parser -> resolver -> NEIR | Blog | View diagram | Time on page |
| 12 | Governance | Policy dan artifact review | Reddit | Discuss trade-offs | Comments |
| 13 | Open source | Cara memilih first contribution | GitHub | Open issue/PR | New contributors |
| 14 | Recap | Lessons learned, limitation, next test | LinkedIn | Follow roadmap | Engagement |
| 15 | Profile | Reusable industry profile | Blog | Explore profile | Profile usage |
| 16 | Plugin | Plugin lifecycle dan extension point | GitHub | Propose plugin | Ideas |
| 17 | MCP | Context engineering untuk agent | Hacker News | Read MCP docs | Qualified visits |
| 18 | Comparison | NAEOS dan adjacent tools | Blog | Discuss fit | Search clicks |
| 19 | Tutorial | Spec -> validate -> generate -> test | YouTube | Complete tutorial | Completion |
| 20 | Contributor | Good first issue walkthrough | GitHub | Pick an issue | Issue activity |
| 21 | Recap | FAQ dari pertanyaan komunitas | Discussions | Ask a question | Questions |
| 22 | Architecture | Traceability dari intent ke artifact | LinkedIn | Read docs | Saves |
| 23 | Roadmap | NEIR, plugin, CI/CD, context API | GitHub | Comment roadmap | Feedback |
| 24 | Founder voice | Trade-off membangun sistem terbuka | Reddit | Challenge assumptions | Comments |
| 25 | Use case | SaaS/AI agent profile, dengan batas klaim | Blog | Try example | Runs |
| 26 | Evidence | Tests, docs, dan repository sebagai proof | LinkedIn | Inspect repo | GitHub referrals |
| 27 | Ecosystem | Ide integrasi yang ingin diuji | Discussions | Propose integration | Proposals |
| 28 | Recap | 4 minggu insight dan unresolved questions | Newsletter/Blog | Subscribe/follow | Returning visitors |
| 29 | Activation | From clone to first contribution | GitHub | Open PR/issue | Contributors |
| 30 | Review | Publish experiment results and next 30 days | All native channels | Join next experiment | Activation rate |

Setiap entry memerlukan asset, owner, tanggal publikasi aktual, dan link evidence sebelum dipublikasikan.

## 15. 90-Day Roadmap

| Fase | Minggu | Fokus | Output |
|---|---|---|---|
| Foundation | 1-4 | Positioning, README, docs, measurement | Message guide, examples, UTM convention, baseline metrics |
| Developer activation | 5-8 | Tutorials, demos, issues, Discussions | 3 reproducible tutorials, contributor ladder, office hour |
| Ecosystem | 9-12 | Profiles, plugins, adapters, research | 1 integration experiment, 1 profile/plugin proposal, public findings |

## 16. Asset Requirements

**Brand:** logo, wordmark, colors, typography, iconography, dan visual language dari `brand/brand.json` dan `brand/brand.css`.  
**Product:** quick start recording, CLI screenshots, pipeline diagram, NEIR diagram, adapter matrix, generated artifact examples.  
**Content:** article template, release template, social variants, architecture diagram template, experiment report.  
**Partnership:** one-page technical overview, architecture deck, contribution deck, integration brief.

Setiap asset harus mencantumkan source repository dan tanggal verifikasi.

## 17. KPI Framework

| Funnel | KPI utama | Cara baca |
|---|---|---|
| Awareness | GitHub visitors, search clicks, qualified page views | Apakah masalah dan kategori ditemukan? |
| Activation | Quick start completion, first successful run, CLI example reproductions | Apakah orang berhasil mencoba? |
| Community | Stars, forks, issues, Discussions, contributors, PRs | Apakah ada partisipasi berulang? |
| Adoption | Active users, projects created, profiles, plugins, integrations | Apakah workflow dipakai ulang? |
| Content | CTR, completion, saves, comments, GitHub referrals | Format mana yang membawa tindakan? |

Tetapkan baseline pada minggu pertama. Jangan memakai impressions sebagai KPI tunggal.

## 18. Experiment Backlog

| Priority | Experiment | Hypothesis | Success signal | Status |
|---|---|---|---|---|
| P0 | README quick-start CTA | CTA tunggal meningkatkan first run | Completion rate | Planned |
| P0 | Four-command tutorial | Alur pendek menurunkan activation friction | Successful runs | Planned |
| P1 | NEIR architecture article | Model pusat menarik architects | Qualified visits + replies | Planned |
| P1 | Seven-adapter demo | Interoperability memicu AI engineer interest | Reproductions | Planned |
| P1 | Good-first-issue campaign | Task konkret meningkatkan contributor conversion | New PRs/issues | Planned |
| P2 | Profile challenge | Reusable profile mendorong ecosystem participation | Profile proposals | Planned |
| P2 | Office hour | Live support meningkatkan tutorial completion | Attendees + completed runs | Planned |
| P3 | Research benchmark | Evidence-driven comparison meningkatkan credibility | Qualified feedback | Planned |

Format pencatatan: hypothesis, audience, channel, message, asset, CTA, metric, threshold, duration, result, learning, next experiment.

## 19. Risks

- **Overclaiming:** klaim adoption, enterprise, performance, compliance, atau production tidak didukung repository. Mitigasi: evidence link dan review sebelum publish.
- **Category confusion:** NAEOS dianggap hanya generator atau assistant. Mitigasi: ulangi pipeline dan NEIR dalam setiap explainer.
- **Activation friction:** dokumentasi luas tetapi first run sulit. Mitigasi: tutorial reproducible dan error-oriented.
- **Channel dilution:** terlalu banyak kanal sebelum ada proof. Mitigasi: prioritaskan GitHub, LinkedIn, Reddit, dan Hacker News.
- **Roadmap mismatch:** marketing mendahului implementasi. Mitigasi: sync setiap milestone dengan `ROADMAP.md` dan code/docs.
- **Metrics vanity:** stars tanpa penggunaan. Mitigasi: ukur successful run, contribution, profile, plugin, dan repeat visit.

## 20. Recommended Next Actions

1. Tambahkan link dokumen ini ke documentation index.
2. Jadikan quick start sebagai campaign asset pertama.
3. Buat tiga example yang dapat direproduksi: `run`, `validate`, dan `ai compile`.
4. Buat label GitHub untuk contributor ladder dan satu issue per eksperimen P0.
5. Publikasikan artikel NEIR dengan diagram dari `ARCHITECTURE-OVERVIEW.md`.
6. Catat baseline KPI dan UTM convention sebelum campaign hari pertama.
7. Review dokumen ini setiap release dan tandai klaim yang berubah.

## Evidence and Claim Policy

Marketing hanya boleh menyatakan kemampuan yang terlihat di repository saat verifikasi dilakukan. Untuk fitur roadmap, gunakan bahasa “direncanakan” atau “sedang dikerjakan”. Untuk customer, adoption, benchmark, partnership, certification, dan production deployment, gunakan “Not currently verified from the NAEOS repository” sampai ada evidence yang dapat ditautkan.
