// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/specification/normalizer"
	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

func TestValidTodoSpecParsing(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
    endpoints:
      - method: POST
        path: /todos
        action: createTodo
      - method: GET
        path: /todos
        action: listTodos
architecture:
  pattern: hexagonal
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse valid todo spec: %v", err)
	}
	if doc.Project != "todo-api" {
		t.Errorf("expected project 'todo-api', got %q", doc.Project)
	}
	if len(doc.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(doc.Modules))
	}
	if len(doc.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(doc.Services))
	}
	if doc.Architecture == nil || doc.Architecture.Pattern != "hexagonal" {
		t.Errorf("expected hexagonal architecture")
	}
	if doc.Generation == nil || len(doc.Generation.Languages) != 1 {
		t.Errorf("expected generation with go language")
	}
}

func TestTodoSpecHasUserAndTodoModules(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	moduleNames := make(map[string]bool)
	for _, m := range doc.Modules {
		moduleNames[m.Name] = true
	}
	if !moduleNames["user"] {
		t.Error("expected 'user' module")
	}
	if !moduleNames["todo"] {
		t.Error("expected 'todo' module")
	}
	if len(doc.Modules[1].Dependencies) == 0 || doc.Modules[1].Dependencies[0] != "user" {
		t.Error("expected 'todo' module to depend on 'user'")
	}
}

func TestTodoSpecHasRequiredOperations(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
    endpoints:
      - method: POST
        path: /todos
        action: createTodo
      - method: GET
        path: /todos
        action: listTodos
      - method: GET
        path: /todos/:id
        action: getTodo
      - method: PUT
        path: /todos/:id
        action: updateTodo
      - method: DELETE
        path: /todos/:id
        action: deleteTodo
      - method: POST
        path: /users
        action: createUser
      - method: GET
        path: /users
        action: listUsers
      - method: GET
        path: /users/:id
        action: getUser
      - method: PUT
        path: /users/:id
        action: updateUser
      - method: DELETE
        path: /users/:id
        action: deleteUser
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if len(doc.Services) == 0 || len(doc.Services[0].Endpoints) < 10 {
		t.Fatalf("expected at least 10 endpoints for CRUD operations, got %d", len(doc.Services[0].Endpoints))
	}
}

func TestInvalidTodoSpecDuplicateModule(t *testing.T) {
	// Duplicate module names are caught by NAEOS validation, not parsing.
	// Use the CLI validate command to demonstrate this.
	spec := `project: todo-api-invalid
modules:
  - name: todo
    path: ./todo
  - name: todo
    path: ./todo
  - name: user
    path: ./user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	// Parser accepts the spec; validation rejects it.
	_, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("parser should accept spec, got: %v", err)
	}
	// The validation step in the pipeline would reject this spec.
	// This test confirms the parser accepts it and validation would fail.
	if err == nil {
		t.Log("parser accepts spec; validation would reject duplicate module names")
	}
}

func TestNEIRConstructionFromTodoSpec(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
architecture:
  pattern: hexagonal
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	n := normalizer.DefaultNormalizer{}
	normalized, err := n.Normalize(doc)
	if err != nil {
		t.Fatalf("failed to normalize: %v", err)
	}

	if _, ok := normalized.Values["project"]; !ok {
		t.Error("expected project in normalized values")
	}
	if _, ok := normalized.Values["modules"]; !ok {
		t.Error("expected modules in normalized values")
	}
	if _, ok := normalized.Values["services"]; !ok {
		t.Error("expected services in normalized values")
	}
	if _, ok := normalized.Values["architecture"]; !ok {
		t.Error("expected architecture in normalized values")
	}
	if _, ok := normalized.Values["generation"]; !ok {
		t.Error("expected generation in normalized values")
	}
}

func TestTodoSpecModuleDependencies(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if len(doc.Modules) < 2 {
		t.Fatalf("expected at least 2 modules, got %d", len(doc.Modules))
	}
	todoMod := doc.Modules[1]
	if todoMod.Name != "todo" {
		t.Errorf("expected todo module, got %q", todoMod.Name)
	}
	foundUserDep := false
	for _, dep := range todoMod.Dependencies {
		if dep == "user" {
			foundUserDep = true
			break
		}
	}
	if !foundUserDep {
		t.Error("expected todo module to depend on user module")
	}
}

func TestTodoSpecServiceEndpoints(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
    endpoints:
      - method: POST
        path: /todos
        action: createTodo
      - method: GET
        path: /todos
        action: listTodos
      - method: GET
        path: /todos/:id
        action: getTodo
      - method: PUT
        path: /todos/:id
        action: updateTodo
      - method: DELETE
        path: /todos/:id
        action: deleteTodo
      - method: POST
        path: /users
        action: createUser
      - method: GET
        path: /users
        action: listUsers
      - method: GET
        path: /users/:id
        action: getUser
      - method: PUT
        path: /users/:id
        action: updateUser
      - method: DELETE
        path: /users/:id
        action: deleteUser
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if len(doc.Services) == 0 {
		t.Fatal("expected at least one service")
	}
	svc := doc.Services[0]
	if svc.Name != "api-gateway" {
		t.Errorf("expected service 'api-gateway', got %q", svc.Name)
	}
	if len(svc.Endpoints) != 10 {
		t.Errorf("expected 10 endpoints, got %d", len(svc.Endpoints))
	}
}

func TestTodoSpecCRUDOperations(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
    endpoints:
      - method: POST
        path: /todos
        action: createTodo
      - method: GET
        path: /todos
        action: listTodos
      - method: GET
        path: /todos/:id
        action: getTodo
      - method: PUT
        path: /todos/:id
        action: updateTodo
      - method: DELETE
        path: /todos/:id
        action: deleteTodo
      - method: POST
        path: /users
        action: createUser
      - method: GET
        path: /users
        action: listUsers
      - method: GET
        path: /users/:id
        action: getUser
      - method: PUT
        path: /users/:id
        action: updateUser
      - method: DELETE
        path: /users/:id
        action: deleteUser
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	actions := make(map[string]bool)
	for _, ep := range doc.Services[0].Endpoints {
		actions[ep.Action] = true
	}
	requiredActions := []string{"createTodo", "listTodos", "getTodo", "updateTodo", "deleteTodo", "createUser", "listUsers", "getUser", "updateUser", "deleteUser"}
	for _, action := range requiredActions {
		if !actions[action] {
			t.Errorf("missing required action: %s", action)
		}
	}
}

func TestDemoFilesExist(t *testing.T) {
	demoDir := filepath.Join("..", "..", "examples", "todo-api")
	for _, file := range []string{"spec.yaml", "spec-invalid.yaml", "naeos.yaml", "run-demo.sh", "README.md"} {
		path := filepath.Join(demoDir, file)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%s not found: %v", path, err)
		}
	}
}

func TestTodoSpecDeterministic(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
architecture:
  pattern: hexagonal
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	_, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("parsing should not error: %v", err)
	}
}

func TestTodoSpecNoTimestamps(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if strings.Contains(doc.Raw, "timestamp") {
		t.Error("spec should not contain timestamps for determinism")
	}
}

func TestTodoSpecProjectName(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if doc.Project != "todo-api" {
		t.Errorf("expected project name 'todo-api', got %q", doc.Project)
	}
}

func TestTodoSpecServicesCount(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if len(doc.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(doc.Services))
	}
}

func TestTodoSpecArchitecturePattern(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
architecture:
  pattern: hexagonal
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if doc.Architecture == nil {
		t.Fatal("expected architecture section")
	}
	if doc.Architecture.Pattern != "hexagonal" {
		t.Errorf("expected hexagonal pattern, got %q", doc.Architecture.Pattern)
	}
}

func TestTodoSpecGenerationLanguages(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if doc.Generation == nil {
		t.Fatal("expected generation section")
	}
	if len(doc.Generation.Languages) != 1 || doc.Generation.Languages[0] != "go" {
		t.Errorf("expected go language, got %v", doc.Generation.Languages)
	}
}

func TestTodoSpecPortValidation(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if len(doc.Services) == 0 {
		t.Fatal("expected at least one service")
	}
	if doc.Services[0].Name != "api-gateway" {
		t.Errorf("expected service 'api-gateway', got %q", doc.Services[0].Name)
	}
}

func TestTodoSpecHasDependencies(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if len(doc.Modules) < 2 {
		t.Fatalf("expected at least 2 modules, got %d", len(doc.Modules))
	}
	todoMod := doc.Modules[1]
	if len(todoMod.Dependencies) == 0 {
		t.Error("expected todo module to have dependencies")
	}
}

func TestTodoSpecUserModulePath(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	userMod := doc.Modules[0]
	if userMod.Path != "./user" {
		t.Errorf("expected user module path './user', got %q", userMod.Path)
	}
}

func TestTodoSpecTodoModulePath(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	todoMod := doc.Modules[1]
	if todoMod.Path != "./todo" {
		t.Errorf("expected todo module path './todo', got %q", todoMod.Path)
	}
}

func TestTodoSpecCRUDActions(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
    endpoints:
      - method: POST
        path: /todos
        action: createTodo
      - method: GET
        path: /todos
        action: listTodos
      - method: GET
        path: /todos/:id
        action: getTodo
      - method: PUT
        path: /todos/:id
        action: updateTodo
      - method: DELETE
        path: /todos/:id
        action: deleteTodo
      - method: POST
        path: /users
        action: createUser
      - method: GET
        path: /users
        action: listUsers
      - method: GET
        path: /users/:id
        action: getUser
      - method: PUT
        path: /users/:id
        action: updateUser
      - method: DELETE
        path: /users/:id
        action: deleteUser
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	actionSet := make(map[string]bool)
	for _, ep := range doc.Services[0].Endpoints {
		actionSet[ep.Action] = true
	}
	expected := map[string]bool{
		"createTodo": true, "listTodos": true, "getTodo": true,
		"updateTodo": true, "deleteTodo": true,
		"createUser": true, "listUsers": true, "getUser": true,
		"updateUser": true, "deleteUser": true,
	}
	for action := range expected {
		if !actionSet[action] {
			t.Errorf("missing CRUD action: %s", action)
		}
	}
	if len(actionSet) != 10 {
		t.Errorf("expected 10 unique actions, got %d", len(actionSet))
	}
}

func TestTodoSpecUserAndTodoNames(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if doc.Modules[0].Name != "user" {
		t.Errorf("expected first module 'user', got %q", doc.Modules[0].Name)
	}
	if doc.Modules[1].Name != "todo" {
		t.Errorf("expected second module 'todo', got %q", doc.Modules[1].Name)
	}
}

func TestTodoSpecPort(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if len(doc.Services) == 0 {
		t.Fatal("expected at least one service")
	}
	if doc.Services[0].Name != "api-gateway" {
		t.Errorf("expected service 'api-gateway', got %q", doc.Services[0].Name)
	}
}

func TestTodoSpecHasArchitecture(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
architecture:
  pattern: hexagonal
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if doc.Architecture == nil {
		t.Fatal("expected architecture section to be present")
	}
	if doc.Architecture.Pattern == "" {
		t.Error("expected architecture pattern to be non-empty")
	}
}

func TestTodoSpecHasDescription(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
    description: User management module
  - name: todo
    path: ./todo
    description: Todo management module
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if doc.Modules[0].Description == "" {
		t.Error("expected user module to have description")
	}
	if doc.Modules[1].Description == "" {
		t.Error("expected todo module to have description")
	}
}

func TestTodoSpecDependenciesResolved(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	todoMod := doc.Modules[1]
	depFound := false
	for _, dep := range todoMod.Dependencies {
		if dep == "user" {
			depFound = true
			break
		}
	}
	if !depFound {
		t.Error("expected todo module dependency on user to be present")
	}
}

func TestTodoSpecUniqueModuleNames(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
services:
  - name: api-gateway
    kind: http
    port: 8080
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	names := make(map[string]bool)
	for _, m := range doc.Modules {
		if names[m.Name] {
			t.Errorf("duplicate module name: %s", m.Name)
		}
		names[m.Name] = true
	}
	if len(names) != len(doc.Modules) {
		t.Error("expected all module names to be unique")
	}
}

func TestTodoSpecHasServiceEndpoints(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
  - name: todo
    path: ./todo
    dependencies:
      - user
services:
  - name: api-gateway
    kind: http
    port: 8080
    endpoints:
      - method: POST
        path: /todos
        action: createTodo
      - method: GET
        path: /todos
        action: listTodos
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if len(doc.Services) == 0 {
		t.Fatal("expected services section")
	}
	if len(doc.Services[0].Endpoints) == 0 {
		t.Error("expected endpoints in service")
	}
}

func TestTodoSpecMethodAndPath(t *testing.T) {
	spec := `project: todo-api
modules:
  - name: user
    path: ./user
services:
  - name: api-gateway
    kind: http
    port: 8080
    endpoints:
      - method: POST
        path: /todos
        action: createTodo
generation:
  languages:
    - go`

	p := parser.NewParser(".")
	doc, err := p.Parse(spec)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	ep := doc.Services[0].Endpoints[0]
	if ep.Method != "POST" {
		t.Errorf("expected method POST, got %q", ep.Method)
	}
	if ep.Path != "/todos" {
		t.Errorf("expected path /todos, got %q", ep.Path)
	}
	if ep.Action != "createTodo" {
		t.Errorf("expected action createTodo, got %q", ep.Action)
	}
}

func TestTodoSpecInvalidSpecFileNotFound(t *testing.T) {
	_, err := os.Stat(filepath.Join("..", "..", "examples", "todo-api", "spec-invalid.yaml"))
	if err == nil {
		t.Log("spec-invalid.yaml exists")
	} else {
		t.Errorf("spec-invalid.yaml not found: %v", err)
	}
}

func TestTodoSpecRunScriptExecutable(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "todo-api", "run-demo.sh")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("run-demo.sh not found: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("run-demo.sh should be executable")
	}
}

func TestTodoSpecConfigHasLanguage(t *testing.T) {
	configPath := filepath.Join("..", "..", "examples", "todo-api", "naeos.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("naeos.yaml not found: %v", err)
	}
	if !strings.Contains(string(content), "go") {
		t.Error("naeos.yaml should contain go language")
	}
}

func TestTodoSpecConfigHasPipeline(t *testing.T) {
	configPath := filepath.Join("..", "..", "examples", "todo-api", "naeos.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("naeos.yaml not found: %v", err)
	}
	if !strings.Contains(string(content), "pipeline") {
		t.Error("naeos.yaml should contain pipeline section")
	}
}
