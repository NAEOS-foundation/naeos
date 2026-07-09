# NAEOS Foundation — NAEOS

> Specify Once. Build Anywhere.

NAEOS (Nusantara Engineering & Architecture Operating System) adalah platform engineering deklaratif yang mengubah spesifikasi menjadi sistem perangkat lunak berkualitas tinggi melalui pipeline yang konsisten, tervalidasi, dan dapat diperluas.

NAEOS bukan sekadar project generator. NAEOS adalah engineering runtime yang memahami spesifikasi, membangun model internal, menyusun rencana eksekusi, menghasilkan artifact, memvalidasi hasil, dan menjaga proyek tetap selaras dengan spesifikasi sepanjang siklus hidupnya.

## Vision
Membangun platform engineering open source yang memungkinkan pengembang dan organisasi mendeskripsikan sistem sekali, kemudian membangun, memvalidasi, dan mengembangkan perangkat lunak di berbagai bahasa, framework, dan platform.

## Motto
Specify Once. Build Anywhere.

## Core Philosophy
- Specification is the Source of Truth
- Everything is Declarative
- Everything is an Artifact
- Everything is Versioned
- Everything is Extensible
- Small Kernel, Powerful Extensions
- Human Review Before Automation

## High-Level Architecture

```text
Specification
      │
      ▼
Parser
      │
      ▼
Normalizer
      │
      ▼
Resolver
      │
      ▼
NEIR
      │
      ├──────────────┐
      ▼              ▼
Planner         Knowledge Graph
      │              │
      └──────┬───────┘
             ▼
     Execution Graph
             │
             ▼
   Generator Engine
             │
             ▼
        Artifacts
             │
             ▼
        Validator
             │
             ▼
   Published Project
```

## Core Components

### Kernel
Kernel menyediakan fondasi runtime:
- Lifecycle Manager
- Event Bus
- Dependency Injection
- Service Registry
- Scheduler
- Configuration
- Logging
- Telemetry

Kernel tidak berisi business logic.

### Specification
Specification menggunakan NAEOS Specification Language sebagai sumber kebenaran utama.

### NEIR
NAEOS Engineering Intermediate Representation adalah model engineering sentral yang merepresentasikan seluruh sistem. NEIR bukan AST, bukan JSON, dan bukan YAML; NEIR adalah struktur komposit yang memuat project, architecture, domain, module, component, service, API, storage, infrastructure, security, AI, documentation, deployment, testing, dan metadata.

NEIR menjadi antarmuka utama bagi planner, generator, validator, dan runtime.

### Planner
Planner menyusun execution graph berdasarkan dependency antar artifact sehingga proses build menjadi deterministik dan dapat dioptimalkan.

### Generator
Generator menghasilkan:
- Source Code
- Documentation
- Dockerfile
- CI/CD Pipeline
- Infrastructure
- Configuration
- Testing
- Release Assets

### Validator
Validator memastikan seluruh artifact sesuai dengan NEIR dan specification.

## Repository Structure

```text
constitution/
governance/
kernel/
policy/
profile/
specification/
Reference Architecture/
internal/
  specification/
    lexer/
    parser/
    normalizer/
    resolver/
  neir/
    model/
      project/
      architecture/
      domain/
      module/
      component/
      service/
      api/
      storage/
      infrastructure/
      security/
      ai/
      documentation/
      deployment/
      testing/
      metadata/
    builder/
    serializer/
    validator/
    version/
  planner/
    graph/
    scheduler/
    optimizer/
  generation/
    engine/
    templates/
    targets/
    renderers/
  governance/
    policy/
    review/
  knowledge/
    graph/
    index/
    provenance/
  runtime/
    engine/
    lifecycle/
    telemetry/
  kernel/
    services/
    registry/
    events/
  shared/
    errors/
    types/
    utils/
    contracts/
templates/
```

## Why this repository exists
Repositori ini berfungsi sebagai sumber kebenaran utama untuk:
- standar engineering yang konsisten antar proyek,
- kebijakan dan aturan yang dapat diaudit,
- dokumentasi yang dapat diproses oleh manusia dan mesin,
- alur kerja yang mendukung traceability dari requirement ke implementation.

## Dokumentasi utama
- [DOCUMENTATION-INDEX.md](DOCUMENTATION-INDEX.md) — indeks lengkap dokumen proyek.
- [GETTING-STARTED.md](GETTING-STARTED.md) — panduan onboarding.
- [CONTRIBUTING.md](CONTRIBUTING.md) — pedoman kontribusi.
- [templates/ADR-template.md](templates/ADR-template.md) — template ADR.
- [templates/RFC-template.md](templates/RFC-template.md) — template RFC.
- [docs/NES-024-Internal-Structure.md](docs/NES-024-Internal-Structure.md) — draft struktur folder internal.
- [docs/NES-025-Implementation-Skeletons.md](docs/NES-025-Implementation-Skeletons.md) — draft skeleton file-level untuk modul internal.

## Development Roadmap

### NAEOS v1
- Specification
- Parser
- Normalizer
- Resolver
- NEIR
- Go Generator

### NAEOS v2
- NEIR
- Go
- Node
- Rust
- Python

### NAEOS v3
- NEIR
- Planner
- AI Optimizer

### NAEOS v4
- NEIR
- Distributed Build
- Cloud Runtime

### NAEOS v5
- NEIR
- Autonomous Engineering

### Alpha 0.1
- NSL
- Parser
- NEIR
- Planner
- Go Generator
- Validator

### Alpha 0.2
- Workspace
- Registry
- Plugin System
- Artifact Store

### Alpha 0.3
- Kernel Runtime
- Event Bus
- Lifecycle
- Telemetry

### Beta
- Knowledge Graph
- Policy Engine
- Multi-language Generator
- AI Context

### Version 1.0
- Stable CLI
- Stable Plugin API
- Stable Specification Schema
- Production Documentation

## CLI Preview

```text
naeos init
naeos build
naeos validate
naeos inspect
naeos doctor
naeos repair
naeos plugin
naeos release
```

## Long-Term Vision
NAEOS dikembangkan sebagai platform engineering modern yang memungkinkan pengembang berpindah dari pendekatan code-first menuju specification-first, sehingga seluruh proyek dapat dibangun, dipelihara, dan divalidasi secara konsisten.

## License
Apache License 2.0

## Status
🚧 Early Engineering Phase

NAEOS saat ini berada pada tahap pembangunan fondasi arsitektur dan implementasi Core Runtime.
