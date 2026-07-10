# NAEOS-KER-004: Kernel Best Practices

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Panduan best practices untuk mengembangkan dan mengintegrasikan komponen dengan NAEOS Kernel. Mencakup:
- Component design principles
- Event design patterns
- Error handling strategies
- Lifecycle management
- Plugin development guidelines
- Configuration management

---

## 2. Component Design Principles

### 2.1 Single Responsibility Principle

✅ **GOOD**: Component dengan single concern

```go
// Good: Compiler focused only on compilation
type Compiler struct {
    parser     Parser
    normalizer Normalizer
    generator  Generator
}

func (c *Compiler) Compile(spec Specification) (Output, error) {
    ast := c.parser.Parse(spec)
    normalized := c.normalizer.Normalize(ast)
    return c.generator.Generate(normalized)
}
```

❌ **BAD**: Component doing too much

```go
// Bad: Compiler also doing policy evaluation and deployment
type SuperCompiler struct {
    // Compilation logic
    parser     Parser
    
    // Policy logic (shouldn't be here!)
    policyEngine PolicyEngine
    
    // Deployment logic (shouldn't be here!)
    deployer   Deployer
}
```

### 2.2 Dependency Injection

✅ **GOOD**: Dependencies injected via constructor

```go
func NewCompiler(
    parser Parser,
    normalizer Normalizer,
    generator Generator,
) *Compiler {
    return &Compiler{
        parser: parser,
        normalizer: normalizer,
        generator: generator,
    }
}
```

❌ **BAD**: Hard-coded dependencies

```go
type Compiler struct{}

func (c *Compiler) Compile(spec Specification) Output {
    parser := NewParser()  // Hard-coded!
    normalizer := NewNormalizer()  // Hard-coded!
    // ...
}
```

### 2.3 Interface Segregation

✅ **GOOD**: Focused interfaces

```go
type Compiler interface {
    Compile(spec Specification) (Output, error)
}

type Validator interface {
    Validate(spec Specification) (ValidationResult, error)
}
```

❌ **BAD**: Large interface

```go
type MegaComponent interface {
    Compile(spec Specification) Output
    Validate(spec Specification) ValidationResult
    Deploy(spec Specification) DeploymentResult
    Monitor(spec Specification) MonitoringData
    // Too many responsibilities!
}
```

---

## 3. Event Design Patterns

### 3.1 Event Naming Conventions

```
Pattern: {domain}.{entity}.{action}

Examples:
- specification.created
- specification.updated
- specification.deleted
- compilation.started
- compilation.completed
- compilation.failed
- policy.violated
- deployment.requested
- deployment.completed
```

### 3.2 Event Payload Structure

✅ **GOOD**: Structured, well-documented payload

```go
type SpecificationCreatedEvent struct {
    ID            string                 // Unique event ID
    Type          string                 // Event type
    Source        string                 // Component that published
    Timestamp     time.Time              // When event occurred
    CorrelationID string                 // For tracing
    
    Payload struct {
        SpecID   string
        Name     string
        Version  string
        Content  []byte
        Author   string
    }
}
```

❌ **BAD**: Unstructured payload

```go
type Event struct {
    Type string
    Data interface{}  // Too generic!
}
```

### 3.3 Event Subscription Patterns

✅ **GOOD**: Explicit subscription with error handling

```go
k.Subscribe("specification.created", func(e Event) error {
    spec := e.Payload["specification"]
    
    if err := processSpec(spec); err != nil {
        return fmt.Errorf("failed to process spec: %w", err)
    }
    
    k.Publish(Event{
        Type: "specification.processed",
        Payload: e.Payload,
    })
    
    return nil
})
```

❌ **BAD**: Implicit subscriptions without error handling

```go
k.Subscribe("*", func(e Event) error {
    // Process everything without context
    // No error handling
    return nil
})
```

---

## 4. Error Handling

### 4.1 Error Wrapping

✅ **GOOD**: Preserve error chain

```go
func (c *Compiler) Compile(spec Specification) (Output, error) {
    ast, err := c.parser.Parse(spec)
    if err != nil {
        return nil, fmt.Errorf("failed to parse specification: %w", err)
    }
    
    normalized, err := c.normalizer.Normalize(ast)
    if err != nil {
        return nil, fmt.Errorf("failed to normalize AST: %w", err)
    }
    
    return c.generator.Generate(normalized)
}
```

❌ **BAD**: Losing error context

```go
func (c *Compiler) Compile(spec Specification) (Output, error) {
    ast, _ := c.parser.Parse(spec)  // Ignoring error!
    normalized, _ := c.normalizer.Normalize(ast)
    return c.generator.Generate(normalized)
}
```

### 4.2 Recovery & Resilience

✅ **GOOD**: Graceful degradation

```go
func (c *Compiler) CompileWithFallback(spec Specification) Output {
    output, err := c.Compile(spec)
    if err != nil {
        // Fall back to previous version
        return c.GetPreviousOutput(spec.ID)
    }
    return output
}
```

❌ **BAD**: Silent failures

```go
func (c *Compiler) Compile(spec Specification) Output {
    // No error handling, might return nil
    return c.doCompile(spec)
}
```

---

## 5. Component Lifecycle Best Practices

### 5.1 Proper Resource Management

✅ **GOOD**: Clean lifecycle hooks

```go
type MyComponent struct {
    db         *sql.DB
    cache      *Cache
    logger     Logger
}

func (c *MyComponent) OnInitialize(ctx context.Context) error {
    // Initialize database connection
    db, err := sql.Open("postgres", c.dsn)
    if err != nil {
        return err
    }
    c.db = db
    
    // Initialize cache
    c.cache = NewCache()
    
    return nil
}

func (c *MyComponent) OnStop(ctx context.Context) error {
    // Close database connection
    if c.db != nil {
        c.db.Close()
    }
    
    // Clear cache
    c.cache.Clear()
    
    return nil
}
```

❌ **BAD**: Resource leaks

```go
type MyComponent struct {
    db *sql.DB
}

func (c *MyComponent) OnInitialize(ctx context.Context) error {
    var err error
    c.db, _ = sql.Open("postgres", dsn)  // Not checking error
    return nil
}

// OnStop never called, resources leak!
```

### 5.2 Graceful Shutdown

✅ **GOOD**: Context-aware shutdown

```go
func (c *MyComponent) OnStop(ctx context.Context) error {
    // Stop accepting new work
    c.stopAcceptingRequests()
    
    // Wait for in-flight requests
    done := make(chan struct{})
    go func() {
        c.waitForInflightRequests()
        close(done)
    }()
    
    // Wait or timeout
    select {
    case <-done:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

---

## 6. Plugin Development Guidelines

### 6.1 Plugin Interface Implementation

✅ **GOOD**: Complete plugin implementation

```go
type MyPlugin struct {
    kernel   kernel.Kernel
    handlers map[string]Handler
}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Initialize(k kernel.Kernel) error {
    p.kernel = k
    p.handlers = make(map[string]Handler)
    
    // Register event handlers
    for topic, handler := range p.getHandlers() {
        k.Subscribe(topic, handler)
    }
    
    return nil
}

func (p *MyPlugin) Topics() []string {
    return []string{
        "specification.created",
        "specification.updated",
    }
}

func (p *MyPlugin) Handle(e kernel.Event) error {
    handler, ok := p.handlers[e.Type]
    if !ok {
        return nil
    }
    return handler(e)
}

func (p *MyPlugin) Shutdown() error {
    // Cleanup
    return nil
}
```

### 6.2 Plugin Configuration

✅ **GOOD**: Plugin with configuration

```yaml
# plugin-config.yaml

plugin:
  name: my-plugin
  version: 1.0.0
  enabled: true
  
  config:
    timeout: 30s
    retries: 3
    cache_enabled: true
    log_level: info
```

```go
type PluginConfig struct {
    Timeout      time.Duration
    Retries      int
    CacheEnabled bool
    LogLevel     string
}

func (p *MyPlugin) LoadConfig(path string) error {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return err
    }
    
    var cfg PluginConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return err
    }
    
    p.config = cfg
    return nil
}
```

---

## 7. Configuration Management

### 7.1 Environment-Specific Configuration

✅ **GOOD**: Environment-aware config

```yaml
# kernel-dev.yaml
kernel:
  environment: development
  log_level: debug
  event_bus:
    type: memory
  cache:
    enabled: false

# kernel-prod.yaml
kernel:
  environment: production
  log_level: warning
  event_bus:
    type: kafka
    brokers: [kafka1:9092, kafka2:9092]
  cache:
    enabled: true
    ttl: 3600
```

```go
config := loadConfig(os.Getenv("ENV"))
// Loads kernel-{ENV}.yaml
```

### 7.2 Configuration Validation

✅ **GOOD**: Validate configuration on load

```go
func (cfg *KernelConfig) Validate() error {
    if cfg.Name == "" {
        return errors.New("kernel name is required")
    }
    
    if cfg.EventBus.Type != "memory" && cfg.EventBus.Type != "kafka" {
        return errors.New("invalid event bus type")
    }
    
    return nil
}

config, err := loadConfig(path)
if err != nil {
    return err
}

if err := config.Validate(); err != nil {
    return err
}
```

---

## 8. Testing Strategies

### 8.1 Component Unit Tests

```go
func TestCompilerCompile(t *testing.T) {
    // Setup
    parser := NewMockParser()
    normalizer := NewMockNormalizer()
    generator := NewMockGenerator()
    
    compiler := NewCompiler(parser, normalizer, generator)
    
    // Test
    spec := &Specification{ID: "test-1"}
    output, err := compiler.Compile(spec)
    
    // Assert
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    
    if output == nil {
        t.Fatal("expected output, got nil")
    }
}
```

### 8.2 Event Bus Integration Tests

```go
func TestEventBusIntegration(t *testing.T) {
    // Setup
    k := kernel.New(testConfig)
    k.Start(context.Background())
    defer k.Stop(context.Background())
    
    received := make([]Event, 0)
    k.Subscribe("test.event", func(e Event) error {
        received = append(received, e)
        return nil
    })
    
    // Test
    k.Publish(Event{Type: "test.event"})
    
    // Allow async processing
    time.Sleep(100 * time.Millisecond)
    
    // Assert
    if len(received) != 1 {
        t.Fatalf("expected 1 event, got %d", len(received))
    }
}
```

---

## 9. Documentation Guidelines

### 9.1 Component Documentation Template

```go
// Package mycomponent provides X functionality for NAEOS.
//
// Example:
//  comp := mycomponent.New(config)
//  if err := kernel.Register(comp); err != nil {
//      panic(err)
//  }
package mycomponent

// Component is the main component type.
// It implements kernel.Component interface.
type Component struct {
    // ...
}

// New creates a new Component with the given config.
func New(cfg Config) *Component {
    return &Component{}
}

// Compile performs compilation.
// It returns an error if compilation fails.
func (c *Component) Compile(spec Specification) (Output, error) {
    // Implementation
}
```

### 9.2 API Documentation

```go
// Topics returns the list of event topics this component handles.
func (c *Component) Topics() []string {
    // Return list of event topics
}

// Handle processes an event.
// It returns an error if event handling fails.
func (c *Component) Handle(event Event) error {
    // Implementation
}
```

---

## 10. Security Best Practices

### 10.1 Input Validation

✅ **GOOD**: Validate all inputs

```go
func (c *Component) Process(input interface{}) error {
    if input == nil {
        return errors.New("input cannot be nil")
    }
    
    spec, ok := input.(*Specification)
    if !ok {
        return errors.New("invalid input type")
    }
    
    if spec.ID == "" {
        return errors.New("specification ID cannot be empty")
    }
    
    return c.process(spec)
}
```

### 10.2 Credential Management

✅ **GOOD**: Use environment variables or secrets

```go
// Load credentials from environment
dbPassword := os.Getenv("DB_PASSWORD")
if dbPassword == "" {
    return errors.New("DB_PASSWORD not set")
}

// Never log sensitive data
log.Printf("Connected to database")  // Good
// log.Printf("Using password: %s", dbPassword)  // Bad!
```

---

## 11. References

- [NAEOS-KER-001.md](NAEOS-KER-001.md) - Kernel Architecture
- [NAEOS-KER-002.md](NAEOS-KER-002.md) - Kernel Implementation
- [NAEOS-KER-003.md](NAEOS-KER-003.md) - Kernel Examples
- [NAEOS-PRO-004.md](../profile/NAEOS-PRO-004.md) - Profile Best Practices
- [NAEOS-POL-004.md](../policy/NAEOS-POL-004.md) - Policy Best Practices
