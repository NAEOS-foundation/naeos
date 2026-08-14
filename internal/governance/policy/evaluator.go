package policy

import (
	"fmt"
	"strconv"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type Rule struct {
	RuleID    string
	Condition string
	Priority  int
	Action    string
	Scope     string
	Enabled   bool
}

type EvaluationResult struct {
	Passed   bool
	RuleID   string
	Message  string
	Action   string
	Priority int
}

type Evaluator interface {
	Evaluate(ctx map[string]any) error
	EvaluateRules(rules []Rule, ctx map[string]any) ([]EvaluationResult, error)
}

type DefaultEvaluator struct{}

func NewEvaluator() Evaluator {
	return DefaultEvaluator{}
}

func (DefaultEvaluator) Evaluate(ctx map[string]any) error {
	if ctx == nil {
		return naeoserr.New(naeoserr.ErrValidation, "context is nil")
	}
	return nil
}

func (DefaultEvaluator) EvaluateRules(rules []Rule, ctx map[string]any) ([]EvaluationResult, error) {
	if ctx == nil {
		return nil, naeoserr.New(naeoserr.ErrValidation, "context is nil")
	}

	var results []EvaluationResult
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		result := evaluateRule(rule, ctx)
		results = append(results, result)
	}

	return results, nil
}

func evaluateRule(rule Rule, ctx map[string]any) EvaluationResult {
	passed := true
	message := "rule passed"

	condition := strings.TrimSpace(rule.Condition)
	if condition == "" {
		passed = true
		message = "no condition specified, default pass"
	} else {
		parts := strings.SplitN(condition, ":", 2)
		if len(parts) == 2 {
			op := strings.TrimSpace(parts[0])
			args := strings.TrimSpace(parts[1])

			switch op {
			case "exists":
				if _, exists := ctx[args]; !exists {
					passed = false
					message = fmt.Sprintf("key %s not found in context", args)
				} else {
					message = fmt.Sprintf("key %s exists in context", args)
				}
			case "not_empty":
				if actual, exists := ctx[args]; exists {
					actualStr := fmt.Sprintf("%v", actual)
					if actualStr == "" {
						passed = false
						message = fmt.Sprintf("key %s is empty", args)
					} else {
						message = fmt.Sprintf("key %s is not empty", args)
					}
				} else {
					passed = false
					message = fmt.Sprintf("key %s not found in context", args)
				}
			case "contains":
				subParts := strings.SplitN(args, ",", 2)
				if len(subParts) == 2 {
					key := strings.TrimSpace(subParts[0])
					substr := strings.TrimSpace(subParts[1])
					if actual, exists := ctx[key]; exists {
						actualStr := fmt.Sprintf("%v", actual)
						if !strings.Contains(actualStr, substr) {
							passed = false
							message = fmt.Sprintf("expected %s to contain %s, got %s", key, substr, actualStr)
						} else {
							message = fmt.Sprintf("condition met: %s contains %s", key, substr)
						}
					} else {
						passed = false
						message = fmt.Sprintf("key %s not found in context", key)
					}
				}
			case "gt":
				subParts := strings.SplitN(args, ",", 2)
				if len(subParts) == 2 {
					key := strings.TrimSpace(subParts[0])
					thresholdStr := strings.TrimSpace(subParts[1])
					if actual, exists := ctx[key]; exists {
						actualStr := fmt.Sprintf("%v", actual)
						actualNum, err1 := strconv.ParseFloat(actualStr, 64)
						thresholdNum, err2 := strconv.ParseFloat(thresholdStr, 64)
						if err1 == nil && err2 == nil {
							if actualNum <= thresholdNum {
								passed = false
								message = fmt.Sprintf("expected %s > %s, got %s", key, thresholdStr, actualStr)
							} else {
								message = fmt.Sprintf("condition met: %s > %s", key, thresholdStr)
							}
						} else {
							passed = false
							message = fmt.Sprintf("cannot compare non-numeric values: %s=%s", key, actualStr)
						}
					} else {
						passed = false
						message = fmt.Sprintf("key %s not found in context", key)
					}
				}
			case "lt":
				subParts := strings.SplitN(args, ",", 2)
				if len(subParts) == 2 {
					key := strings.TrimSpace(subParts[0])
					thresholdStr := strings.TrimSpace(subParts[1])
					if actual, exists := ctx[key]; exists {
						actualStr := fmt.Sprintf("%v", actual)
						actualNum, err1 := strconv.ParseFloat(actualStr, 64)
						thresholdNum, err2 := strconv.ParseFloat(thresholdStr, 64)
						if err1 == nil && err2 == nil {
							if actualNum >= thresholdNum {
								passed = false
								message = fmt.Sprintf("expected %s < %s, got %s", key, thresholdStr, actualStr)
							} else {
								message = fmt.Sprintf("condition met: %s < %s", key, thresholdStr)
							}
						} else {
							passed = false
							message = fmt.Sprintf("cannot compare non-numeric values: %s=%s", key, actualStr)
						}
					} else {
						passed = false
						message = fmt.Sprintf("key %s not found in context", key)
					}
				}
			case "in":
				subParts := strings.SplitN(args, ",", 2)
				if len(subParts) == 2 {
					key := strings.TrimSpace(subParts[0])
					optionsStr := strings.TrimSpace(subParts[1])
					options := strings.Split(optionsStr, ",")
					if actual, exists := ctx[key]; exists {
						actualStr := fmt.Sprintf("%v", actual)
						found := false
						for _, opt := range options {
							if strings.TrimSpace(opt) == actualStr {
								found = true
								break
							}
						}
						if !found {
							passed = false
							message = fmt.Sprintf("expected %s to be one of [%s], got %s", key, optionsStr, actualStr)
						} else {
							message = fmt.Sprintf("condition met: %s is in allowed values", key)
						}
					} else {
						passed = false
						message = fmt.Sprintf("key %s not found in context", key)
					}
				}
			default:
				key := op
				expected := args
				if actual, exists := ctx[key]; exists {
					actualStr := fmt.Sprintf("%v", actual)
					if actualStr != expected {
						passed = false
						message = fmt.Sprintf("expected %s=%s, got %s", key, expected, actualStr)
					} else {
						message = fmt.Sprintf("condition met: %s=%s", key, expected)
					}
				} else {
					passed = false
					message = fmt.Sprintf("key %s not found in context", key)
				}
			}
		}
	}

	return EvaluationResult{
		Passed:   passed,
		RuleID:   rule.RuleID,
		Message:  message,
		Action:   rule.Action,
		Priority: rule.Priority,
	}
}

func DefaultRules() []Rule {
	return []Rule{
		{RuleID: "project-required", Condition: "exists:project", Priority: 1, Action: "block", Scope: "spec", Enabled: true},
		{RuleID: "modules-required", Condition: "exists:modules", Priority: 1, Action: "block", Scope: "spec", Enabled: true},
		{RuleID: "architecture-pattern-valid", Condition: "in:architecture_pattern,hexagonal,layered,clean,event-driven,cqrs,microkernel,monolith,monolithic,microservices,serverless", Priority: 2, Action: "warn", Scope: "spec", Enabled: true},
		{RuleID: "deployment-strategy-valid", Condition: "in:deployment_strategy,rolling,blue-green,canary,recreate", Priority: 2, Action: "warn", Scope: "spec", Enabled: true},
		{RuleID: "service-port-positive", Condition: "gt:service_port,0", Priority: 3, Action: "warn", Scope: "spec", Enabled: true},
	}
}
