package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestExprString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr string
		want string
	}{
		{name: "ident", expr: "int", want: "int"},
		{name: "star", expr: "*string", want: "*string"},
		{name: "selector", expr: "context.Context", want: "context.Context"},
		{name: "array", expr: "[]byte", want: "[]byte"},
		{name: "map", expr: "map[string]int", want: "map[string]int"},
		{name: "interface", expr: "any", want: "any"},
		{name: "func", expr: "func()", want: "func(...)"},
		{name: "pointer selector", expr: "*net.Conn", want: "*net.Conn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpr(tt.expr)
			if err != nil {
				t.Fatalf("parse expr %q: %v", tt.expr, err)
			}
			got := exprString(expr)
			if got != tt.want {
				t.Errorf("exprString(%q) = %q, want %q", tt.expr, got, tt.want)
			}
		})
	}
}

func TestParseFunc(t *testing.T) {
	t.Parallel()

	src := `package test
func Foo(a string, b int) error { return nil }
func (s *Service) Bar(ctx context.Context) {}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		want FuncInfo
	}{
		{
			name: "Foo",
			want: FuncInfo{
				Name:    "Foo",
				Params:  []string{"string", "int"},
				Results: []string{"error"},
			},
		},
		{
			name: "Bar",
			want: FuncInfo{
				Name:     "Bar",
				Params:   []string{"context.Context"},
				IsMethod: true,
				RecvType: "Service",
			},
		},
	}

	var i int
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if i >= len(tests) {
			break
		}
		t.Run(tests[i].name, func(t *testing.T) {
			got := parseFunc(fn)
			if got.Name != tests[i].want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tests[i].want.Name)
			}
			if got.IsMethod != tests[i].want.IsMethod {
				t.Errorf("IsMethod = %v, want %v", got.IsMethod, tests[i].want.IsMethod)
			}
			if got.RecvType != tests[i].want.RecvType {
				t.Errorf("RecvType = %q, want %q", got.RecvType, tests[i].want.RecvType)
			}
			if len(got.Params) != len(tests[i].want.Params) {
				t.Errorf("Params = %v, want %v", got.Params, tests[i].want.Params)
			}
			if len(got.Results) != len(tests[i].want.Results) {
				t.Errorf("Results = %v, want %v", got.Results, tests[i].want.Results)
			}
		})
		i++
	}
}

func TestDeduplicateFuncs(t *testing.T) {
	t.Parallel()

	funcs := []FuncInfo{
		{Name: "Alpha"},
		{Name: "Beta", IsMethod: true, RecvType: "S"},
		{Name: "Alpha"},
		{Name: "Gamma", IsMethod: true, RecvType: "S"},
	}

	got := deduplicateFuncs(funcs)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].Name != "Alpha" {
		t.Errorf("got[0].Name = %q, want %q", got[0].Name, "Alpha")
	}
	if got[2].Name != "S_Gamma" {
		t.Errorf("got[2].Name = %q, want %q", got[2].Name, "S_Gamma")
	}
}

func TestTemplateOutput(t *testing.T) {
	t.Parallel()

	info := &PackageInfo{
		Name: "mypkg",
		Funcs: []FuncInfo{
			{Name: "DoSomething", Params: []string{"string"}, Results: []string{"error"}},
		},
	}

	var buf strings.Builder
	if err := testTmpl.Execute(&buf, info); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "package mypkg") {
		t.Error("output missing package declaration")
	}
	if !strings.Contains(out, "TestDoSomething") {
		t.Error("output missing TestDoSomething function")
	}
	if strings.Contains(out, "TODO") {
		t.Error("output should not contain TODO")
	}
	if !strings.Contains(out, `t.Log("generated test for DoSomething")`) {
		t.Error("output missing generated test log")
	}
}

func TestAnalyzePackage_Invalid(t *testing.T) {
	t.Parallel()

	_, err := analyzePackage("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}
