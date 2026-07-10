# NAEOS-KER-003: Kernel Examples & Use Cases

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Contoh konkret implementasi NAEOS Kernel dalam berbagai skenario. Mencakup:
- Minimal kernel setup
- Component integration examples
- Event-driven workflows
- Plugin development examples
- Real-world deployment scenarios

---

## 2. Minimal Kernel Setup

### 2.1 Basic Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/naeos/kernel"
)

func main() {
    // Create kernel configuration
    config := kernel.KernelConfig{
        Name: "my-naeos",
        Version: "1.0.0",
        Knowledge: kernel.KnowledgeConfig{
            ArtifactPath: "./artifacts",
        },
        Policy: kernel.PolicyConfig{
            ConstitutionPath: "./constitution",
        },
    }
    
    // Initialize kernel
    k, err := kernel.New(config)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to initialize kernel: %v\n", err)
        os.Exit(1)
    }
    
    ctx := context.Background()
    
    // Start kernel
    if err := k.Start(ctx); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to start kernel: %v\n", err)
        os.Exit(1)
    }
    defer k.Stop(ctx)
    
    fmt.Println("Kernel started successfully")
}
```

### 2.2 Configuration File

```yaml
# kernel.yaml

kernel:
  name: naeos-core
  version: 1.0.0
  
  knowledge:
    artifact_path: ./artifacts
    cache_enabled: true
    cache_size: 1GB
    
  policy:
    constitution_path: ./constitution
    enforcement_mode: blocking
    
  runtime:
    max_components: 1000
    
  event_bus:
    type: memory
    buffer_size: 10000
    
  telemetry:
    logging_level: info
```

---

## 3. Component Integration

### 3.1 Compiler Component

Register dan integrate Compiler dengan Kernel:

```go
package main

import (
    "context"
    "github.com/naeos/kernel"
    "github.com/naeos/components/compiler"
)

func main() {
    // Initialize kernel
    k, err := kernel.New(config)
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // Create compiler component
    comp := compiler.NewCompiler(compiler.Config{
        OutputFormat: "go",
        CachePath: "./compiler-cache",
    })
    
    // Register with kernel
    if err := k.Register(comp); err != nil {
        panic(err)
    }
    
    // Subscribe to specification changes
    k.eventBus.Subscribe("specification.created", func(e kernel.Event) error {
        fmt.Printf("New specification: %s\n", e.Payload["id"])
        // Trigger compilation
        return comp.Compile(e.Payload["specification"])
    })
    
    k.Start(ctx)
    defer k.Stop(ctx)
}
```

### 3.2 Validator Component

```go
package main

import (
    "github.com/naeos/components/validator"
    "github.com/naeos/kernel"
)

func main() {
    k, _ := kernel.New(config)
    ctx := context.Background()
    
    // Create validator
    val := validator.NewValidator(validator.Config{
        SchemaPath: "./schemas",
        StrictMode: true,
    })
    
    // Register validator
    k.Register(val)
    
    // Subscribe to artifacts for validation
    k.eventBus.Subscribe("artifact.uploaded", func(e kernel.Event) error {
        artifact := e.Payload["artifact"]
        result, err := val.Validate(artifact)
        
        if !result.Valid {
            fmt.Printf("Validation failed: %s\n", result.Error)
            k.eventBus.Publish(kernel.Event{
                Type: "validation.failed",
                Payload: result,
            })
        }
        return err
    })
    
    k.Start(ctx)
    defer k.Stop(ctx)
}
```

### 3.3 AI Runtime Component

```go
package main

import (
    "github.com/naeos/components/ai"
    "github.com/naeos/kernel"
)

func main() {
    k, _ := kernel.New(config)
    ctx := context.Background()
    
    // Create AI runtime
    aiRuntime := ai.NewRuntime(ai.Config{
        ModelPath: "./models",
        GPUEnabled: true,
        MaxInference: 100,
    })
    
    // Register AI runtime
    k.Register(aiRuntime)
    
    // Use AI for policy generation
    k.eventBus.Subscribe("policy.generation.requested", func(e kernel.Event) error {
        prompt := e.Payload["prompt"].(string)
        
        // Generate policy using AI
        result, err := aiRuntime.Generate(ai.GenerationRequest{
            Type: "policy",
            Prompt: prompt,
        })
        
        k.eventBus.Publish(kernel.Event{
            Type: "policy.generated",
            Payload: map[string]interface{}{
                "policy": result.Output,
            },
        })
        return err
    })
    
    k.Start(ctx)
    defer k.Stop(ctx)
}
```

---

## 4. Event-Driven Workflows

### 4.1 Specification → Compile → Deploy

```go
func setupWorkflow(k kernel.Kernel) {
    // Step 1: Spec uploaded
    k.Subscribe("specification.uploaded", func(e kernel.Event) error {
        specID := e.Payload["id"].(string)
        fmt.Printf("Spec uploaded: %s\n", specID)
        
        // Trigger validation
        k.Publish(kernel.Event{
            Type: "specification.validation.requested",
            Payload: e.Payload,
        })
        return nil
    })
    
    // Step 2: Validation
    k.Subscribe("specification.validation.requested", func(e kernel.Event) error {
        spec := e.Payload["specification"]
        fmt.Printf("Validating specification\n")
        
        // Validate
        valid := validateSpec(spec)
        
        if valid {
            k.Publish(kernel.Event{
                Type: "specification.compilation.requested",
                Payload: e.Payload,
            })
        } else {
            k.Publish(kernel.Event{
                Type: "specification.validation.failed",
                Payload: e.Payload,
            })
        }
        return nil
    })
    
    // Step 3: Compilation
    k.Subscribe("specification.compilation.requested", func(e kernel.Event) error {
        spec := e.Payload["specification"]
        fmt.Printf("Compiling specification\n")
        
        // Compile
        compiled := compileSpec(spec)
        
        k.Publish(kernel.Event{
            Type: "specification.compiled",
            Payload: map[string]interface{}{
                "compiled": compiled,
            },
        })
        return nil
    })
    
    // Step 4: Deployment
    k.Subscribe("specification.compiled", func(e kernel.Event) error {
        compiled := e.Payload["compiled"]
        fmt.Printf("Deploying\n")
        
        // Deploy
        deploy(compiled)
        
        k.Publish(kernel.Event{
            Type: "deployment.completed",
            Payload: e.Payload,
        })
        return nil
    })
}
```

### 4.2 Policy Enforcement Workflow

```go
func setupPolicyWorkflow(k kernel.Kernel) {
    // Monitor policy violations
    k.Subscribe("policy.evaluation.completed", func(e kernel.Event) error {
        result := e.Payload["result"].(PolicyResult)
        
        if !result.Compliant {
            fmt.Printf("Policy violation: %s\n", result.ViolatedPolicy)
            
            // Determine action based on severity
            switch result.Severity {
            case "critical":
                // Block and alert immediately
                k.Publish(kernel.Event{
                    Type: "alert.critical",
                    Payload: result,
                })
                
            case "warning":
                // Log and notify
                k.Publish(kernel.Event{
                    Type: "alert.warning",
                    Payload: result,
                })
                
            case "info":
                // Just log
                k.Publish(kernel.Event{
                    Type: "alert.info",
                    Payload: result,
                })
            }
        }
        return nil
    })
}
```

---

## 5. Real-World Scenarios

### 5.1 Microservices Architecture

NAEOS Kernel coordinating multiple microservices:

```go
func setupMicroservices(k kernel.Kernel) error {
    // Service 1: API Gateway
    apiGateway := NewAPIGatewayComponent()
    k.Register(apiGateway)
    
    // Service 2: Compiler Service
    compilerService := NewCompilerComponent()
    k.Register(compilerService)
    
    // Service 3: Policy Engine
    policyEngine := NewPolicyEngineComponent()
    k.Register(policyEngine)
    
    // Service 4: Deployment Service
    deploymentService := NewDeploymentComponent()
    k.Register(deploymentService)
    
    // Orchestrate workflow
    k.Subscribe("request.received", func(e kernel.Event) error {
        request := e.Payload["request"]
        
        // Route to compiler
        k.Publish(kernel.Event{
            Type: "compile.request",
            Payload: map[string]interface{}{
                "spec": request.Specification,
            },
        })
        return nil
    })
    
    return nil
}
```

### 5.2 Cloud Deployment

```go
func setupCloudDeployment(k kernel.Kernel) error {
    // Register cloud components
    k8s := NewKubernetesComponent()
    terraform := NewTerraformComponent()
    monitoring := NewMonitoringComponent()
    
    k.Register(k8s)
    k.Register(terraform)
    k.Register(monitoring)
    
    // Deploy to cloud
    k.Subscribe("deployment.requested", func(e kernel.Event) error {
        spec := e.Payload["specification"]
        env := e.Payload["environment"]
        
        // Generate infrastructure config
        k.Publish(kernel.Event{
            Type: "infrastructure.generation.requested",
            Payload: map[string]interface{}{
                "spec": spec,
                "env": env,
            },
        })
        return nil
    })
    
    // Generate infrastructure
    k.Subscribe("infrastructure.generation.requested", func(e kernel.Event) error {
        spec := e.Payload["spec"]
        env := e.Payload["env"]
        
        // Use Terraform to generate IaC
        config := terraform.Generate(spec, env)
        
        k.Publish(kernel.Event{
            Type: "infrastructure.ready",
            Payload: map[string]interface{}{
                "config": config,
            },
        })
        return nil
    })
    
    // Deploy to Kubernetes
    k.Subscribe("infrastructure.ready", func(e kernel.Event) error {
        config := e.Payload["config"]
        
        k8s.Deploy(config)
        
        k.Publish(kernel.Event{
            Type: "deployment.completed",
            Payload: e.Payload,
        })
        return nil
    })
    
    return nil
}
```

### 5.3 CI/CD Pipeline Integration

```go
func setupCIPipeline(k kernel.Kernel) error {
    ciRunner := NewCIPipelineComponent()
    k.Register(ciRunner)
    
    // Webhook from GitHub
    k.Subscribe("github.push", func(e kernel.Event) error {
        repo := e.Payload["repository"]
        commit := e.Payload["commit"]
        
        fmt.Printf("New push: %s@%s\n", repo, commit)
        
        // Run CI pipeline
        k.Publish(kernel.Event{
            Type: "ci.pipeline.start",
            Payload: map[string]interface{}{
                "repo": repo,
                "commit": commit,
            },
        })
        return nil
    })
    
    // CI Pipeline steps
    k.Subscribe("ci.pipeline.start", func(e kernel.Event) error {
        // Step 1: Build
        result := ciRunner.Build(e.Payload["repo"])
        
        if result.Success {
            k.Publish(kernel.Event{
                Type: "ci.tests.start",
                Payload: e.Payload,
            })
        } else {
            k.Publish(kernel.Event{
                Type: "ci.failed",
                Payload: map[string]interface{}{
                    "error": result.Error,
                },
            })
        }
        return nil
    })
    
    // Step 2: Tests
    k.Subscribe("ci.tests.start", func(e kernel.Event) error {
        result := ciRunner.RunTests(e.Payload["repo"])
        
        if result.Success {
            k.Publish(kernel.Event{
                Type: "ci.policy.check",
                Payload: e.Payload,
            })
        }
        return nil
    })
    
    // Step 3: Policy check
    k.Subscribe("ci.policy.check", func(e kernel.Event) error {
        result := ciRunner.CheckPolicies(e.Payload["repo"])
        
        if result.Compliant {
            k.Publish(kernel.Event{
                Type: "ci.success",
                Payload: e.Payload,
            })
        }
        return nil
    })
    
    return nil
}
```

---

## 6. Plugin Examples

### 6.1 Custom Metrics Plugin

```go
package metrics_plugin

import (
    "github.com/naeos/kernel"
    "time"
)

type MetricsPlugin struct {
    metrics map[string]int64
}

func (p *MetricsPlugin) Name() string {
    return "metrics-plugin"
}

func (p *MetricsPlugin) Initialize(k kernel.Kernel) error {
    p.metrics = make(map[string]int64)
    
    // Subscribe to all events
    k.Subscribe("*", func(e kernel.Event) error {
        p.metrics[e.Type]++
        return nil
    })
    
    return nil
}

func (p *MetricsPlugin) Topics() []string {
    return []string{"*"}
}

func (p *MetricsPlugin) Handle(event kernel.Event) error {
    // Handle metrics
    return nil
}

func (p *MetricsPlugin) GetMetrics() map[string]int64 {
    return p.metrics
}
```

### 6.2 Audit Logging Plugin

```go
package audit_plugin

import (
    "fmt"
    "github.com/naeos/kernel"
    "time"
)

type AuditPlugin struct {
    logFile *os.File
}

func (p *AuditPlugin) Name() string {
    return "audit-plugin"
}

func (p *AuditPlugin) Initialize(k kernel.Kernel) error {
    var err error
    p.logFile, err = os.OpenFile(
        "audit.log",
        os.O_CREATE|os.O_WRONLY|os.O_APPEND,
        0644,
    )
    return err
}

func (p *AuditPlugin) Topics() []string {
    return []string{"*"}
}

func (p *AuditPlugin) Handle(event kernel.Event) error {
    entry := fmt.Sprintf(
        "[%s] %s: %v\n",
        time.Now().Format(time.RFC3339),
        event.Type,
        event.Payload,
    )
    
    _, err := p.logFile.WriteString(entry)
    return err
}

func (p *AuditPlugin) Shutdown() error {
    return p.logFile.Close()
}
```

---

## 7. Error Handling Examples

### 7.1 Component Initialization Error

```go
func setupWithErrorHandling(k kernel.Kernel) error {
    compiler := compiler.NewCompiler(config)
    
    if err := k.Register(compiler); err != nil {
        return fmt.Errorf("failed to register compiler: %w", err)
    }
    
    return nil
}
```

### 7.2 Event Handling Errors

```go
k.Subscribe("specification.created", func(e kernel.Event) error {
    spec := e.Payload["specification"]
    
    if err := validateSpec(spec); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    return nil
})
```

---

## 8. References

- [NAEOS-KER-001.md](NAEOS-KER-001.md) - Kernel Architecture
- [NAEOS-KER-002.md](NAEOS-KER-002.md) - Kernel Implementation & Setup
- [NAEOS-POL-003.md](../policy/NAEOS-POL-003.md) - Policy Examples
