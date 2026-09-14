// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"testing"
)

func TestDefault(t *testing.T) {
	v := Default()
	if v.NEIRVersion != "0.1.0" {
		t.Fatalf("expected NEIRVersion 0.1.0, got %s", v.NEIRVersion)
	}
	if v.SchemaVersion != "1.0" {
		t.Fatalf("expected SchemaVersion 1.0, got %s", v.SchemaVersion)
	}
}

func TestParseSemVer(t *testing.T) {
	v, err := ParseSemVer("1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Fatalf("expected 1.2.3, got %v", v)
	}
}

func TestParseSemVerWithVPrefix(t *testing.T) {
	v, err := ParseSemVer("v1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Fatalf("expected 1.2.3, got %v", v)
	}
}

func TestParseSemVerInvalid(t *testing.T) {
	_, err := ParseSemVer("not-a-version")
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestParseSemVerTooFewParts(t *testing.T) {
	_, err := ParseSemVer("1.2")
	if err == nil {
		t.Fatal("expected error for version with too few parts")
	}
}

func TestParseSemVerNegative(t *testing.T) {
	_, err := ParseSemVer("-1.0.0")
	if err == nil {
		t.Fatal("expected error for negative version")
	}
}

func TestSemVerString(t *testing.T) {
	v := SemVer{Major: 1, Minor: 2, Patch: 3}
	if v.String() != "1.2.3" {
		t.Fatalf("expected 1.2.3, got %s", v.String())
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b   string
		expect int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.1.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
	}
	for _, tt := range tests {
		a, _ := ParseSemVer(tt.a)
		b, _ := ParseSemVer(tt.b)
		result := Compare(a, b)
		if result != tt.expect {
			t.Fatalf("Compare(%s, %s) = %d, want %d", tt.a, tt.b, result, tt.expect)
		}
	}
}

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		required, actual string
		expect           bool
	}{
		{"1.0.0", "1.0.0", true},
		{"1.0.0", "1.1.0", true},
		{"1.0.0", "1.2.3", true},
		{"1.0.0", "2.0.0", false},
		{"1.0.0", "0.9.0", false},
		{"0.1.0", "0.1.0", true},
		{"0.1.0", "0.2.0", false},
		{"0.1.0", "0.0.9", false},
	}
	for _, tt := range tests {
		req, _ := ParseSemVer(tt.required)
		act, _ := ParseSemVer(tt.actual)
		result := IsCompatible(req, act)
		if result != tt.expect {
			t.Fatalf("IsCompatible(%s, %s) = %v, want %v", tt.required, tt.actual, result, tt.expect)
		}
	}
}

func TestValidate(t *testing.T) {
	v := VersionInfo{NEIRVersion: "0.1.0", SchemaVersion: "1.0.0"}
	if err := v.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateEmptyNEIRVersion(t *testing.T) {
	v := VersionInfo{NEIRVersion: "", SchemaVersion: "1.0.0"}
	if err := v.Validate(); err == nil {
		t.Fatal("expected error for empty NEIRVersion")
	}
}

func TestValidateInvalidNEIRVersion(t *testing.T) {
	v := VersionInfo{NEIRVersion: "invalid", SchemaVersion: "1.0.0"}
	if err := v.Validate(); err == nil {
		t.Fatal("expected error for invalid NEIRVersion")
	}
}

func TestIsCompatibleWith(t *testing.T) {
	a := VersionInfo{NEIRVersion: "1.0.0", SchemaVersion: "1.0.0"}
	b := VersionInfo{NEIRVersion: "1.1.0", SchemaVersion: "1.0.0"}
	if !a.IsCompatibleWith(b) {
		t.Fatal("expected compatible versions")
	}

	c := VersionInfo{NEIRVersion: "2.0.0", SchemaVersion: "1.0.0"}
	if a.IsCompatibleWith(c) {
		t.Fatal("expected incompatible versions")
	}
}
