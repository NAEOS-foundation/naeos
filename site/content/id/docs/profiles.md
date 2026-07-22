---
title: Profil Industri
description: Profil yang sudah dikonfigurasi sebelumnya untuk proyek SaaS, AI Agent, FinTech, Healthcare, dan Government.
weight: 14
---

NAeos menyediakan profil spesifik industri yang mengonfigurasi modul, layanan, pola arsitektur, dan pengaturan keamanan untuk jenis proyek umum.

## Profil yang Tersedia

| Profil | Industri | Deskripsi |
|--------|----------|-----------|
| **SaaS** | Software-as-a-Service | Arsitektur multi-tenant, billing, analytics, manajemen pengguna |
| **AI Agent** | Kecerdasan Buatan | Integrasi LLM, tool calling, sistem memori, orkestrasi agen |
| **FinTech** | Teknologi Keuangan | Compliance, audit trail, enkripsi, pemrosesan transaksi |
| **Healthcare** | Teknologi Kesehatan | Compliance HIPAA, audit logging, perlindungan PHI, alur kerja klinis |
| **Government** | Sektor Publik | Security clearance, audit, aksesibilitas, integrasi antar lembaga |

## Penggunaan CLI

```bash
# Lihat semua profil yang tersedia
naeos profile list

# Lihat detail profil
naeos profile show saas

# Cari profil berdasarkan kata kunci
naeos profile search "multi-tenant"

# Terapkan profil ke spesifikasi
naeos profile apply saas --output spec.yaml
```

## Penggunaan Go API

```go
import "github.com/NAEOS-foundation/naeos/internal/profiles"

registry := profiles.NewRegistry()

// Daftar semua profil
allProfiles := registry.List()
for _, p := range allProfiles {
    fmt.Printf("%s: %s\n", p.Name, p.Description)
}

// Cari profil
results := registry.Search("fintech")

// Dapatkan profil tertentu
saas := registry.Get("saas")

// Terapkan profil untuk menghasilkan spec YAML
specYAML := saas.ToSpecYAML()
```

## Struktur Profil

Setiap profil mendefinisikan template proyek lengkap:

```go
type Profile struct {
    Name         string
    Description  string
    Industry     string
    Modules      []Module
    Services     []Service
    Architecture *Architecture
    Security     *Security
    Deployment   *Deployment
    Testing      *Testing
}
```

### Contoh: Profil SaaS

Profil SaaS mencakup:

- **Modul**: auth, billing, analytics, notification, user-management
- **Layanan**: api-gateway (HTTP), worker (async), admin (HTTP)
- **Arsitektur**: hexagonal dengan komunikasi event-driven
- **Keamanan**: OAuth2, rate limiting, validasi input
- **Deployment**: Kubernetes dengan autoscaling horizontal
- **Testing**: unit, integration, e2e

## Konversi ke Spec YAML

```go
profile := registry.Get("saas")
specYAML := profile.ToSpecYAML()
// Mengembalikan string YAML yang siap digunakan sebagai spesifikasi
```

## Profil Kustom

Buat dan daftarkan profil Anda sendiri:

```go
registry := profiles.NewRegistry()

customProfile := &profiles.Profile{
    Name:        "manufacturing",
    Description: "Profil untuk sistem manufaktur dan IoT",
    Industry:    "manufacturing",
    Modules: []profiles.Module{
        {Name: "inventory", Path: "./inventory"},
        {Name: "supply-chain", Path: "./supply-chain"},
        {Name: "iot-gateway", Path: "./iot-gateway"},
    },
    Services: []profiles.Service{
        {Name: "api", Kind: "http", Port: 8080},
        {Name: "mqtt-broker", Kind: "worker", Port: 1883},
    },
}

registry.Register(customProfile)
```

## Marketplace

Jelajahi profil komunitas di [Marketplace](/id/docs/plugin-sdk/):

```bash
naeos marketplace search profile
naeos marketplace install community-profile-name
```

Lihat juga: [Ekosistem](/id/ecosystem/), [Plugin SDK](/id/docs/plugin-sdk/)
