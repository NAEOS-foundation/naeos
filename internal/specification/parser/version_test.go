package parser

import (
	"testing"
)

func TestParseSchemaVersion(t *testing.T) {
	tests := []struct {
		input   string
		want    SchemaVersion
		wantErr bool
	}{
		{"0.1.0", SchemaVersion{0, 1, 0}, false},
		{"0.3.0", SchemaVersion{0, 3, 0}, false},
		{"v1.2.3", SchemaVersion{1, 2, 3}, false},
		{"1.0", SchemaVersion{1, 0, 0}, false},
		{"2", SchemaVersion{2, 0, 0}, false},
		{"", SchemaVersion{}, true},
		{"abc", SchemaVersion{}, true},
		{"1.2.3.4", SchemaVersion{}, true},
	}

	for _, tt := range tests {
		got, err := ParseSchemaVersion(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseSchemaVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("ParseSchemaVersion(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestSchemaVersionComparisons(t *testing.T) {
	v1 := SchemaVersion{0, 1, 0}
	v2 := SchemaVersion{0, 3, 0}
	v3 := SchemaVersion{1, 0, 0}

	if !v2.GreaterThan(v1) {
		t.Error("0.3.0 should be greater than 0.1.0")
	}
	if v1.GreaterThan(v2) {
		t.Error("0.1.0 should not be greater than 0.3.0")
	}
	if v1.LessThan(v2) != v2.GreaterThan(v1) {
		t.Error("LessThan and GreaterThan should be symmetric")
	}
	if !v1.CompatibleWith(v1) {
		t.Error("0.1.0 should be compatible with 0.1.0")
	}
	if !v2.CompatibleWith(v1) {
		t.Error("0.3.0 should be compatible with 0.1.0")
	}
	if v1.CompatibleWith(v3) {
		t.Error("0.1.0 should not be compatible with 1.0.0")
	}
}

func TestCheckSpecVersion(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"0.1.0", true},
		{"0.3.0", true},
		{"1.0.0", true},
		{"", true},
		{"0.0.1", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		result := CheckSpecVersion(tt.version)
		if result.Valid != tt.valid {
			t.Errorf("CheckSpecVersion(%q).Valid = %v, want %v (msg: %s)", tt.version, result.Valid, tt.valid, result.Message)
		}
	}
}

func TestExtractVersionFromData(t *testing.T) {
	t.Run("with version", func(t *testing.T) {
		data := map[string]any{"version": "0.3.0"}
		got := ExtractVersionFromData(data)
		if got != "0.3.0" {
			t.Errorf("got %q, want 0.3.0", got)
		}
	})

	t.Run("without version", func(t *testing.T) {
		data := map[string]any{"project": "test"}
		got := ExtractVersionFromData(data)
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("non-map", func(t *testing.T) {
		got := ExtractVersionFromData("not a map")
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

func TestSchemaVersionString(t *testing.T) {
	v := SchemaVersion{0, 3, 0}
	if v.String() != "0.3.0" {
		t.Errorf("String() = %q, want 0.3.0", v.String())
	}
}
