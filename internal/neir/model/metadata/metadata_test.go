package metadata

import (
	"testing"
	"time"
)

func TestMetadata_ZeroValue(t *testing.T) {
	var m Metadata
	if m.NEIRVersion != "" {
		t.Error("expected empty NEIRVersion")
	}
	if m.CreatedAt != nil {
		t.Error("expected nil CreatedAt")
	}
	if m.Tags != nil {
		t.Error("expected nil Tags")
	}
}

func TestMetadata_Full(t *testing.T) {
	now := time.Now()
	m := Metadata{
		NEIRVersion:    "2.0",
		SchemaVersion:  "1.0",
		ProjectVersion: "0.1.0",
		CreatedAt:      &now,
		ModifiedAt:     &now,
		Tags:           []string{"backend", "api"},
		Labels:         map[string]string{"team": "platform"},
		Source:         &SourceRef{Kind: "file", Ref: "naeos.yaml"},
	}
	if m.NEIRVersion != "2.0" {
		t.Errorf("expected 2.0, got %s", m.NEIRVersion)
	}
	if len(m.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(m.Tags))
	}
	if m.Labels["team"] != "platform" {
		t.Errorf("expected platform, got %s", m.Labels["team"])
	}
	if m.Source.Kind != "file" {
		t.Errorf("expected file, got %s", m.Source.Kind)
	}
	if !m.CreatedAt.Equal(now) {
		t.Error("CreatedAt not set correctly")
	}
}

func TestMetadata_SourceRefNil(t *testing.T) {
	var m Metadata
	if m.Source != nil {
		t.Error("expected nil Source")
	}
}
