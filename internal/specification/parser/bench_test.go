package parser

import (
	"testing"
)

func BenchmarkParseSimple(b *testing.B) {
	spec := `project: bench-app
modules:
  - name: core
    path: ./core
services:
  - name: api
    kind: http
    port: 8080
`

	p := NewParser(".")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(spec)
	}
}

func BenchmarkParseComplex(b *testing.B) {
	spec := `project: complex-app
version: 0.3.0
modules:
  - name: auth
    path: ./auth
    description: Authentication module
    dependencies: [core, user]
  - name: api
    path: ./api
    description: REST API
    dependencies: [auth, core]
  - name: core
    path: ./core
    description: Core library
  - name: user
    path: ./user
    description: User management
    dependencies: [core]
  - name: payment
    path: ./payment
    description: Payment processing
    dependencies: [core, user]
services:
  - name: gateway
    kind: http
    port: 8080
    description: API Gateway
    endpoints:
      - method: GET
        path: /health
        action: healthCheck
      - method: POST
        path: /api/auth/login
        action: login
      - method: POST
        path: /api/auth/logout
        action: logout
      - method: GET
        path: /api/users
        action: listUsers
      - method: POST
        path: /api/users
        action: createUser
  - name: worker
    kind: worker
    description: Background worker
  - name: grpc-service
    kind: grpc
    port: 9090
    description: gRPC service
architecture:
  pattern: hexagonal
  principles: [loose-coupling, high-cohesion, single-responsibility]
deployment:
  strategy: rolling
  environments: [development, staging, production]
testing:
  strategy: unit+integration
  coverage: "80"
generation:
  languages: [go, typescript, python]
  outputDir: ./generated
  moduleDir: ./modules
`

	p := NewParser(".")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(spec)
	}
}

func BenchmarkParseWithVariables(b *testing.B) {
	spec := `project: ${project_name}
modules:
  - name: ${module_name}
    path: ./${module_name}
    dependencies: [$ref{core.name}]
services:
  - name: ${service_name}
    kind: http
    port: $env{SERVICE_PORT}
`

	p := NewParser(".")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(spec)
	}
}

func BenchmarkValidateModules(b *testing.B) {
	modules := []Module{
		{Name: "auth", Dependencies: []string{"core"}},
		{Name: "api", Dependencies: []string{"auth", "core"}},
		{Name: "core", Dependencies: []string{}},
		{Name: "user", Dependencies: []string{"core"}},
		{Name: "payment", Dependencies: []string{"core", "user"}},
	}

	v := NewSpecValidator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateModules(modules)
	}
}

func BenchmarkValidateServices(b *testing.B) {
	services := []Service{
		{Name: "api", Kind: "http", Port: 8080},
		{Name: "grpc", Kind: "grpc", Port: 9090},
		{Name: "worker", Kind: "worker", Port: 0},
	}

	v := NewSpecValidator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateServices(services)
	}
}

func BenchmarkVariableResolver(b *testing.B) {
	resolver := NewVariableResolver()
	resolver.SetVar("project", "my-app")
	resolver.SetVar("module", "auth")
	resolver.SetVars(map[string]string{
		"service": "api",
		"port":    "8080",
	})

	input := "project: ${project}, module: ${module}, service: ${service}, port: ${port}"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.Resolve(input)
	}
}

func BenchmarkSchemaVersionCheck(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckSpecVersion("0.3.0")
	}
}

func BenchmarkDetectCycles(b *testing.B) {
	graph := map[string][]string{
		"a": {"b", "c"},
		"b": {"c"},
		"c": {"d"},
		"d": {},
		"e": {"f"},
		"f": {"e"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detectCycles(graph)
	}
}
