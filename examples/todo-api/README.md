# NAEOS Todo API Killer Demo

## What This Demo Demonstrates

The NAEOS Todo API demo showcases the core value proposition of NAEOS through a complete specification-first engineering workflow:

- **Specification-First Engineering**: The Todo API is defined entirely through a NAEOS specification (`spec.yaml`), serving as the single source of truth.
- **NEIR (NAEOS Engineering Intermediate Representation)**: The specification is parsed, normalized, resolved, and transformed into a structured NEIR model representing entities, fields, relationships, operations, and constraints.
- **Validation**: Real NAEOS validation is executed against the specification, confirming schema validity, semantic checks, and structural integrity before any artifacts are produced.
- **Artifact Generation**: The validated specification generates Go source code, tests, configuration files, documentation, and project scaffolding through the NAEOS generation pipeline.
- **Traceability**: Every stage of the pipeline produces traceable artifacts with run IDs, specification hashes, and NEIR hashes, enabling full provenance tracking from requirement to generated artifact.
- **AI-Ready Context**: The same specification model provides structured engineering context consumable by AI development tools via the `naeos context` and `naeos ai compile` commands.

## Quick Start

### Build the NAEOS CLI

```bash
go build -o naeos ./cmd/naeos
```

### Run the Demo

```bash
./examples/todo-api/run-demo.sh
```

Or with a custom binary:

```bash
NAEOS_BIN=/path/to/naeos ./examples/todo-api/run-demo.sh
```

### Individual Commands

```bash
# Parse and validate the specification
naeos validate --config examples/todo-api/naeos.yaml --input-file examples/todo-api/spec.yaml --output json

# Generate AI context bundle
naeos context --input-file examples/todo-api/spec.yaml --output markdown

# Run the full pipeline and generate artifacts
naeos run --config examples/todo-api/naeos.yaml --input-file examples/todo-api/spec.yaml --output json

# Compile for AI agent (requires NAEOS_LLM_API_KEY)
naeos ai compile --input-file examples/todo-api/spec.yaml --target opencode
```

## Expected Result

The demo produces a complete engineering run with:

- **Specification parsed** and validated successfully
- **NEIR model constructed** with 2 modules (user, todo), 1 service, and 10 endpoints
- **Validation passed** with all schema, semantic, and structural checks
- **Artifacts generated** including Go source code, tests, configuration, and documentation
- **Traceability chain** established with run_id, specification_hash, and neir_hash

### Sample Output

```text
NAEOS Todo API Killer Demo
========================================

[1/6] Loading specification
  ✓ todo.yaml

[2/6] Building NEIR
  ✓ project: todo-api

[3/6] Validating
  ✓ valid — project: todo-api

[4/6] Building engineering context
  ✓ structured context generated

[5/6] Generating artifacts
  ✓ artifacts generated: 42
  ✓ run_id: pipe-...
  ✓ spec_hash: ...
  ✓ neir_hash: ...

[6/6] Traceability
  ✓ specification → NEIR
  ✓ NEIR → artifacts
  ✓ artifacts → tests
  ✓ traceability chain established

========================================
  Failure Demonstration
========================================

Running validation on intentionally invalid specification...

✗ validation failed
  [PIPELINE_FAILED] VALIDATION_ERROR: validation failed:
  - duplicate module name "todo" at index 0 and 1 — module names must be unique

NAEOS stopped before artifact generation.

========================================
  NAEOS Demo completed successfully.
========================================
```

## Architecture

The demo illustrates the complete NAEOS engineering pipeline:

```text
Specification (spec.yaml)
      │
      ▼
Parse → Normalize → Resolve → Build NEIR → Validate
                                              │
                                              ▼
                           Schedule → Generate → Artifacts
                                              │
                                              ▼
                            AI Context · Governance · Evidence / Audit
```

### Specification Structure

The Todo API specification defines:

- **Project**: `todo-api`
- **Modules**: `user` (user management) and `todo` (todo management), with `todo` depending on `user`
- **Service**: `api-gateway` (HTTP service on port 8080) with 10 CRUD endpoints
- **Architecture**: Hexagonal pattern
- **Generation**: Go language output

### Domain Model

- **User**: id, name, email (managed by the `user` module)
- **Todo**: id, user_id, title, description, completed, created_at, updated_at (managed by the `todo` module)
- **Relationship**: User 1 ──── N Todo

## Failure Scenario

The `spec-invalid.yaml` demonstrates NAEOS's validation capabilities by using duplicate module names. Running validation on this specification produces a real validation failure:

```bash
naeos validate --config examples/todo-api/naeos.yaml --input-file examples/todo-api/spec-invalid.yaml --output text
```

This demonstrates that NAEOS is not simply a code generator — it acts as an engineering control/validation layer that stops before artifact generation when the specification is invalid.

## Why This Matters

NAEOS solves the fundamental problem of disconnected specifications and implementations. By making the specification the single source of truth, NAEOS ensures that:

1. **Requirements are traceable** through every stage of the engineering pipeline
2. **Validation prevents invalid systems** from being generated
3. **Artifacts are reproducible** from the same specification
4. **AI tools receive structured context** rather than raw text
5. **Engineering discipline is enforced** through automated validation and policy evaluation

## Limitations

- The NAEOS specification format uses `services` with `endpoints` for API definitions; domain entity fields (User, Todo) are not explicitly represented as typed entities in the current spec format but are captured through module descriptions and service endpoints.
- Port numbers in service definitions may not be fully reflected in the NEIR model due to a known parser type assertion issue with `int` vs `int64` in YAML deserialization.
- The AI compile feature (`naeos ai compile`) requires an LLM API key and is not included in the default demo flow.

## Documentation

- [NAEOS Getting Started](../../GETTING-STARTED.md)
- [NAEOS Repository README](../../README.md)
- [NAEOS Architecture Overview](../../ARCHITECTURE-OVERVIEW.md)
