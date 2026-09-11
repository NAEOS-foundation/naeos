package monitoring

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDefaultSLOParams(t *testing.T) {
	slo := DefaultSLO("orders-api")
	if slo.Target != 0.99 {
		t.Errorf("expected 0.99 target, got %f", slo.Target)
	}
	if slo.Window != 30*24*time.Hour {
		t.Errorf("expected 30d window, got %v", slo.Window)
	}
	if _, ok := slo.BurnRateThresholds["critical"]; !ok {
		t.Error("expected critical threshold")
	}
}

func TestSLOErrorBudget(t *testing.T) {
	slo := DefaultSLO("api")
	want := time.Duration(float64(slo.Window) * 0.01)
	got := slo.ErrorBudget()
	if got != want {
		t.Errorf("expected budget %v, got %v", want, got)
	}
}

func TestSLOEvaluate(t *testing.T) {
	slo := DefaultSLO("api")

	report := slo.Evaluate(1000, 5, time.Hour)
	if report.ErrorRatio != 0.005 {
		t.Errorf("expected 0.005 ratio, got %f", report.ErrorRatio)
	}
	if report.Breaching {
		t.Error("5 errors per hour should not breach 99%/30d SLO")
	}
	if report.Requests != 1000 || report.Errors != 5 {
		t.Errorf("unexpected counts %d/%d", report.Requests, report.Errors)
	}
}

func TestSLOEvaluateBreaching(t *testing.T) {
	slo := DefaultSLO("api")
	report := slo.Evaluate(100, 10, time.Hour)
	if !report.Breaching {
		t.Error("10% error rate should breach budget")
	}
	if report.BudgetUsed == 0 {
		t.Error("expected budget used > 0")
	}
}

func TestSLOEvaluateZeroBudget(t *testing.T) {
	slo := DefaultSLO("api")
	slo.Target = 1.0
	report := slo.Evaluate(100, 1, time.Hour)
	if report.ErrorBudgetSeconds != 0 {
		t.Errorf("expected zero budget, got %f", report.ErrorBudgetSeconds)
	}
}

func TestSLOAlertRulesCount(t *testing.T) {
	slo := DefaultSLO("api")
	rules := slo.AlertRules()
	if len(rules) != 2 {
		t.Fatalf("expected 2 alert rules (critical/warning), got %d", len(rules))
	}

	// verify ordering is deterministic (critical first)
	if rules[0].Severity != "critical" {
		t.Errorf("expected critical first, got %s", rules[0].Severity)
	}
	if !strings.Contains(rules[0].Expr, "naeos_slo_errors_total") {
		t.Errorf("expected expression referencing error total: %s", rules[0].Expr)
	}
	if rules[0].For == "" {
		t.Error("expected 'for' duration set")
	}
}

func TestSLOAlertRulesSkipZeroThreshold(t *testing.T) {
	slo := DefaultSLO("api")
	slo.BurnRateThresholds["critical"] = 0
	rules := slo.AlertRules()
	for _, r := range rules {
		if r.Severity == "critical" {
			t.Error("expected critical rule skipped for zero threshold")
		}
	}
}

func TestSLORulesYAML(t *testing.T) {
	slo := DefaultSLO("api")
	yaml := slo.RulesYAML()

	if !strings.HasPrefix(yaml, "groups:\n") {
		t.Errorf("expected groups header, got %s", yaml)
	}
	for _, r := range slo.AlertRules() {
		if !strings.Contains(yaml, fmt.Sprintf("alert: %s", r.Alert)) {
			t.Errorf("expected alert %s in YAML", r.Alert)
		}
	}
}

func TestSLOAlertRulesSortable(t *testing.T) {
	slo := DefaultSLO("api")
	rules := slo.AlertRules()
	for i := 1; i < len(rules); i++ {
		if rules[i-1].Severity > rules[i].Severity {
			t.Errorf("rules not sorted by severity: %s before %s", rules[i-1].Severity, rules[i].Severity)
		}
	}
}

func TestMultipledBurnRateMax(t *testing.T) {
	slo := DefaultSLO("api")
	want := 14.4
	got := slo.MultipledBurnRate(1)
	if got != want {
		t.Errorf("expected %g, got %g", want, got)
	}
}

func TestMultipledBurnRateFallback(t *testing.T) {
	slo := DefaultSLO("api")
	slo.BurnRateThresholds = nil
	got := slo.MultipledBurnRate(2)
	if got != 2 {
		t.Errorf("expected fallback 2, got %g", got)
	}
}

func TestErrorBudgetUsed(t *testing.T) {
	slo := DefaultSLO("api")
	budget := slo.ErrorBudget()
	used := slo.ErrorBudgetUsed(budget, budget)
	if used < 0.99 || used > 1.01 {
		t.Errorf("expected ~1.0 budget used, got %f", used)
	}
	if slo.ErrorBudgetUsed(0, 0) != 0 {
		t.Error("expected 0 for empty window")
	}
}
