# NES-028 — CLI Reference

> Status: Stable
> Last Updated: 2026-07-13

Complete reference for the `naeos` CLI tool.

---

## Global Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--verbose` | `bool` | `false` | Enable verbose logging |

```bash
naeos --verbose run --config config.yaml --input spec.yaml
```

---

## Commands

### `naeos version`

Show NAEOS version.

```bash
naeos version
# Output: naeos <version>
```

---

### `naeos init`

Generate a default NAEOS config file.

```bash
naeos init                          # Creates config.example.yaml
naeos init -o myconfig.yaml         # Custom output path
```

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--output` | `-o` | `string` | `config.example.yaml` | Output path |

**Generated config:**

```yaml
pipeline:
  name: naeos-dev
  mode: development
  verbose: true
  output_dir: ./out
```

---

### `naeos run`

Execute the full NAEOS pipeline.

```bash
naeos run --config config.yaml --input spec.yaml
naeos run --config config.yaml --input "my specification text"
naeos run --config config.yaml --input spec.yaml --output json
naeos run --config config.yaml --input spec.yaml --output yaml --output-file result.yaml
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to JSON or YAML config file |
| `--input` | `string` | `""` | One of `--input`/`--input-file` | Inline specification text |
| `--input-file` | `string` | `""` | One of `--input`/`--input-file` | Path to specification file |
| `--output` | `string` | `text` | No | Output format: `text`, `json`, `yaml` |
| `--output-file` | `string` | `""` | No | Write output to file |
| `--language` | `string` | `""` | No | Target language (repeatable: `--language go --language typescript`) |

**Text output example:**

```
pipeline=my-app mode=development verbose=true output_dir=./out
artifacts=12 tasks=8
```

**JSON output keys:** `pipeline`, `mode`, `verbose`, `output_dir`, `artifacts`, `tasks`

**Multi-language example:**

```bash
# Generate Go + TypeScript artifacts
naeos run --config config.yaml --input spec.yaml --language go --language typescript

# Generate Python artifacts only
naeos run --config config.yaml --input spec.yaml --language python

# All 5 supported languages
naeos run --config config.yaml --input spec.yaml \
  --language go --language typescript --language python \
  --language java --language rust
```

---

### `naeos validate` (alias: `v`)

Validate a specification using the NAEOS pipeline.

```bash
naeos validate --config config.yaml --input spec.yaml
naeos v --config config.yaml --input spec.yaml          # alias
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--input` | `string` | `""` | One of `--input`/`--input-file` | Inline specification text |
| `--input-file` | `string` | `""` | One of `--input`/`--input-file` | Path to specification file |
| `--output` | `string` | `text` | No | Output format: `text`, `json`, `yaml` |
| `--output-file` | `string` | `""` | No | Write output to file |
| `--language` | `string` | `""` | No | Target language (repeatable) |

**Output keys:** `pipeline`, `mode`, `verbose`, `output_dir`, `status` (`"valid"`), `project`, `source_len`

---

### `naeos inspect`

Inspect the NAEOS pipeline result.

```bash
naeos inspect --config config.yaml --input spec.yaml
naeos inspect --config config.yaml --input spec.yaml --output json
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--input` | `string` | `""` | Yes | Specification text or file path |
| `--output` | `string` | `text` | No | Output format: `text`, `json`, `yaml` |
| `--output-file` | `string` | `""` | No | Write output to file |

**Output keys:** `pipeline`, `mode`, `verbose`, `output_dir`, `project`, `input`, `artifacts`, `tasks`, `source_words`

---

### `naeos doctor`

Run diagnostics on the NAEOS configuration.

```bash
naeos doctor --config config.yaml
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |

**Output keys:** `config`, `pipeline_name`, `mode`, `output_dir`, `output_dir_exists`

---

### `naeos repair`

Repair the NAEOS output directory (creates directory and README.md if missing).

```bash
naeos repair --config config.yaml
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |

---

### `naeos scaffold`

Generate a starter project scaffold.

```bash
naeos scaffold --name my-project
naeos scaffold --name my-project --output ./scaffold
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--name` | `string` | `""` | Yes | Project name |
| `--output` | `string` | `""` | No | Output directory (defaults to `--name`) |
| `--language` | `string` | `""` | No | Target language (repeatable) |

**Generated structure:**

```
<scaffold>/
  README.md
  spec.yaml
  Makefile
  .gitignore
  Dockerfile
  .github/workflows/ci.yml
  go.mod
  config.yaml
  config.json
  cmd/app/main.go
  internal/core/
    README.md
    package.go
    config.yaml
    handler.go
    repository.go
    service.go
    domain/model.go
    http/handler.go
    http/router.go
    middleware/logging.go
    config/config.go
    config/load.go
    handler_test.go
```

---

### `naeos export`

Export generated artifacts to a directory.

```bash
naeos export --config config.yaml --input spec.yaml
naeos export --config config.yaml --input spec.yaml --output-dir ./artifacts
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--input` | `string` | `""` | Yes | Specification text or file path |
| `--output-dir` | `string` | `""` | No | Output directory (falls back to config `output_dir`) |
| `--language` | `string` | `""` | No | Target language (repeatable) |

---

### `naeos preview`

Preview generated artifacts without writing them to disk.

```bash
naeos preview --config config.yaml --input spec.yaml
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--input` | `string` | `""` | Yes | Specification text or file path |

---

### `naeos benchmark`

Run performance benchmarks on the pipeline and generators.

```bash
naeos benchmark --config config.yaml --input spec.yaml
naeos benchmark --config config.yaml --input spec.yaml --iterations 100
naeos benchmark --config config.yaml --input spec.yaml --output json
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--input` | `string` | `""` | One of `--input`/`--input-file` | Inline specification text |
| `--input-file` | `string` | `""` | One of `--input`/`--input-file` | Path to specification file |
| `--iterations` | `int` | `10` | No | Number of benchmark iterations |
| `--output` | `string` | `text` | No | Output format: `text`, `json`, `yaml` |

**Output keys:** `pipeline`, `iterations`, `avg_ms`, `p50_ms`, `p95_ms`, `p99_ms`, `ops_per_sec`

```bash
# Benchmark with 50 iterations, JSON output
naeos benchmark --config config.yaml --input-file spec.yaml --iterations 50 --output json
```

---

### `naeos config`

Manage NAEOS configuration files.

```bash
naeos config show --config config.yaml
naeos config validate --config config.yaml
naeos config set --config config.yaml --key pipeline.mode --value production
```

**Subcommands:**

| Subcommand | Description |
|---|---|
| `config show` | Display the resolved configuration |
| `config validate` | Validate configuration file |
| `config set` | Set a configuration value |

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--key` | `string` | `""` | For `set` | Dot-notation config key |
| `--value` | `string` | `""` | For `set` | Value to set |

```bash
# Show resolved config
naeos config show --config config.yaml

# Validate config syntax
naeos config validate --config config.yaml

# Set a value
naeos config set --config config.yaml --key pipeline.verbose --value true
```

---

### `naeos deploy`

Deploy generated artifacts to a target environment.

```bash
naeos deploy --config config.yaml --input spec.yaml --environment staging
naeos deploy --config config.yaml --input spec.yaml --environment production --dry-run
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--input` | `string` | `""` | One of `--input`/`--input-file` | Inline specification text |
| `--input-file` | `string` | `""` | One of `--input`/`--input-file` | Path to specification file |
| `--environment` | `string` | `""` | Yes | Target environment: `staging`, `production` |
| `--dry-run` | `bool` | `false` | No | Preview deployment without executing |

**Output keys:** `deployment_id`, `environment`, `status`, `resources`

```bash
# Dry-run to preview what would be deployed
naeos deploy --config config.yaml --input-file spec.yaml --environment production --dry-run

# Actual deployment
naeos deploy --config config.yaml --input-file spec.yaml --environment staging
```

---

### `naeos distributed`

Manage distributed pipeline execution across nodes.

```bash
naeos distributed status --config config.yaml
naeos distributed nodes --config config.yaml
naeos distributed schedule --config config.yaml --input spec.yaml --nodes 3
```

**Subcommands:**

| Subcommand | Description |
|---|---|
| `distributed status` | Show distributed execution status |
| `distributed nodes` | List connected worker nodes |
| `distributed schedule` | Schedule a pipeline across nodes |

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--input` | `string` | `""` | For `schedule` | Inline specification text |
| `--input-file` | `string` | `""` | For `schedule` | Path to specification file |
| `--nodes` | `int` | `1` | For `schedule` | Number of worker nodes to use |

```bash
# Check cluster status
naeos distributed status --config config.yaml

# List available nodes
naeos distributed nodes --config config.yaml

# Schedule across 3 nodes
naeos distributed schedule --config config.yaml --input-file spec.yaml --nodes 3
```

---

### `naeos events`

Manage and inspect the kernel event bus.

```bash
naeos events list --config config.yaml
naeos events stream --config config.yaml --topic build.*
naeos events history --config config.yaml --limit 50
```

**Subcommands:**

| Subcommand | Description |
|---|---|
| `events list` | List available event topics |
| `events stream` | Stream events in real time |
| `events history` | Show past events |

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--topic` | `string` | `""` | For `stream` | Topic filter (supports glob patterns) |
| `--limit` | `int` | `20` | For `history` | Maximum events to show |

```bash
# List all topics
naeos events list --config config.yaml

# Stream build events
naeos events stream --config config.yaml --topic build.*

# Show last 50 events
naeos events history --config config.yaml --limit 50
```

---

### `naeos health`

Check health of NAEOS services and dependencies.

```bash
naeos health --config config.yaml
naeos health --config config.yaml --output json
naeos health --config config.yaml --check database --check cache
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--check` | `string` | `""` | No | Specific check to run (repeatable) |
| `--output` | `string` | `text` | No | Output format: `text`, `json`, `yaml` |

**Output keys:** `status`, `version`, `uptime`, `checks`

```bash
# Full health check
naeos health --config config.yaml

# Check specific services
naeos health --config config.yaml --check database --check cache --output json
```

---

### `naeos history`

Show pipeline execution history.

```bash
naeos history --config config.yaml
naeos history --config config.yaml --limit 20
naeos history --config config.yaml --status failed
naeos history --config config.yaml --output json
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--limit` | `int` | `10` | No | Maximum records to show |
| `--status` | `string` | `""` | No | Filter by status: `succeeded`, `failed`, `running` |
| `--output` | `string` | `text` | No | Output format: `text`, `json`, `yaml` |

**Output keys:** `runs`, `total`, `filtered`

```bash
# Last 10 runs
naeos history --config config.yaml

# Failed runs only
naeos history --config config.yaml --status failed --limit 20

# JSON output for scripting
naeos history --config config.yaml --output json
```

---

### `naeos import`

Import specifications or artifacts from external sources.

```bash
naeos import --source https://example.com/spec.yaml --output ./specs
naeos import --source ./archive.tar.gz --output ./imported
naeos import --source openapi --url https://petstore.swagger.io/v2/swagger.json --output ./specs
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--source` | `string` | `""` | Yes | Source type: `url`, `file`, `openapi`, `swagger` |
| `--url` | `string` | `""` | For `openapi`/`swagger` | URL to fetch from |
| `--output` | `string` | `""` | Yes | Output directory |

**Output keys:** `source`, `imported`, `files`

```bash
# Import from URL
naeos import --source url --url https://example.com/spec.yaml --output ./specs

# Import OpenAPI spec and convert to NAEOS format
naeos import --source openapi --url https://petstore.swagger.io/v2/swagger.json --output ./specs

# Import from local archive
naeos import --source file --url ./archive.tar.gz --output ./imported
```

---

### `naeos kernel`

Inspect the NAEOS kernel and service registry. Contains 5 subcommands.

#### Global Kernel Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--config` | `string` | `""` | Path to config file |
| `--output` | `string` | `text` | Output format: `text`, `json`, `yaml` |
| `--topic` | `string` | `""` | Kernel event topic |
| `--payload` | `string` | `""` | Event payload to publish |

#### `naeos kernel services`

List registered kernel services.

```bash
naeos kernel services --config config.yaml
```

#### `naeos kernel metrics`

Show kernel telemetry metrics.

```bash
naeos kernel metrics --config config.yaml
# Output: events=<N> last_event=<name>
```

#### `naeos kernel events`

List active kernel event topics.

```bash
naeos kernel events --config config.yaml
```

#### `naeos kernel publish`

Publish an event to the kernel event bus.

```bash
naeos kernel publish --config config.yaml --topic build.start --payload '{"module":"auth"}'
```

#### `naeos kernel subscribe`

Subscribe to a kernel event topic.

```bash
naeos kernel subscribe --config config.yaml --topic build.start
naeos kernel subscribe --config config.yaml --topic build.start --payload '{"test":"data"}'
```

---

## Cloud Commands

### `naeos cloud plan`

Generate a cloud deployment plan from a specification.

```bash
naeos cloud plan --config config.yaml --input spec.yaml --provider aws --region us-east-1
naeos cloud plan --config config.yaml --input spec.yaml --provider gcp --region us-central1 --output json
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--input` | `string` | `""` | One of `--input`/`--input-file` | Inline specification text |
| `--input-file` | `string` | `""` | One of `--input`/`--input-file` | Path to specification file |
| `--provider` | `string` | `""` | Yes | Cloud provider: `aws`, `gcp`, `azure`, `digitalocean` |
| `--region` | `string` | `""` | No | Target region |
| `--output` | `string` | `text` | No | Output format: `text`, `json`, `yaml` |

**Output keys:** `plan_id`, `provider`, `region`, `resources`, `estimated_monthly_cost`

```bash
# Plan for AWS
naeos cloud plan --config config.yaml --input-file spec.yaml --provider aws --region us-east-1

# Plan for GCP with JSON output
naeos cloud plan --config config.yaml --input-file spec.yaml --provider gcp --region us-central1 --output json
```

---

### `naeos cloud status`

Check the status of a cloud deployment.

```bash
naeos cloud status --deployment-id deploy-abc123
naeos cloud status --deployment-id deploy-abc123 --output json
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--deployment-id` | `string` | `""` | Yes | Deployment identifier |
| `--output` | `string` | `text` | No | Output format: `text`, `json`, `yaml` |

**Output keys:** `deployment_id`, `provider`, `status`, `resources`, `last_updated`

```bash
# Check deployment status
naeos cloud status --deployment-id deploy-abc123

# JSON for scripting
naeos cloud status --deployment-id deploy-abc123 --output json
```

---

## AI Commands

### `naeos ai enrich`

Enrich a specification with AI-generated suggestions and best practices.

```bash
naeos ai enrich --config config.yaml --input spec.yaml
naeos ai enrich --config config.yaml --input spec.yaml --focus security
naeos ai enrich --config config.yaml --input spec.yaml --output json --output-file enriched.json
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--config` | `string` | `""` | Yes | Path to config file |
| `--input` | `string` | `""` | One of `--input`/`--input-file` | Inline specification text |
| `--input-file` | `string` | `""` | One of `--input`/`--input-file` | Path to specification file |
| `--focus` | `string` | `""` | No | Enrichment focus: `security`, `performance`, `testing`, `all` |
| `--output` | `string` | `text` | No | Output format: `text`, `json`, `yaml` |
| `--output-file` | `string` | `""` | No | Write output to file |

**Output keys:** `spec`, `suggestions`, `focus`, `enriched_at`

```bash
# General enrichment
naeos ai enrich --config config.yaml --input-file spec.yaml

# Security-focused
naeos ai enrich --config config.yaml --input-file spec.yaml --focus security

# Save enriched spec
naeos ai enrich --config config.yaml --input-file spec.yaml --output yaml --output-file enriched-spec.yaml
```

---

## Plugin Commands

### `naeos plugin test`

Test a plugin in isolation before installing.

```bash
naeos plugin test --source ./my-plugin
naeos plugin test --source ./my-plugin --args '{"key":"value"}'
naeos plugin test --source https://naeos-foundation.github.io/naeos/plugins/registry.json --output json
```

| Flag | Type | Default | Required | Description |
|---|---|---|---|---|
| `--source` | `string` | `""` | Yes | Plugin path or registry URL |
| `--args` | `string` | `""` | No | JSON arguments to pass to the plugin |
| `--output` | `string` | `text` | No | Output format: `text`, `json`, `yaml` |

**Output keys:** `plugin`, `version`, `kind`, `passed`, `tests`, `duration_ms`

```bash
# Test local plugin
naeos plugin test --source ./my-plugin

# Test with arguments
naeos plugin test --source ./my-plugin --args '{"input":"test-spec.yaml"}' --output json

# Test remote plugin
naeos plugin test --source https://naeos-foundation.github.io/naeos/plugins/registry.json
```

---

## Common Workflows

### Quick Start

```bash
# 1. Generate config
naeos init -o config.yaml

# 2. Create spec.yaml (or use scaffold)
naeos scaffold --name my-app

# 3. Validate spec
naeos validate --config config.yaml --input-file my-app/spec.yaml

# 4. Run pipeline
naeos run --config config.yaml --input-file my-app/spec.yaml

# 5. Export artifacts
naeos export --config config.yaml --input-file my-app/spec.yaml
```

### Diagnostics

```bash
naeos doctor --config config.yaml        # Check config health
naeos repair --config config.yaml        # Fix output directory
naeos health --config config.yaml        # Service health check
naeos config validate --config config.yaml  # Config syntax check
```

### Kernel Inspection

```bash
naeos kernel services --config config.yaml    # List services
naeos kernel events --config config.yaml      # List topics
naeos kernel metrics --config config.yaml     # View metrics
```

### Multi-Language SDK Generation

```bash
# 1. Validate with language targets
naeos validate --config config.yaml --input-file spec.yaml --language go --language typescript

# 2. Run pipeline with multi-language output
naeos run --config config.yaml --input-file spec.yaml --language go --language typescript --language python

# 3. Export artifacts
naeos export --config config.yaml --input-file spec.yaml --language go --language typescript --language python

# 4. Preview what would be generated
naeos preview --config config.yaml --input-file spec.yaml --language go
```

### Cloud Deployment

```bash
# 1. Generate deployment plan
naeos cloud plan --config config.yaml --input-file spec.yaml --provider aws --region us-east-1

# 2. Deploy to staging
naeos deploy --config config.yaml --input-file spec.yaml --environment staging

# 3. Check deployment status
naeos cloud status --deployment-id deploy-abc123

# 4. Deploy to production
naeos deploy --config config.yaml --input-file spec.yaml --environment production
```

### Plugin Workflow

```bash
# 1. Test plugin before install
naeos plugin test --source ./my-plugin

# 2. Install plugin
curl -X POST http://localhost:8080/api/v1/plugins \
  -d '{"name":"my-plugin","source":"./my-plugin"}'

# 3. Verify installation
naeos health --config config.yaml --check plugins
```

---

## Output Formats

All commands that accept `--output` support:

| Format | Description |
|---|---|
| `text` | Human-readable key=value pairs (default) |
| `json` | Pretty-printed JSON |
| `yaml` | YAML output |
