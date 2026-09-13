// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package contracts

import (
	"testing"
)

type testContract struct {
	valid bool
}

func (c testContract) Validate() error {
	if !c.valid {
		return errInvalid
	}
	return nil
}

type testSchemaAware struct {
	version string
}

func (s testSchemaAware) SchemaVersion() string {
	return s.version
}

type testVersioned struct {
	version string
}

func (v testVersioned) Version() string {
	return v.version
}

type testIdentifiable struct {
	id string
}

func (i testIdentifiable) ID() string {
	return i.id
}

type testNamed struct {
	name string
}

func (n testNamed) Name() string {
	return n.name
}

var errInvalid = &testError{"invalid contract"}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

func TestContractInterface(t *testing.T) {
	var c Contract = testContract{valid: true}
	if err := c.Validate(); err != nil {
		t.Errorf("valid contract should return nil, got %v", err)
	}

	c = testContract{valid: false}
	if err := c.Validate(); err == nil {
		t.Error("invalid contract should return error")
	}
}

func TestSchemaAwareInterface(t *testing.T) {
	var s SchemaAware = testSchemaAware{version: "1.0.0"}
	if got := s.SchemaVersion(); got != "1.0.0" {
		t.Errorf("SchemaVersion() = %q, want %q", got, "1.0.0")
	}
}

func TestVersionedInterface(t *testing.T) {
	var v Versioned = testVersioned{version: "2.1.0"}
	if got := v.Version(); got != "2.1.0" {
		t.Errorf("Version() = %q, want %q", got, "2.1.0")
	}
}

func TestIdentifiableInterface(t *testing.T) {
	var i Identifiable = testIdentifiable{id: "ID-001"}
	if got := i.ID(); got != "ID-001" {
		t.Errorf("ID() = %q, want %q", got, "ID-001")
	}
}

func TestNamedInterface(t *testing.T) {
	var n Named = testNamed{name: "my-service"}
	if got := n.Name(); got != "my-service" {
		t.Errorf("Name() = %q, want %q", got, "my-service")
	}
}

func TestHealthStatus(t *testing.T) {
	h := Healthy("all good")
	if !h.Healthy || h.Message != "all good" {
		t.Error("unexpected healthy status")
	}
	u := Unhealthy("broken")
	if u.Healthy || u.Message != "broken" {
		t.Error("unexpected unhealthy status")
	}
}

func TestPaginatedResult(t *testing.T) {
	p := PaginatedResult{Page: 2, PageSize: 10, TotalItems: 25, TotalPages: 3}
	if !p.HasNext() {
		t.Error("expected HasNext")
	}
	if !p.HasPrev() {
		t.Error("expected HasPrev")
	}
	p2 := PaginatedResult{Page: 1, TotalPages: 1}
	if p2.HasNext() {
		t.Error("expected no next")
	}
	if p2.HasPrev() {
		t.Error("expected no prev")
	}
}

func TestInMemoryRegistry(t *testing.T) {
	r := NewInMemoryRegistry()
	n := &testNamed{name: "comp1"}
	if err := r.Register(n); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(n); err == nil {
		t.Error("expected duplicate error")
	}
	got, ok := r.Get("comp1")
	if !ok || got != n {
		t.Error("expected to find comp1")
	}
	list := r.List()
	if len(list) != 1 {
		t.Errorf("expected 1, got %d", len(list))
	}
	if err := r.Unregister("comp1"); err != nil {
		t.Fatal(err)
	}
	if err := r.Unregister("comp1"); err == nil {
		t.Error("expected not found error")
	}
}

func TestInMemoryRegistryNotNamed(t *testing.T) {
	r := NewInMemoryRegistry()
	err := r.Register("not-named")
	if err == nil {
		t.Error("expected error for non-named component")
	}
}

func TestValidationResult(t *testing.T) {
	vr := ValidationResult{Valid: false, Errors: []ValidationError{{Field: "f", Message: "m"}}}
	if !vr.HasErrors() {
		t.Error("expected errors")
	}
	if len(vr.ErrorMessages()) != 1 {
		t.Errorf("expected 1 error message, got %d", len(vr.ErrorMessages()))
	}
	vw := ValidationResult{Valid: true, Warnings: []ValidationWarning{{Field: "f", Message: "w"}}}
	if !vw.HasWarnings() {
		t.Error("expected warnings")
	}
	if len(vw.WarningMessages()) != 1 {
		t.Errorf("expected 1 warning message, got %d", len(vw.WarningMessages()))
	}
}

func TestStringValidator(t *testing.T) {
	sv := StringValidator{MinLength: 2, MaxLength: 5, Enum: []string{"a", "b", "abc"}}
	vr := sv.ValidateValue(123)
	if vr.Valid {
		t.Error("expected invalid for non-string")
	}
	vr = sv.ValidateValue("x")
	if vr.Valid {
		t.Error("expected invalid for too short")
	}
	vr = sv.ValidateValue("toolong")
	if vr.Valid {
		t.Error("expected invalid for too long")
	}
	vr = sv.ValidateValue("d")
	if vr.Valid {
		t.Error("expected invalid for not in enum")
	}
	vr = sv.ValidateValue("abc")
	if !vr.Valid {
		t.Error("expected valid for 'abc'")
	}
}

func TestIntValidator(t *testing.T) {
	iv := IntValidator{Min: 1, Max: 10}
	vr := iv.ValidateValue("not-an-int")
	if vr.Valid {
		t.Error("expected invalid for string")
	}
	vr = iv.ValidateValue(0)
	if vr.Valid {
		t.Error("expected invalid for below min")
	}
	vr = iv.ValidateValue(11)
	if vr.Valid {
		t.Error("expected invalid for above max")
	}
	vr = iv.ValidateValue(5)
	if !vr.Valid {
		t.Error("expected valid")
	}
	vr = iv.ValidateValue(float64(5))
	if !vr.Valid {
		t.Error("expected valid for float64")
	}
}

func TestAllContract(t *testing.T) {
	ok := All(func() error { return nil }, func() error { return nil })
	if err := ok(); err != nil {
		t.Error("expected nil error")
	}
	fail := All(func() error { return nil }, func() error { return errInvalid })
	if err := fail(); err == nil {
		t.Error("expected error")
	}
}

func TestAnyContract(t *testing.T) {
	ok := Any(func() error { return errInvalid }, func() error { return nil })
	if err := ok(); err != nil {
		t.Error("expected nil error")
	}
	fail := Any(func() error { return errInvalid }, func() error { return errInvalid })
	if err := fail(); err == nil {
		t.Error("expected error")
	}
}

func TestNotContract(t *testing.T) {
	ok := Not(func() error { return errInvalid })
	if err := ok(); err != nil {
		t.Error("expected nil")
	}
	fail := Not(func() error { return nil })
	if err := fail(); err == nil {
		t.Error("expected error")
	}
}

func TestRequireNonEmpty(t *testing.T) {
	if err := RequireNonEmpty("name", ""); err == nil {
		t.Error("expected error")
	}
	if err := RequireNonEmpty("name", "  "); err == nil {
		t.Error("expected error for whitespace")
	}
	if err := RequireNonEmpty("name", "ok"); err != nil {
		t.Error("expected nil")
	}
}

func TestRequireRange(t *testing.T) {
	if err := RequireRange("port", 0, 1, 65535); err == nil {
		t.Error("expected error")
	}
	if err := RequireRange("port", 99999, 1, 65535); err == nil {
		t.Error("expected error")
	}
	if err := RequireRange("port", 8080, 1, 65535); err != nil {
		t.Error("expected nil")
	}
}

func TestRequireOneOf(t *testing.T) {
	if err := RequireOneOf("lang", "rust", "go", "ts"); err == nil {
		t.Error("expected error")
	}
	if err := RequireOneOf("lang", "go", "go", "ts"); err != nil {
		t.Error("expected nil")
	}
}

func TestRequireGreaterThan(t *testing.T) {
	if err := RequireGreaterThan("count", 0, 0); err == nil {
		t.Error("expected error")
	}
	if err := RequireGreaterThan("count", 1, 0); err != nil {
		t.Error("expected nil")
	}
}

func TestValidationError(t *testing.T) {
	e := ValidationError{Field: "name", Message: "required", Code: "REQUIRED"}
	if e.Error() != "name: required" {
		t.Errorf("unexpected error string: %s", e.Error())
	}
}
