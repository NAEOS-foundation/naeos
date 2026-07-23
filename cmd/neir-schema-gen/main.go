package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
)

func main() {
	outDir := "site/static/schemaregistry"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	t := reflect.TypeOf(model.NEIR{})
	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	schema := g.generate(t, false)

	defs := make(map[string]any)
	for name, def := range g.definitions {
		defs[name] = def
	}

	root := map[string]any{
		"$schema":     "http://json-schema.org/draft-07/schema#",
		"$id":         "https://naeos.dev/schemaregistry/neir.json",
		"title":       "NEIR Specification",
		"description": "JSON Schema for the NAEOS Engineering Intelligence Representation (NEIR) specification format",
		"type":        "object",
		"properties":  schema,
		"definitions": defs,
	}

	required := g.requiredRoot()
	if len(required) > 0 {
		root["required"] = required
	}

	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal: %v\n", err)
		os.Exit(1)
	}

	v1Dir := filepath.Join(outDir, "v1")
	if err := os.MkdirAll(v1Dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir v1: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(filepath.Join(v1Dir, "neir.json"), data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write v1/neir.json: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(filepath.Join(outDir, "latest.json"), data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write latest.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated NEIR JSON Schema:")
	fmt.Printf("  %s\n", filepath.Join(v1Dir, "neir.json"))
	fmt.Printf("  %s\n", filepath.Join(outDir, "latest.json"))
}

type generator struct {
	definitions map[string]map[string]any
	visited     map[string]bool
	enums       map[string][]string
}

func (g *generator) requiredRoot() []string {
	return []string{"project", "modules"}
}

func (g *generator) generate(t reflect.Type, optional bool) map[string]any {
	props := make(map[string]any)

	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		name, opts := parseJSONTag(f.Tag.Get("json"))
		if name == "" || name == "-" {
			continue
		}

		_, omitEmpty := opts["omitempty"]
		fieldOpt := optional || omitEmpty

		schema := g.fieldSchema(f.Type, fieldOpt)

		props[name] = schema
	}

	if extra, ok := g.definitions[t.Name()]; ok {
		for k, v := range extra {
			props[k] = v
		}
	}

	return props
}

func (g *generator) fieldSchema(t reflect.Type, optional bool) map[string]any {
	switch t.Kind() {
	case reflect.Ptr:
		return g.fieldSchema(t.Elem(), true)
	case reflect.Slice:
		return g.arraySchema(t.Elem())
	case reflect.Map:
		return g.mapSchema(t)
	case reflect.Struct:
		return g.structSchema(t, optional)
	default:
		return g.primitiveSchema(t, optional)
	}
}

func (g *generator) structSchema(t reflect.Type, optional bool) map[string]any {
	if t == reflect.TypeOf(time.Time{}) {
		return map[string]any{
			"type":   "string",
			"format": "date-time",
		}
	}

	typeName := typeName(t)
	if g.visited[typeName] {
		return map[string]any{
			"$ref": fmt.Sprintf("#/definitions/%s", typeName),
		}
	}

	g.visited[typeName] = true

	props := make(map[string]any)
	var required []string

	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		name, opts := parseJSONTag(f.Tag.Get("json"))
		if name == "" || name == "-" {
			continue
		}

		_, omitEmpty := opts["omitempty"]
		schema := g.fieldSchema(f.Type, omitEmpty)
		props[name] = schema

		if !omitEmpty {
			required = append(required, name)
		}

		if enumVals, ok := enumForField(t, f, g.enums); ok {
			schema["enum"] = enumVals
		}
	}

	def := map[string]any{
		"type":       "object",
		"title":      typeName,
		"properties": props,
	}
	if len(required) > 0 {
		def["required"] = required
	}

	g.definitions[typeName] = def

	ref := map[string]any{
		"$ref": fmt.Sprintf("#/definitions/%s", typeName),
	}
	if optional {
		return ref
	}

	return ref
}

func (g *generator) arraySchema(elem reflect.Type) map[string]any {
	return map[string]any{
		"type":  "array",
		"items": g.fieldSchema(elem, false),
	}
}

func (g *generator) mapSchema(t reflect.Type) map[string]any {
	valSchema := g.fieldSchema(t.Elem(), false)
	return map[string]any{
		"type":                 "object",
		"additionalProperties": valSchema,
	}
}

func (g *generator) primitiveSchema(t reflect.Type, optional bool) map[string]any {
	s := map[string]any{
		"type": jsonType(t),
	}

	if enumVals, ok := g.enums[typeName(t)]; ok {
		s["enum"] = enumVals
	}

	return s
}

func jsonType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	default:
		if t.Name() == "Time" && t.PkgPath() == "time" {
			return "string"
		}
		return "string"
	}
}

func typeName(t reflect.Type) string {
	pkg := strings.TrimPrefix(t.PkgPath(), "github.com/NAEOS-foundation/naeos/internal/neir/model/")
	if pkg == t.PkgPath() {
		pkg = strings.TrimPrefix(t.PkgPath(), "github.com/NAEOS-foundation/naeos/internal/neir/model")
	}
	pkg = strings.ReplaceAll(pkg, "/", ".")
	return pkg + "." + t.Name()
}

func parseJSONTag(tag string) (string, map[string]bool) {
	if tag == "" {
		return "", nil
	}
	parts := strings.Split(tag, ",")
	opts := make(map[string]bool)
	for _, p := range parts[1:] {
		opts[p] = true
	}
	return parts[0], opts
}

func knownEnums() map[string][]string {
	return map[string][]string{
		"architecture.Pattern":    {"layered", "clean", "hexagonal", "microkernel", "event-driven", "cqrs", "monolith"},
		"service.ServiceKind":     {"http", "grpc", "worker", "cli", "job"},
		"api.Protocol":            {"http", "grpc", "graphql", "websocket"},
		"storage.StorageType":     {"sql", "nosql", "file", "cache", "queue", "blob"},
		"deployment.Strategy":     {"rolling", "blue-green", "canary", "recreate"},
		"testing.TestingStrategy": {"unit", "integration", "e2e", "contract"},
		"infrastructure.Provider": {"aws", "gcp", "azure", "local"},
		"component.ComponentKind": {"handler", "service", "repository", "middleware", "model", "config", "worker", "scheduler"},
		"docs.DocKind":            {"guide", "reference", "adr", "rfc", "changelog"},
		"language.Language":       {"go", "typescript", "python", "java", "rust"},
	}
}

func enumForField(parent reflect.Type, f reflect.StructField, enums map[string][]string) ([]string, bool) {
	ft := f.Type
	if ft.Kind() == reflect.Ptr {
		ft = ft.Elem()
	}
	if ft.Kind() == reflect.Slice {
		return nil, false
	}

	key := typeName(ft)
	vals, ok := enums[key]
	return vals, ok
}
