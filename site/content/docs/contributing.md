---
title: Contributing
description: Guide for contributing to the NAEOS project.
weight: 20
---

This guide covers everything you need to know to contribute to NAEOS.

## Development Setup

### Prerequisites

- Go 1.25 or later
- Git
- golangci-lint (recommended)

### Getting Started

```bash
# Clone the repository
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos

# Build the project
go build ./cmd/naeos/

# Run tests
go test -race -count=1 -timeout 300s ./...

# Run linter
golangci-lint run ./...
```

## Project Structure

```text
naeos/
├── cmd/naeos/           # CLI commands (35+ files)
│   ├── main.go          # Root command
│   ├── run_cmd.go       # Run pipeline
│   ├── compile_cmd.go   # Compile to AI tools
│   ├── context_cmd.go   # Generate context bundles
│   └── ...              # 30+ command files
├── internal/
│   ├── specification/   # Parser, normalizer, resolver
│   ├── neir/            # NEIR model and builder
│   ├── compiler/        # AI instruction compiler
│   ├── context/         # Context bundle generator
│   ├── generation/      # Code generation
│   ├── governance/      # Policy and review
│   ├── artifacts/       # Artifact store
│   ├── profiles/        # Industry profiles
│   ├── migration/       # Schema migration
│   ├── security/        # Security rules
│   └── shared/          # Shared utilities
├── pkg/
│   ├── pipeline/        # Main pipeline
│   ├── kernel/          # System kernel
│   ├── config/          # Configuration
│   └── plugin/          # Plugin system
├── docs/                # NES specifications (56 files)
├── governance/          # Governance documents
├── constitution/        # Engineering constitution
├── specification/       # Specification documents
└── site/                # Hugo website
```

## Coding Standards

### Go Style

- Follow `gofmt` and `go vet` conventions
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused (ideally <50 lines)
- Return errors; never panic in library code
- Use `sync.Mutex`/`sync.RWMutex` for thread safety
- Prefer standard library patterns over third-party packages

### Testing

- Write table-driven tests with `tt` struct
- Use `t.Parallel()` where safe
- Maintain test coverage above 80%
- Mock external dependencies
- Test edge cases and error paths

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

### Documentation

- Update documentation for new features
- Add examples for CLI commands
- Keep CHANGELOG updated
- Write commit messages that explain **why**, not just **what**

## Pull Request Process

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/my-feature`)
3. **Write** tests for your changes
4. **Run** the full test suite:
   ```bash
   go test -race -count=1 -timeout 300s ./...
   golangci-lint run ./...
   ```
5. **Update** documentation if needed
6. **Commit** with a descriptive message
7. **Push** and create a Pull Request

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```text
feat: add new parser for HCL configuration
fix: resolve circular dependency detection in validator
docs: update CLI reference with new flags
test: add integration tests for pipeline
refactor: simplify NEIR builder API
chore: update dependencies
```

## Code Review

- All PRs require at least one review
- Tests must pass in CI
- Documentation must be updated for new features
- Follow existing coding conventions
- Keep PRs focused — one feature or fix per PR

## Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include steps to reproduce for bugs
- Specify Go version and OS for environment-specific issues
- Check existing issues before creating new ones

## License

By contributing to NAEOS, you agree that your contributions will be licensed under the Apache License 2.0.
