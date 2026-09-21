# NAEOS 30-Day Hardening Content Plan (09-23 → 10-22)

Plan konten berdasarkan AGENTS.md Mode 15 (30-day marketing engine) dan temuan
eksperimen *policy bypass* (PR #143) serta hardening yang menyusul (PR #154).
Mulai **besok, 2026-09-23**.

Channel disingkat: LI = LinkedIn, X = Twitter/X, GH = GitHub, RD = Reddit,
DW = Dev.to, MD = Medium, YT = YouTube, DI = Discord.

---

## Anchor Fakta (wajib diverifikasi ulang sebelum posting)

Baca dulu `experiments/policy-bypass/reports/EXPERIMENT-REPORT.md` + issue
sebelum menulis angka apa pun. Angka di dokumen ini adalah snapshot 2026-09-22.

- 17 skenario, 5 layer: evaluator / control-plane / reviewer / prompt / pipeline.
- 12 temuan awal; 11 tersisa setelah #144 ditutup (PR #154, main 09-22).
- Issue publik: #144 (CLOSED), #145–#152 (tracking sisanya).
  - HIGH: #145 pipeline ctx, #146 disabled rule senyap, #147 policy kosong, #148 fail-open.
  - MEDIUM: #149 evaluator NaN/Inf/nil/whitespace, #150 heuristik reviewer.
  - INFO/defense: #151 AGENTS advisory, #152 fail-closed terverifikasi.
- 7 AI adapters (copilot, claude, cursor, gemini, codex, windsurf, opencode).
- rilis v3.6.0 (unified control plane + supply-chain hardening).

## Aturan Keras

1. Angka (12/11/total per layer) dibaca dari laporan terkini — bukan hafalan.
2. Jangan sebut "closed" sebelum PR benar-benar merged di main.
3. Setiap post wajib punya minimal satu link GitHub yang bisa diverifikasi.
4. Rekap mingguan = delta riil (closed / merged / "belum ada perubahan").
5. Tidak ada klaim metrik yang tidak terverifikasi (AGENTS.md Mode 19).

---

## Minggu 1 — 09-23..09-29: Audit & Transparansi

| DATE | OBJECTIVE | CONTENT | CHANNEL | FORMAT | HOOK | CTA | ASSET REQUIRED | EXPECTED SIGNAL | FOLLOW-UP ACTION |
|---|---|---|---|---|---|---|---|---|---|
| 23 | Rilis temuan | Eskperimen governance, 11 tersisa, #145–#152 publik | LI, X | Post + thread | "Kami attack governance kami sendiri." | Baca issue | Report + issue list | Share/comment | Kumpulkan prioritas komunitas |
| 24 | Edukasi | Deep dive evaluator: NaN/NaN/Inf/nil/whitespace (#149) | DW, MD | Article | "NaN bisa meloloskan policy `gt`." | Lihat evaluator.go | Code + report | Komentar teknis | Buat issue lanjutan bila perlu |
| 25 | Edukasi | Deep dive control-plane: fail-open vs fail-closed (#148, #152) | HN | Essay | "Fail-open = semua lolos; inilah kenapa default harus fail-closed." | Baca decision logic | control.go | HN discussion | Siapkan follow-up essay |
| 26 | Edukasi | Deep dive reviewer: `to-do` vs `TODO` (#150) | DW | Article | "Heuristik bisa dibodohi. Ini caranya." | Coba reviewer | reviewer.go | Claps/comments | Rencanakan call-for-contributors |
| 27 | Eksposur | Yang ditutup vs yang belum: #144 vs #145 | LI | Post | "Satu sudah ditutup. Yang satunya belum." | Baca PR #154 | PR + issue | DMs | Identifikasi design partner |
| 28 | Komunitas | Prioritas publik: 11 temuan, mana dulu? | GH Discussions | Voting | "Roadmap hardening ditentukan feedback, bukan slide." | Vote #145–#152 | Discussions | Engagement | Prioritaskan dari hasil |
| 29 | Recap | Rekap minggu 1 + thread 7 hari | X, GH | Thread | "Satu minggu menguji guardrail sendiri." | Follow issues | Laporan + thread | Followers | Struktur minggu 2 |

## Minggu 2 — 09-30..10-06: Benahi HIGH (#145–#148)

| DATE | OBJECTIVE | CONTENT | CHANNEL | FORMAT | HOOK | CTA | ASSET REQUIRED | EXPECTED SIGNAL | FOLLOW-UP ACTION |
|---|---|---|---|---|---|---|---|---|---|
| 30 | Kredibilitas | #152: fail-closed terverifikasi (defense win) | LI | Post | "Satu bagian yang benar: deny tak bisa di-spoof." | Baca control.go | control.go + test | Credibility | Package jadi case study |
| 1 | Eksplorasi | #147: policy kosong diam-diam lolos | X | Poll | "Kosong ≠ aman. Setuju?" | Reaksi + komentar | Pipeline code | Poll + replies | Dokumentasikan consent |
| 2 | Fiks (bila merged) | Fix #147: policy kosong kini warning | LI | Post | "Closed. Kosong kini berisik." | Lihat commit | PR link | PR visibility | Lanjut #146 |
| 3 | Eksplorasi | #146: `Enabled:false` = penjaga hilang tanpa jejak | DW | Article | "Menonaktifkan rule menghapus pengaman tanpa tanda." | Diskusi desain | evaluator.go | Comments | Konsultasi desain |
| 4 | Fiks (bila merged) | Fix #146: disabled rule kini tercatat | LI | Post | "Closed. Disabled kini ter-observability." | Lihat PR | PR link | Followers | Lanjut #145 |
| 5 | Status | Progress: N/N HIGH selesai | GH | Milestone | "Dari temuan ke PR ke close." | Track milestone | Milestone | Issues updated | Atur ekspektasi |
| 6 | Recap | Rekap minggu 2: delta riil | X, GH | Thread | "2 HIGH closed, 2 tersisa." | Cek backlog | Backlog | Engagement | Struktur minggu 3 |

## Minggu 3 — 10-07..10-13: MEDIUM + Engineering Practice

| DATE | OBJECTIVE | CONTENT | CHANNEL | FORMAT | HOOK | CTA | ASSET REQUIRED | EXPECTED SIGNAL | FOLLOW-UP ACTION |
|---|---|---|---|---|---|---|---|---|---|
| 7 | Edukasi | Fix #149 (NaN/Inf/nil/whitespace) | DW, MD | Article | "Truthy-kosong adalah bug dalam policy." | Lihat evaluator fix | PR + evaluator.go | Reads | Benchmark perbaikan |
| 8 | Edukasi | Fix #150 (normalisasi heuristik) | DW | Article | "Bobol TODO diotomatiskan — normalisasi mengalahkan substring." | Lihat reviewer fix | PR + reviewer.go | Comments | Buka good-first-issue |
| 9 | Story | 1 temuan → 1 PR → merged dalam 48 jam | LI | Case | "CRITICAL → PR #154 → merged." | Baca cerita | PR #154 | DMs/license | Replikasi untuk #145–#148 |
| 10 | Open source | Eskperimen & report sebagai data publik | HN | Essay | "Governance platform yang mempublikasikan kelemahannya." | Clone harness | experiments/ | Clones | Track GitHub referrals |
| 11 | Roadmap | Backlog hardening live | GH, LI | Post | "Roadmap = daftar issue, bukan slide." | Comment | ROADMAP.md | Feedback | Sinkronisasi issue |
| 12 | Kontribusi | Good-first-issue di evaluator/reviewer | GH | Kampanye | "Betulkan heuristik? Mulai dari sini." | Pick issues | Issue labels | PR baru | Review/merge cepat |
| 13 | Recap | Rekap minggu 3 dengan angka riil | X, GH | Thread | "Standing: closed/total per layer." | Cek report | Report | Followers | Struktur minggu 4 |

## Minggu 4 — 10-14..10-20: Ecosystem & Konfirmasi

| DATE | OBJECTIVE | CONTENT | CHANNEL | FORMAT | HOOK | CTA | ASSET REQUIRED | EXPECTED SIGNAL | FOLLOW-UP ACTION |
|---|---|---|---|---|---|---|---|---|---|
| 14 | Konfirmasi | Harness adversarial di CI (report tiap PR) | LI | Post | "Laporan bypass dibuat tiap change, bukan sekali." | Lihat workflow | policy-bypass.yml | Awareness | Amplify ke komunitas |
| 15 | Edukasi | 7 adapter, satu konteks | DW, HN | Comparison | "Satu spec → Copilot..OpenCode." | Coba `ai compile` | Adapter list + CLI | Experiments | Co-marketing tooling |
| 16 | Trust | SBOM + signed release dalam konteks cerita hardening | LI | Post | "Supply chain: SBOM, NOTICE, tanda tangan Ed25519." | Verify asset | .goreleaser.yaml | Enterprise DM | Enterprise follow-up |
| 17 | Komunitas | AMA: maintainer + founder | DI | Live | "Tanya apa saja soal temuan & arah." | Gabung sesi | Agenda | Attendees | Convert ke kontributor |
| 18 | Conversion | Case study: reviewer menjalankan demo + eskperimen | DW, MD | Case | "Dari skeptis → ikut eskperimen." | Baca case study | demo-cli | Leads | Landing CTA |
| 19 | Recap | 30 hari data: angka riil + kontributor | GH, X | Report | "Eskperimen dalam 30 hari." | Cek report | Metrics dashboard | Commits/issues | Rencana 30 hari berikut |
| 20 | Preview | Next-30: bagian mana difix duluan | LI | Post | "Roadmap berikutnya = prioritas issue." | Comment roadmap | Roadmap Review | Votes | Kick-off siklus baru |

## Hari 1 Draf Lengkap — 09-23

### Saran Objective
Insiden-publik: memaparkan hasil eskperimen secara jujur, membangun kredibilitas
"open source = transparent", dan mengundang komunitas ikut menentukan prioritas.

### Hook
> Kami menguji guardrail kami sendiri, dan ia meloloskan 12 dari 17 percobaan.

### Post LinkedIn (EN)
```
We tested NAEOS' own governance the way an attacker would.

17 scenarios across 5 enforcement layers — policy evaluation, the control
plane, artifact review, prompt generation, and pipeline gates.

Result: 12 bypass paths found, 11 still open. Not as a surprise, but as
tracked public work: issues #145-#152.

One CRITICAL case — a prompt override silently stripping policy guidance
from generated AGENTS.md — is already closed (PR #154).

Why publish this? A governance platform earns trust by showing its
weaknesses and the fixes, not by hiding them.

Review the audit and the issues:
github.com/NAEOS-foundation/naeos
```

### Thread X (1–4)
```
1/4 Kami menjalankan adversarial harness terhadap governance NAEOS.
2/4 17 skenario, 5 layer. Ditemukan 12 jalur bypass; 11 masih terbuka sebagai public issues #145-#152.
3/4 Yang CRITICAL sudah ditutup (PR #154): override prompt tak bisa lagi menghapus policy dari AGENTS.md.
4/4 Kalau kamu tim engineering yang pakai AI: mana dari 11 temuan yang paling penting menurutmu? github.com/NAEOS-foundation/naeos
```

### Post Reddit (r/ExperiencedDevs) — jika dipakai
Format founder-journey, bukan advertorial. Judul saran:
"I built an adversarial test harness for my own AI-governance tool — then
published the 11 bypasses it found."

### Asset
- `experiments/policy-bypass/reports/EXPERIMENT-REPORT.md`
- Daftar issue #145–#152
- PR #154

### CTA
Baca laporan + komentari prioritas di GitHub Discussions.

### Expected signal & follow-up
Signal: share/comment pada LinkedIn, technical comments pada thread.
Follow-up: catat saran prioritas → tetapkan urutan kerja #145–#148 di minggu 2.

---

## Catatan Operasional

- Setiap campaign menuliskan Hypothesis, Audience, Channel, Message, Asset,
  CTA, Metric, Success threshold (Mode 28); kolom di atas sudah memuatnya.
- Kanban: update `brand/marketing/BACKLOG.md` saat item berpindah status.
- Pendamping: `brand/marketing/content-calendar.json` (slot tanggal nyata).