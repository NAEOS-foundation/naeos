# NES-038 Shared Types & Contracts

## 1. Status
- Status: Draft
- Version: 0.1
- Owner: NAEOS Core Team

## 2. Purpose
Dokumentasi referensi untuk package `internal/shared` — tipe data dan kontrak bersama yang digunakan oleh seluruh komponen internal NAEOS.

## 3. Scope
Dokumen ini mencakup shared types (`internal/shared/types`) dan shared contracts (`internal/shared/contracts`).

> **Catatan status implementasi:** package `internal/shared/types` belum ada di kode saat ini — shared types di dokumen ini adalah spesifikasi desain (rencana). Package `internal/shared/contracts` sudah diimplementasikan di `internal/shared/contracts/contracts.go`.

## 4. Normative References
- NES-000 Foundation — prinsip fondasi
- NES-023 NEIR — model engineering sentral

## 5. Shared Types

```mermaid
classDiagram
    class ID {
        <<type>>
        +string
    }

    class Reference {
        +ID id
        +string name
    }

    class ErrorInfo {
        +string Code
        +string Message
    }

    class Artifact {
        +string Path
        +bytes Content
    }

    class Task {
        +string ID
        +string Name
        +string[] Dependencies
        +int Priority
    }

    class PolicyRule {
        +string RuleID
        +string Condition
        +int Priority
        +string Action
        +string Scope
    }

    class KnowledgeEntry {
        +string Topic
        +string Component
        +string Version
        +string Rationale
    }

    class TelemetryEvent {
        +string Name
        +int64 Timestamp
        +map Payload
    }

    class ValidationResult {
        +bool Valid
        +ErrorInfo[] Errors
    }

    class ReviewResult {
        +bool Approved
        +string[] Comments
    }

    class SpecDocument {
        +string Raw
        +string Project
        +ModuleDef[] Modules
        +ServiceDef[] Services
    }

    class ModuleDef {
        +string Name
        +string Path
    }

    class ServiceDef {
        +string Name
        +string Kind
        +int Port
    }

    class Contract {
        <<interface>>
        +Validate() error
    }

    class SchemaAware {
        <<interface>>
        +SchemaVersion() string
    }

    class Versioned {
        <<interface>>
        +Version() string
    }

    class Identifiable {
        <<interface>>
        +ID() string
    }

    class Named {
        <<interface>>
        +Name() string
    }

    Reference --> ID
    ValidationResult --> ErrorInfo
    SpecDocument --> ModuleDef
    SpecDocument --> ServiceDef

    Kernel ..> TelemetryEvent : emits
    Pipeline ..> Artifact : produces
    Pipeline ..> Task : schedules
    Policy ..> PolicyRule : evaluates
    Validator ..> ValidationResult : returns
    Reviewer ..> ReviewResult : returns
    Parser ..> SpecDocument : creates
```

Lokasi (direncanakan): `internal/shared/types/types.go` — belum diimplementasikan; lihat catatan Status Implementasi di Scope.

### 5.1 ID

```go
type ID string
```

Tipe string yang merepresentasikan identifikasi unik dalam sistem.

### 5.2 Reference

```go
type Reference struct {
    ID   ID
    Name string
}
```

Referensi ke entitas lain dalam sistem. Digunakan untuk cross-referencing antar komponen.

### 5.3 ErrorInfo

```go
type ErrorInfo struct {
    Code    string
    Message string
}
```

Informasi error terstruktur. Code bersifat machine-readable, Message bersifat human-readable.

### 5.4 Artifact

```go
type Artifact struct {
    Path    string
    Content []byte
}
```

Representasi artefak yang dihasilkan oleh pipeline. Path adalah lokasi output, Content adalah isi file.

### 5.5 Task

```go
type Task struct {
    ID           string
    Name         string
    Dependencies []string
    Priority     int
}
```

Tugas dalam execution plan. Dependencies berisi ID task yang harus selesai terlebih dahulu.

### 5.6 PolicyRule

```go
type PolicyRule struct {
    RuleID    string
    Condition string
    Priority  int
    Action    string
    Scope     string
}
```

Aturan policy yang dievaluasi oleh governance engine.

Field | Deskripsi
------|----------
RuleID | ID unik aturan
Condition | Ekspresi kondisi
Priority | Prioritas evaluasi (lebih rendah = lebih tinggi)
Action | Aksi jika kondisi terpenuhi
Scope | Cakupan aturan (project, module, service)

### 5.7 KnowledgeEntry

```go
type KnowledgeEntry struct {
    Topic     string
    Component string
    Version   string
    Rationale string
}
```

Entri pengetahuan yang terstruktur.

### 5.8 TelemetryEvent

```go
type TelemetryEvent struct {
    Name      string
    Timestamp int64
    Payload   map[string]any
}
```

Event telemetry untuk observabilitas.

### 5.9 ValidationResult

```go
type ValidationResult struct {
    Valid  bool
    Errors []ErrorInfo
}
```

Hasil validasi. `Valid` true jika tidak ada error.

### 5.10 ReviewResult

```go
type ReviewResult struct {
    Approved bool
    Comments []string
}
```

Hasil review artefak.

### 5.11 SpecDocument

```go
type SpecDocument struct {
    Raw      string
    Project  string
    Modules  []ModuleDef
    Services []ServiceDef
}
```

Dokumen spesifikasi hasil parsing.

### 5.12 ModuleDef

```go
type ModuleDef struct {
    Name string
    Path string
}
```

Definisi modul dalam spesifikasi.

### 5.13 ServiceDef

```go
type ServiceDef struct {
    Name string
    Kind string
    Port int
}
```

Definisi service dalam spesifikasi.

## 6. Shared Contracts

Lokasi: `internal/shared/contracts/contracts.go`

### 6.1 Contract

```go
type Contract interface {
    Validate() error
}
```

Kontrak dasar untuk semua entitas yang dapat divalidasi.

### 6.2 SchemaAware

```go
type SchemaAware interface {
    SchemaVersion() string
}
```

Kontrak untuk entitas yang memiliki versi skema.

### 6.3 Versioned

```go
type Versioned interface {
    Version() string
}
```

Kontrak untuk entitas yang memiliki versi.

### 6.4 Identifiable

```go
type Identifiable interface {
    ID() string
}
```

Kontrak untuk entitas yang memiliki ID unik.

### 6.5 Named

```go
type Named interface {
    Name() string
}
```

Kontrak untuk entitas yang memiliki nama.

## 7. Usage Patterns

### Validating Entities

```go
type UserService struct {
    contracts.Contract
    contracts.Versioned
    contracts.Named
}

func (s *UserService) Validate() error {
    if s.Name() == "" {
        return fmt.Errorf("service name must not be empty")
    }
    return nil
}
```

### Using Shared Types

```go
// Validation result
result := types.ValidationResult{
    Valid: false,
    Errors: []types.ErrorInfo{
        {Code: "E001", Message: "project name required"},
    },
}

// Artifact
artifact := types.Artifact{
    Path:    "cmd/main.go",
    Content: []byte("package main\n..."),
}

// Task
task := types.Task{
    ID:           "gen-main",
    Name:         "Generate main.go",
    Dependencies: []string{"validate-spec", "build-model"},
    Priority:     1,
}
```

## 8. Constraints

- Shared types bersifat pure data — tidak ada logic bisnis.
- Shared contracts bersifat minimal — hanya mendefinisikan kemampuan dasar.
- Package ini tidak boleh memiliki dependensi ke package internal lainnya (hanya dependensi standar Go).
- Perubahan pada shared types harus mempertimbangkan dampak ke seluruh komponen.
