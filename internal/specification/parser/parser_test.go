package parser

import "testing"

func TestParserParsesStructuredSpec(t *testing.T) {
	input := `
project: acme-api
version: 1.0
modules:
  - name: auth
    path: ./internal/auth
services:
  - name: gateway
    kind: http
    port: 8080
`

	doc, err := NewParser().Parse(input)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if doc.Project != "acme-api" {
		t.Fatalf("expected project name acme-api, got %q", doc.Project)
	}
	if len(doc.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(doc.Modules))
	}
	if doc.Modules[0].Name != "auth" {
		t.Fatalf("expected module name auth, got %q", doc.Modules[0].Name)
	}
	if len(doc.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(doc.Services))
	}
	if doc.Services[0].Name != "gateway" {
		t.Fatalf("expected service name gateway, got %q", doc.Services[0].Name)
	}
}
