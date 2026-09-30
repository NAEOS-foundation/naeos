# NES-039 SDK Multi-Language

## 1. Status
- Status: Draft
- Version: 0.1
- Owner: NAEOS Core Team

## 2. Purpose
Dokumentasi spesifikasi untuk SDK multi-language — kemampuan NAEOS menghasilkan proyek dalam berbagai bahasa pemrograman dari satu spesifikasi.

## 3. Scope
Dokumen ini mencakup arsitektur adapter, kontrak SDK per bahasa, konfigurasi bahasa, dan panduan penggunaan.

## 4. Normative References
- NES-040 Output Adapter Architecture
- NES-019 SDK
- NAEOS-SPEC-008 Compiler Model
- NAEOS-GOV-001 Project Charter — Architecture Drives Engineering.

## 5. Supported Languages

| Language | Status | Package Manager | Build File | Base Image |
|----------|--------|----------------|------------|------------|
| Go | Stable | go mod | go.mod | golang:1.22-alpine |
| TypeScript | Stable | npm | package.json | node:22-alpine |
| Python | Stable | pip/pyproject | pyproject.toml | python:3.12-slim |
| Java | Stable | maven | pom.xml | eclipse-temurin:21-jdk-alpine |
| Rust | Stable | cargo | Cargo.toml | rust:1.78-alpine |

## 6. Architecture

```mermaid
flowchart TD
    Spec["Specification (YAML/JSON)"] --> NEIR["NEIR Model (with GenerationConfig)"]
    NEIR --> LangSelector["Language Selector"]
    LangSelector --> Registry["Output Adapter Registry"]
    Registry --> Go["GoAdapter"]
    Registry --> TS["TSAdapter"]
    Registry --> Py["PyAdapter"]
    Registry --> Java["JavaAdapter"]
    Registry --> Rust["RustAdapter"]
    Go --> Artifacts["Combined Artifacts"]
    TS --> Artifacts
    Py --> Artifacts
    Java --> Artifacts
    Rust --> Artifacts
```

```
Specification (YAML/JSON)
    ↓
NEIR Model (with GenerationConfig)
    ↓
Language Selector (from config or NEIR)
    ↓
┌─────────────────────────────────────────┐
│            Output Adapter Registry       │
├──────────┬──────────┬──────────┬────────┤
│ GoAdapter│ TSAdapter│ PyAdapter│ ...    │
├──────────┼──────────┼──────────┼────────┤
│ Generate │ Generate │ Generate │ ...    │
│ Project  │ Project  │ Project  │        │
│ Module   │ Module   │ Module   │        │
│ Service  │ Service  │ Service  │        │
│ Docker   │ Docker   │ Docker   │        │
│ CI       │ CI       │ CI       │        │
└──────────┴──────────┴──────────┴────────┘
    ↓
Artifacts (language-specific)
```

## 7. Configuration

### 7.1 Pipeline Config

```yaml
pipeline:
  name: my-project
  mode: full
  language:
    - go
    - typescript
  output_dir: ./out
```

### 7.2 NEIR GenerationConfig

```go
type GenerationConfig struct {
    Languages []language.Language `json:"languages,omitempty"`
    OutputDir string             `json:"output_dir,omitempty"`
}
```

### 7.3 Specification

```yaml
project: my-project
generation:
  languages:
    - go
    - typescript
```

## 8. OutputAdapter Interface

```go
type OutputAdapter interface {
    Language() language.Language
    GenerateProject(projectName string) []Artifact
    GenerateModule(moduleName, modulePath, projectName string) []Artifact
    GenerateService(serviceName, serviceKind string, servicePort int, projectName string) []Artifact
    GenerateDockerfile(projectName string) []Artifact
    GenerateCI(projectName string) []Artifact
    GenerateDockerCompose(projectName string) []Artifact
    GenerateArchitectureDoc(projectName, pattern string) []Artifact
}
```

## 9. Generated Artifacts per Language

### 9.1 Go
- `go.mod`, `cmd/app/main.go`, `Dockerfile` (multi-stage), CI workflow
- Per module: handler.go, repository.go, service.go, domain/model.go, http/handler.go, http/router.go, middleware/logging.go, config/config.go, handler_test.go
- Per service: server.go, server_test.go, config.yaml

### 9.2 TypeScript
- `package.json`, `tsconfig.json`, `src/index.ts`, Dockerfile (multi-stage), CI workflow
- Per module: index.ts, handler.ts, service.ts, repository.ts, types.ts, handler.test.ts
- Per service: index.ts, server.ts

### 9.3 Python
- `pyproject.toml`, `__init__.py`, `__main__.py`, Dockerfile (multi-stage), CI workflow
- Per module: __init__.py, handler.py, service.py, repository.py, models.py, test_*.py
- Per service: __init__.py, server.py

### 9.4 Java
- `pom.xml`, `App.java`, Dockerfile (multi-stage), CI workflow
- Per module: Handler.java, Service.java, Repository.java, Model.java, HandlerTest.java
- Per service: Server.java

### 9.5 Rust
- `Cargo.toml`, `src/main.rs`, `src/lib.rs`, Dockerfile (multi-stage), CI workflow
- Per module: mod.rs, handler.rs, service.rs, repository.rs, models.rs, *_test.rs
- Per service: server.rs

## 10. Multi-Language Generation

Ketika beberapa bahasa dikonfigurasi, NAEOS menghasilkan output terpisah per bahasa:

```
out/
├── go/
│   ├── go.mod
│   ├── cmd/app/main.go
│   ├── Dockerfile
│   └── ...
├── typescript/
│   ├── package.json
│   ├── tsconfig.json
│   ├── src/index.ts
│   ├── Dockerfile
│   └── ...
└── python/
    ├── pyproject.toml
    ├── src/__init__.py
    ├── Dockerfile
    └── ...
```

## 11. Usage Example

```yaml
# specification.yaml
project: my-ecommerce
generation:
  languages:
    - go
    - typescript
    - python

modules:
  - name: user
    path: ./internal/user
  - name: order
    path: ./internal/order

services:
  - name: api-gateway
    kind: http
    port: 8080
```

```bash
# Generate for all configured languages
naeos run specification.yaml

# Generate for specific language only
naeos run specification.yaml --language go
```

## 12. Adding a New Language

Untuk menambahkan bahasa baru:

1. Buat file adapter di `internal/generation/adapters/<language>.go`.
2. Implementasikan `OutputAdapter` interface.
3. Register adapter di `init()` function.
4. Tambahkan bahasa ke `internal/neir/model/language/language.go`.
5. Tambahkan base image dan build file ke language package.
6. Tulis dokumentasi NES untuk bahasa baru.

## 13. Acceptance Criteria
- Single specification produces working projects in all 5 supported languages.
- Each generated project builds and runs without modification.
- Dockerfile uses correct base image per language.
- CI workflow uses correct setup action per language.
- New languages can be added by implementing the OutputAdapter interface.
