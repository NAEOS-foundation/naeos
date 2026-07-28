---
title: Plugin SDK
description: Perluas NAEOS dengan plugin kustom, generator, dan validator.
---

## Ikhtisar

NAEOS menyediakan Plugin SDK untuk memperluas platform dengan fungsionalitas kustom. Plugin terintegrasi langsung ke dalam pipeline 9-tahap dan dapat menambahkan generator kode baru, validator, deployer, analis, dan hook siklus hidup.

## Interface Plugin

Setiap plugin harus mengimplementasikan interface `Plugin`:

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Initialize(ctx *PluginContext) error
    Execute(action string, params map[string]any) (any, error)
    Shutdown() error
}
```

### BasePlugin

SDK menyediakan `BasePlugin` — struct yang dapat di-embed dengan implementasi default:

```go
type MyPlugin struct {
    pluginhost.BasePlugin
}
```

## Tipe Plugin

| Tipe | Action | Deskripsi |
|------|--------|-----------|
| **Generator** | `generate` | Hasilkan kode dalam bahasa kustom |
| **Validator** | `validate` | Aturan validasi kustom |
| **Deployer** | `deploy` | Deploy ke platform kustom |
| **Analyzer** | `analyze` | Analisis dan pelaporan kustom |
| **Hook** | — | Hook siklus hidup pipeline |

## Memulai

### Prasyarat

- Go 1.25+ (untuk plugin native)
- TinyGo 0.35+ (untuk plugin WASM)

### Plugin Native (Go)

```go
type MyValidator struct {
    pluginhost.BasePlugin
}

func (v *MyValidator) Execute(action string, params map[string]any) (any, error) {
    // logika validasi
    return issues, nil
}
```

Build sebagai shared library:

```bash
go build -buildmode=plugin -o my-validator.so .
```

### Plugin WASM (TinyGo)

```go
//export generate
func generate(modelPtr, modelLen uint32) uint64 {
    model := sdk.ReadNEIR(modelPtr, modelLen)
    return sdk.WriteResult(processModel(model))
}

func main() {}
```

Build:

```bash
tinygo build -o plugin.wasm -target=wasi -scheduler=none .
```

## Manifes Plugin

Setiap plugin yang dipublikasi harus menyertakan `plugin.yaml`:

```yaml
name: my-generator
version: "1.0.0"
description: Generate Rust service scaffolding
author: NAEOS Foundation
license: Apache-2.0
actions:
  - name: generate
    description: Generate Rust service code
    params:
      output_dir: string
    returns: array
config:
  template_dir:
    type: string
    required: false
```

## Konteks Plugin

Saat `Initialize` dipanggil, plugin menerima `PluginContext` dengan:

| Field | Tipe | Deskripsi |
|-------|------|-----------|
| `ConfigDir` | `string` | Direktori konfigurasi plugin |
| `OutputDir` | `string` | Direktori output untuk artefak |
| `Verbose` | `bool` | Apakah logging verbose diaktifkan |
| `Config` | `map[string]any` | Konfigurasi spesifik plugin |
| `Logger` | `Logger` | Logger terstruktur |
| `Metrics` | `MetricsCollector` | Kolektor metrik |
| `EventBus` | `EventEmitter` | Emit dan subscribe event pipeline |

## Event Siklus Hidup

Plugin dapat subscribe ke event pipeline:

| Event | Waktu |
|-------|-------|
| `before_parse` | Sebelum parsing spesifikasi |
| `after_parse` | Setelah parsing spesifikasi |
| `before_generate` | Sebelum generasi kode |
| `after_generate` | Setelah generasi kode |
| `on_pipeline_complete` | Pipeline selesai |
| `on_pipeline_failed` | Pipeline gagal |

## Konfigurasi Plugin

Plugin menerima konfigurasi dari file spesifikasi:

```yaml
plugins:
  - name: my-generator
    config:
      template_dir: ./templates
      features: [typescript]
```

Akses di plugin:

```go
tmplDir, _ := ctx.Config["template_dir"].(string)
```

## Memasang Plugin

```bash
# Dari marketplace
naeos plugin install my-generator

# Dari file lokal
naeos plugin install ./path/to/plugin.wasm

# Dari registry kontainer
naeos plugin install ghcr.io/naeos-foundation/plugins/my-generator:latest
```

### Mengelola Plugin

```bash
naeos plugin list          # daftar plugin
naeos plugin update        # perbarui plugin
naeos plugin remove        # hapus plugin
naeos plugin info          # inspeksi metadata
naeos plugin enable/disable
```

## Menguji Plugin

```bash
naeos test --plugin my-generator
naeos test --plugin my-generator --input-file test-spec.yaml
```

## Publikasi ke Marketplace

```bash
naeos plugin package ./my-generator --output my-generator.tar.gz
naeos marketplace publish my-generator.tar.gz
```

Marketplace memverifikasi checksum SHA-256, kesesuaian skema manifes, dan keunikan versi.

## Hot-Reload

Plugin host mendukung hot-reloading saat pengembangan:

```bash
naeos run --plugin-dir ./plugins --watch
```

File `*.so` atau `*.wasm` yang berubah akan dimuat ulang secara otomatis.

## Referensi SDK

| Fungsi | Deskripsi |
|--------|-----------|
| `pluginhost.NewManager(config)` | Buat manager plugin baru |
| `manager.Load(path)` | Muat plugin dari file |
| `manager.Register(plugin)` | Daftarkan instance plugin |
| `manager.List()` | Daftar semua plugin |
| `sdk.ReadNEIR(ptr, len)` | Deserialisasi NEIR dari memori WASM |
| `sdk.WriteResult(data)` | Serialisasi hasil ke memori WASM |

## Praktik Terbaik

- **Mulai kecil** — Embed `BasePlugin`, implementasi `Execute` dulu
- **Gunakan semantic versioning** — Rilis sebagai `v1.0.0`
- **Sertakan manifes** — Selalu kirim `plugin.yaml`
- **Log dengan bermakna** — Gunakan `ctx.Logger` dengan key-value terstruktur
- **Tangani error** — Kembalikan objek `Issue` deskriptif
- **Uji dengan `naeos test`** — Validasi plugin terhadap spesifikasi nyata
- **Jaga WASM tetap ramping** — Minimalisir impor untuk loading cepat
- **Gunakan hot-reload** — Saat development, gunakan `--plugin-dir ./plugins --watch`

## Pemecahan Masalah

| Masalah | Solusi |
|---------|--------|
| Plugin tidak dimuat | Cek `naeos plugin list`. Verifikasi format biner (`.so` / `.wasm`). |
| Plugin WASM crash | Build ulang dengan `-scheduler=none`. |
| Konfigurasi tidak terbaca | Verifikasi struktur YAML `plugins[].config`. |
| Hot-reload tidak bekerja | Pastikan file di direktori yang diawasi. Ekstensi `*.so` / `*.wasm`. |

## Lihat Juga

- [Plugin Marketplace](/id/plugins/) — Jelajahi plugin komunitas
- [Template Marketplace](/id/templates/) — Template untuk plugin baru
- [Pipeline Engine](/id/docs/pipeline-engine/) — Integrasi plugin dengan pipeline
