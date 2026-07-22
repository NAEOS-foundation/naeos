package schemaregistry

import (
	"testing"
)

func TestRegisterAndGet(t *testing.T) {
	r := New()
	err := r.Register("test-schema", "v1.0.0", `{"type": "object"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry, err := r.Get("test-schema", "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Name != "test-schema" {
		t.Errorf("expected name test-schema, got %s", entry.Name)
	}
	if entry.Version != "v1.0.0" {
		t.Errorf("expected version v1.0.0, got %s", entry.Version)
	}
	if entry.Schema != `{"type": "object"}` {
		t.Errorf("unexpected schema: %s", entry.Schema)
	}
}

func TestGetLatest(t *testing.T) {
	r := New()
	r.Register("test", "v1.0.0", `{"type": "string"}`)
	r.Register("test", "v2.0.0", `{"type": "integer"}`)

	entry, err := r.Get("test", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Version != "v2.0.0" {
		t.Errorf("expected latest v2.0.0, got %s", entry.Version)
	}
}

func TestRegisterInvalidInputs(t *testing.T) {
	r := New()

	if err := r.Register("", "v1.0.0", "{}"); err == nil {
		t.Error("expected error for empty name")
	}
	if err := r.Register("test", "", "{}"); err == nil {
		t.Error("expected error for empty version")
	}
	if err := r.Register("test", "v1.0.0", ""); err == nil {
		t.Error("expected error for empty schema")
	}
	if err := r.Register("test", "1.0", "{}"); err == nil {
		t.Error("expected error for invalid semver")
	}
}

func TestList(t *testing.T) {
	r := New()
	r.Register("a", "v1.0.0", "{}")
	r.Register("b", "v1.0.0", "{}")

	names := r.List()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	if names[0] != "a" || names[1] != "b" {
		t.Errorf("expected sorted [a b], got %v", names)
	}
}

func TestVersions(t *testing.T) {
	r := New()
	r.Register("test", "v1.0.0", "{}")
	r.Register("test", "v2.0.0", "{}")
	r.Register("test", "v2.1.0", "{}")

	versions, err := r.Versions("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	if versions[0] != "v2.1.0" {
		t.Errorf("expected v2.1.0 first, got %s", versions[0])
	}
}

func TestVersionsNotFound(t *testing.T) {
	r := New()
	_, err := r.Versions("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent schema")
	}
}

func TestDelete(t *testing.T) {
	r := New()
	r.Register("test", "v1.0.0", "{}")
	r.Register("test", "v2.0.0", "{}")

	if err := r.Delete("test", "v1.0.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	versions, _ := r.Versions("test")
	if len(versions) != 1 {
		t.Errorf("expected 1 version after delete, got %d", len(versions))
	}

	if err := r.Delete("nonexistent", "v1.0.0"); err == nil {
		t.Error("expected error for nonexistent schema")
	}
	if err := r.Delete("test", "v9.9.9"); err == nil {
		t.Error("expected error for nonexistent version")
	}
}

func TestDeleteLastVersion(t *testing.T) {
	r := New()
	r.Register("test", "v1.0.0", "{}")
	r.Delete("test", "v1.0.0")

	names := r.List()
	if len(names) != 0 {
		t.Errorf("expected schema to be removed after deleting last version")
	}
}

func TestOverwriteVersion(t *testing.T) {
	r := New()
	r.Register("test", "v1.0.0", `{"type": "string"}`)
	r.Register("test", "v1.0.0", `{"type": "number"}`)

	entry, _ := r.Get("test", "v1.0.0")
	if entry.Schema != `{"type": "number"}` {
		t.Errorf("expected overwritten schema, got %s", entry.Schema)
	}
}
