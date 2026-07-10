# NAEOS-KER-002: Kernel Implementation & Setup Guide

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Panduan lengkap untuk mengimplementasikan dan setup NAEOS Kernel. Mencakup:
- Kernel architecture review
- Initialization & bootstrap process
- Component registration & discovery
- Lifecycle management
- Configuration & customization
- Event bus setup
- Plugin integration

---

## 2. Kernel Architecture Deep Dive

### 2.1 Three Kernel Tiers

```
┌─────────────────────────────────────┐
│    NAEOS Kernel (Core Runtime)      │
├─────────────────────────────────────┤
│                                     │
│  ┌─────────────────────────────────┐│
│  │  Knowledge Kernel               ││
│  │  - Artifact loading             ││
│  │  - Knowledge graph building     ││
│  │  - Metadata management          ││
│  │  - Query API                    ││
│  └─────────────────────────────────┘│
│                                     │
│  ┌─────────────────────────────────┐│
│  │  Policy Kernel                  ││
│  │  - Constitution loading         ││
│  │  - Policy compilation           ││
│  │  - Policy evaluation            ││
│  │  - Governance enforcement       ││
│  └─────────────────────────────────┘│
│                                     │
│  ┌─────────────────────────────────┐│
│  │  Runtime Kernel                 ││
│  │  - Component lifecycle          ││
│  │  - Dependency injection         ││
│  │  - Event orchestration          ││
│  │  - Telemetry collection         ││
│  └─────────────────────────────────┘│
│                                     │
│  ┌─────────────────────────────────┐│
│  │  Event Bus                      ││
│  │  - Event pub/sub                ││
│  │  - Event routing                ││
│  │  - Event sequencing             ││
│  └─────────────────────────────────┘│
│                                     │
└─────────────────────────────────────┘
         ↓        ↓        ↓
    Compiler  Validator  AI Runtime
```

### 2.2 Component Model

```go
// Core kernel interface
type Kernel interface {
    // Knowledge Kernel
    Load(artifacts []Artifact) error
    Resolve(id string) (interface{}, error)
    Search(query Query) ([]interface{}, error)
    
    // Policy Kernel
    EnforcePolicy(policy Policy) error
    ValidateAgainstPolicy(obj interface{}) error
    
    // Runtime Kernel
    Register(component Component) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    
    // Event Bus
    Publish(event Event) error
    Subscribe(topic string, handler EventHandler) error
}
```

---

## 3. Initialization Process

### 3.1 Bootstrap Sequence

```
Step 1: Create Kernel instance
        ↓
Step 2: Load configuration
        ↓
Step 3: Initialize Knowledge Kernel
        ├─ Load artifact registry
        ├─ Build knowledge graph
        └─ Initialize query engine
        ↓
Step 4: Initialize Policy Kernel
        ├─ Load constitution
        ├─ Compile policies
        └─ Setup policy evaluator
        ↓
Step 5: Initialize Runtime Kernel
        ├─ Setup service registry
        ├─ Setup dependency resolver
        └─ Setup lifecycle manager
        ↓
Step 6: Initialize Event Bus
        ├─ Setup pub/sub system
        └─ Register core handlers
        ↓
Step 7: Load and register plugins
        ├─ Discover plugins
        ├─ Initialize plugins
        └─ Subscribe to events
        ↓
Step 8: Start kernel
        └─ Emit startup events
```

### 3.2 Go Implementation

```go
package kernel

import (
    "context"
    "fmt"
    "github.com/naeos/kernel/pkg/knowledge"
    "github.com/naeos/kernel/pkg/policy"
    "github.com/naeos/kernel/pkg/runtime"
    "github.com/naeos/kernel/pkg/events"
)

type KernelImpl struct {
    config      KernelConfig
    knowledge   knowledge.Kernel
    policy      policy.Kernel
    runtime     runtime.Kernel
    eventBus    events.EventBus
    plugins     []Plugin
    logger      Logger
}

// New creates and initializes a kernel
func New(config KernelConfig) (*KernelImpl, error) {
    k := &KernelImpl{
        config: config,
        logger: config.Logger,
    }
    
    // Step 1-3: Initialize Knowledge Kernel
    if err := k.initKnowledge(); err != nil {
        return nil, fmt.Errorf("failed to init knowledge kernel: %w", err)
    }
    
    // Step 4: Initialize Policy Kernel
    if err := k.initPolicy(); err != nil {
        return nil, fmt.Errorf("failed to init policy kernel: %w", err)
    }
    
    // Step 5: Initialize Runtime Kernel
    if err := k.initRuntime(); err != nil {
        return nil, fmt.Errorf("failed to init runtime kernel: %w", err)
    }
    
    // Step 6: Initialize Event Bus
    if err := k.initEventBus(); err != nil {
        return nil, fmt.Errorf("failed to init event bus: %w", err)
    }
    
    // Step 7: Load plugins
    if err := k.loadPlugins(); err != nil {
        return nil, fmt.Errorf("failed to load plugins: %w", err)
    }
    
    return k, nil
}

func (k *KernelImpl) initKnowledge() error {
    k.logger.Info("Initializing Knowledge Kernel")
    
    kk, err := knowledge.New(k.config.Knowledge)
    if err != nil {
        return err
    }
    
    // Load artifacts
    artifacts, err := loadArtifacts(k.config.ArtifactPath)
    if err != nil {
        return err
    }
    
    if err := kk.Load(artifacts); err != nil {
        return err
    }
    
    k.knowledge = kk
    return nil
}

func (k *KernelImpl) initPolicy() error {
    k.logger.Info("Initializing Policy Kernel")
    
    pk, err := policy.New(k.config.Policy, k.knowledge)
    if err != nil {
        return err
    }
    
    // Load constitution
    constitution, err := loadConstitution(k.config.ConstitutionPath)
    if err != nil {
        return err
    }
    
    if err := pk.Load(constitution); err != nil {
        return err
    }
    
    k.policy = pk
    return nil
}

func (k *KernelImpl) initRuntime() error {
    k.logger.Info("Initializing Runtime Kernel")
    
    rk, err := runtime.New(k.config.Runtime)
    if err != nil {
        return err
    }
    
    k.runtime = rk
    return nil
}

func (k *KernelImpl) initEventBus() error {
    k.logger.Info("Initializing Event Bus")
    
    eb, err := events.New(k.config.EventBus)
    if err != nil {
        return err
    }
    
    k.eventBus = eb
    return nil
}

func (k *KernelImpl) loadPlugins() error {
    k.logger.Info("Loading plugins")
    
    plugins, err := discoverPlugins(k.config.PluginPath)
    if err != nil {
        return err
    }
    
    for _, p := range plugins {
        if err := p.Initialize(k); err != nil {
            k.logger.Warn("Failed to initialize plugin %s: %v", p.Name(), err)
            continue
        }
        k.plugins = append(k.plugins, p)
    }
    
    return nil
}

// Start starts the kernel
func (k *KernelImpl) Start(ctx context.Context) error {
    k.logger.Info("Starting NAEOS Kernel")
    
    // Start runtime kernel
    if err := k.runtime.Start(ctx); err != nil {
        return err
    }
    
    // Publish startup event
    k.eventBus.Publish(Event{
        Type: "kernel.startup",
        Timestamp: time.Now(),
    })
    
    return nil
}

// Stop stops the kernel gracefully
func (k *KernelImpl) Stop(ctx context.Context) error {
    k.logger.Info("Stopping NAEOS Kernel")
    
    // Publish shutdown event
    k.eventBus.Publish(Event{
        Type: "kernel.shutdown",
        Timestamp: time.Now(),
    })
    
    // Stop runtime kernel
    return k.runtime.Stop(ctx)
}
```

---

## 4. Configuration

### 4.1 YAML Configuration

```yaml
# kernel-config.yaml

kernel:
  name: "naeos-core"
  version: "1.0.0"
  
  knowledge:
    artifact_path: "./artifacts"
    cache_enabled: true
    cache_size: "1GB"
    index_type: "elasticsearch"
    
  policy:
    constitution_path: "./constitution"
    policy_path: "./policies"
    enforcement_mode: "blocking"
    
  runtime:
    max_components: 1000
    shutdown_timeout: "30s"
    health_check_interval: "10s"
    
  event_bus:
    type: "kafka"
    brokers:
      - "localhost:9092"
    topics:
      - "kernel.events"
    buffer_size: 10000
    
  plugins:
    enabled: true
    plugin_path: "./plugins"
    auto_load: true
    
  telemetry:
    enabled: true
    metrics_enabled: true
    tracing_enabled: true
    logging_level: "info"
```

### 4.2 Go Configuration

```go
type KernelConfig struct {
    Name     string
    Version  string
    
    Knowledge KnowledgeConfig
    Policy    PolicyConfig
    Runtime   RuntimeConfig
    EventBus  EventBusConfig
    Plugins   PluginsConfig
    Telemetry TelemetryConfig
    Logger    Logger
}

type KnowledgeConfig struct {
    ArtifactPath string
    CacheEnabled bool
    CacheSize    string
    IndexType    string
}

type PolicyConfig struct {
    ConstitutionPath string
    PolicyPath       string
    EnforcementMode  string // "blocking", "reporting"
}

type RuntimeConfig struct {
    MaxComponents       int
    ShutdownTimeout     time.Duration
    HealthCheckInterval time.Duration
}

type EventBusConfig struct {
    Type       string
    Brokers    []string
    Topics     []string
    BufferSize int
}
```

---

## 5. Component Registration

### 5.1 Service Registry

```go
// Register a component
func (k *KernelImpl) Register(component Component) error {
    // Validate component
    if err := validateComponent(component); err != nil {
        return err
    }
    
    // Register in runtime kernel
    if err := k.runtime.Register(component); err != nil {
        return err
    }
    
    // Subscribe to relevant events
    for _, topic := range component.Topics() {
        k.eventBus.Subscribe(topic, component.Handle)
    }
    
    return nil
}

// Component interface
type Component interface {
    ID() string
    Name() string
    Version() string
    Dependencies() []string
    Topics() []string
    Initialize(ctx context.Context) error
    Handle(event Event) error
    Shutdown(ctx context.Context) error
}
```

### 5.2 Service Discovery

```go
// Resolve a component by ID
func (k *KernelImpl) Resolve(id string) (Component, error) {
    return k.runtime.Resolve(id)
}

// Resolve with dependencies
func (k *KernelImpl) ResolveWithDeps(id string) (Component, []Component, error) {
    component, err := k.runtime.Resolve(id)
    if err != nil {
        return nil, nil, err
    }
    
    var deps []Component
    for _, depID := range component.Dependencies() {
        dep, err := k.runtime.Resolve(depID)
        if err != nil {
            return nil, nil, err
        }
        deps = append(deps, dep)
    }
    
    return component, deps, nil
}
```

---

## 6. Lifecycle Management

### 6.1 Component Lifecycle

```
New Component
    ↓
Registered
    ↓
Initialized
    ↓
Running
    ├─ Active (handling events)
    └─ Idle (waiting for events)
    ↓
Stopping
    ↓
Stopped
```

### 6.2 Lifecycle Hooks

```go
type ComponentLifecycle interface {
    OnRegister() error
    OnInitialize(ctx context.Context) error
    OnStart(ctx context.Context) error
    OnStop(ctx context.Context) error
}

// Implementation
func (c *MyComponent) OnRegister() error {
    log.Println("Component registered")
    return nil
}

func (c *MyComponent) OnInitialize(ctx context.Context) error {
    log.Println("Component initializing")
    // Setup resources
    return nil
}

func (c *MyComponent) OnStart(ctx context.Context) error {
    log.Println("Component starting")
    // Start services
    return nil
}

func (c *MyComponent) OnStop(ctx context.Context) error {
    log.Println("Component stopping")
    // Cleanup resources
    return nil
}
```

---

## 7. Event Bus & Publishing

### 7.1 Event Publishing

```go
// Publish an event
err := k.eventBus.Publish(Event{
    Type: "component.deployed",
    Source: "deployment-service",
    Payload: map[string]interface{}{
        "component_id": "compiler-v2",
        "version": "2.1.0",
    },
    Timestamp: time.Now(),
})

// Async publishing
go func() {
    if err := k.eventBus.PublishAsync(event); err != nil {
        log.Error("Failed to publish event", err)
    }
}()
```

### 7.2 Event Subscription

```go
// Subscribe to events
k.eventBus.Subscribe("component.deployed", func(event Event) error {
    fmt.Printf("Component deployed: %s\n", event.Payload["component_id"])
    return nil
})

// Subscribe with handler
type MyHandler struct {
    kernel Kernel
}

func (h *MyHandler) Handle(event Event) error {
    // Process event
    return nil
}

k.eventBus.Subscribe("policy.violated", h.Handle)
```

---

## 8. Plugin System

### 8.1 Plugin Interface

```go
type Plugin interface {
    Name() string
    Version() string
    Initialize(kernel Kernel) error
    Topics() []string
    Handle(event Event) error
    Shutdown() error
}

// Example plugin
type CompilerPlugin struct {
    kernel Kernel
}

func (p *CompilerPlugin) Name() string {
    return "compiler-plugin"
}

func (p *CompilerPlugin) Version() string {
    return "1.0.0"
}

func (p *CompilerPlugin) Initialize(kernel Kernel) error {
    p.kernel = kernel
    
    // Register compiler component
    compiler := NewCompiler()
    return kernel.Register(compiler)
}

func (p *CompilerPlugin) Topics() []string {
    return []string{"specification.created", "specification.updated"}
}

func (p *CompilerPlugin) Handle(event Event) error {
    // Handle specification changes
    spec := event.Payload["specification"]
    // Compile specification
    return nil
}
```

### 8.2 Plugin Discovery

```bash
# Plugins should be in plugins/ directory
plugins/
├── compiler/
│   ├── plugin.so
│   └── config.yaml
├── validator/
│   ├── plugin.so
│   └── config.yaml
└── ai-runtime/
    ├── plugin.so
    └── config.yaml
```

---

## 9. Health & Diagnostics

### 9.1 Health Checks

```go
type HealthStatus struct {
    Status    string
    Knowledge HealthComponent
    Policy    HealthComponent
    Runtime   HealthComponent
    EventBus  HealthComponent
}

func (k *KernelImpl) GetHealth(ctx context.Context) (HealthStatus, error) {
    status := HealthStatus{
        Status: "healthy",
    }
    
    // Check knowledge kernel
    status.Knowledge = k.knowledge.GetHealth(ctx)
    if status.Knowledge.Status != "healthy" {
        status.Status = "degraded"
    }
    
    // Check policy kernel
    status.Policy = k.policy.GetHealth(ctx)
    if status.Policy.Status != "healthy" {
        status.Status = "degraded"
    }
    
    // Check runtime kernel
    status.Runtime = k.runtime.GetHealth(ctx)
    if status.Runtime.Status != "healthy" {
        status.Status = "unhealthy"
    }
    
    return status, nil
}
```

### 9.2 Telemetry

```go
type KernelMetrics struct {
    ComponentsRegistered  int64
    EventsPublished       int64
    EventsProcessed       int64
    AverageBusLatency     time.Duration
    MemoryUsage           uint64
    CPUUsage              float64
}

func (k *KernelImpl) GetMetrics() KernelMetrics {
    return KernelMetrics{
        ComponentsRegistered: k.runtime.ComponentCount(),
        EventsPublished: k.eventBus.EventsPublished(),
        EventsProcessed: k.eventBus.EventsProcessed(),
        AverageBusLatency: k.eventBus.AverageLatency(),
    }
}
```

---

## 10. References

- [NAEOS-KER-001.md](NAEOS-KER-001.md) - Kernel Architecture
- [docs/NES-002-Kernel.md](../docs/NES-002-Kernel.md) - Kernel Specification
- [NAEOS-POL-001.md](../policy/NAEOS-POL-001.md) - Policy System
