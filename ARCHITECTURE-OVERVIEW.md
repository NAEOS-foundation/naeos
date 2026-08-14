# Architecture Overview

This document provides a high-level conceptual architecture of NAEOS.

## System Architecture

```mermaid
graph TB
    subgraph Input
        YAML[Specification YAML/JSON]
        CLI[CLI Commands]
    end

    subgraph Core["Core Runtime"]
        KB[Kernel Bus]
        Engine[Engine / Parser]
        Validator[Validator]
        Compiler[Compiler]
        Scheduler[Scheduler]
    end

    subgraph Reasoning["Reasoning Layer"]
        RG[Reasoning Graph]
        KG[Knowledge Graph]
        AI[AI Integration]
    end

    subgraph Generation["Generation Layer"]
        GenModel[Generator]
        Adapters[Output Adapters]
        Go[Go Adapter]
        TS[TypeScript Adapter]
        PY[Python Adapter]
        JA[Java Adapter]
        RU[Rust Adapter]
    end

    subgraph Output
        NEIR[NEIR Model]
        Artifacts[Generated Artifacts]
        Docs[Documentation]
    end

    YAML --> Engine
    CLI --> KB
    Engine --> KB
    KB --> Validator
    KB --> Compiler
    KB --> Scheduler
    Validator --> Compiler
    Compiler --> RG
    RG --> GenModel
    KG --> GenModel
    AI --> GenModel
    GenModel --> Adapters
    Adapters --> Go
    Adapters --> TS
    Adapters --> PY
    Adapters --> JA
    Adapters --> RU
    Go --> NEIR
    TS --> NEIR
    PY --> NEIR
    JA --> NEIR
    RU --> NEIR
    NEIR --> Artifacts
    NEIR --> Docs
```

## Data Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Pipeline
    participant Engine
    participant Validator
    participant Compiler
    participant Generator
    participant Adapter

    User->>CLI: naeos run --config config.yaml --input spec.yaml
    CLI->>Pipeline: New(cfg) + Run(input)
    Pipeline->>Engine: Parse(spec)
    Engine-->>Pipeline: Source AST
    Pipeline->>Validator: Validate(ast)
    Validator-->>Pipeline: ValidationReport
    Pipeline->>Compiler: Compile(ast)
    Compiler-->>Pipeline: NEIR
    Pipeline->>Generator: Generate(NEIR, languages)
    Generator->>Adapter: GenerateForNEIR(NEIR)
    Adapter-->>Generator: []Artifact
    Generator-->>Pipeline: NEIRResult{NEIR, Artifacts}
    Pipeline-->>CLI: result
    CLI-->>User: stdout or exported files
```

## Layered Architecture

```mermaid
graph TB
    subgraph Layer1["1. Specification Layer"]
        NES[NES Documents]
        SPEC[SPEC Documents]
        GOV[Governance Docs]
    end

    subgraph Layer2["2. Validation Layer"]
        VAL[Policy Validator]
        RULES[Rule Engine]
        DEPS[Dependency Graph]
    end

    subgraph Layer3["3. Reasoning Layer"]
        RG[Reasoning Graph]
        KG[Knowledge Graph]
        TRACE[Traceability]
    end

    subgraph Layer4["4. Generation Layer"]
        GEN[Generator]
        ADP[Adapters]
        TPL[Template Engine]
    end

    subgraph Layer5["5. Output Layer"]
        NEIR[NEIR Model]
        FILES[Generated Files]
        DOCS[Documentation]
    end

    NES --> VAL
    SPEC --> VAL
    GOV --> VAL
    VAL --> RULES
    RULES --> DEPS
    DEPS --> RG
    RG --> KG
    KG --> TRACE
    TRACE --> GEN
    GEN --> ADP
    ADP --> TPL
    TPL --> NEIR
    NEIR --> FILES
    NEIR --> DOCS
```

## NEIR Model Structure

```mermaid
classDiagram
    class NEIR {
        +Metadata metadata
        +Project project
        +Module[] modules
        +Service[] services
        +GenerationConfig generation
    }

    class Metadata {
        +string neir_version
        +string schema_version
        +string project_version
        +Time created_at
    }

    class Project {
        +string name
        +string description
        +string version
    }

    class Module {
        +string name
        +string path
        +string description
    }

    class Service {
        +string name
        +string kind
        +int port
    }

    class GenerationConfig {
        +string[] languages
        +string output_dir
        +bool enabled
    }

    NEIR --> Metadata
    NEIR --> Project
    NEIR --> Module
    NEIR --> Service
    NEIR --> GenerationConfig
```

## Purpose

NAEOS is designed to connect four major layers:

1. **Governance** — establishes organizational rules and processes.
2. **Specification** — defines requirements, design, and contracts.
3. **Constitution** — holds normative principles that cannot be violated.
4. **Policy Compiler** — transforms policies into executable rules.

## Conceptual Flow

Requirements and intents enter the specification layer.
After that, policies and governance are mapped to rules that can be validated.
The final output is implementation artifacts, documentation, and consistent execution rules.

## Main Components

- **Governance layer**: organizational rules, standards, and processes
- **Specification layer**: NES and SPEC documents defining system behavior
- **Constitution layer**: normative principles enforced by the system
- **Policy layer**: rules compiled from governance and specification
- **Validation and compiler pipeline**: transforms specs into NEIR
- **Output adapters**: generate code in Go, TypeScript, Python, Java, Rust
- **Reasoning graph**: decision traceability and knowledge management

## Design Principles

- Human readable specifications
- Machine readable NEIR output
- Vendor neutral (multi-language, multi-cloud)
- Extensible via adapters and plugins
- Deterministic pipeline execution

## Repository Structure

```text
naeos/
├── cmd/naeos/          # CLI entry point
├── pkg/pipeline/       # Pipeline orchestration
├── pkg/kernel/         # Kernel services, event bus, lifecycle
├── pkg/config/         # Configuration loading and validation
├── internal/
│   ├── neir/           # NEIR model and sub-packages
│   │   ├── model/      # Domain models (ai, api, architecture, ...)
│   │   └── validator/  # Validation engine
│   ├── specification/  # Parser, normalizer, resolver
│   ├── compiler/       # NEIR compiler
│   ├── generation/     # Generation engine and adapters
│   ├── scheduler/      # Task scheduler
│   ├── pluginsdk/      # Plugin SDK (scaffold, sandbox, wasm)
│   └── shared/         # Shared contracts and utilities
├── specification/      # NES/SPEC documents
├── docs/               # Documentation
└── examples/           # Example specifications
```
