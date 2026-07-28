//go:build integration

package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/audit"
	"github.com/NAEOS-foundation/naeos/internal/compliance"
)

type localExporter struct {
	dir string
}

func (e *localExporter) Upload(path string, data []byte) error {
	fullPath := filepath.Join(e.dir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0600)
}

func (e *localExporter) List(path string) ([]string, error) {
	var out []string
	entries, err := os.ReadDir(filepath.Join(e.dir, path))
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		out = append(out, entry.Name())
	}
	return out, nil
}

func TestAuditComplianceCloudIntegration(t *testing.T) {
	store := audit.NewMemoryAuditor()

	now := time.Now()

	events := []audit.AuditEvent{
		{
			ID: "evt-1", Timestamp: now.Add(-10 * time.Minute),
			UserID: "alice", Action: "create", Resource: "user", ResourceID: "user-42",
			IP: "10.0.0.1", Status: "success", Details: "created user alice",
		},
		{
			ID: "evt-2", Timestamp: now.Add(-9 * time.Minute),
			UserID: "bob", Action: "login", Resource: "session", ResourceID: "sess-1",
			IP: "10.0.0.2", Status: "success", Details: "login ok",
		},
		{
			ID: "evt-3", Timestamp: now.Add(-8 * time.Minute),
			UserID: "alice", Action: "create", Resource: "user", ResourceID: "user-43",
			IP: "10.0.0.1", Status: "success",
		},
		{
			ID: "evt-4", Timestamp: now.Add(-7 * time.Minute),
			UserID: "bob", Action: "read", Resource: "record", ResourceID: "rec-100",
			Status: "success",
		},
		{
			ID: "evt-5", Timestamp: now.Add(-6 * time.Minute),
			UserID: "bob", Action: "write", Resource: "record", ResourceID: "rec-100",
			Status: "failed", Details: "permission denied",
		},
		{
			ID: "evt-6", Timestamp: now.Add(-5 * time.Minute),
			UserID: "system", Action: "update", Resource: "config", ResourceID: "tls-config",
			Status: "success",
		},
		{
			ID: "evt-7", Timestamp: now.Add(-4 * time.Minute),
			UserID: "bob", Action: "delete", Resource: "user", ResourceID: "user-99",
			Status: "success", Details: "GDPR erasure request processed",
		},
	}

	for _, e := range events {
		if err := store.Log(e); err != nil {
			t.Fatalf("failed to log event: %v", err)
		}
	}

	stored := store.Events()
	if len(stored) != len(events) {
		t.Fatalf("expected %d stored events, got %d", len(events), len(stored))
	}

	t.Run("soc2 compliance", func(t *testing.T) {
		report := compliance.GenerateReport(compliance.FrameworkSOC2, stored)
		if report == nil {
			t.Fatal("expected non-nil SOC2 report")
		}
		if report.TotalControls != 8 {
			t.Errorf("expected 8 SOC2 controls, got %d", report.TotalControls)
		}
		if report.PassedControls == 0 {
			t.Error("expected at least some passed controls")
		}
		if report.NotApplicable < 0 {
			t.Error("not_applicable should not be negative")
		}

		statusMap := make(map[string]*compliance.ControlStatus)
		for i := range report.ControlStatuses {
			cs := &report.ControlStatuses[i]
			statusMap[cs.ControlID] = cs
		}

		cc61, ok := statusMap["CC6.1"]
		if !ok {
			t.Fatal("expected CC6.1 control")
		}
		if !cc61.Passed {
			t.Error("CC6.1 should pass (user:create events exist)")
		}

		cc71, ok := statusMap["CC7.1"]
		if !ok {
			t.Fatal("expected CC7.1 control")
		}
		if !cc71.Passed {
			t.Error("CC7.1 should pass (failed/error events exist)")
		}
	})

	t.Run("hipaa compliance", func(t *testing.T) {
		report := compliance.GenerateReport(compliance.FrameworkHIPAA, stored)
		if report == nil {
			t.Fatal("expected non-nil HIPAA report")
		}
		if report.TotalControls != 11 {
			t.Errorf("expected 11 HIPAA controls, got %d", report.TotalControls)
		}

		statusMap := make(map[string]*compliance.ControlStatus)
		for i := range report.ControlStatuses {
			cs := &report.ControlStatuses[i]
			statusMap[cs.ControlID] = cs
		}

		for _, ctrl := range []string{"164.308(a)(1)", "164.308(a)(6)", "164.308(a)(8)"} {
			cs, ok := statusMap[ctrl]
			if !ok {
				t.Fatalf("expected control %s", ctrl)
			}
			if !cs.Passed {
				t.Errorf("control %s should pass with existing audit events", ctrl)
			}
		}
	})

	t.Run("gdpr compliance", func(t *testing.T) {
		report := compliance.GenerateReport(compliance.FrameworkGDPR, stored)
		if report == nil {
			t.Fatal("expected non-nil GDPR report")
		}
		if report.TotalControls != 8 {
			t.Errorf("expected 8 GDPR controls, got %d", report.TotalControls)
		}

		statusMap := make(map[string]*compliance.ControlStatus)
		for i := range report.ControlStatuses {
			cs := &report.ControlStatuses[i]
			statusMap[cs.ControlID] = cs
		}

		art17, ok := statusMap["Art. 17"]
		if !ok {
			t.Fatal("expected Art. 17 control")
		}
		if !art17.Passed {
			t.Error("Art. 17 (right to erasure) should pass (data:delete event exists)")
		}
	})

	t.Run("cloud export audit events", func(t *testing.T) {
		tmpDir := t.TempDir()
		exporter := &localExporter{dir: tmpDir}

		path, err := audit.UploadToCloud(exporter, "audit/", stored)
		if err != nil {
			t.Fatalf("UploadToCloud failed: %v", err)
		}
		if path == "" {
			t.Fatal("expected non-empty upload path")
		}

		if !strings.HasPrefix(path, "audit/") {
			t.Errorf("expected path to have prefix audit/, got %s", path)
		}

		fullPath := filepath.Join(tmpDir, path)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("failed to read exported file: %v", err)
		}

		var exported []audit.AuditEvent
		if err := json.Unmarshal(data, &exported); err != nil {
			t.Fatalf("exported file is not valid JSON: %v", err)
		}
		if len(exported) != len(events) {
			t.Errorf("expected %d exported events, got %d", len(events), len(exported))
		}

		entries, err := exporter.List("audit/")
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(entries) != 1 {
			t.Errorf("expected 1 file in audit/ prefix, got %d", len(entries))
		}
	})

	t.Run("compliance with empty events", func(t *testing.T) {
		for _, fw := range []compliance.Framework{compliance.FrameworkSOC2, compliance.FrameworkHIPAA, compliance.FrameworkGDPR} {
			report := compliance.GenerateReport(fw, nil)
			if report == nil {
				t.Errorf("expected non-nil report for %s", fw)
				continue
			}
			if report.TotalControls == 0 {
				t.Errorf("expected controls for %s", fw)
			}
		}
	})

	t.Run("compliance with unknown framework", func(t *testing.T) {
		report := compliance.GenerateReport("unknown-framework", stored)
		if report != nil {
			t.Error("expected nil report for unknown framework")
		}
	})
}
