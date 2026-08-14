---
title: Kompiler AI
slug: ai-compiler
description: Ubah NEIR menjadi set instruksi AI untuk 7 asisten coding.
---

## Ikhtisar

Kompiler AI NAEOS mengubah spesifikasi NAEOS menjadi file instruksi khusus platform untuk asisten coding AI. Ini memastikan alat AI Anda memahami arsitektur proyek, konvensi, dan dependensi — mengurangi rekayasa prompt manual dan meningkatkan kualitas kode.

## Platform yang Didukung

| Platform | Target ID | File |
|----------|-----------|------|
| GitHub Copilot | `copilot` | `.github/copilot-instructions.md` |
| Claude Code | `claude` | `CLAUDE.md` |
| Cursor | `cursor` | `.cursorrules` |
| Gemini CLI | `gemini` | `.gemini/CONFIG.md` |
| Codex | `codex` | `AGENTS.md` |
| OpenCode | `opencode` | `AGENTS.md` |
| Windsurf | `windsurf` | `.windsurfrules` |

## Cara Kerja

Kompiler mengirim spesifikasi Anda ke LLM (OpenAI, Anthropic, atau Ollama) beserta prompt khusus target, lalu mengalirkan kembali file instruksi yang dihasilkan:

```text
┌──────────┐     ┌────────────┐     ┌──────────────────┐
│ Spesifikasi│────→│   Kompiler │────→│  CLAUDE.md       │
│ (YAML/   │     │     AI     │     │  .cursorrules    │
│  JSON)   │     │ (berbasis-  │     │  AGENTS.md       │
│          │     │   LLM)     │     └──────────────────┘
└──────────┘     └────────────┘
```

## Penggunaan

```bash
# Kompilasi untuk agen AI tertentu
naeos ai compile --input-file spec.yaml --target claude

# Gunakan provider LLM tertentu (default: openai)
naeos ai compile --input-file spec.yaml --target opencode --provider anthropic
```

Atur `NAEOS_LLM_API_KEY` (dan opsional `NAEOS_LLM_PROVIDER`) sebelum menjalankan:

```bash
export NAEOS_LLM_API_KEY="kunci-api-anda"
export NAEOS_LLM_PROVIDER="anthropic"
```

File instruksi dicetak ke stdout — arahkan ke file target Anda:

```bash
naeos ai compile --input-file spec.yaml --target claude > CLAUDE.md
```

Untuk generasi konteks offline yang deterministik tanpa LLM, gunakan:

```bash
# Hasilkan bundel konteks AI di direktori saat ini
naeos context --input-file spec.yaml
```

## Contoh Output

Saat Anda mengompilasi proyek microservices, `CLAUDE.md` yang dihasilkan mungkin berisi:

```markdown
# Project: ecommerce-platform
## Architecture: Microservices
## Language Stack: Go, TypeScript, PostgreSQL

### Modules
- api-gateway → depends on: user-service, product-service
- user-service → depends on: database
- product-service → depends on: database, search-engine

### Conventions
- Go: standard library + chi router, sqlx for DB
- TypeScript: Express with Zod validation
- All services expose Prometheus metrics on :9090
```

## Praktik Terbaik

- Jalankan `naeos ai compile` sebagai bagian dari skrip setup proyek Anda
- Commit file instruksi yang dihasilkan ke repositori Anda
- Kompilasi ulang setiap kali arsitektur Anda berubah
- Gunakan `naeos context` untuk pendekatan bundel offline
- Arahkan output ke file khusus platform (mis. `CLAUDE.md`, `AGENTS.md`)