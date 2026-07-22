package parser

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func FuzzParse(f *testing.F) {
	f.Add(`project: test`)
	f.Add(`project: test
modules:
  - name: core
    path: ./core`)
	f.Add(`project: test
modules:
  - name: auth
    dependencies: [core]
  - name: core`)
	f.Add(`{invalid yaml`)
	f.Add(`project: ""`)
	f.Add(`modules: []`)
	f.Add(`services:
  - name: api
    port: 8080
    endpoints:
      - method: GET
        path: /test`)
	f.Add(`version: 0.3.0
project: versioned
architecture:
  pattern: hexagonal`)

	f.Fuzz(func(t *testing.T, input string) {
		p := NewParser(".")
		doc, err := p.Parse(input)

		if err != nil {
			if doc != nil {
				t.Error("doc should be nil on error")
			}
			return
		}

		if doc == nil {
			t.Error("doc should not be nil on success")
			return
		}

		if doc.Project == "" && len(doc.Modules) == 0 && len(doc.Services) == 0 {
			if doc.Raw != input {
				t.Error("raw should match input")
			}
		}
	})
}

func FuzzParseYAMLNode(f *testing.F) {
	f.Add(`project: test`)
	f.Add(`key: value`)
	f.Add(`list:
  - item1
  - item2`)
	f.Add(`nested:
  inner: value`)
	f.Add(`null_val: null`)
	f.Add(`bool_val: true`)
	f.Add(`int_val: 42`)
	f.Add(`float_val: 3.14`)

	f.Fuzz(func(t *testing.T, input string) {
		var root yaml.Node
		if err := yaml.Unmarshal([]byte(input), &root); err != nil {
			return
		}

		if len(root.Content) == 0 {
			return
		}

		result, err := parseYAMLNode(root.Content[0])
		if err != nil {
			return
		}

		_ = result
	})
}

func FuzzVariableResolver(f *testing.F) {
	f.Add("${var}", "value")
	f.Add("$env{HOME}", "/home/user")
	f.Add("no variables here", "")
	f.Add("${a}-${b}", "x")
	f.Add("$env{NONEXISTENT}", "")

	f.Fuzz(func(t *testing.T, input, value string) {
		resolver := NewVariableResolver()
		resolver.SetVar("var", value)
		resolver.SetVar("a", value)
		resolver.SetVar("b", value)

		result, err := resolver.Resolve(input)
		if err != nil {
			t.Fatal(err)
		}

		if result == "" && input != "" {
			if !strings.Contains(input, "${") && !strings.Contains(input, "$env{") {
				if result != input {
					t.Errorf("expected %q, got %q", input, result)
				}
			}
		}
	})
}

func FuzzSchemaVersionParse(f *testing.F) {
	f.Add("0.1.0")
	f.Add("0.3.0")
	f.Add("1.0.0")
	f.Add("v2.0.0")
	f.Add("invalid")
	f.Add("")
	f.Add("1.2.3.4")
	f.Add("0.0")
	f.Add("-1.0.0")
	f.Add("abc.def.ghi")

	f.Fuzz(func(t *testing.T, input string) {
		result, err := ParseSchemaVersion(input)
		if err != nil {
			return
		}

		if result.Major < 0 {
			t.Error("major version should not be negative")
		}
		if result.Minor < 0 {
			t.Error("minor version should not be negative")
		}
		if result.Patch < 0 {
			t.Error("patch version should not be negative")
		}
	})
}

func FuzzValidateModules(f *testing.F) {
	f.Add("auth", "core")
	f.Add("api", "auth")
	f.Add("core", "")
	f.Add("", "something")
	f.Add("a", "b")
	f.Add("b", "a")

	f.Fuzz(func(t *testing.T, name, dep string) {
		modules := []Module{
			{Name: name, Dependencies: []string{dep}},
		}

		v := NewSpecValidator()
		issues := v.ValidateModules(modules)

		for _, issue := range issues {
			if issue.Severity != "error" && issue.Severity != "warning" {
				t.Errorf("invalid severity: %s", issue.Severity)
			}
			if issue.Rule == "" {
				t.Error("rule should not be empty")
			}
		}
	})
}
