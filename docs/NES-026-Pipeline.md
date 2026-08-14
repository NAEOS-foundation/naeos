# NAEOS Pipeline Documentation

## Status
- Status: Stable
- Version: 1.0.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Overview

NAEOS Pipeline adalah komponen sentral yang mengorkestrasi seluruh alur transformasi dari spesifikasi menjadi artefak yang di-deploy. Pipeline mengintegrasikan seluruh komponen internal menjadi satu alur kerja yang utuh.

---

## 2. Architecture

```mermaid
graph TB
    subgraph Pipeline["NAEOS Pipeline"]
        direction TB
        Parser["Parser"] --> Normalizer["Normalizer"]
        Normalizer --> Resolver["Resolver"]
        Resolver --> Builder["Builder"]
        Builder --> Validator["Validator"]
        Validator --> GraphNode["Graph (PlannerGraph)"]
        GraphNode --> Policy["Policy Evaluator"]
        Policy --> Scheduler["Scheduler"]
        Scheduler --> Generator["Generator"]
        Generator --> Reviewer["Reviewer"]
    end

    subgraph Supporting["Supporting Services"]
        Kernel["Kernel — lifecycle & telemetry"]
        GraphService["Graph — dependency DAG"]
        Registry["Registry — component discovery"]
    end

    Pipeline --> Supporting

    style Pipeline fill:#e8f4fd,stroke:#2196F3
    style Supporting fill:#fff3e0,stroke:#FF9800
```

```
┌─────────────────────────────────────────────────────────────────┐
│                     NAEOS Pipeline                               │
│                                                                  │
│  ┌──────────┐   ┌────────────┐   ┌──────────┐   ┌──────────┐  │
│  │  Parser   │──▶│ Normalizer │──▶│ Resolver │──▶│ Builder  │  │
│  └──────────┘   └────────────┘   └──────────┘   └──────────┘  │
│                                                      │          │
│                                                      ▼          │
│  ┌──────────┐   ┌────────────┐   ┌──────────┐   ┌──────────┐  │
│  │Reviewer  │◀──│ Generator  │◀──│Scheduler │◀──│Validator │  │
│  └──────────┘   └────────────┘   └──────────┘   └──────────┘  │
│       │                │                │                        │
│       ▼                ▼                ▼                        │
│  ┌──────────┐   ┌────────────┐   ┌──────────┐                 │
│  │  Kernel  │   │   Graph    │   │ Registry │                 │
│  └──────────┘   └────────────┘   └──────────┘                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. Pipeline Flow

### 3.1 Run Flow

```mermaid
flowchart TD
    Input["Input (YAML/JSON Specification)"] --> Parser["Parser — parse input into SpecDocument"]
    Parser --> Normalizer["Normalizer — normalize to structured values"]
    Normalizer --> Resolver["Resolver — resolve references and defaults"]
    Resolver --> Builder["Builder — build NEIR model"]
    Builder --> Validator["Validator — validate NEIR model"]
    Validator --> ExecGraph["Execution Graph — build dependency graph"]
    ExecGraph --> PolicyEval["Policy Evaluator — evaluate governance rules"]
    PolicyEval --> Scheduler["Scheduler — generate execution tasks"]
    Scheduler --> Generator["Generator — generate artifacts"]
    Generator --> Reviewer["Reviewer — review generated artifacts"]
    Reviewer --> Output["Output (NEIR + Artifacts + Tasks + Reviews)"]
```

```
Input (YAML/JSON Specification)
    │
    ▼
┌─────────────────┐
│     Parser      │  Parse input into SpecDocument
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Normalizer    │  Normalize to structured values
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Resolver     │  Resolve references and defaults
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│     Builder     │  Build NEIR model
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Validator    │  Validate NEIR model
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Execution Graph │  Build dependency graph
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│Policy Evaluator │  Evaluate governance rules
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Scheduler     │  Generate execution tasks
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Generator     │  Generate artifacts
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Reviewer     │  Review generated artifacts
└────────┬────────┘
         │
         ▼
Output (NEIR + Artifacts + Tasks + Reviews)
```

---

## 4. Package API

### 4.1 Config

```go
type Config struct {
    Name       string
    Mode       string
    Verbose    bool
    OutputDir  string
    Parser     parser.Parser
    Normalizer normalizer.Normalizer
    Resolver   resolver.Resolver
    Builder    builder.Builder
    Validator  validator.Validator
    Scheduler  scheduler.Scheduler
    Generator  engine.GeneratorEngine
    Graph      *graph.PlannerGraph
    Registry   *registry.Registry
    Evaluator  policy.Evaluator
    Reviewer   review.Reviewer
    Kernel     *kernel.Kernel
    Policies   []policy.Rule
}
```

### 4.2 Result

```go
type Result struct {
    Source    string
    NEIR      *model.NEIR
    Artifacts []engine.Artifact
    Tasks     []scheduler.Task
    Graph     *graph.PlannerGraph
    Reviews   []*review.ReviewResult
}
```

---

## 5. Methods

### 5.1 New

```go
func New(cfg Config) (*Pipeline, error)
```

Creates a new pipeline with the given configuration. Any nil components will be initialized with default implementations.

**Example:**
```go
p, err := pipeline.New(pipeline.Config{
    Name:      "my-pipeline",
    Mode:      "standard",
    OutputDir: "./output",
})
```

### 5.2 Run

```go
func (p *Pipeline) Run(input string) (*Result, error)
```

Executes the full pipeline: parse → normalize → resolve → build → validate → build graph → policy evaluation → schedule → generate → review → write artifacts.

**Parameters:**
- `input` (string): YAML or JSON specification.

**Returns:**
- `*Result`: Contains NEIR model, generated artifacts, execution tasks, and review results.
- `error`: Any error during pipeline execution.

**Example:**
```go
spec := `
project: my-api
modules:
  - name: auth
    path: ./internal/auth
services:
  - name: gateway
    kind: http
    port: 8080
`

result, err := p.Run(spec)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Generated %d artifacts\n", len(result.Artifacts))
fmt.Printf("Scheduled %d tasks\n", len(result.Tasks))
```

### 5.3 Validate

```go
func (p *Pipeline) Validate(input string) (*Result, error)
```

Runs only the validation portion of the pipeline (parse → normalize → resolve → build → validate).

### 5.4 Kernel Integration Methods

```go
func (p *Pipeline) RegisteredKernelServices() []string
func (p *Pipeline) KernelMetrics() kernel.Metrics
func (p *Pipeline) KernelTopics() []string
func (p *Pipeline) Publish(topic string, payload any) error
func (p *Pipeline) Subscribe(topic string, handler func(any)) error
```

### 5.5 Accessor Methods

```go
func (p *Pipeline) Registry() *registry.Registry
func (p *Pipeline) Graph() *graph.PlannerGraph
```

---

## 6. Components

### 6.1 Parser

**Package:** `internal/specification/parser`

Parses YAML/JSON input into a `SpecDocument` struct.

```go
type SpecDocument struct {
    Raw          string
    Data         any
    Project      string
    Modules      []Module
    Services     []Service
    Architecture *Architecture
    Deployment   *Deployment
    Testing      *Testing
}
```

### 6.2 Normalizer

**Package:** `internal/specification/normalizer`

Normalizes parsed spec into structured `NormalizedSpec`.

### 6.3 Resolver

**Package:** `internal/specification/resolver`

Resolves references and applies defaults.

### 6.4 Builder

**Package:** `internal/neir/builder`

Builds the NEIR model from resolved spec.

### 6.5 Validator

**Package:** `internal/neir/validator`

Validates the NEIR model with comprehensive rules:
- Project name validation
- Module validation (name, path, dependencies)
- Service validation (name, port, kind)
- Architecture validation
- Duplicate detection
- Port range validation

### 6.6 Scheduler

**Package:** `internal/planner/scheduler`

Generates execution tasks from NEIR model:
- validate → build → modules → services → config → validate-output

### 6.7 Generator

**Package:** `internal/generation/engine`

Generates artifacts from NEIR model:
- README.md, Dockerfile, CI workflow, go.mod, main.go
- Module files (handler.go, service.go, repository.go, etc.)
- Service files (server.go, config.yaml)
- Deployment files (docker-compose.yml)
- Testing files (main_test.go, handler_test.go)
- Documentation files (architecture.md)

### 6.8 Reviewer

**Package:** `internal/governance/review`

Reviews generated artifacts with rules:
- `no-todo`: Checks for TODO comments
- `no-placeholder`: Checks for placeholder content
- `has-package-declaration`: Validates Go files have package declaration
- `has-license-header`: Checks for license header

### 6.9 Policy Evaluator

**Package:** `internal/governance/policy`

Evaluates governance rules with operators:
- `exists:key` - Check if key exists
- `not_empty:key` - Check if key is not empty
- `contains:key,substr` - Check if value contains substring
- `gt:key,value` - Greater than comparison
- `lt:key,value` - Less than comparison
- `in:key,v1,v2` - Check if value is in list

### 6.10 Graph

**Package:** `internal/planner/graph`

Builds execution dependency graph from NEIR model.

### 6.11 Registry

**Package:** `internal/registry`

Thread-safe service registry for component discovery.

---

## 7. Kernel Integration

Pipeline automatically registers all components with the kernel:

```
parser, normalizer, resolver, builder, validator,
scheduler, generator, graph, registry, evaluator, reviewer, pipeline
```

The kernel manages lifecycle and provides telemetry events:
- `kernel.start` - Pipeline started
- `kernel.stop` - Pipeline stopped
- `pipeline.validate` - Validation completed
- `pipeline.run` - Full pipeline executed

---

## 8. Configuration File

```yaml
# config.yaml
pipeline:
  name: my-pipeline
  mode: standard
  verbose: true
  output_dir: ./output
```

---

## 9. Example Usage

### 9.1 Basic Usage

```go
package main

import (
    "fmt"
    "github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func main() {
    p, err := pipeline.New(pipeline.Config{
        Name:      "demo",
        OutputDir: "./output",
    })
    if err != nil {
        panic(err)
    }

    spec := `
project: demo-api
modules:
  - name: auth
    path: ./internal/auth
services:
  - name: gateway
    kind: http
    port: 8080
`

    result, err := p.Run(spec)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Project: %s\n", result.NEIR.Project.Name)
    fmt.Printf("Modules: %d\n", len(result.NEIR.Modules))
    fmt.Printf("Services: %d\n", len(result.NEIR.Services))
    fmt.Printf("Artifacts: %d\n", len(result.Artifacts))
    fmt.Printf("Tasks: %d\n", len(result.Tasks))
    fmt.Printf("Reviews: %d\n", len(result.Reviews))
}
```

### 9.2 With Policies

```go
p, _ := pipeline.New(pipeline.Config{
    Name: "secure-pipeline",
    Policies: []policy.Rule{
        {RuleID: "require-auth", Condition: "exists:modules", Priority: 1, Action: "block", Enabled: true},
    },
})

result, err := p.Run(spec)
```

### 9.3 With Event Handling

```go
p, _ := pipeline.New(pipeline.Config{Name: "event-pipeline"})

p.Subscribe("specification.created", func(payload any) {
    fmt.Println("Specification created!")
})

result, err := p.Run(spec)
```

---

## 10. References

- [NES-002-Kernel-API.md](NES-002-Kernel-API.md) - Kernel API Reference
- [NES-023-NEIR.md](NES-023-NEIR.md) - NEIR Model
- [NAEOS-KER-001.md](../kernel/NAEOS-KER-001.md) - Kernel Architecture
