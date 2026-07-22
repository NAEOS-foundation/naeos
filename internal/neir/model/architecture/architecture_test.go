package architecture

import "testing"

func TestPatternConstants(t *testing.T) {
	tests := []struct {
		p    Pattern
		want string
	}{
		{PatternLayered, "layered"},
		{PatternClean, "clean"},
		{PatternHexagonal, "hexagonal"},
		{PatternMicrokernel, "microkernel"},
		{PatternEventDriven, "event-driven"},
		{PatternCQRS, "cqrs"},
		{PatternMonolith, "monolith"},
	}
	for _, tt := range tests {
		if string(tt.p) != tt.want {
			t.Errorf("Pattern(%s) = %s, want %s", tt.want, string(tt.p), tt.want)
		}
	}
}

func TestArchitecture_ZeroValue(t *testing.T) {
	var a Architecture
	if a.Pattern != "" {
		t.Error("expected empty Pattern")
	}
	if a.Layers != nil {
		t.Error("expected nil Layers")
	}
}

func TestArchitecture_Full(t *testing.T) {
	a := Architecture{
		Pattern:     PatternHexagonal,
		Style:       "domain-driven",
		Description: "Hexagonal architecture",
		Principles:  []string{"single-responsibility", "dependency-inversion"},
		Layers: []Layer{
			{Name: "domain", Description: "Core domain", Modules: []string{"core"}},
		},
		Attributes: map[string]string{"key": "val"},
	}
	if a.Pattern != PatternHexagonal {
		t.Errorf("expected hexagonal, got %s", a.Pattern)
	}
	if len(a.Principles) != 2 {
		t.Errorf("expected 2 principles, got %d", len(a.Principles))
	}
	if len(a.Layers) != 1 {
		t.Errorf("expected 1 layer, got %d", len(a.Layers))
	}
	if a.Layers[0].Name != "domain" {
		t.Errorf("expected domain, got %s", a.Layers[0].Name)
	}
}
