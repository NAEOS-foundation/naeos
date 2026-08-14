---
title: Plugin SDK
description: Perluas NAEOS dengan plugin kustom, generator, dan validator.
---

## Ikhtisar

NAEOS menyediakan Plugin SDK untuk memperluas platform dengan fungsionalitas kustom. Plugin terintegrasi langsung ke dalam pipeline 11-tahap dan dapat menambahkan generator kode baru, validator, deployer, analis, dan hook siklus hidup.

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

Plugin native adalah paket `main` Go yang dibangun dengan `-buildmode=plugin`.
Plugin wajib mengekspor variabel metadata `PluginName` dan nilai `NaeosPlugin`
yang mengimplementasikan `pluginhost.Plugin`:

```go
package main

import (
    "fmt"

    "github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

type MyValidator struct {
    pluginhost.BasePlugin
}

func New() *MyValidator {
    return &MyValidator{
        BasePlugin: pluginhost.BasePlugin{
            NameVal:        "my-validator",
            VersionVal:     "1.0.0",
            DescriptionVal: "Custom validation rules",
        },
    }
}

var (
    pluginName        = "my-validator"
    pluginVersion     = "1.0.0"
    pluginDescription = "Custom validation rules"
    pluginAuthor      = "you"

    PluginName        = &pluginName
    PluginVersion     = &pluginVersion
    PluginDescription = &pluginDescription
    PluginAuthor      = &pluginAuthor
)

var NaeosPlugin pluginhost.Plugin = New()

func (v *MyValidator) Initialize(ctx *pluginhost.PluginContext) error {
    ctx.Logger.Info("MyValidator initialized")
    return nil
}

func (v *MyValidator) Execute(action string, params map[string]any) (any, error) {
    if action != "validate" {
        return nil, fmt.Errorf("unsupported action: %s", action)
    }
    return []map[string]string{
        {"severity": "warning", "message": "custom check passed"},
    }, nil
}

func (v *MyValidator) Shutdown() error {
    return nil
}
```

Build sebagai shared library:

```bash
go build -buildmode=plugin -o my-validator.so .
```

Karena plugin mengimpor paket internal NAEOS, aturan internal Go mengharuskan
modul plugin berada di bawah prefiks `github.com/NAEOS-foundation/naeos` dan
me-resolve `naeos` ke checkout sumber yang sesuai dengan CLI Anda:

```
replace github.com/NAEOS-foundation/naeos => /path/to/naeos
```

`naeos plugin init` sudah menyiapkan semua ini untuk Anda.

### Plugin WASM (TinyGo)

Plugin WASM berjalan sebagai modul WASI. Saat eksekusi, host mengirim request
(`{"method": "<action>", "params": {...}}`) melalui stdin dan mengharapkan
respons JSON `{"ok": true, "result": ...}` atau `{"ok": false, "error": "..."}`
pada stdout. Bahasa apa pun yang mengompilasi ke WASI dapat menerapkan protokol
ini. Dengan TinyGo:

```go
package main

import (
    "encoding/json"
    "os"
)

func main() {
    var req struct {
        Method string `json:"method"`
        Params map[string]any `json:"params"`
    }
    if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
        json.NewEncoder(os.Stdout).Encode(map[string]any{
            "ok": false, "error": "invalid request: " + err.Error(),
        })
        os.Exit(1)
    }

    switch req.Method {
    case "ping":
        json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true, "result": "pong"})
    default:
        json.NewEncoder(os.Stdout).Encode(map[string]any{
            "ok": false, "error": "unknown action: " + req.Method,
        })
        os.Exit(1)
    }
}
```

Build:

```bash
tinygo build -o plugin.wasm -target=wasi -scheduler=none .
```

## Manifes Plugin

Setiap plugin menyertakan manifes `naeos.yaml` (dibuat otomatis oleh
`naeos plugin init`):

```yaml
name: my-generator
version: "0.1.0"
description: Generate Rust service scaffolding
author: NAEOS Foundation
type: wasm              # "wasm" atau "native"
tags: []
```

Manifes ini wajib untuk `naeos marketplace publish`. Aksi dideklarasikan di
kode melalui metode `Execute`, bukan di manifes.

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

Muat, inisialisasi, dan periksa kesehatan plugin dengan test runner plugin:

```bash
naeos plugin test my-generator
naeos plugin test my-generator --plugin-dir ./my-plugins
```

`naeos test` juga menjalankan tes untuk kode yang di-generate per bahasa (Go, TypeScript, Python, Java, Rust).

## Publikasi ke Marketplace

```bash
# Publikasikan paket plugin Anda (direktori berisi naeos.yaml)
naeos marketplace publish ./my-generator
```

Perintah publish memvalidasi bahwa paket berisi manifes `naeos.yaml` dengan
kolom `name`, `version`, dan `type` sebelum dipublikasikan.

## Hot-Reload

Plugin host menyediakan watcher tingkat pustaka untuk pengembangan. `PluginWatcher` menggunakan `fsnotify` untuk mengawasi direktori plugin dan memuat ulang plugin `*.so` atau `*.wasm` yang berubah secara otomatis (debounce 500 ms):

```go
pw := pluginhost.NewPluginWatcher("./plugins", manager)
if err := pw.Start(ctx); err != nil {
    // handle error
}
```

Pada tingkat CLI, gunakan `naeos plugin test --plugin-dir ./plugins` untuk memeriksa ulang plugin setelah rebuild.

## Referensi SDK

| Fungsi | Deskripsi |
|--------|-----------|
| `pluginhost.NewManager(pluginDir)` | Buat manager plugin baru |
| `manager.Install(path)` | Pasang plugin `.so` (membaca metadata yang diekspor) |
| `manager.LoadAll(ctx)` | Muat dan inisialisasi semua plugin terpasang |
| `manager.Register(plugin)` | Daftarkan instance plugin in-process |
| `manager.Execute(ctx, name, action, params)` | Eksekusi aksi satu plugin |
| `manager.List()` | Daftar semua plugin terpasang |
| `manager.GetInfo(name)` | Ambil metadata satu plugin |
| `manager.Cleanup()` | Shutdown semua plugin dan lepaskan resource |
| Protokol WASM (stdin → stdout JSON) | Request: `{"method": "<action>", "params": {...}}`; sukses: `{"ok": true, "result": ...}`; gagal: `{"ok": false, "error": "..."}` |

## Praktik Terbaik

- **Mulai kecil** — Embed `BasePlugin`, implementasi `Execute` dulu
- **Gunakan semantic versioning** — Rilis sebagai `v1.0.0`
- **Sertakan manifes** — Selalu kirim `naeos.yaml`
- **Log dengan bermakna** — Gunakan `ctx.Logger` dengan key-value terstruktur
- **Tangani error** — Kembalikan struktur hasil yang deskriptif alih-alih panic
- **Uji dengan `naeos plugin test`** — Validasi plugin terhadap spesifikasi nyata
- **Jaga WASM tetap ramping** — Minimalisir impor untuk loading cepat
- **Gunakan hot-reload** — Saat development, periksa ulang plugin setelah rebuild dengan `naeos plugin test --plugin-dir ./plugins`

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
