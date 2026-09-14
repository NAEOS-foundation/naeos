// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package hcl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return path
}

func basicSpec() *Spec {
	return &Spec{
		Project:  Project{Name: "myapp", Version: "1.0"},
		Services: map[string]Service{"web": {Image: "nginx:latest", Port: 80, Type: "http"}},
		Infra:    Infra{Engine: "docker"},
	}
}

// ===========================================================================
// ParseError tests
// ===========================================================================

func TestParseError_Error(t *testing.T) {
	e := &ParseError{FileName: "test.hcl", Line: 5, Column: 12, Message: "bad token"}
	got := e.Error()
	if !strings.Contains(got, "test.hcl:5:12") {
		t.Errorf("expected filename:line:col, got %s", got)
	}
	if !strings.Contains(got, "bad token") {
		t.Errorf("expected message, got %s", got)
	}
}

func TestParseError_ErrorNoFile(t *testing.T) {
	e := &ParseError{Line: 3, Column: 1, Message: "oops"}
	got := e.Error()
	if strings.Contains(got, ":") == false {
		t.Error("expected formatted string")
	}
}

// ===========================================================================
// Parse tests
// ===========================================================================

func TestParse_Minimal(t *testing.T) {
	input := []byte(`project "test" {
  version = "0.1"
}
service "api" {
  image = "myimg:latest"
  port  = 8080
  type  = "http"
}
`)
	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Project.Name != "test" {
		t.Errorf("project name = %q", spec.Project.Name)
	}
	if spec.Project.Version != "0.1" {
		t.Errorf("version = %q", spec.Project.Version)
	}
	svc, ok := spec.Services["api"]
	if !ok {
		t.Fatal("missing service api")
	}
	if svc.Image != "myimg:latest" {
		t.Errorf("image = %q", svc.Image)
	}
	if svc.Port != 8080 {
		t.Errorf("port = %d", svc.Port)
	}
	if svc.Type != "http" {
		t.Errorf("type = %q", svc.Type)
	}
}

func TestParse_WithDescription(t *testing.T) {
	input := []byte(`project "demo" {
  description = "A demo project"
}
`)
	spec, err := Parse(input, "demo.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Project.Description != "A demo project" {
		t.Errorf("description = %q", spec.Project.Description)
	}
}

func TestParse_ProjectNameOverride(t *testing.T) {
	input := []byte(`project "p1" {
  name = "overridden"
}
`)
	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Project.Name != "overridden" {
		t.Errorf("expected overridden name, got %q", spec.Project.Name)
	}
}

func TestParse_MultipleServices(t *testing.T) {
	input := []byte(`project "multi" {}
service "web" {
  image = "nginx"
  port  = 80
  type  = "http"
}
service "db" {
  image = "postgres:14"
  port  = 5432
  type  = "tcp"
}
`)
	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spec.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(spec.Services))
	}
	if spec.Services["web"].Port != 80 {
		t.Errorf("web port = %d", spec.Services["web"].Port)
	}
	if spec.Services["db"].Port != 5432 {
		t.Errorf("db port = %d", spec.Services["db"].Port)
	}
}

func TestParse_Infra(t *testing.T) {
	input := []byte(`infra "main" {
  engine = "podman"
}
`)
	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Infra.Engine != "podman" {
		t.Errorf("engine = %q", spec.Infra.Engine)
	}
}

func TestParse_CommentsIgnored(t *testing.T) {
	input := []byte(`# this is a comment
// so is this
project "c" {
  version = "1"
}
`)
	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Project.Name != "c" {
		t.Errorf("project name = %q", spec.Project.Name)
	}
}

func TestParse_EmptyLines(t *testing.T) {
	input := []byte(`

project "e" {
  version = "2"
}

`)
	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Project.Version != "2" {
		t.Errorf("version = %q", spec.Project.Version)
	}
}

func TestParse_InvalidPort(t *testing.T) {
	input := []byte(`project "bad" {}
service "x" {
  port = "notanumber"
  type = "http"
}
`)
	_, err := Parse(input, "test.hcl")
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
	if !strings.Contains(err.Error(), "invalid port") {
		t.Errorf("error should mention invalid port: %v", err)
	}
}

func TestParse_InvalidReplicas(t *testing.T) {
	input := []byte(`project "bad" {}
service "x" {
  type     = "http"
  replicas = "abc"
}
`)
	_, err := Parse(input, "test.hcl")
	if err == nil {
		t.Fatal("expected error for invalid replicas")
	}
	if !strings.Contains(err.Error(), "invalid replicas") {
		t.Errorf("error should mention invalid replicas: %v", err)
	}
}

func TestParse_EnvSubBlock(t *testing.T) {
	input := []byte(`project "envproj" {}
service "web" {
  image = "nginx"
  type  = "http"
  env "vars" {
    DATABASE_URL = "postgres://localhost/db"
    LOG_LEVEL    = "debug"
  }
}
`)
	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc := spec.Services["web"]
	if svc.Env == nil {
		t.Fatal("expected env map")
	}
	if svc.Env["DATABASE_URL"] != "postgres://localhost/db" {
		t.Errorf("DATABASE_URL = %q", svc.Env["DATABASE_URL"])
	}
	if svc.Env["LOG_LEVEL"] != "debug" {
		t.Errorf("LOG_LEVEL = %q", svc.Env["LOG_LEVEL"])
	}
}

func TestParse_Volumes(t *testing.T) {
	input := []byte(`project "vol" {}
service "data" {
  type    = "volume"
  volumes = ["/host/data", "/container/data"]
}
`)
	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc := spec.Services["data"]
	if len(svc.Volumes) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(svc.Volumes))
	}
	if svc.Volumes[0] != "/host/data" {
		t.Errorf("volume[0] = %q", svc.Volumes[0])
	}
}

func TestParse_Depends(t *testing.T) {
	input := []byte(`project "dep" {}
service "db" {
  type = "tcp"
}
service "api" {
  type    = "http"
  depends = ["db"]
}
`)
	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	api := spec.Services["api"]
	if len(api.Depends) != 1 || api.Depends[0] != "db" {
		t.Errorf("depends = %v", api.Depends)
	}
}

func TestParse_SingleVolume(t *testing.T) {
	input := []byte(`project "sv" {}
service "app" {
  type    = "volume"
  volumes = "/data"
}
`)
	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc := spec.Services["app"]
	if len(svc.Volumes) != 1 || svc.Volumes[0] != "/data" {
		t.Errorf("volumes = %v", svc.Volumes)
	}
}

func TestParse_SingleDepend(t *testing.T) {
	input := []byte(`project "sd" {}
service "worker" {
  type   = "worker"
  depends = "db"
}
`)
	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc := spec.Services["worker"]
	if len(svc.Depends) != 1 || svc.Depends[0] != "db" {
		t.Errorf("depends = %v", svc.Depends)
	}
}

func TestParse_Replicas(t *testing.T) {
	input := []byte(`project "rep" {}
service "web" {
  type     = "http"
  replicas = 3
}
`)
	spec, err := Parse(input, "test.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Services["web"].Replicas != 3 {
		t.Errorf("replicas = %d", spec.Services["web"].Replicas)
	}
}

func TestParse_EmptyInput(t *testing.T) {
	spec, err := Parse([]byte(""), "empty.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Project.Name != "" {
		t.Errorf("expected empty project name")
	}
}

// ===========================================================================
// ParseFile tests
// ===========================================================================

func TestParseFile_Success(t *testing.T) {
	content := `project "filetest" {
  version = "2.0"
}
service "api" {
  image = "myimg"
  port  = 3000
  type  = "http"
}
`
	path := writeTempFile(t, "test.hcl", content)
	spec, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Project.Name != "filetest" {
		t.Errorf("project name = %q", spec.Project.Name)
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/file.hcl")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// ===========================================================================
// ToYAML tests
// ===========================================================================

func TestToYAML_Basic(t *testing.T) {
	spec := basicSpec()
	out, err := ToYAML(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "name: myapp") {
		t.Error("missing project name")
	}
	if !strings.Contains(s, "version: 1.0") {
		t.Error("missing version")
	}
	if !strings.Contains(s, "image: nginx:latest") {
		t.Error("missing image")
	}
	if !strings.Contains(s, "engine: docker") {
		t.Error("missing engine")
	}
}

func TestToYAML_Minimal(t *testing.T) {
	spec := &Spec{
		Project:  Project{Name: "m"},
		Services: map[string]Service{"a": {Type: "grpc"}},
	}
	out, err := ToYAML(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "name: m") {
		t.Error("missing project name")
	}
	if !strings.Contains(s, "type: grpc") {
		t.Error("missing type")
	}
}

func TestToYAML_EmptySpec(t *testing.T) {
	spec := &Spec{Services: make(map[string]Service)}
	out, err := ToYAML(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "name: ") {
		t.Error("missing name line")
	}
}

// ===========================================================================
// Validate tests
// ===========================================================================

func TestValidate_ValidSpec(t *testing.T) {
	spec := basicSpec()
	errs := Validate(spec)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidate_MissingProjectName(t *testing.T) {
	spec := &Spec{
		Services: map[string]Service{"x": {Type: "http"}},
	}
	errs := Validate(spec)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "project name is required") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing project name error, got %v", errs)
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	spec := &Spec{
		Project:  Project{Name: "p"},
		Services: map[string]Service{"x": {Port: 99999, Type: "http"}},
	}
	errs := Validate(spec)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "invalid port") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected invalid port error, got %v", errs)
	}
}

func TestValidate_MissingType(t *testing.T) {
	spec := &Spec{
		Project:  Project{Name: "p"},
		Services: map[string]Service{"x": {Image: "img"}},
	}
	errs := Validate(spec)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "missing a type") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing type error, got %v", errs)
	}
}

func TestValidate_NegativeReplicas(t *testing.T) {
	spec := &Spec{
		Project:  Project{Name: "p"},
		Services: map[string]Service{"x": {Type: "http", Replicas: -1}},
	}
	errs := Validate(spec)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "negative replica") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected negative replica error, got %v", errs)
	}
}

func TestValidate_SelfDependency(t *testing.T) {
	spec := &Spec{
		Project:  Project{Name: "p"},
		Services: map[string]Service{"x": {Type: "http", Depends: []string{"x"}}},
	}
	errs := Validate(spec)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "depends on itself") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected self-dependency error, got %v", errs)
	}
}

func TestValidate_UnknownDependency(t *testing.T) {
	spec := &Spec{
		Project:  Project{Name: "p"},
		Services: map[string]Service{"x": {Type: "http", Depends: []string{"ghost"}}},
	}
	errs := Validate(spec)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown service") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected unknown dependency error, got %v", errs)
	}
}

func TestValidate_PortZero(t *testing.T) {
	spec := &Spec{
		Project:  Project{Name: "p"},
		Services: map[string]Service{"x": {Type: "http", Port: 0}},
	}
	errs := Validate(spec)
	for _, e := range errs {
		if strings.Contains(e.Message, "invalid port") {
			t.Errorf("port 0 should be valid: %v", e.Message)
		}
	}
}

func TestValidate_Port65535(t *testing.T) {
	spec := &Spec{
		Project:  Project{Name: "p"},
		Services: map[string]Service{"x": {Type: "http", Port: 65535}},
	}
	errs := Validate(spec)
	for _, e := range errs {
		if strings.Contains(e.Message, "invalid port") {
			t.Errorf("port 65535 should be valid: %v", e.Message)
		}
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	spec := &Spec{
		Services: map[string]Service{
			"a": {Port: 70000},
			"b": {},
		},
	}
	errs := Validate(spec)
	if len(errs) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %v", len(errs), errs)
	}
}

// ===========================================================================
// ToJSON tests
// ===========================================================================

func TestToJSON_Basic(t *testing.T) {
	spec := basicSpec()
	out, err := ToJSON(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `"name": "myapp"`) {
		t.Error("missing project name in JSON")
	}
	if !strings.Contains(s, `"image": "nginx:latest"`) {
		t.Error("missing image in JSON")
	}
	if !strings.Contains(s, `"engine": "docker"`) {
		t.Error("missing engine in JSON")
	}
}

func TestToJSON_EmptySpec(t *testing.T) {
	spec := &Spec{Services: make(map[string]Service)}
	out, err := ToJSON(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "project") {
		t.Error("missing project key")
	}
}

func TestToJSON_WithEnv(t *testing.T) {
	spec := &Spec{
		Project: Project{Name: "e"},
		Services: map[string]Service{
			"web": {Type: "http", Env: map[string]string{"K": "V"}},
		},
	}
	out, err := ToJSON(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), `"K": "V"`) {
		t.Error("missing env in JSON")
	}
}

// ===========================================================================
// ToHCL tests
// ===========================================================================

func TestToHCL_Basic(t *testing.T) {
	spec := basicSpec()
	hcl := ToHCL(spec)
	if !strings.Contains(hcl, `project "myapp"`) {
		t.Error("missing project block")
	}
	if !strings.Contains(hcl, `service "web"`) {
		t.Error("missing service block")
	}
	if !strings.Contains(hcl, `engine = "docker"`) {
		t.Error("missing engine")
	}
}

func TestToHCL_WithEnv(t *testing.T) {
	spec := &Spec{
		Project: Project{Name: "e", Version: "1"},
		Services: map[string]Service{
			"app": {Image: "img", Port: 80, Type: "http", Env: map[string]string{"A": "1", "B": "2"}},
		},
	}
	hcl := ToHCL(spec)
	if !strings.Contains(hcl, `env "vars"`) {
		t.Error("missing env block")
	}
	if !strings.Contains(hcl, `A = "1"`) {
		t.Error("missing env var A")
	}
	if !strings.Contains(hcl, `B = "2"`) {
		t.Error("missing env var B")
	}
}

func TestToHCL_WithVolumesAndDepends(t *testing.T) {
	spec := &Spec{
		Project: Project{Name: "vd"},
		Services: map[string]Service{
			"api": {Type: "http", Volumes: []string{"/a", "/b"}, Depends: []string{"db"}},
			"db":  {Type: "tcp"},
		},
	}
	hcl := ToHCL(spec)
	if !strings.Contains(hcl, `volumes = ["/a", "/b"]`) {
		t.Error("missing volumes")
	}
	if !strings.Contains(hcl, `depends = ["db"]`) {
		t.Error("missing depends")
	}
}

func TestToHCL_Replicas(t *testing.T) {
	spec := &Spec{
		Project:  Project{Name: "r"},
		Services: map[string]Service{"web": {Type: "http", Replicas: 5}},
	}
	hcl := ToHCL(spec)
	if !strings.Contains(hcl, "replicas = 5") {
		t.Error("missing replicas")
	}
}

func TestToHCL_MinimalProject(t *testing.T) {
	spec := &Spec{
		Project:  Project{Name: "min"},
		Services: map[string]Service{},
	}
	hcl := ToHCL(spec)
	if !strings.Contains(hcl, `project "min"`) {
		t.Error("missing project block")
	}
	if strings.Contains(hcl, "service") {
		t.Error("should not contain service block")
	}
}

func TestToHCL_RoundTrip(t *testing.T) {
	spec := basicSpec()
	hcl := ToHCL(spec)
	parsed, err := Parse([]byte(hcl), "roundtrip.hcl")
	if err != nil {
		t.Fatalf("round-trip parse error: %v", err)
	}
	if parsed.Project.Name != spec.Project.Name {
		t.Errorf("project name mismatch: %q vs %q", parsed.Project.Name, spec.Project.Name)
	}
	if parsed.Project.Version != spec.Project.Version {
		t.Errorf("version mismatch: %q vs %q", parsed.Project.Version, spec.Project.Version)
	}
	svc := parsed.Services["web"]
	if svc.Image != spec.Services["web"].Image {
		t.Errorf("image mismatch: %q vs %q", svc.Image, spec.Services["web"].Image)
	}
	if svc.Port != spec.Services["web"].Port {
		t.Errorf("port mismatch: %d vs %d", svc.Port, spec.Services["web"].Port)
	}
}

func TestToHCL_RoundTripWithEnv(t *testing.T) {
	spec := &Spec{
		Project: Project{Name: "rt", Version: "3.0"},
		Services: map[string]Service{
			"web": {Image: "nginx", Port: 80, Type: "http", Env: map[string]string{"X": "Y"}},
		},
	}
	hcl := ToHCL(spec)
	parsed, err := Parse([]byte(hcl), "rt.hcl")
	if err != nil {
		t.Fatalf("round-trip parse error: %v", err)
	}
	svc := parsed.Services["web"]
	if svc.Env["X"] != "Y" {
		t.Errorf("env round-trip failed: %v", svc.Env)
	}
}

func TestToHCL_MinimalEmpty(t *testing.T) {
	spec := &Spec{
		Project:  Project{Name: "empty"},
		Services: map[string]Service{},
	}
	hcl := ToHCL(spec)
	if !strings.Contains(hcl, `project "empty"`) {
		t.Error("missing project")
	}
	if strings.Contains(hcl, "infra") {
		t.Error("should not contain infra")
	}
}

// ===========================================================================
// MergeSpecs tests
// ===========================================================================

func TestMergeSpecs_NoConflict(t *testing.T) {
	a := &Spec{
		Project:  Project{Name: "a", Version: "1"},
		Services: map[string]Service{"web": {Image: "nginx", Port: 80, Type: "http"}},
	}
	b := &Spec{
		Project:  Project{Name: "b", Version: "2"},
		Services: map[string]Service{"db": {Image: "postgres", Port: 5432, Type: "tcp"}},
	}
	merged := MergeSpecs(a, b)
	if merged.Project.Name != "b" {
		t.Errorf("project name = %q", merged.Project.Name)
	}
	if merged.Project.Version != "2" {
		t.Errorf("version = %q", merged.Project.Version)
	}
	if _, ok := merged.Services["web"]; !ok {
		t.Error("missing service web")
	}
	if _, ok := merged.Services["db"]; !ok {
		t.Error("missing service db")
	}
}

func TestMergeSpecs_SrcOverrides(t *testing.T) {
	a := &Spec{
		Project:  Project{Name: "a", Version: "1"},
		Services: map[string]Service{"web": {Image: "old", Port: 80, Type: "http"}},
	}
	b := &Spec{
		Project:  Project{Name: "a", Version: "2"},
		Services: map[string]Service{"web": {Image: "new", Port: 90, Type: "grpc"}},
	}
	merged := MergeSpecs(a, b)
	if merged.Project.Version != "2" {
		t.Errorf("version = %q", merged.Project.Version)
	}
	web := merged.Services["web"]
	if web.Image != "new" {
		t.Errorf("image = %q", web.Image)
	}
	if web.Port != 90 {
		t.Errorf("port = %d", web.Port)
	}
	if web.Type != "grpc" {
		t.Errorf("type = %q", web.Type)
	}
}

func TestMergeSpecs_EnvMerge(t *testing.T) {
	a := &Spec{
		Project:  Project{Name: "e"},
		Services: map[string]Service{"x": {Type: "http", Env: map[string]string{"A": "1"}}},
	}
	b := &Spec{
		Services: map[string]Service{"x": {Env: map[string]string{"B": "2"}}},
	}
	merged := MergeSpecs(a, b)
	env := merged.Services["x"].Env
	if env["A"] != "1" {
		t.Errorf("env A = %q", env["A"])
	}
	if env["B"] != "2" {
		t.Errorf("env B = %q", env["B"])
	}
}

func TestMergeSpecs_VolumeConcat(t *testing.T) {
	a := &Spec{
		Project:  Project{Name: "v"},
		Services: map[string]Service{"x": {Type: "http", Volumes: []string{"/a"}}},
	}
	b := &Spec{
		Services: map[string]Service{"x": {Volumes: []string{"/b"}}},
	}
	merged := MergeSpecs(a, b)
	vols := merged.Services["x"].Volumes
	if len(vols) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(vols))
	}
	if vols[0] != "/a" || vols[1] != "/b" {
		t.Errorf("volumes = %v", vols)
	}
}

func TestMergeSpecs_DependsConcat(t *testing.T) {
	a := &Spec{
		Project:  Project{Name: "d"},
		Services: map[string]Service{"x": {Type: "http", Depends: []string{"a"}}},
	}
	b := &Spec{
		Services: map[string]Service{"x": {Depends: []string{"b"}}},
	}
	merged := MergeSpecs(a, b)
	deps := merged.Services["x"].Depends
	if len(deps) != 2 {
		t.Fatalf("expected 2 depends, got %d", len(deps))
	}
}

func TestMergeSpecs_InfraOverride(t *testing.T) {
	a := &Spec{Infra: Infra{Engine: "docker"}}
	b := &Spec{Infra: Infra{Engine: "podman"}}
	merged := MergeSpecs(a, b)
	if merged.Infra.Engine != "podman" {
		t.Errorf("engine = %q", merged.Infra.Engine)
	}
}

func TestMergeSpecs_InfraFallback(t *testing.T) {
	a := &Spec{Infra: Infra{Engine: "docker"}}
	b := &Spec{}
	merged := MergeSpecs(a, b)
	if merged.Infra.Engine != "docker" {
		t.Errorf("engine = %q", merged.Infra.Engine)
	}
}

func TestMergeSpecs_BothEmpty(t *testing.T) {
	a := &Spec{Services: make(map[string]Service)}
	b := &Spec{Services: make(map[string]Service)}
	merged := MergeSpecs(a, b)
	if merged.Project.Name != "" {
		t.Errorf("expected empty project name")
	}
}

// ===========================================================================
// SpecDiff tests
// ===========================================================================

func TestSpecDiff_Identical(t *testing.T) {
	a := basicSpec()
	b := basicSpec()
	diffs := SpecDiff(a, b)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d: %v", len(diffs), diffs)
	}
}

func TestSpecDiff_ProjectNameChanged(t *testing.T) {
	a := &Spec{Project: Project{Name: "old"}, Services: make(map[string]Service)}
	b := &Spec{Project: Project{Name: "new"}, Services: make(map[string]Service)}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "project.name" && d.Kind == DiffChanged {
			found = true
			if d.OldValue != "old" || d.NewValue != "new" {
				t.Errorf("old=%q new=%q", d.OldValue, d.NewValue)
			}
		}
	}
	if !found {
		t.Errorf("expected project.name diff, got %v", diffs)
	}
}

func TestSpecDiff_VersionChanged(t *testing.T) {
	a := &Spec{Project: Project{Name: "x", Version: "1"}, Services: make(map[string]Service)}
	b := &Spec{Project: Project{Name: "x", Version: "2"}, Services: make(map[string]Service)}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "project.version" {
			found = true
		}
	}
	if !found {
		t.Error("expected version diff")
	}
}

func TestSpecDiff_DescriptionChanged(t *testing.T) {
	a := &Spec{Project: Project{Name: "x", Description: "old"}, Services: make(map[string]Service)}
	b := &Spec{Project: Project{Name: "x", Description: "new"}, Services: make(map[string]Service)}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "project.description" {
			found = true
		}
	}
	if !found {
		t.Error("expected description diff")
	}
}

func TestSpecDiff_InfraEngineChanged(t *testing.T) {
	a := &Spec{Infra: Infra{Engine: "docker"}, Services: make(map[string]Service)}
	b := &Spec{Infra: Infra{Engine: "podman"}, Services: make(map[string]Service)}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "infra.engine" {
			found = true
		}
	}
	if !found {
		t.Error("expected infra.engine diff")
	}
}

func TestSpecDiff_ServiceAdded(t *testing.T) {
	a := &Spec{Services: map[string]Service{"a": {Type: "http"}}}
	b := &Spec{Services: map[string]Service{"a": {Type: "http"}, "b": {Type: "tcp"}}}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "services.b" && d.Kind == DiffAdded {
			found = true
		}
	}
	if !found {
		t.Error("expected services.b added diff")
	}
}

func TestSpecDiff_ServiceRemoved(t *testing.T) {
	a := &Spec{Services: map[string]Service{"a": {Type: "http"}, "b": {Type: "tcp"}}}
	b := &Spec{Services: map[string]Service{"a": {Type: "http"}}}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "services.b" && d.Kind == DiffRemoved {
			found = true
		}
	}
	if !found {
		t.Error("expected services.b removed diff")
	}
}

func TestSpecDiff_ServiceImageChanged(t *testing.T) {
	a := &Spec{Services: map[string]Service{"x": {Image: "old"}}}
	b := &Spec{Services: map[string]Service{"x": {Image: "new"}}}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "services.x.image" && d.Kind == DiffChanged {
			found = true
		}
	}
	if !found {
		t.Error("expected services.x.image diff")
	}
}

func TestSpecDiff_ServicePortChanged(t *testing.T) {
	a := &Spec{Services: map[string]Service{"x": {Port: 80}}}
	b := &Spec{Services: map[string]Service{"x": {Port: 443}}}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "services.x.port" && d.Kind == DiffChanged {
			found = true
		}
	}
	if !found {
		t.Error("expected services.x.port diff")
	}
}

func TestSpecDiff_ServiceTypeChanged(t *testing.T) {
	a := &Spec{Services: map[string]Service{"x": {Type: "http"}}}
	b := &Spec{Services: map[string]Service{"x": {Type: "grpc"}}}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "services.x.type" && d.Kind == DiffChanged {
			found = true
		}
	}
	if !found {
		t.Error("expected services.x.type diff")
	}
}

func TestSpecDiff_ReplicasChanged(t *testing.T) {
	a := &Spec{Services: map[string]Service{"x": {Replicas: 1}}}
	b := &Spec{Services: map[string]Service{"x": {Replicas: 5}}}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "services.x.replicas" {
			found = true
		}
	}
	if !found {
		t.Error("expected replicas diff")
	}
}

func TestSpecDiff_VolumesChanged(t *testing.T) {
	a := &Spec{Services: map[string]Service{"x": {Volumes: []string{"/a"}}}}
	b := &Spec{Services: map[string]Service{"x": {Volumes: []string{"/a", "/b"}}}}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "services.x.volumes" {
			found = true
		}
	}
	if !found {
		t.Error("expected volumes diff")
	}
}

func TestSpecDiff_DependsChanged(t *testing.T) {
	a := &Spec{Services: map[string]Service{"x": {Depends: []string{"a"}}}}
	b := &Spec{Services: map[string]Service{"x": {Depends: []string{"a", "b"}}}}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "services.x.depends" {
			found = true
		}
	}
	if !found {
		t.Error("expected depends diff")
	}
}

func TestSpecDiff_EnvChanged(t *testing.T) {
	a := &Spec{Services: map[string]Service{"x": {Env: map[string]string{"K": "V1"}}}}
	b := &Spec{Services: map[string]Service{"x": {Env: map[string]string{"K": "V2"}}}}
	diffs := SpecDiff(a, b)
	found := false
	for _, d := range diffs {
		if d.Path == "services.x.env" {
			found = true
		}
	}
	if !found {
		t.Error("expected env diff")
	}
}

func TestSpecDiff_MultipleChanges(t *testing.T) {
	a := &Spec{
		Project:  Project{Name: "a", Version: "1"},
		Services: map[string]Service{"x": {Type: "http"}, "y": {Type: "tcp"}},
		Infra:    Infra{Engine: "docker"},
	}
	b := &Spec{
		Project:  Project{Name: "b", Version: "2"},
		Services: map[string]Service{"x": {Type: "grpc"}, "z": {Type: "worker"}},
		Infra:    Infra{Engine: "podman"},
	}
	diffs := SpecDiff(a, b)
	if len(diffs) < 5 {
		t.Errorf("expected at least 5 diffs, got %d: %v", len(diffs), diffs)
	}
}

// ===========================================================================
// DiffEntry.String tests
// ===========================================================================

func TestDiffEntry_String(t *testing.T) {
	tests := []struct {
		entry    DiffEntry
		contains string
	}{
		{DiffEntry{Kind: DiffAdded, Path: "x", NewValue: "1"}, "+ x = 1"},
		{DiffEntry{Kind: DiffRemoved, Path: "y", OldValue: "2"}, "- y = 2"},
		{DiffEntry{Kind: DiffChanged, Path: "z", OldValue: "old", NewValue: "new"}, "~ z: old -> new"},
	}
	for _, tt := range tests {
		got := tt.entry.String()
		if got != tt.contains {
			t.Errorf("String() = %q, want %q", got, tt.contains)
		}
	}
}

func TestDiffEntry_StringDefault(t *testing.T) {
	e := DiffEntry{Kind: "unknown"}
	if e.String() != "" {
		t.Errorf("expected empty string for unknown kind, got %q", e.String())
	}
}

// ===========================================================================
// ParseError: newParseError
// ===========================================================================

func TestNewParseError(t *testing.T) {
	e := newParseError("test.hcl", 10, 5, "bad", "the bad line")
	if e.FileName != "test.hcl" {
		t.Errorf("FileName = %q", e.FileName)
	}
	if e.Line != 10 || e.Column != 5 {
		t.Errorf("Line=%d Column=%d", e.Line, e.Column)
	}
	if e.Context != "the bad line" {
		t.Errorf("Context = %q", e.Context)
	}
}
