---
title: Berkontribusi
description: Panduan untuk berkontribusi ke proyek NAEOS.
weight: 20
---

Panduan ini mencakup semua yang perlu Anda ketahui untuk berkontribusi ke NAEOS.

## Persiapan Development

### Prasyarat

- Go 1.25 atau lebih baru
- Git
- golangci-lint (disarankan)

### Memulai

```bash
# Clone repositori
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos

# Build proyek
go build ./cmd/naeos/

# Jalankan test
go test -race -count=1 -timeout 300s ./...

# Jalankan linter
golangci-lint run ./...
```

## Struktur Proyek

```text
naeos/
├── cmd/naeos/           # CLI commands (35+ file)
│   ├── main.go          # Root command
│   ├── run_cmd.go       # Jalankan pipeline
│   ├── compile_cmd.go   # Kompilasi ke AI tools
│   ├── context_cmd.go   # Hasilkan context bundles
│   └── ...              # 30+ file command lainnya
├── internal/
│   ├── specification/   # Parser, normalizer, resolver
│   ├── neir/            # Model NEIR dan builder
│   ├── compiler/        # Kompiler instruksi AI
│   ├── context/         # Generator context bundle
│   ├── generation/      # Generasi kode
│   ├── governance/      # Kebijakan dan review
│   ├── artifacts/       # Artifact store
│   ├── profiles/        # Profil industri
│   ├── migration/       # Migrasi skema
│   ├── security/        # Aturan keamanan
│   └── shared/          # Utilitas bersama
├── pkg/
│   ├── pipeline/        # Pipeline utama
│   ├── kernel/          # Kernel sistem
│   ├── config/          # Konfigurasi
│   └── plugin/          # Sistem plugin
├── docs/                # Spesifikasi NES (56 file)
├── governance/          # Dokumen governance
├── constitution/        # Konstitusi engineering
├── specification/       # Dokumen spesifikasi
└── site/                # Website Hugo
```

## Standar Coding

### Gaya Go

- Ikuti konvensi `gofmt` dan `go vet`
- Gunakan nama variabel dan fungsi yang bermakna
- Tambahkan komentar untuk fungsi dan tipe yang di-export
- Pertahankan fungsi kecil dan fokus (idealnya <50 baris)
- Kembalikan error; jangan pernah panic di library code
- Gunakan `sync.Mutex`/`sync.RWMutex` untuk thread safety
- Gunakan pola standard library daripada package pihak ketiga

### Testing

- Tulis test berbasis tabel dengan struct `tt`
- Gunakan `t.Parallel()` di mana aman
- Pertahankan test coverage di atas 80%
- Mock dependensi eksternal
- Test kasus batas dan path error

```go
func TestValidator(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    bool
        wantErr bool
    }{
        {"valid spec", validSpec, true, false},
        {"missing project", missingProject, false, true},
        {"empty modules", emptyModules, true, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got, err := Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("Validate() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Dokumentasi

- Perbarui dokumentasi untuk fitur baru
- Tambahkan contoh untuk CLI commands
- Pertahankan CHANGELOG tetap terupdate
- Tulis commit message yang menjelaskan **mengapa**, bukan hanya **apa**

## Proses Pull Request

1. **Fork** repositori
2. **Buat** branch fitur (`git checkout -b feature/my-feature`)
3. **Tulis** test untuk perubahan Anda
4. **Jalankan** test suite lengkap:
   ```bash
   go test -race -count=1 -timeout 300s ./...
   golangci-lint run ./...
   ```
5. **Perbarui** dokumentasi jika diperlukan
6. **Commit** dengan pesan yang deskriptif
7. **Push** dan buat Pull Request

## Commit Messages

Gunakan format [Conventional Commits](https://www.conventionalcommits.org/):

```text
feat: tambahkan parser baru untuk konfigurasi HCL
fix: selesaikan deteksi circular dependency di validator
docs: perbarui referensi CLI dengan flag baru
test: tambahkan integration test untuk pipeline
refactor: sederhanakan API NEIR builder
chore: perbarui dependensi
```

## Code Review

- Semua PR membutuhkan minimal satu review
- Test harus lulus di CI
- Dokumentasi harus diperbarui untuk fitur baru
- Ikuti konvensi coding yang ada
- Pertahankan PR fokus — satu fitur atau fix per PR

## Melaporkan Isu

- Gunakan GitHub Issues untuk laporan bug dan permintaan fitur
- Sertakan langkah replikasi untuk bug
- Tentukan versi Go dan OS untuk isu spesifik lingkungan
- Periksa isu yang ada sebelum membuat yang baru

## Lisensi

Dengan berkontribusi ke NAEOS, Anda setuju bahwa kontribusi Anda akan dilisensikan di bawah Lisensi Apache 2.0.
