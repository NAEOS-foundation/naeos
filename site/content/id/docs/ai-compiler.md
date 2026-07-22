---
title: Kompiler AI
description: Ubah NEIR menjadi set instruksi AI untuk 6 asisten coding.
---

## Ikhtisar

Kompiler AI NAEOS mengubah model NEIR menjadi file instruksi khusus platform untuk asisten coding AI. Ini memastikan alat AI Anda memahami arsitektur proyek, konvensi, dan dependensi — mengurangi rekayasa prompt manual dan meningkatkan kualitas kode.

## Platform yang Didukung

| Platform | File | Status |
|----------|------|--------|
| GitHub Copilot | `.github/copilot-instructions.md` | ✅ |
| Claude Code | `CLAUDE.md` | ✅ |
| Cursor | `.cursorrules` | ✅ |
| Gemini CLI | `.gemini/CONFIG.md` | ✅ |
| Codex | `AGENTS.md` | ✅ |
| OpenCode | `AGENTS.md` | ✅ |

## Cara Kerja

Pipeline kompilasi:
1. **Ekstrak** — Ambil bagian relevan dari NEIR (arsitektur, modul, dependensi, konvensi)
2. **Terjemahkan** — Ubah ke format dan sintaks khusus platform
3. **Optimalkan** — Prioritaskan informasi berdasarkan relevansi dan anggaran token
4. **Format** — Terapkan aturan pemformatan khusus platform
5. **Keluarkan** — Tulis ke lokasi file target

## Yang Disertakan

Setiap file instruksi biasanya mencakup: ikhtisar proyek, struktur modul, definisi layanan, tumpukan bahasa, konvensi kode, dependensi utama, aturan arsitektur, dan kebijakan keamanan.

## Penggunaan

```bash
# Kompilasi untuk semua platform
naeos compile --all --input-file spec.yaml

# Kompilasi untuk platform tertentu
naeos compile --platform copilot --input-file spec.yaml

# Hasilkan bundel konteks AI
naeos context --input-file spec.yaml --format bundle
```

## Praktik Terbaik

- Jalankan `naeos compile` sebagai bagian dari skrip setup proyek Anda
- Commit file instruksi yang dihasilkan ke repositori Anda
- Kompilasi ulang setiap kali arsitektur Anda berubah
- Gunakan `naeos context` untuk pendekatan bundel
- Kombinasikan dengan `naeos watch` untuk kompilasi ulang otomatis
