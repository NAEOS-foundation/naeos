package parser

import (
	"testing"
)

func TestResolveVariable(t *testing.T) {
	r := NewVariableResolver()
	r.SetVar("name", "test-project")
	r.SetVar("port", "8080")

	tests := []struct {
		input    string
		expected string
	}{
		{"hello ${name}", "hello test-project"},
		{"port: ${port}", "port: 8080"},
		{"no vars here", "no vars here"},
		{"${name}-${port}", "test-project-8080"},
		{"${missing} stays", "${missing} stays"},
	}

	for _, tt := range tests {
		got, err := r.Resolve(tt.input)
		if err != nil {
			t.Fatalf("Resolve(%q) error: %v", tt.input, err)
		}
		if got != tt.expected {
			t.Errorf("Resolve(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestResolveEnv(t *testing.T) {
	t.Setenv("NAEOS_TEST_VAR", "env-value")

	r := NewVariableResolver()
	got, err := r.Resolve("value: $env{NAEOS_TEST_VAR}")
	if err != nil {
		t.Fatal(err)
	}
	if got != "value: env-value" {
		t.Errorf("got %q, want %q", got, "value: env-value")
	}
}

func TestResolveEnvMissing(t *testing.T) {
	r := NewVariableResolver()
	got, err := r.Resolve("$env{NONEXISTENT_VAR_12345}")
	if err != nil {
		t.Fatal(err)
	}
	if got != "$env{NONEXISTENT_VAR_12345}" {
		t.Errorf("unresolved env should stay as-is, got %q", got)
	}
}

func TestResolveMap(t *testing.T) {
	r := NewVariableResolver()
	r.SetVar("proj", "myapp")

	input := map[string]any{
		"project": "${proj}",
		"port":    8080,
		"nested": map[string]any{
			"name": "${proj}-service",
		},
		"list": []any{"${proj}", "static"},
	}

	got, err := r.ResolveMap(input)
	if err != nil {
		t.Fatal(err)
	}

	if got["project"] != "myapp" {
		t.Errorf("project = %v, want myapp", got["project"])
	}
	if got["port"] != 8080 {
		t.Errorf("port = %v, want 8080", got["port"])
	}

	nested := got["nested"].(map[string]any)
	if nested["name"] != "myapp-service" {
		t.Errorf("nested.name = %v, want myapp-service", nested["name"])
	}

	list := got["list"].([]any)
	if list[0] != "myapp" {
		t.Errorf("list[0] = %v, want myapp", list[0])
	}
	if list[1] != "static" {
		t.Errorf("list[1] = %v, want static", list[1])
	}
}

func TestValidatorModules(t *testing.T) {
	v := NewSpecValidator()

	t.Run("duplicate modules", func(t *testing.T) {
		modules := []Module{
			{Name: "auth", Path: "./auth"},
			{Name: "auth", Path: "./auth2"},
		}
		issues := v.ValidateModules(modules)
		found := false
		for _, issue := range issues {
			if issue.Rule == "module-duplicate" {
				found = true
			}
		}
		if !found {
			t.Error("expected duplicate module issue")
		}
	})

	t.Run("circular dependency", func(t *testing.T) {
		modules := []Module{
			{Name: "a", Dependencies: []string{"b"}},
			{Name: "b", Dependencies: []string{"a"}},
		}
		issues := v.ValidateModules(modules)
		found := false
		for _, issue := range issues {
			if issue.Rule == "circular-dependency" {
				found = true
			}
		}
		if !found {
			t.Error("expected circular dependency issue")
		}
	})

	t.Run("dangling dependency", func(t *testing.T) {
		modules := []Module{
			{Name: "a", Dependencies: []string{"nonexistent"}},
		}
		issues := v.ValidateModules(modules)
		found := false
		for _, issue := range issues {
			if issue.Rule == "dependency-not-found" {
				found = true
			}
		}
		if !found {
			t.Error("expected dependency-not-found issue")
		}
	})

	t.Run("valid modules", func(t *testing.T) {
		modules := []Module{
			{Name: "auth", Dependencies: []string{}},
			{Name: "api", Dependencies: []string{"auth"}},
		}
		issues := v.ValidateModules(modules)
		for _, issue := range issues {
			if issue.Severity == "error" {
				t.Errorf("unexpected error: %s", issue.Message)
			}
		}
	})
}

func TestValidatorServices(t *testing.T) {
	v := NewSpecValidator()

	t.Run("port conflict", func(t *testing.T) {
		services := []Service{
			{Name: "api", Port: 8080},
			{Name: "web", Port: 8080},
		}
		issues := v.ValidateServices(services)
		found := false
		for _, issue := range issues {
			if issue.Rule == "service-port-conflict" {
				found = true
			}
		}
		if !found {
			t.Error("expected port conflict issue")
		}
	})

	t.Run("port out of range", func(t *testing.T) {
		services := []Service{
			{Name: "bad", Port: 99999},
		}
		issues := v.ValidateServices(services)
		found := false
		for _, issue := range issues {
			if issue.Rule == "service-port-range" {
				found = true
			}
		}
		if !found {
			t.Error("expected port range issue")
		}
	})

	t.Run("valid services", func(t *testing.T) {
		services := []Service{
			{Name: "api", Port: 8080},
			{Name: "web", Port: 3000},
		}
		issues := v.ValidateServices(services)
		for _, issue := range issues {
			if issue.Severity == "error" {
				t.Errorf("unexpected error: %s", issue.Message)
			}
		}
	})
}

func TestDetectCycles(t *testing.T) {
	t.Run("no cycle", func(t *testing.T) {
		graph := map[string][]string{
			"a": {"b"},
			"b": {"c"},
			"c": {},
		}
		cycles := detectCycles(graph)
		if len(cycles) != 0 {
			t.Errorf("expected no cycles, got %d", len(cycles))
		}
	})

	t.Run("simple cycle", func(t *testing.T) {
		graph := map[string][]string{
			"a": {"b"},
			"b": {"a"},
		}
		cycles := detectCycles(graph)
		if len(cycles) == 0 {
			t.Error("expected cycle to be detected")
		}
	})

	t.Run("three-node cycle", func(t *testing.T) {
		graph := map[string][]string{
			"a": {"b"},
			"b": {"c"},
			"c": {"a"},
		}
		cycles := detectCycles(graph)
		if len(cycles) == 0 {
			t.Error("expected cycle to be detected")
		}
	})
}
