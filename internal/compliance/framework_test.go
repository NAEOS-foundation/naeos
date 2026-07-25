package compliance

import (
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/audit"
)

func TestFrameworksDefined(t *testing.T) {
	if len(Frameworks) != 3 {
		t.Errorf("expected 3 frameworks, got %d", len(Frameworks))
	}

	for name, def := range Frameworks {
		if def.Name == "" {
			t.Errorf("framework %q: empty name", name)
		}
		if len(def.Controls) == 0 {
			t.Errorf("framework %q: no controls defined", name)
		}
	}
}

func TestSOC2Controls(t *testing.T) {
	def, ok := Frameworks[FrameworkSOC2]
	if !ok {
		t.Fatal("SOC2 framework not found")
	}

	expectedControls := []string{"CC1.1", "CC2.1", "CC3.1", "CC4.1", "CC5.1", "CC6.1", "CC7.1", "CC8.1"}
	if len(def.Controls) != len(expectedControls) {
		t.Errorf("expected %d SOC2 controls, got %d", len(expectedControls), len(def.Controls))
	}

	for _, exp := range expectedControls {
		var found bool
		for _, c := range def.Controls {
			if c.ID == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SOC2 control %q not found", exp)
		}
	}
}

func TestHIPPAControls(t *testing.T) {
	def, ok := Frameworks[FrameworkHIPAA]
	if !ok {
		t.Fatal("HIPAA framework not found")
	}

	if len(def.Controls) != 11 {
		t.Errorf("expected 11 HIPAA controls, got %d", len(def.Controls))
	}
}

func TestGDPRControls(t *testing.T) {
	def, ok := Frameworks[FrameworkGDPR]
	if !ok {
		t.Fatal("GDPR framework not found")
	}

	if len(def.Controls) != 8 {
		t.Errorf("expected 8 GDPR controls, got %d", len(def.Controls))
	}
}

func TestGenerateReportNoEvents(t *testing.T) {
	report := GenerateReport(FrameworkSOC2, nil)
	if report == nil {
		t.Fatal("expected non-nil report")
	}

	if report.Framework != FrameworkSOC2 {
		t.Errorf("expected SOC2, got %s", report.Framework)
	}
	if report.TotalControls != 8 {
		t.Errorf("expected 8 controls, got %d", report.TotalControls)
	}
	if len(report.ControlStatuses) != 8 {
		t.Errorf("expected 8 statuses, got %d", len(report.ControlStatuses))
	}
}

func TestGenerateReportWithEvents(t *testing.T) {
	events := []audit.AuditEvent{
		{UserID: "u1", Action: "create", Resource: "user", Status: "success"},
		{UserID: "u2", Action: "delete", Resource: "user", Status: "failed"},
		{UserID: "u1", Action: "update", Resource: "config", Status: "success"},
	}

	report := GenerateReport(FrameworkSOC2, events)
	if report == nil {
		t.Fatal("expected non-nil report")
	}

	if report.PassedControls == 0 {
		t.Error("expected at least some controls to pass with events")
	}
}

func TestGenerateReportUnsupported(t *testing.T) {
	report := GenerateReport("unsupported", nil)
	if report != nil {
		t.Error("expected nil for unsupported framework")
	}
}

func TestGenerateReportGDPR(t *testing.T) {
	events := []audit.AuditEvent{
		{UserID: "u1", Action: "delete", Resource: "user", Status: "success"},
	}

	report := GenerateReport(FrameworkGDPR, events)
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Framework != FrameworkGDPR {
		t.Errorf("expected GDPR, got %s", report.Framework)
	}

	// Art. 17 (right to erasure) should pass with delete:user events
	var art17Passed bool
	for _, cs := range report.ControlStatuses {
		if cs.ControlID == "Art. 17" && cs.Passed {
			art17Passed = true
			break
		}
	}
	if !art17Passed {
		t.Error("expected Art. 17 to pass with delete:user events")
	}
}

func TestGenerateReportHIPAA(t *testing.T) {
	events := []audit.AuditEvent{
		{UserID: "u1", Action: "create", Resource: "user", Status: "success"},
		{UserID: "u2", Action: "read", Resource: "report", Status: "error"},
	}

	report := GenerateReport(FrameworkHIPAA, events)
	if report == nil {
		t.Fatal("expected non-nil report")
	}

	// 164.308(a)(6) (incident response) should pass with error events
	var irPassed bool
	for _, cs := range report.ControlStatuses {
		if cs.ControlID == "164.308(a)(6)" && cs.Passed {
			irPassed = true
			break
		}
	}
	if !irPassed {
		t.Error("expected 164.308(a)(6) to pass with error events")
	}
}

func TestGenerateReportTimestamp(t *testing.T) {
	report := GenerateReport(FrameworkSOC2, nil)
	if report.GeneratedAt.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if time.Since(report.GeneratedAt) > time.Minute {
		t.Error("expected recent timestamp")
	}
}

func TestGenerateReportEvidence(t *testing.T) {
	report := GenerateReport(FrameworkSOC2, []audit.AuditEvent{
		{UserID: "u1", Action: "create", Resource: "user", Status: "success"},
	})
	if len(report.Evidence) == 0 {
		t.Error("expected evidence entries")
	}
	for _, ev := range report.Evidence {
		if ev.Type == "" {
			t.Error("expected evidence type")
		}
		if ev.Source == "" {
			t.Error("expected evidence source")
		}
	}
}
