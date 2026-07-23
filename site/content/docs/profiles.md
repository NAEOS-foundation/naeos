---
title: Industry Profiles
description: Pre-configured profiles for SaaS, AI Agent, FinTech, Healthcare, and Government projects.
weight: 14
---

NAEOS provides industry-specific profiles that pre-configure modules, services, architecture patterns, and security settings for common project types.

## Available Profiles

| Profile | Industry | Description |
|---------|----------|-------------|
| **SaaS** | Software-as-a-Service | Multi-tenant architecture, billing, analytics, user management |
| **AI Agent** | Artificial Intelligence | LLM integration, tool calling, memory systems, agent orchestration |
| **FinTech** | Financial Technology | Compliance, audit trail, encryption, transaction processing |
| **Healthcare** | Health Technology | HIPAA compliance, audit logging, PHI protection, clinical workflows |
| **Government** | Public Sector | Security clearance, audit, accessibility, inter-agency integration |

## CLI Usage

```bash
# List all available profiles
naeos profile list

# Show profile details
naeos profile show saas

# Search profiles by keyword
naeos profile search "multi-tenant"

# Apply a profile to a specification
naeos profile apply saas --output spec.yaml
```

## Go API Usage

```go
import "github.com/NAEOS-foundation/naeos/internal/profiles"

registry := profiles.NewRegistry()

// List all profiles
allProfiles := registry.List()
for _, p := range allProfiles {
    fmt.Printf("%s: %s\n", p.Name, p.Description)
}

// Search for a profile
results := registry.Search("fintech")

// Get a specific profile
saas := registry.Get("saas")

// Apply profile to generate a spec YAML
specYAML := saas.ToSpecYAML()
```

## Profile Structure

Each profile defines a complete project template:

```go
type Profile struct {
    Name         string
    Description  string
    Industry     string
    Modules      []Module
    Services     []Service
    Architecture *Architecture
    Security     *Security
    Deployment   *Deployment
    Testing      *Testing
}
```

### Example: SaaS Profile

The SaaS profile includes:

- **Modules**: auth, billing, analytics, notification, user-management
- **Services**: api-gateway (HTTP), worker (async), admin (HTTP)
- **Architecture**: hexagonal with event-driven communication
- **Security**: OAuth2, rate limiting, input validation
- **Deployment**: Kubernetes with horizontal autoscaling
- **Testing**: unit, integration, e2e

## Converting to Spec YAML

```go
profile := registry.Get("saas")
specYAML := profile.ToSpecYAML()
// Returns a YAML string ready to use as a specification
```

## Custom Profiles

Create and register your own profiles:

```go
registry := profiles.NewRegistry()

customProfile := &profiles.Profile{
    Name:        "manufacturing",
    Description: "Profile for manufacturing and IoT systems",
    Industry:    "manufacturing",
    Modules: []profiles.Module{
        {Name: "inventory", Path: "./inventory"},
        {Name: "supply-chain", Path: "./supply-chain"},
        {Name: "iot-gateway", Path: "./iot-gateway"},
    },
    Services: []profiles.Service{
        {Name: "api", Kind: "http", Port: 8080},
        {Name: "mqtt-broker", Kind: "worker", Port: 1883},
    },
}

registry.Register(customProfile)
```

## Marketplace

Browse community profiles in the [Marketplace](/docs/plugin-sdk/):

```bash
naeos marketplace search profile
naeos marketplace install community-profile-name
```

See also: [Ecosystem](/ecosystem/), [Plugin SDK](/docs/plugin-sdk/)
