package docs

import "testing"

func TestDocKindConstants(t *testing.T) {
	tests := []struct {
		k    DocKind
		want string
	}{
		{KindGuide, "guide"},
		{KindReference, "reference"},
		{KindADR, "adr"},
		{KindRFC, "rfc"},
		{KindChangelog, "changelog"},
	}
	for _, tt := range tests {
		if string(tt.k) != tt.want {
			t.Errorf("DocKind(%s) = %s, want %s", tt.want, string(tt.k), tt.want)
		}
	}
}

func TestDocumentation_ZeroValue(t *testing.T) {
	var d Documentation
	if d.Guides != nil {
		t.Error("expected nil Guides")
	}
	if d.Attributes != nil {
		t.Error("expected nil Attributes")
	}
}

func TestDocumentation_Full(t *testing.T) {
	d := Documentation{
		Guides: []Doc{
			{Title: "Getting Started", Path: "/docs/start", Kind: KindGuide},
		},
		ADRs: []Doc{
			{Title: "ADR-001", Path: "/adrs/001", Kind: KindADR},
		},
		RFCs: []Doc{
			{Title: "RFC-001", Path: "/rfcs/001", Kind: KindRFC},
		},
		Attributes: map[string]string{"key": "val"},
	}
	if len(d.Guides) != 1 {
		t.Errorf("expected 1 guide, got %d", len(d.Guides))
	}
	if d.Guides[0].Kind != KindGuide {
		t.Errorf("expected guide kind, got %s", d.Guides[0].Kind)
	}
	if len(d.ADRs) != 1 {
		t.Errorf("expected 1 adr, got %d", len(d.ADRs))
	}
	if d.RFCs[0].Title != "RFC-001" {
		t.Errorf("expected RFC-001, got %s", d.RFCs[0].Title)
	}
	if d.Attributes["key"] != "val" {
		t.Errorf("expected val, got %s", d.Attributes["key"])
	}
}

func TestDoc_ZeroValue(t *testing.T) {
	var d Doc
	if d.Title != "" {
		t.Error("expected empty Title")
	}
}
