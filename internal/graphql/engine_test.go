// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testSchema() *Schema {
	schema := &Schema{
		Types: map[string]*TypeDef{
			"User": {
				Name: "User",
				Fields: map[string]*FieldDef{
					"id":    {Name: "id", Type: "string"},
					"name":  {Name: "name", Type: "string"},
					"email": {Name: "email", Type: "string"},
				},
			},
		},
		Queries: &OperationDef{
			Fields: map[string]*FieldDef{
				"hello": {
					Name: "hello",
					Type: "string",
					Resolve: func(ctx *Context, args map[string]any) (any, error) {
						return "world", nil
					},
				},
				"user": {
					Name: "user",
					Type: "User",
					Args: map[string]*ArgDef{
						"id": {Name: "id", Type: "string", Required: true},
					},
					Resolve: func(ctx *Context, args map[string]any) (any, error) {
						return map[string]any{
							"id":    args["id"],
							"name":  "John",
							"email": "john@example.com",
						}, nil
					},
				},
				"users": {
					Name: "users",
					Type: "[User]",
					Resolve: func(ctx *Context, args map[string]any) (any, error) {
						return []any{
							map[string]any{"id": "1", "name": "Alice"},
							map[string]any{"id": "2", "name": "Bob"},
						}, nil
					},
				},
				"add": {
					Name: "add",
					Type: "number",
					Args: map[string]*ArgDef{
						"a": {Name: "a", Type: "number", Required: true},
						"b": {Name: "b", Type: "number", Required: true},
					},
					Resolve: func(ctx *Context, args map[string]any) (any, error) {
						a, _ := args["a"].(int)
						b, _ := args["b"].(int)
						return a + b, nil
					},
				},
			},
		},
	}
	return schema
}

func TestBasicQuery(t *testing.T) {
	schema := testSchema()
	executor := NewExecutor(schema)
	ctx := &Context{Variables: make(map[string]any)}

	resp := executor.Execute(ctx, `{ hello }`)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map data, got %T", resp.Data)
	}
	if data["hello"] != "world" {
		t.Errorf("expected 'world', got %v", data["hello"])
	}
}

func TestNestedQuery(t *testing.T) {
	schema := testSchema()
	executor := NewExecutor(schema)
	ctx := &Context{Variables: make(map[string]any)}

	resp := executor.Execute(ctx, `{ user(id: "1") { id name email } }`)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}
	data := resp.Data.(map[string]any)
	user := data["user"].(map[string]any)
	if user["name"] != "John" {
		t.Errorf("expected name 'John', got %v", user["name"])
	}
	if user["email"] != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got %v", user["email"])
	}
	if user["id"] != "1" {
		t.Errorf("expected id '1', got %v", user["id"])
	}
}

func TestNestedQuerySelectiveFields(t *testing.T) {
	schema := testSchema()
	executor := NewExecutor(schema)
	ctx := &Context{Variables: make(map[string]any)}

	resp := executor.Execute(ctx, `{ user(id: "1") { name } }`)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}
	data := resp.Data.(map[string]any)
	user := data["user"].(map[string]any)
	if _, hasEmail := user["email"]; hasEmail {
		t.Error("should not have email field when not requested")
	}
	if user["name"] != "John" {
		t.Errorf("expected name 'John', got %v", user["name"])
	}
}

func TestMultipleQueries(t *testing.T) {
	schema := testSchema()
	executor := NewExecutor(schema)
	ctx := &Context{Variables: make(map[string]any)}

	resp := executor.Execute(ctx, `{ hello user(id: "1") { name } }`)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}
	data := resp.Data.(map[string]any)
	if data["hello"] != "world" {
		t.Errorf("expected 'world', got %v", data["hello"])
	}
	user := data["user"].(map[string]any)
	if user["name"] != "John" {
		t.Errorf("expected 'John', got %v", user["name"])
	}
}

func TestVariableSupport(t *testing.T) {
	schema := testSchema()
	executor := NewExecutor(schema)
	ctx := &Context{
		Variables: map[string]any{
			"userId": "42",
		},
	}

	resp := executor.Execute(ctx, `{ user(id: $userId) { name } }`)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}
	data := resp.Data.(map[string]any)
	user := data["user"].(map[string]any)
	if user["name"] != "John" {
		t.Errorf("expected 'John', got %v", user["name"])
	}
}

func TestNumericArgs(t *testing.T) {
	schema := testSchema()
	executor := NewExecutor(schema)
	ctx := &Context{Variables: make(map[string]any)}

	resp := executor.Execute(ctx, `{ add(a: 3, b: 4) }`)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}
	data := resp.Data.(map[string]any)
	if data["add"] != 7 {
		t.Errorf("expected 7, got %v", data["add"])
	}
}

func TestDepthLimiting(t *testing.T) {
	schema := testSchema()
	executor := NewExecutor(schema)
	executor.SetMaxDepth(0)
	ctx := &Context{Variables: make(map[string]any)}

	resp := executor.Execute(ctx, `{ user(id: "1") { name email } }`)
	if len(resp.Errors) == 0 {
		t.Error("expected depth limit error")
	}
}

func TestMiddleware(t *testing.T) {
	schema := testSchema()
	executor := NewExecutor(schema)

	called := false
	executor.Use(func(ctx *Context, next func(ctx *Context) (*Response, error)) (*Response, error) {
		called = true
		return next(ctx)
	})

	ctx := &Context{Variables: make(map[string]any)}
	executor.Execute(ctx, `{ hello }`)

	if !called {
		t.Error("middleware was not called")
	}
}

func TestMiddlewareChain(t *testing.T) {
	schema := testSchema()
	executor := NewExecutor(schema)

	order := []string{}
	executor.Use(func(ctx *Context, next func(ctx *Context) (*Response, error)) (*Response, error) {
		order = append(order, "before-1")
		resp, err := next(ctx)
		order = append(order, "after-1")
		return resp, err
	})
	executor.Use(func(ctx *Context, next func(ctx *Context) (*Response, error)) (*Response, error) {
		order = append(order, "before-2")
		resp, err := next(ctx)
		order = append(order, "after-2")
		return resp, err
	})

	ctx := &Context{Variables: make(map[string]any)}
	executor.Execute(ctx, `{ hello }`)

	if len(order) != 4 || order[0] != "before-1" || order[1] != "before-2" || order[2] != "after-2" || order[3] != "after-1" {
		t.Errorf("middleware order wrong: %v", order)
	}
}

func TestUnknownField(t *testing.T) {
	schema := testSchema()
	executor := NewExecutor(schema)
	ctx := &Context{Variables: make(map[string]any)}

	resp := executor.Execute(ctx, `{ unknown }`)
	if len(resp.Errors) == 0 {
		t.Error("expected error for unknown field")
	}
}

func TestEmptyQuery(t *testing.T) {
	schema := testSchema()
	executor := NewExecutor(schema)
	ctx := &Context{Variables: make(map[string]any)}

	resp := executor.Execute(ctx, `{}`)
	if len(resp.Errors) > 0 {
		t.Errorf("empty query should not error: %v", resp.Errors)
	}
}

func TestIntrospect(t *testing.T) {
	schema := testSchema()
	executor := NewExecutor(schema)
	resp := executor.Introspect()
	if resp.Data == nil {
		t.Error("expected introspection data")
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		input string
		want  any
	}{
		{`"hello"`, "hello"},
		{"true", true},
		{"false", false},
		{"null", nil},
		{"42.5", 42.5},
		{"100", 100},
		{"abc", "abc"},
	}
	for _, tt := range tests {
		got := parseValue(tt.input)
		if got != tt.want {
			t.Errorf("parseValue(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestHandler(t *testing.T) {
	schema := testSchema()
	handler := Handler(schema)

	body := `{"query": "{ hello }"}`
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/graphql", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "world") {
		t.Errorf("expected 'world' in response: %s", w.Body.String())
	}
}

func TestHandlerGETIntrospect(t *testing.T) {
	schema := testSchema()
	handler := Handler(schema)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/graphql?introspect=true", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "__schema") {
		t.Errorf("expected introspection data: %s", w.Body.String())
	}
}

func TestHandlerMethodNotAllowed(t *testing.T) {
	schema := testSchema()
	handler := Handler(schema)

	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/graphql", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlerInvalidBody(t *testing.T) {
	schema := testSchema()
	handler := Handler(schema)

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/graphql", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
