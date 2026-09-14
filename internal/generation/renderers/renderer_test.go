// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package renderers

import (
	"os"
	"strings"
	"testing"
	"text/template"
)

func TestRenderSimpleTemplate(t *testing.T) {
	r := NewRenderer()
	result, err := r.Render("hello {{.Name}}", map[string]string{"Name": "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", string(result))
	}
}

func TestRenderNamedTemplate(t *testing.T) {
	r := NewRenderer()
	result, err := r.RenderNamed("greeting", "hi {{.Name}}", map[string]string{"Name": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "hi test" {
		t.Fatalf("expected 'hi test', got '%s'", string(result))
	}
}

func TestRenderInvalidTemplate(t *testing.T) {
	r := NewRenderer()
	_, err := r.Render("hello {{.Name", nil)
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}

func TestRenderTemplateWithFunctions(t *testing.T) {
	funcs := template.FuncMap{
		"upper": func(s string) string {
			result := ""
			for _, c := range s {
				if c >= 'a' && c <= 'z' {
					result += string(c - 32)
				} else {
					result += string(c)
				}
			}
			return result
		},
	}
	result, err := RenderWithFuncs("test", "hello {{upper .Name}}", map[string]string{"Name": "world"}, funcs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "hello WORLD" {
		t.Fatalf("expected 'hello WORLD', got '%s'", string(result))
	}
}

func TestRenderTemplateWithNestedData(t *testing.T) {
	type Data struct {
		Project string
		Modules []string
	}
	data := Data{
		Project: "my-project",
		Modules: []string{"auth", "user", "api"},
	}
	tmpl := "project: {{.Project}}\nmodules: {{range .Modules}}{{.}} {{end}}"
	result, err := RenderTemplate("test", tmpl, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "project: my-project\nmodules: auth user api "
	if string(result) != expected {
		t.Fatalf("expected '%s', got '%s'", expected, string(result))
	}
}

func TestTemplateData(t *testing.T) {
	data := TemplateData{
		Project: "acme-api",
		Module:  "auth",
		Port:    8080,
	}
	tmpl := "{{.Project}} {{.Module}} :{{.Port}}"
	result, err := RenderTemplate("test", tmpl, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "acme-api auth :8080" {
		t.Fatalf("unexpected result: %s", string(result))
	}
}

func TestRenderEmptyTemplate(t *testing.T) {
	r := NewRenderer()
	result, err := r.Render("", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got '%s'", string(result))
	}
}

func TestRenderNilData(t *testing.T) {
	r := NewRenderer()
	result, err := r.Render("static text", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "static text" {
		t.Fatalf("expected 'static text', got '%s'", string(result))
	}
}

func TestRenderWithFuncsInvalidTemplate(t *testing.T) {
	funcs := template.FuncMap{
		"upper": strings.ToUpper,
	}
	_, err := RenderWithFuncs("test", "hello {{.Name", nil, funcs)
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}

func TestRenderWithFuncsExecuteError(t *testing.T) {
	funcs := template.FuncMap{}
	_, err := RenderWithFuncs("test", "hello {{.Name}}", nil, funcs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderWithFuncsNilFuncMap(t *testing.T) {
	_, err := RenderWithFuncs("test", "hello {{.Name}}", map[string]string{"Name": "world"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderNamedEmptyTemplate(t *testing.T) {
	r := NewRenderer()
	result, err := r.RenderNamed("empty", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got '%s'", string(result))
	}
}

func TestRenderNamedInvalidTemplate(t *testing.T) {
	r := NewRenderer()
	_, err := r.RenderNamed("bad", "unclosed {{", nil)
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}

func TestRenderTemplateDataFull(t *testing.T) {
	data := TemplateData{
		Project:  "my-proj",
		Module:   "auth",
		Service:  "api",
		Port:     9090,
		Kind:     "grpc",
		Version:  "2.0.0",
		Package:  "com.example",
		Language: "go",
		Attributes: map[string]string{
			"env": "prod",
		},
	}
	tmpl := "{{.Project}}/{{.Module}} {{.Service}}:{{.Port}} {{.Kind}} {{.Version}} {{.Package}} {{.Language}} env={{index .Attributes \"env\"}}"
	result, err := RenderTemplate("test", tmpl, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "my-proj/auth api:9090 grpc 2.0.0 com.example go env=prod"
	if string(result) != expected {
		t.Fatalf("expected '%s', got '%s'", expected, string(result))
	}
}

func TestRenderWithFuncsEmptyFuncMap(t *testing.T) {
	funcs := template.FuncMap{}
	result, err := RenderWithFuncs("test", "hello world", nil, funcs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", string(result))
	}
}

func TestRenderTemplateConditional(t *testing.T) {
	type Data struct {
		Show bool
		Name string
	}
	tmpl := "{{if .Show}}{{.Name}}{{else}}hidden{{end}}"
	result, err := RenderTemplate("test", tmpl, Data{Show: true, Name: "visible"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "visible" {
		t.Fatalf("expected 'visible', got '%s'", string(result))
	}
	result, err = RenderTemplate("test", tmpl, Data{Show: false, Name: "visible"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "hidden" {
		t.Fatalf("expected 'hidden', got '%s'", string(result))
	}
}

func TestNewRendererReturnsDefaultRenderer(t *testing.T) {
	r := NewRenderer()
	if _, ok := r.(DefaultRenderer); !ok {
		t.Fatal("expected DefaultRenderer type")
	}
}

func TestGoRenderer(t *testing.T) {
	r := NewGoRenderer()
	result, err := r.RenderCode("go", `package main

func main(){{"{"}}
println("hello")
{{"}"}}`, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(result), "package main") {
		t.Error("expected go code")
	}
}

func TestGoRendererFormat(t *testing.T) {
	r := NewGoRenderer()
	input := []byte("package main\nfunc main(){{\"{\"}}println(\"hello\"){{\"}\"}}\n")
	_, _ = r.FormatCode("go", input)
}

func TestTypeScriptRenderer(t *testing.T) {
	r := NewTypeScriptRenderer()
	result, err := r.RenderCode("ts", "const x = {{.Value}}", map[string]any{"Value": 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "const x = 42" {
		t.Errorf("unexpected result: %s", string(result))
	}
}

func TestTypeScriptRendererFormat(t *testing.T) {
	r := NewTypeScriptRenderer()
	result, err := r.FormatCode("ts", []byte("code"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "code" {
		t.Error("expected unchanged code")
	}
}

func TestPythonRenderer(t *testing.T) {
	r := NewPythonRenderer()
	result, err := r.RenderCode("py", "x = {{.Value}}", map[string]any{"Value": "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "x = hello" {
		t.Errorf("unexpected result: %s", string(result))
	}
}

func TestPythonRendererFormat(t *testing.T) {
	r := NewPythonRenderer()
	result, err := r.FormatCode("py", []byte("code"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "code" {
		t.Error("expected unchanged code")
	}
}

func TestNewCodeRenderer(t *testing.T) {
	tests := []struct {
		lang string
		want string
	}{
		{"go", "GoRenderer"},
		{"golang", "GoRenderer"},
		{"typescript", "TypeScriptRenderer"},
		{"ts", "TypeScriptRenderer"},
		{"tsx", "TypeScriptRenderer"},
		{"python", "PythonRenderer"},
		{"py", "PythonRenderer"},
		{"rust", ""},
	}
	for _, tc := range tests {
		r := NewCodeRenderer(tc.lang)
		if tc.want == "" {
			if r != nil {
				t.Errorf("expected nil for %q", tc.lang)
			}
		} else {
			if r == nil {
				t.Errorf("expected renderer for %q", tc.lang)
			}
		}
	}
}

func TestFileGenerator(t *testing.T) {
	fg := NewFileGenerator()
	fg.AddFile("main.go", []byte("package main"))
	fg.AddExecutable("run.sh", []byte("#!/bin/bash"))
	if fg.Count() != 2 {
		t.Errorf("expected 2, got %d", fg.Count())
	}
	if !fg.Contains("main.go") {
		t.Error("expected to contain main.go")
	}
	if fg.Contains("other.go") {
		t.Error("should not contain other.go")
	}
	paths := fg.Paths()
	if len(paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(paths))
	}
	files := fg.Files()
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestFileGeneratorWriteToDir(t *testing.T) {
	dir := t.TempDir()
	fg := NewFileGenerator()
	fg.AddFile("sub/hello.txt", []byte("world"))
	if err := fg.WriteToDir(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dir + "/sub/hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "world" {
		t.Errorf("expected 'world', got '%s'", string(data))
	}
}

func TestMultiRenderer(t *testing.T) {
	mr := NewMultiRenderer()
	mr.Register("go", NewGoRenderer())
	mr.Register("ts", NewTypeScriptRenderer())

	langs := mr.Languages()
	if len(langs) != 2 {
		t.Errorf("expected 2 languages, got %d", len(langs))
	}

	result, err := mr.Render("ts", "x = {{.V}}", map[string]any{"V": 1})
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != "x = 1" {
		t.Errorf("unexpected: %s", string(result))
	}

	_, err = mr.Render("rust", "code", nil)
	if err == nil {
		t.Error("expected error for unknown language")
	}

	_, err = mr.Format("rust", []byte("code"))
	if err == nil {
		t.Error("expected error for unknown language")
	}
}

func TestRenderPipeline(t *testing.T) {
	rp := NewRenderPipeline()
	if rp.StepCount() != 0 {
		t.Error("expected 0 steps")
	}

	rp.AddStep("step1", func(data any) ([]byte, error) {
		return []byte("hello"), nil
	})
	rp.AddStep("step2", func(data any) ([]byte, error) {
		return []byte("world"), nil
	})

	if rp.StepCount() != 2 {
		t.Errorf("expected 2 steps, got %d", rp.StepCount())
	}

	results, err := rp.Execute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if string(results[0]) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(results[0]))
	}
	if string(results[1]) != "world" {
		t.Errorf("expected 'world', got '%s'", string(results[1]))
	}
}

func TestRenderPipelineError(t *testing.T) {
	rp := NewRenderPipeline()
	rp.AddStep("fail", func(data any) ([]byte, error) {
		return nil, errTest
	})
	_, err := rp.Execute(nil)
	if err == nil {
		t.Error("expected error")
	}
}

var errTest = &testErr{"test error"}

type testErr struct{ msg string }

func (e *testErr) Error() string { return e.msg }

func TestTemplateLibrary(t *testing.T) {
	tl := NewTemplateLibrary()
	tl.Register("greeting", "hello {{.Name}}")
	tl.Register("farewell", "bye {{.Name}}")

	if !tl.Has("greeting") {
		t.Error("expected to have greeting")
	}
	if tl.Has("unknown") {
		t.Error("should not have unknown")
	}

	names := tl.Names()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}

	tmpl, ok := tl.Get("greeting")
	if !ok || tmpl != "hello {{.Name}}" {
		t.Error("unexpected template")
	}

	result, err := tl.Render("greeting", map[string]string{"Name": "world"})
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != "hello world" {
		t.Errorf("unexpected: %s", string(result))
	}

	_, err = tl.Render("unknown", nil)
	if err == nil {
		t.Error("expected error for unknown template")
	}

	_, ok = tl.Get("unknown")
	if ok {
		t.Error("expected not found")
	}
}
