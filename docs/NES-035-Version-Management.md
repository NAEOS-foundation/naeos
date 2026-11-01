# NES-035 Version Management

## 1. Status
- Status: Draft
- Version: 0.1
- Owner: NAEOS Core Team

## 2. Purpose
Dokumentasi referensi untuk package `internal/neir/version` — manajemen versi SemVer untuk NEIR dan komponen NAEOS.

## 3. Scope
Dokumen ini mencakup parsing SemVer, perbandingan versi, kompatibilitas versi, dan validasi informasi versi.

## 4. Normative References
- NES-023 NEIR — model NEIR dan pipeline
- NES-023-NEIR-Model — referensi model Go

## 5. Types

### 5.1 VersionInfo

```go
type VersionInfo struct {
    NEIRVersion    string
    SchemaVersion  string
    ProjectVersion string
}
```

Field | Tipe | Deskripsi | Contoh
------|------|-----------|-------
NEIRVersion | string | Versi model NEIR menurut NES-023 (SemVer) | "2.0.0"
SchemaVersion | string | Versi skema input spesifikasi yang diterima parser (SemVer) | "0.5.0"
ProjectVersion | string | Versi proyek yang dihasilkan | "0.1.0"

### 5.2 SemVer

```go
type SemVer struct {
    Major int
    Minor int
    Patch int
}
```

Representasi terurai dari versi Semantic Versioning.

## 6. Functions

### 6.1 Default

```go
func Default() VersionInfo
```

Mengembalikan VersionInfo dengan nilai default:
- NEIRVersion: "2.0.0" (versi stabil NES-023)
- SchemaVersion: "0.5.0" (parser.CurrentSpecVersion)
- ProjectVersion: "0.1.0"

Nilai default harus selalu lolos `VersionInfo.Validate()`.

### 6.2 ParseSemVer

```go
func ParseSemVer(s string) (SemVer, error)
```

Mem parsing string versi SemVer menjadi struct SemVer.

Aturan:
- Format: `MAJOR.MINOR.PATCH` atau `vMAJOR.MINOR.PATCH`
- Komponen harus numerik dan non-negatif.
- Jika format tidak valid, mengembalikan error.

```go
v, err := version.ParseSemVer("1.2.3")
// v.Major=1, v.Minor=2, v.Patch=3

v, err := version.ParseSemVer("v2.0.0")
// v.Major=2, v.Minor=0, v.Patch=0

v, err := version.ParseSemVer("invalid")
// error: invalid semver format
```

### 6.3 Compare

```go
func Compare(a, b SemVer) int
```

Membandingkan dua versi SemVer.

Pengembalian:
- `-1` jika a < b
- `0` jika a == b
- `1` urutan perbandingan: Major, lalu Minor, lalu Patch.

### 6.4 IsCompatible

```go
func IsCompatible(required, actual SemVer) bool
```

Memeriksa apakah versi `actual` kompatibel dengan `required`.

Aturan:
- Jika required.Major == 0 (pre-1.0): Major dan Minor harus sama persis.
- Jika required.Major >= 1: Major harus sama, actual.Minor >= required.Minor.

```go
// Pre-1.0: harus sama persis
version.IsCompatible(SemVer{0,1,0}, SemVer{0,1,0}) // true
version.IsCompatible(SemVer{0,1,0}, SemVer{0,2,0}) // false

// Post-1.0: backward compatible
version.IsCompatible(SemVer{1,0,0}, SemVer{1,1,0}) // true
version.IsCompatible(SemVer{1,1,0}, SemVer{1,0,0}) // false
```

## 7. Methods

### 7.1 SemVer.String

```go
func (v SemVer) String() string
```

Mengembalikan representasi string dari SemVer: "MAJOR.MINOR.PATCH".

### 7.2 VersionInfo.Validate

```go
func (vi VersionInfo) Validate() error
```

Memvalidasi VersionInfo:
- NEIRVersion tidak boleh kosong dan harus valid SemVer.
- SchemaVersion tidak boleh kosong dan harus valid SemVer.

### 7.3 VersionInfo.IsCompatibleWith

```go
func (vi VersionInfo) IsCompatibleWith(other VersionInfo) bool
```

Memeriksa kompatibilitas antara dua VersionInfo berdasarkan NEIRVersion.

## 8. Constraints

- Semua komponen versi harus non-negatif.
- Format string harus sesuai SemVer (MAJOR.MINOR.PATCH).
- Tidak ada support untuk prerelease atau build metadata.
