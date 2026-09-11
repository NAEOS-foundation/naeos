package monitoring

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// SLO defines a service level objective with a target and alerting burn rate.
type SLO struct {
	ID          string
	Name        string
	Service     string
	Description string
	// Target is the SLO objective as a ratio (0.99 = 99%).
	Target float64
	// Window is the SLO evaluation window.
	Window time.Duration
	// BurnRateThresholds map severity → burn rate over which an alert fires.
	BurnRateThresholds map[string]float64
	// Multiplier is the alert duration multiplier (default 1).
	Multiplier float64
}

type SLOAlertRule struct {
	Alert     string
	Expr      string
	For       string
	Severity  string
	Summary   string
	Labels    map[string]string
	Window    string
	BurnRate  float64
	Threshold float64
}

// ErrorBudget returns the total error budget in seconds over SLO.Window.
func (s SLO) ErrorBudget() time.Duration {
	return time.Duration(float64(s.Window) * (1 - s.Target))
}

// ErrorBudgetUsed returns the error budget consumed by measuredError over window.
func (s SLO) ErrorBudgetUsed(measuredError time.Duration, over time.Duration) float64 {
	if over <= 0 {
		return 0
	}
	budget := s.ErrorBudget()
	if budget <= 0 {
		return 1
	}
	return mathRatio(float64(measuredError), float64(budget))
}

// ErrorRateFromSamples computes the error ratio from request and error counts.
type SLOReport struct {
	SLO                SLO
	Requests           uint64
	Errors             uint64
	ErrorRatio         float64
	ErrorBudgetSeconds float64
	BudgetUsed         float64
	Breaching          bool
	BurnRate           float64
	MeasuredOver       time.Duration
}

func (s SLO) Evaluate(requests, errors uint64, over time.Duration) SLOReport {
	ratio := 0.0
	if requests > 0 {
		ratio = mathRatio(float64(errors), float64(requests))
	}
	measuredErr := time.Duration(float64(over) * ratio)
	budget := s.ErrorBudget()
	used := 0.0
	if budget > 0 {
		used = float64(measuredErr) / float64(budget)
	}

	// Burn rate = measured error rate / (1 - target). A burn rate of 1 means
	// the error budget is depleted exactly over the SLO window; >1 means the
	// budget would be exhausted faster than allowed.
	budgetFraction := 1.0 - s.Target
	burnRate := 0.0
	if budgetFraction > 0 {
		burnRate = mathRatio(ratio, budgetFraction)
	}

	return SLOReport{
		SLO:                s,
		Requests:           requests,
		Errors:             errors,
		ErrorRatio:         ratio,
		ErrorBudgetSeconds: budget.Seconds(),
		BudgetUsed:         used,
		Breaching:          burnRate >= 1,
		BurnRate:           burnRate,
		MeasuredOver:       over,
	}
}

// MultipledBurnRate returns the total burn rate for the worst configured threshold.
func (s SLO) MultipledBurnRate(fallback float64) float64 {
	worst := 0.0
	for _, v := range s.BurnRateThresholds {
		if v > worst {
			worst = v
		}
	}
	if worst == 0 {
		return fallback
	}
	return worst
}

// AlertRules generates Prometheus alerting rules for each burn-rate threshold.
func (s SLO) AlertRules() []SLOAlertRule {
	if s.Window == 0 {
		return nil
	}
	if s.Multiplier == 0 {
		s.Multiplier = 1
	}

	windowMinutes := int(s.Window.Minutes())
	if windowMinutes < 1 {
		windowMinutes = 1
	}

	rules := make([]SLOAlertRule, 0, len(s.BurnRateThresholds))
	levels := make([]string, 0, len(s.BurnRateThresholds))
	for level := range s.BurnRateThresholds {
		levels = append(levels, level)
	}
	sort.Strings(levels)

	for _, level := range levels {
		burnRate := s.BurnRateThresholds[level]
		if burnRate <= 0 {
			continue
		}
		mult := int(s.Multiplier)
		expr := fmt.Sprintf(
			"sum(rate(naeos_slo_errors_total{service=%q}[%dm%d])) / sum(rate(naeos_slo_requests_total{service=%q}[%dm%d])) > %g",
			s.Service, windowMinutes, mult, s.Service, windowMinutes, mult, burnRate*amountToBudget(s),
		)
		forDur := fmt.Sprintf("%dm", windowMinutes*mult)

		rules = append(rules, SLOAlertRule{
			Alert:     fmt.Sprintf("SLO%sBurnRate%s", s.ID, strings.ToUpper(level)),
			Expr:      expr,
			For:       forDur,
			Severity:  level,
			Summary:   fmt.Sprintf("%s SLO %s (%.2f%%) burning at %gx", s.Service, s.Name, s.Target*100, burnRate),
			Labels:    map[string]string{"service": s.Service, "slo": s.ID},
			Window:    s.Window.String(),
			BurnRate:  burnRate,
			Threshold: s.Target,
		})
	}
	return rules
}

func amountToBudget(s SLO) float64 {
	return 1 - s.Target
}

func (s SLO) RulesYAML() string {
	var sb strings.Builder
	sb.WriteString("groups:\n")
	sb.WriteString("  - name: naeos-slo\n")
	sb.WriteString("    rules:\n")
	for _, r := range s.AlertRules() {
		sb.WriteString(fmt.Sprintf("      - alert: %s\n", r.Alert))
		sb.WriteString(fmt.Sprintf("        expr: %s\n", r.Expr))
		sb.WriteString(fmt.Sprintf("        for: %s\n", r.For))
		sb.WriteString("        labels:\n")
		sk := make([]string, 0, len(r.Labels))
		for k := range r.Labels {
			sk = append(sk, k)
		}
		sort.Strings(sk)
		for _, k := range sk {
			sb.WriteString(fmt.Sprintf("          %s: %q\n", k, r.Labels[k]))
		}
		sb.WriteString("        annotations:\n")
		sb.WriteString(fmt.Sprintf("          summary: %q\n", r.Summary))
	}
	return sb.String()
}

func mathRatio(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

// DefaultSLO returns a reference SLO: 99% availability over 30 days.
func DefaultSLO(service string) SLO {
	return SLO{
		ID:      "P99",
		Name:    "availability",
		Service: service,
		Target:  0.99,
		Window:  30 * 24 * time.Hour,
		BurnRateThresholds: map[string]float64{
			"critical": 14.4, // consume 5% of budget per hour
			"warning":  3.0,  // consume 50% of budget in ~4h window
		},
		Multiplier: 1,
	}
}
