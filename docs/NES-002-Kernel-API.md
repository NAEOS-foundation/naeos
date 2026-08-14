# NAEOS Kernel API Reference

## Status
- Status: Stable
- Version: 1.0.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Overview

NAEOS Kernel adalah runtime inti yang mengorkestrasi seluruh komponen NAEOS. Document ini mendeskripsikan API aktual dari implementasi Go yang ada di `pkg/kernel/`.

---

## 2. Package Structure

```
pkg/kernel/
├── kernel.go       # Core kernel implementation
├── lifecycle.go    # Lifecycle interface
├── events.go       # Event bus interface
├── telemetry.go    # Telemetry and metrics
└── kernel_test.go  # Tests
```

---

## 3. Core Types

### 3.1 Kernel

```go
type Kernel struct {
    mutex       sync.RWMutex
    services    map[string]any
    subscribers map[string][]func(any)
    metrics     Metrics
    started     bool
}
```

The central orchestrator that manages services, events, and lifecycle.

### 3.2 Lifecycle Interface

```go
type Lifecycle interface {
    Initialize() error
    Start() error
    Stop() error
}
```

Any service implementing this interface will have its lifecycle managed by the kernel.

### 3.3 EventBus Interface

```go
type EventBus interface {
    Publish(topic string, payload any)
    Subscribe(topic string, handler func(any)) error
}
```

The pub/sub system for inter-component communication.

### 3.4 TelemetryEvent

```go
type TelemetryEvent struct {
    Name      string
    Timestamp int64
    Payload   map[string]any
}
```

### 3.5 Metrics

```go
type Metrics struct {
    Events    int
    LastEvent TelemetryEvent
}
```

---

## 4. Kernel Methods

### 4.1 NewKernel

```go
func NewKernel() *Kernel
```

Creates a new kernel instance.

**Returns:** A pointer to a new `Kernel` with empty service registry and subscriber map.

**Example:**
```go
k := kernel.NewKernel()
```

---

### 4.2 Register

```go
func (k *Kernel) Register(name string, service any) error
```

Registers a service with the kernel.

**Parameters:**
- `name` (string): Unique name for the service. Must not be empty.
- `service` (any): The service instance. Must not be nil.

**Returns:**
- `nil` on success
- Error if name is empty, service is nil, or name is already registered.

**Example:**
```go
err := k.Register("compiler", compiler.NewCompiler())
if err != nil {
    log.Fatal(err)
}
```

---

### 4.3 Resolve

```go
func (k *Kernel) Resolve(name string) (any, error)
```

Retrieves a registered service by name.

**Parameters:**
- `name` (string): The service name to resolve.

**Returns:**
- The service instance if found.
- Error if service not found.

**Example:**
```go
compiler, err := k.Resolve("compiler")
if err != nil {
    log.Fatal(err)
}
comp := compiler.(*Compiler)
```

---

### 4.4 RegisteredServices

```go
func (k *Kernel) RegisteredServices() []string
```

Returns a sorted list of all registered service names.

**Returns:** `[]string` of service names, sorted alphabetically.

**Example:**
```go
services := k.RegisteredServices()
// ["compiler", "knowledge", "policy", "validator"]
```

---

### 4.5 Start

```go
func (k *Kernel) Start() error
```

Starts the kernel and all registered services that implement the `Lifecycle` interface.

**Behavior:**
1. Calls `Initialize()` on each lifecycle service
2. Calls `Start()` on each lifecycle service
3. Sets kernel state to started

**Returns:** Error if kernel is already started, or if any service fails to initialize/start.

**Example:**
```go
if err := k.Start(); err != nil {
    log.Fatal(err)
}
```

---

### 4.6 Stop

```go
func (k *Kernel) Stop() error
```

Stops the kernel and all registered services that implement the `Lifecycle` interface.

**Behavior:**
1. Calls `Stop()` on each lifecycle service
2. Sets kernel state to stopped

**Returns:** Error if kernel is not running, or if any service fails to stop.

**Example:**
```go
defer k.Stop()
```

---

### 4.7 Subscribe

```go
func (k *Kernel) Subscribe(topic string, handler func(any)) error
```

Subscribes a handler to a topic.

**Parameters:**
- `topic` (string): The topic to subscribe to. Must not be empty.
- `handler` (func(any)): The callback function. Must not be nil.

**Returns:** Error if topic is empty or handler is nil.

**Example:**
```go
err := k.Subscribe("specification.created", func(payload any) {
    spec := payload.(*Specification)
    fmt.Printf("New specification: %s\n", spec.Name)
})
```

---

### 4.8 Publish

```go
func (k *Kernel) Publish(topic string, payload any)
```

Publishes an event to all subscribers of a topic.

**Parameters:**
- `topic` (string): The topic to publish to.
- `payload` (any): The event payload.

**Behavior:**
- Copies subscriber list to avoid locking during handler execution
- Calls each handler with the payload
- Handlers are called synchronously in order

**Example:**
```go
k.Publish("specification.created", &Specification{
    Name:    "my-api",
    Version: "1.0.0",
})
```

---

### 4.9 Topics

```go
func (k *Kernel) Topics() []string
```

Returns a sorted list of all topics with active subscribers.

**Returns:** `[]string` of topic names, sorted alphabetically.

**Example:**
```go
topics := k.Topics()
// ["compilation.completed", "specification.created", "validation.failed"]
```

---

### 4.10 EmitTelemetry

```go
func (k *Kernel) EmitTelemetry(event TelemetryEvent) error
```

Records a telemetry event.

**Parameters:**
- `event` (TelemetryEvent): The telemetry event. Name must not be empty.

**Returns:** Error if event name is empty.

**Example:**
```go
err := k.EmitTelemetry(TelemetryEvent{
    Name:      "compilation.started",
    Timestamp: time.Now().Unix(),
    Payload: map[string]any{
        "spec_id": "spec-001",
    },
})
```

---

### 4.11 Metrics

```go
func (k *Kernel) Metrics() Metrics
```

Returns the current kernel metrics.

**Returns:** A copy of the `Metrics` struct.

**Example:**
```go
m := k.Metrics()
fmt.Printf("Total events: %d\n", m.Events)
fmt.Printf("Last event: %s\n", m.LastEvent.Name)
```

---

## 5. Lifecycle Management

Services implementing the `Lifecycle` interface are automatically managed:

```
Register → Initialize → Start → Running → Stop
```

### Example Implementation

```go
type MyService struct {
    started bool
}

func (s *MyService) Initialize() error {
    fmt.Println("Initializing service...")
    return nil
}

func (s *MyService) Start() error {
    s.started = true
    fmt.Println("Service started")
    return nil
}

func (s *MyService) Stop() error {
    s.started = false
    fmt.Println("Service stopped")
    return nil
}

// Register and start
k := kernel.NewKernel()
k.Register("my-service", &MyService{})
k.Start()
defer k.Stop()
```

---

## 6. Event System

### 6.1 Topic-Based Communication

```go
// Publisher
k.Publish("specification.created", spec)

// Subscriber
k.Subscribe("specification.created", func(payload any) {
    spec := payload.(*Specification)
    processSpec(spec)
})
```

### 6.2 Event Flow

```
Component A                 Kernel                   Component B
    │                         │                          │
    │  Publish("topic", data) │                          │
    │────────────────────────>│                          │
    │                         │  Handler(payload)        │
    │                         │─────────────────────────>│
    │                         │                          │
```

### 6.3 Multiple Subscribers

```go
// Multiple handlers can subscribe to the same topic
k.Subscribe("specification.created", handler1)
k.Subscribe("specification.created", handler2)
// Both handlers will be called when the event is published
```

---

## 7. Thread Safety

The kernel uses `sync.RWMutex` for thread-safe operations:

- **Read operations** (Resolve, RegisteredServices, Topics, Metrics): Use `RLock`
- **Write operations** (Register, Start, Stop, Subscribe): Use `Lock`
- **Publish**: Uses `RLock` to copy handlers, then calls handlers outside lock

---

## 8. Error Handling

| Method | Error Conditions |
|--------|-----------------|
| `Register` | Empty name, nil service, duplicate name |
| `Resolve` | Service not found |
| `Start` | Already started, service init/start failure |
| `Stop` | Not running, service stop failure |
| `Subscribe` | Empty topic, nil handler |
| `EmitTelemetry` | Empty event name |

---

## 9. Usage Patterns

### 9.1 Minimal Setup

```go
k := kernel.NewKernel()
k.Register("service1", &Service1{})
k.Start()
defer k.Stop()
```

### 9.2 Event-Driven Architecture

```go
k := kernel.NewKernel()

// Register services
k.Register("compiler", &Compiler{})
k.Register("validator", &Validator{})

// Setup event flow
k.Subscribe("specification.created", func(payload any) {
    k.Publish("validation.requested", payload)
})

k.Subscribe("validation.requested", func(payload any) {
    // Validate
    k.Publish("validation.completed", result)
})

k.Start()
```

### 9.3 Service Resolution

```go
k := kernel.NewKernel()
k.Register("compiler", NewCompiler())

// Later, in another component
compiler, err := k.Resolve("compiler")
if err != nil {
    return err
}
comp := compiler.(*Compiler)
output := comp.Compile(spec)
```

---

## 10. References

- [NAEOS-KER-001.md](../kernel/NAEOS-KER-001.md) - Kernel Architecture
- [NAEOS-KER-002.md](../kernel/NAEOS-KER-002.md) - Kernel Implementation & Setup
- [NES-002-Kernel.md](NES-002-Kernel.md) - Kernel Specification
