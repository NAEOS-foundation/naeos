package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFuncRegistryUpper(t *testing.T) {
	r := NewFuncRegistry()
	result := r.Resolve("$fn{upper(hello)}")
	if result != "HELLO" {
		t.Errorf("expected HELLO, got %s", result)
	}
}

func TestFuncRegistryLower(t *testing.T) {
	r := NewFuncRegistry()
	result := r.Resolve("$fn{lower(WORLD)}")
	if result != "world" {
		t.Errorf("expected world, got %s", result)
	}
}

func TestFuncRegistrySlug(t *testing.T) {
	r := NewFuncRegistry()
	result := r.Resolve("$fn{slug(Hello World)}")
	if result != "hello-world" {
		t.Errorf("expected hello-world, got %s", result)
	}
}

func TestFuncRegistryDefault(t *testing.T) {
	r := NewFuncRegistry()
	result := r.Resolve("$fn{default(,fallback)}")
	if result != "fallback" {
		t.Errorf("expected fallback, got %s", result)
	}
}

func TestFuncRegistryDefaultHasValue(t *testing.T) {
	r := NewFuncRegistry()
	result := r.Resolve("$fn{default(value,fallback)}")
	if result != "value" {
		t.Errorf("expected value, got %s", result)
	}
}

func TestFuncRegistryDefaultNoComma(t *testing.T) {
	r := NewFuncRegistry()
	result := r.Resolve("$fn{default(only)}")
	if result != "only" {
		t.Errorf("expected only, got %s", result)
	}
}

func TestFuncRegistryLen(t *testing.T) {
	r := NewFuncRegistry()
	result := r.Resolve("$fn{len(hello)}")
	if result != "5" {
		t.Errorf("expected 5, got %s", result)
	}
}

func TestFuncRegistryLenEmpty(t *testing.T) {
	r := NewFuncRegistry()
	result := r.Resolve("$fn{len(   )}")
	if result != "0" {
		t.Errorf("expected 0, got %s", result)
	}
}

func TestFuncRegistryCoalesce(t *testing.T) {
	r := NewFuncRegistry()
	result := r.Resolve("$fn{coalesce(, ,first)}")
	if result != "first" {
		t.Errorf("expected first, got %s", result)
	}
}

func TestFuncRegistryCoalesceAllEmpty(t *testing.T) {
	r := NewFuncRegistry()
	result := r.Resolve("$fn{coalesce(,,)}")
	if result != "" {
		t.Errorf("expected empty, got %s", result)
	}
}

func TestFuncRegistryCustom(t *testing.T) {
	r := NewFuncRegistry()
	r.Register("double", func(args string) string {
		return args + args
	})
	result := r.Resolve("$fn{double(x)}")
	if result != "xx" {
		t.Errorf("expected xx, got %s", result)
	}
}

func TestFuncRegistryUnregistered(t *testing.T) {
	r := NewFuncRegistry()
	result := r.Resolve("$fn{unknown(x)}")
	if result != "$fn{unknown(x)}" {
		t.Errorf("expected unchanged, got %s", result)
	}
}

func TestFuncRegistryMultipleInOrder(t *testing.T) {
	r := NewFuncRegistry()
	result := r.Resolve("$fn{upper(a)} $fn{lower(B)}")
	if result != "A b" {
		t.Errorf("expected 'A b', got %s", result)
	}
}

func TestConditionalResolverEqualsTrue(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("env", "prod")
	input := "$if{env == prod}\nactive\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "active" {
		t.Errorf("expected 'active', got %q", result)
	}
}

func TestConditionalResolverEqualsFalse(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("env", "dev")
	input := "$if{env == prod}\nactive\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestConditionalResolverEqualsMissingKey(t *testing.T) {
	r := NewConditionalResolver()
	input := "$if{missing == val}\ncontent\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestConditionalResolverEqualsQuotedValue(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("mode", "secure")
	input := `$if{mode == "secure"}`
	input += "\nenabled\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "enabled" {
		t.Errorf("expected enabled, got %q", result)
	}
}

func TestConditionalResolverNotEqualsTrue(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("env", "dev")
	input := "$if{env != prod}\nactive\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "active" {
		t.Errorf("expected active, got %q", result)
	}
}

func TestConditionalResolverNotEqualsFalse(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("env", "prod")
	input := "$if{env != prod}\nactive\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestConditionalResolverNotEqualsMissing(t *testing.T) {
	r := NewConditionalResolver()
	input := "$if{missing != val}\ncontent\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "content" {
		t.Errorf("expected content, got %q", result)
	}
}

func TestConditionalResolverDefinedTrue(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("feature_x", "true")
	input := "$if{defined:feature_x}\npresent\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "present" {
		t.Errorf("expected present, got %q", result)
	}
}

func TestConditionalResolverDefinedFalse(t *testing.T) {
	r := NewConditionalResolver()
	input := "$if{defined:missing}\npresent\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestConditionalResolverNegation(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("debug_disabled", "")
	input := "$if{!debug_disabled}\nactive\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "active" {
		t.Errorf("expected active, got %q", result)
	}
}

func TestConditionalResolverNegationTruthy(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("debug", "true")
	input := "$if{!debug}\nactive\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) == "active" {
		t.Error("expected empty when negating truthy value")
	}
}

func TestConditionalResolverTruthiness(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("flag", "true")
	input := "$if{flag}\nyes\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "yes" {
		t.Errorf("expected yes, got %q", result)
	}
}

func TestConditionalResolverEnvOne(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("count", "1")
	input := "$if{count}\nyes\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "yes" {
		t.Errorf("expected yes, got %q", result)
	}
}

func TestConditionalResolverEnvNotEmpty(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("name", "anything")
	input := "$if{name}\nyes\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "yes" {
		t.Errorf("expected yes, got %q", result)
	}
}

func TestConditionalResolverEnvFalsey(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("flag", "false")
	input := "$if{flag}\nyes\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "yes" {
		t.Errorf("expected yes for 'false' string, got %q", result)
	}
}

func TestConditionalResolverMissingKey(t *testing.T) {
	r := NewConditionalResolver()
	input := "$if{missing}\nyes\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "" {
		t.Errorf("expected empty for missing key, got %q", result)
	}
}

func TestConditionalResolverBlockNotAtStart(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("env", "prod")
	input := "header\n$if{env == prod}\nbody\n$endif\nfooter"
	result := r.Resolve(input)
	expected := "header\nbody\nfooter"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestConditionalResolverMultipleBlocks(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnv("feature_x", "enabled")
	input := "$if{feature_x == enabled}\nA\n$endif\n$if{feature_y == enabled}\nB\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "A" {
		t.Errorf("expected A only, got %q", result)
	}
}

func TestConditionalResolverSetEnvs(t *testing.T) {
	r := NewConditionalResolver()
	r.SetEnvs(map[string]string{"feature": "enabled"})
	input := "$if{feature == enabled}\nyes\n$endif"
	result := r.Resolve(input)
	if strings.TrimSpace(result) != "yes" {
		t.Errorf("expected yes, got %q", result)
	}
}

func TestIncludeResolverPathTraversal(t *testing.T) {
	r := NewIncludeResolver("")
	_, err := r.ResolveIncludes(`$include{../etc/passwd}`)
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestIncludeResolverFileNotFound(t *testing.T) {
	dir := t.TempDir()
	r := NewIncludeResolver(dir)
	_, err := r.ResolveIncludes(`$include{nonexistent.yaml}`)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestIncludeResolverSimple(t *testing.T) {
	dir := t.TempDir()
	includePath := filepath.Join(dir, "vars.yaml")
	os.WriteFile(includePath, []byte("key: value\n"), 0o644)

	r := NewIncludeResolver(dir)
	result, err := r.ResolveIncludes(`config: $include{vars.yaml}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "key: value") {
		t.Errorf("expected included content, got %q", result)
	}
}

func TestIncludeResolverCaching(t *testing.T) {
	dir := t.TempDir()
	includePath := filepath.Join(dir, "shared.yaml")
	os.WriteFile(includePath, []byte("cached"), 0o644)

	r := NewIncludeResolver(dir)
	result1, _ := r.ResolveIncludes(`$include{shared.yaml}`)
	result2, err := r.ResolveIncludes(`$include{shared.yaml}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result1 != result2 {
		t.Errorf("cached result should match")
	}
}

func TestIncludeResolverNested(t *testing.T) {
	dir := t.TempDir()
	inner := filepath.Join(dir, "inner.yaml")
	os.WriteFile(inner, []byte("nested"), 0o644)
	outer := filepath.Join(dir, "outer.yaml")
	os.WriteFile(outer, []byte("$include{inner.yaml}"), 0o644)

	r := NewIncludeResolver(dir)
	result, err := r.ResolveIncludes(`$include{outer.yaml}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "nested") {
		t.Errorf("expected nested content, got %q", result)
	}
}

func TestIncludeResolverChain(t *testing.T) {
	dir := t.TempDir()
	prev := ""
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("level_%d.yaml", i)
		path := filepath.Join(dir, name)
		var content string
		if prev != "" {
			content = fmt.Sprintf("$include{%s}", prev)
		} else {
			content = "leaf"
		}
		os.WriteFile(path, []byte(content), 0o644)
		prev = name
	}
	r := NewIncludeResolver(dir)
	result, err := r.ResolveIncludes("$include{level_9.yaml}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "leaf" {
		t.Errorf("expected leaf, got %s", result)
	}
}

func TestIncludeResolverDepthExceeded(t *testing.T) {
	dir := t.TempDir()
	files := make([]string, 12)
	for i := range files {
		name := fmt.Sprintf("f_%d.yaml", i)
		files[i] = filepath.Join(dir, name)
	}
	for i := 0; i < 11; i++ {
		os.WriteFile(files[i], []byte(fmt.Sprintf("$include{f_%d.yaml}", i+1)), 0o644)
	}
	os.WriteFile(files[11], []byte("leaf"), 0o644)

	r := NewIncludeResolver(dir)
	_, err := r.ResolveIncludes("$include{f_0.yaml}")
	if err == nil {
		t.Error("expected depth exceeded error")
	}
}

func TestIncludeResolverNoBaseDir(t *testing.T) {
	r := NewIncludeResolver("")
	_, err := r.ResolveIncludes(`$include{/nonexistent/file.yaml}`)
	if err == nil {
		t.Error("expected error")
	}
}

func TestImportResolverPathTraversal(t *testing.T) {
	r := NewImportResolver("")
	_, err := r.ResolveImports(`$import{../etc/passwd}`)
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestImportResolverFileNotFound(t *testing.T) {
	dir := t.TempDir()
	r := NewImportResolver(dir)
	_, err := r.ResolveImports(`$import{nonexistent.yaml}`)
	if err == nil {
		t.Error("expected error")
	}
}

func TestImportResolverSimple(t *testing.T) {
	dir := t.TempDir()
	importPath := filepath.Join(dir, "lib.yaml")
	os.WriteFile(importPath, []byte("data: hello"), 0o644)

	r := NewImportResolver(dir)
	result, err := r.ResolveImports(`$import{lib.yaml}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "data: hello" {
		t.Errorf("expected 'data: hello', got %q", result)
	}
}

func TestImportResolverWithSection(t *testing.T) {
	dir := t.TempDir()
	importPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(importPath, []byte("server:\n  port: 8080\n  host: localhost"), 0o644)

	r := NewImportResolver(dir)
	result, err := r.ResolveImports(`$import{config.yaml::server}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "port: 8080") {
		t.Errorf("expected server section, got %q", result)
	}
}

func TestImportResolverSectionNotFound(t *testing.T) {
	dir := t.TempDir()
	importPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(importPath, []byte("server:\n  port: 8080"), 0o644)

	r := NewImportResolver(dir)
	_, err := r.ResolveImports(`$import{config.yaml::missing}`)
	if err == nil {
		t.Error("expected error for missing section")
	}
}

func TestImportResolverCaching(t *testing.T) {
	dir := t.TempDir()
	importPath := filepath.Join(dir, "lib.yaml")
	os.WriteFile(importPath, []byte("cached: data"), 0o644)

	r := NewImportResolver(dir)
	result1, _ := r.ResolveImports(`$import{lib.yaml}`)
	result2, err := r.ResolveImports(`$import{lib.yaml}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result1 != result2 {
		t.Errorf("cached result should match")
	}
}

func TestImportResolverDepthExceeded(t *testing.T) {
	dir := t.TempDir()
	self := filepath.Join(dir, "self.yaml")
	os.WriteFile(self, []byte("$import{self.yaml}"), 0o644)

	r := NewImportResolver(dir)
	_, err := r.ResolveImports(`$import{self.yaml}`)
	if err == nil {
		t.Error("expected depth exceeded error")
	}
}

func TestImportResolverAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	absPath := filepath.Join(dir, "abs.yaml")
	os.WriteFile(absPath, []byte("absolute: path"), 0o644)

	r := NewImportResolver(dir)
	result, err := r.ResolveImports(`$import{` + absPath + `}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "absolute: path" {
		t.Errorf("expected content, got %q", result)
	}
}

func TestImportResolverNoBaseDir(t *testing.T) {
	r := NewImportResolver("")
	_, err := r.ResolveImports(`$import{/nonexistent/file.yaml}`)
	if err == nil {
		t.Error("expected error")
	}
}
