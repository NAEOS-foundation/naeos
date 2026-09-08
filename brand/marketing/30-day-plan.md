# NAEOS 30-Day Marketing Engine

Plan sesuai AGENTS.md Mode 15. Channel disingkat:
HN = Hacker News, X = Twitter/X, LI = LinkedIn, RD = Reddit, GH = GitHub,
DW = Dev.to, MD = Medium, YT = YouTube, IH = Indie Hackers.

Aset yang disingkat:
- `cli-demo` = `examples/demo-cli/` (runnable, verified)
- `HN-cli` = `brand/marketing/hn-post-cli-demo.md`
- `X-cli` = `brand/marketing/x-thread-2-cli-demo.md`
- `LI-cli` = `brand/marketing/linkedin-post-cli-demo.md`
- `RD-cli` = `brand/marketing/reddit-post-cli-demo.md`
- `SVG` = `brand/marketing/*.svg` (pipeline, hero, banners)
- `script` = `brand/marketing/demo-script.md` / `video-demo-storyboard.md`

Semua klaim teknis hanya dari repositori (Positioning inti: "Specify Once.
Build Anywhere").

---

## Minggu 1 — CLI Demo Launch & Kata Kunci

| DAY | OBJECTIVE | CONTENT | CHANNEL | FORMAT | HOOK | CTA | ASSET REQUIRED | EXPECTED SIGNAL | FOLLOW-UP ACTION |
|---|---|---|---|---|---|---|---|---|---|
| 1 | Awareness | "AI coding tools cepat. Tapi apakah sistem di belakangnya koheren?" + demo yang bisa dijalankan | HN | Show HN post | "AI coding tools are fast. But can you verify what they build?" | Jalankan `run-demo.sh` | HN-cli | Clone & first run | Ambil feedback jadi issue/Discussion |
| 2 | Awareness | Announcement resmi demo CLI (bilingual) | X | Thread 1/7 | "2 commands. Spec → Go + TypeScript." | Coba di mesin Anda | X-cli | Replies, screenshots | Pantau mention, balas cepat |
| 3 | First run | Tutorial: spec → validate → generate → inspect artifacts | DW/MD | Tutorial | "Ini yang terjadi saat NAEOS validasi spec Anda" | Ikuti quick start | `getting-started-tutorial` (site) | Tutorial views, claps | Tambah komentar/tanya jawab |
| 4 | Credibility | Announcement #95 cross-post ke publikasi | LI | Announcement | "Spesifikasi dulu. AI siap." | Jelajahi repo | SVG | LI engagement | Cek klik GitHub |
| 5 | Community | "Kenapa saya bangun ini" — founder journey | RD | Founder post | "Saya eksperimmen pendekatan berbeda pada AI-assisted engineering" | Kritik teknis diterima | RD-cli | Komentar teknis | Kumpulkan 3 kritik terbaik jadi artikel |
| 6 | Education | NEIR: bagaimana spec menjadi model engineering | DW/MD | Series episode 1 | "Satu model untuk semua tool: NEIR" | Baca arsitektur | Architecture diagram | Reads | Plan episode 2 |
| 7 | Recap | Weekly technical recap + demo hasil | X/GH | Thread + discussion | "Minggu ini NAEOS: recape 7 hari" | Cek changelog | SVG/changelog | Engaged followers | Struktur untuk minggu 2 |

---

## Minggu 2 — Specification-Driven Engineering (Fondasi Konsep)

| DAY | OBJECTIVE | CONTENT | CHANNEL | FORMAT | HOOK | CTA | ASSET REQUIRED | EXPECTED SIGNAL | FOLLOW-UP ACTION |
|---|---|---|---|---|---|---|---|---|---|
| 8 | Education | "Kenapa spec itu penting. Kenapa code saja tidak cukup" | MD/DW | Long-form | "Kode bukan satu-satunya sumber kebenaran" | Baca dan bagikan | Repo sources | Reads, backlinks | Tambah ke SEO keyword map |
| 9 | Awareness | "AI-native engineering butuh struktur lebih kuat" | X | Thread | "AI butuh sistem yang bisa divalidasi, bukan hanya menghasilkan" | Tonton penjelasan 3-menit | YT short script | Thread views | Repurposing ke LI |
| 10 | Activation | Demo dimodifikasi: jalankan pipeline dengan output custom | GH/Discussions | Tutorial + challenge | "Ulangi demo — ubah spec Anda sendiri" | Bagikan hasil Anda | cli-demo | User-run hasil | Buat showcase postingan hasil user |
| 11 | Governance | Policies, artifact review, audit trail (untuk enterprise/regulated) | LI | Article | "AI cepat, tapi siapa yang memverifikasi?" | Lihat governance docs | Governance diagram | LI leads DM | Identifikasi design partner |
| 12 | Ecosystem | Plugin & profile: perluas NAEOS di luar pipeline inti | DW/MD | How-to | "Pakai NAEOS sebagai platform, bukan tool satu-trik" | Explore plugins | Plugin registry | Registry contributors | Jadwalkan plugin spotlight |
| 13 | Community | First Contribution challenge | GH | Discussion + issue | "Kontribusi pertama Anda ke NAEOS" | Pilih good-first-issue | Good first issues | Issues/PR baru | Review & merge cepat |
| 14 | Recap | Minggu 2: konsep, tutorial, ecosystem, governance | X/GH | Thread | "Fondasi konsep sudah jelas — sekarang demo" | Cek roadmap | SVG | Followers growth | Struktur minggu 3 |

---

## Minggu 3 — AI Tool Integration & Arsitektur

| DAY | OBJECTIVE | CONTENT | CHANNEL | FORMAT | HOOK | CTA | ASSET REQUIRED | EXPECTED SIGNAL | FOLLOW-UP ACTION |
|---|---|---|---|---|---|---|---|---|---|
| 15 | Education | Parser → Normalizer → Resolver → NEIR (end-to-end) | MD/DW | Architecture deep dive | "Perjalanan spec hingga menjadi sistem" | Baca peta arsitektur | `ARCHITECTURE-OVERVIEW.md` | Reads, queries | Buat diagram SVG versi kasual |
| 16 | Integration | Memetakan NAEOS ke GitHub Copilot, Cursor, Claude Code, Gemini CLI, Codex, OpenCode (verified adapters) | DW/HN | Technical comparison | "Satu spec, 6 tool AI, satu konteks" | Lihat adapter yang ada | AI adapters list | Experiments | Co-marketing with tool communities |
| 17 | Activation | Mencoba `naeos ai compile` untuk context tool | GH/Discussions | Tutorial | "Compile konteks AI untuk tool yang Anda pakai" | Bagikan output Anda | `naeos ai compile` example | User experiments | Kumpulkan contoh nyata |
| 18 | Awareness | "AI coding tools hanya mengoptimalkan codegen. NAEOS strukturkan sistem di sekitarnya" | X | Thread | "Bukan pengganti Copilot, tapi sistem yang menyatukannya" | Pahami perbedaannya | LI template | Engagement | Siapkan article "NAEOS vs scaffolding tools" (ada di site/blog) |
| 19 | Governance | SBOM, verification, deployment controls | LI | Article | "Rantai pasok software butuh verifikasi" | Lihat evidence/sign tooling | SBOM assets | Enterprise DM | Schedule enterprise demo |
| 20 | Community | Open source journey: lessons, trade-offs, failure | RD | Founder post part 2 | "Saya juga pernah membuat keputusan yang salah" | Bagikan pengalaman Anda | Founder notes | Komentar reflektif | Curate ke blog |
| 21 | Recap | Minggu 3: AI integration, arsitektur, governance | X/GH | Thread | "AI tool integration dalam 3 menit" | Cek wiki/Docs | SVG + arch diagram | Engaged followers | Struktur minggu 4 |

---

## Minggu 4 — Reconciliation, Review, Roadmap

| DAY | OBJECTIVE | CONTENT | CHANNEL | FORMAT | HOOK | CTA | ASSET REQUIRED | EXPECTED SIGNAL | FOLLOW-UP ACTION |
|---|---|---|---|---|---|---|---|---|---|
| 22 | Adoption | Case study: reviewer menjalankan demo CLI sebagai percobaan | DW/MD | Case study | "Enam alasan sebuah tim mencoba spec-driven engineering" | Baca studi kasus | `first-cli-demo-case-study.md` (site) | Case study reads | Plan landing page CTA |
| 23 | Education | "Apa yang terjadi jika arsitektur menjadi machine-readable" | HN | Technical essay | "Arsitektur yang bisa diaudit bukan mimpi lagi" | Baca essay | NEIR essay draft | HN discussions | Pitch ke podcast/forum arsitektur |
| 24 | Activation | 30 days: 30 penonton mencobaLagi funnel activation | GH/Discussions | Challenge | "30 hari: jalankanlah NAEOS sekali" | Kirim bukti run | cli-demo | 30 demo runs | Publikasi hasil agregat |
| 25 | Partnership | University/research collab interest | LI | Partnership | "NAEOS adalah open source & untuk riset" | DM untuk kolaborasi | Overview deck | University outreach | Follow-up email |
| 26 | Recap+Ecosystem | Review semua adapter, plugin, profile yang hidup | GH | Report | "Ekosistem NAEOS bulan ini" | Lihat registry | Registry data | Registry commits | Plan berikutnya |
| 27 | Roadmap | Publikasi roadmap publik + 'what's next' | X/GH | Thread | "Roadmap NAEOS: apa yang kami bangun berikutnya" | Komentar roadmap | `ROADMAP.md` | Roadmap feedback | Prioritaskan atas feedback |
| 28 | Reflection | Founders' lessons: bagaimana kami membangun in public | RD | Founder essay | "Ini yang kami pelajari dalam 28 hari" | Baca & beri masukan | Founder notes | Komentar dalam | Jadwalkan AMA |
| 29 | AMA/community | AMA sesi dengan maintainer/founder (video/teks) | GH/Discord | Live discussion | "Tanya apa saja soal NAEOS" | Gabung sesi | Video/script | Active attendees | Convert ke new contributors |
| 30 | Recap | 30 days results + next 30 days preview | X/GH | Thread | "30 hari NAEOS — angka & pelajaran" | Langganan / cek backlog | Metrics dashboard, BACKLOG.md | Subscribers, issues | Kick-off 30-day berikutnya |

---

## Catatan Operasional

- Setiap campaign harus menuliskan Hypothesis, Audience, Channel, Message,
  Asset, CTA, Metric, Success threshold (Mode 28).
- Satu cerita (CLI demo) dapat dipakai lintas channel dengan *format native*
  berbeda, bukan salin-tempel.
- Semua data metrik diambil dari GitHub API / analytics; tidak ada klaim
  metrik yang tidak terverifikasi.
- Backlog prioritas: lihat `brand/marketing/BACKLOG.md`.