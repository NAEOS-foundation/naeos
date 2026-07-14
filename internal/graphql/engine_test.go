package graphql

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected int
		hasError bool
	}{
		{
			name:     "simple query",
			query:    "{ health }",
			expected: 1,
			hasError: false,
		},
		{
			name:     "multiple fields",
			query:    "{ health status version }",
			expected: 3,
			hasError: false,
		},
		{
			name:     "empty query",
			query:    "{ }",
			expected: 0,
			hasError: false,
		},
		{
			name:     "query with arguments",
			query:    "{ user(name: John) }",
			expected: 1,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, errs := ParseQuery(tt.query)
			if tt.hasError && len(errs) == 0 {
				t.Error("expected errors, got none")
			}
			if !tt.hasError && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if len(ast.Selections) != tt.expected {
				t.Errorf("expected %d selections, got %d", tt.expected, len(ast.Selections))
			}
		})
	}
}

func TestParseSelection(t *testing.T) {
	sel, err := parseSelection("health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.Name != "health" {
		t.Errorf("expected name 'health', got %s", sel.Name)
	}

	sel, err = parseSelection(`user(name: "John")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.Name != "user" {
		t.Errorf("expected name 'user', got %s", sel.Name)
	}
	if sel.Arguments["name"] != `"John"` {
		t.Errorf("expected argument name 'John', got %s", sel.Arguments["name"])
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{`"hello"`, "hello"},
		{"true", true},
		{"false", false},
		{"null", nil},
		{"123", "123"},
	}

	for _, tt := range tests {
		result := parseValue(tt.input)
		if result != tt.expected {
			t.Errorf("parseValue(%s) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestIntrospect(t *testing.T) {
	schema := &Schema{
		Types: map[string]*TypeDef{
			"Health": {
				Name: "Health",
				Fields: map[string]*FieldDef{
					"status": {Name: "status", Type: "String"},
				},
			},
		},
	}

	executor := NewExecutor(schema)
	resp := executor.Introspect()

	if resp.Data == nil {
		t.Error("expected data in response")
	}
}

func TestGraphQLHandler(t *testing.T) {
	schema := &Schema{
		Queries: &OperationDef{
			Fields: map[string]*FieldDef{
				"health": {
					Name: "health",
					Type: "String",
					Resolve: func(ctx *Context, args map[string]any) (any, error) {
						return "healthy", nil
					},
				},
			},
		},
	}

	handler := Handler(schema)

	t.Run("GET introspect", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/graphql?introspect=true", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("POST query", func(t *testing.T) {
		body := `{"query": "{ health }"}`
		req := httptest.NewRequest("POST", "/graphql", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if !strings.Contains(w.Body.String(), "healthy") {
			t.Errorf("expected 'healthy' in response, got %s", w.Body.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/graphql", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", w.Code)
		}
	})
}
