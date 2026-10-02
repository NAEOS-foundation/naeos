// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type Rule struct {
	RuleID    string `json:"rule_id" yaml:"rule_id"`
	Condition string `json:"condition" yaml:"condition"`
	Priority  int    `json:"priority" yaml:"priority"`
	Action    string `json:"action" yaml:"action"`
	Scope     string `json:"scope" yaml:"scope"`
	Enabled   bool   `json:"enabled" yaml:"enabled"`
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
		if strings.TrimSpace(rule.Condition) == "" {
			return nil, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("rule %q has empty condition", rule.RuleID))
		}
		result := evaluateRule(rule, ctx)
		results = append(results, result)
	}

	return results, nil
}

func parseFiniteFloat(value string) (float64, error) {
	n, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, fmt.Errorf("non-finite numeric value")
	}
	return n, nil
}

// resolveContextPath resolves a condition key against a policy context.
//
// A key may be dotted ("architecture.pattern") to reach into nested objects.
// When an intermediate value is a slice the lookup fans out and collects the
// remaining path from every element, so "services.port" yields every service
// port. A fan-out that matches nothing is reported as not found, so a rule
// referencing a field no element carries fails closed instead of silently
// passing.
func resolveContextPath(ctx map[string]any, path string) (any, bool) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, false
	}

	segments := strings.Split(path, ".")
	value, found := ctx[segments[0]]
	if !found {
		return nil, false
	}

	for _, segment := range segments[1:] {
		switch current := value.(type) {
		case map[string]any:
			next, present := current[segment]
			if !present {
				return nil, false
			}
			value = next
		case []any:
			fanned := make([]any, 0, len(current))
			for _, element := range current {
				elementMap, isMap := element.(map[string]any)
				if !isMap {
					continue
				}
				if field, present := elementMap[segment]; present {
					fanned = append(fanned, field)
				}
			}
			if len(fanned) == 0 {
				return nil, false
			}
			value = fanned
		case []map[string]any:
			fanned := make([]any, 0, len(current))
			for _, element := range current {
				if field, present := element[segment]; present {
					fanned = append(fanned, field)
				}
			}
			if len(fanned) == 0 {
				return nil, false
			}
			value = fanned
		default:
			return nil, false
		}
	}

	return value, true
}

// contextValues flattens a resolved context value into the scalars a comparison
// applies to. Scalars yield themselves; slices yield their elements, so a
// comparison against "services.port" must hold for every service.
func contextValues(value any) []any {
	switch typed := value.(type) {
	case []any:
		return typed
	case []map[string]any:
		out := make([]any, len(typed))
		for i, element := range typed {
			out[i] = element
		}
		return out
	default:
		return []any{value}
	}
}

// everyStringValue applies check to every value a condition key resolves to.
// A dotted path that fans out over a slice must satisfy the comparison for
// every element, so "in:architecture.pattern,..." checks the pattern and
// "gt:services.port,0" requires every service to declare a positive port. It
// returns whether all held plus the first offending value, for the message.
func everyStringValue(value any, check func(actualStr string) bool) (bool, string) {
	values := contextValues(value)
	if len(values) == 0 {
		return false, ""
	}
	for _, element := range values {
		actualStr := fmt.Sprintf("%v", element)
		if !check(actualStr) {
			return false, actualStr
		}
	}
	return true, ""
}

// numericCheck builds a predicate that parses the resolved value and applies
// cmp to it. A value that is not a finite number is reported as not comparable.
func numericCheck(threshold string, cmp func(actual, threshold float64) bool) (check func(string) (bool, bool), ok bool) {
	thresholdNum, err := parseFiniteFloat(threshold)
	if err != nil {
		return func(string) (bool, bool) { return false, false }, false
	}
	return func(actualStr string) (bool, bool) {
		actualNum, err := parseFiniteFloat(actualStr)
		if err != nil {
			return false, false
		}
		return true, cmp(actualNum, thresholdNum)
	}, true
}

// everyNumericValue applies a numeric comparison to every value a condition key
// resolves to, keeping "not comparable" distinct from "compares false". That
// distinction is load bearing: a non-finite operand must be rejected as such
// rather than reported as an ordinary mismatch, so the rule fails closed with
// an accurate reason. offending is the first value that broke the comparison.
func everyNumericValue(value any, check func(actualStr string) (ok, holds bool)) (comparable, held bool, offending string) {
	values := contextValues(value)
	if len(values) == 0 {
		return false, false, ""
	}
	for _, element := range values {
		actualStr := fmt.Sprintf("%v", element)
		ok, satisfied := check(actualStr)
		if !ok {
			return false, false, actualStr
		}
		if !satisfied {
			return true, false, actualStr
		}
	}
	return true, true, ""
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
				if actual, exists := resolveContextPath(ctx, args); !exists || actual == nil || strings.TrimSpace(fmt.Sprintf("%v", actual)) == "" {
					passed = false
					message = fmt.Sprintf("key %s is absent from context", args)
				} else {
					message = fmt.Sprintf("key %s exists in context", args)
				}
			case "not_empty":
				if actual, exists := resolveContextPath(ctx, args); exists {
					actualStr := fmt.Sprintf("%v", actual)
					if actual == nil || strings.TrimSpace(actualStr) == "" {
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
					if actual, exists := resolveContextPath(ctx, key); exists {
						if held, offending := everyStringValue(actual, func(s string) bool {
							return strings.Contains(s, substr)
						}); !held {
							passed = false
							message = fmt.Sprintf("expected %s to contain %s, got %s", key, substr, offending)
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
					if actual, exists := resolveContextPath(ctx, key); exists {
						check, ok := numericCheck(thresholdStr, func(a, t float64) bool { return a > t })
						comparable, held, offending := false, false, ""
						if ok {
							comparable, held, offending = everyNumericValue(actual, check)
						}
						switch {
						case !ok || !comparable:
							passed = false
							message = fmt.Sprintf("cannot compare non-finite or non-numeric values: %s=%s", key, offending)
						case !held:
							passed = false
							message = fmt.Sprintf("expected %s > %s, got %s", key, thresholdStr, offending)
						default:
							message = fmt.Sprintf("condition met: %s > %s", key, thresholdStr)
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
					if actual, exists := resolveContextPath(ctx, key); exists {
						check, ok := numericCheck(thresholdStr, func(a, t float64) bool { return a < t })
						comparable, held, offending := false, false, ""
						if ok {
							comparable, held, offending = everyNumericValue(actual, check)
						}
						switch {
						case !ok || !comparable:
							passed = false
							message = fmt.Sprintf("cannot compare non-finite or non-numeric values: %s=%s", key, offending)
						case !held:
							passed = false
							message = fmt.Sprintf("expected %s < %s, got %s", key, thresholdStr, offending)
						default:
							message = fmt.Sprintf("condition met: %s < %s", key, thresholdStr)
						}
					} else {
						passed = false
						message = fmt.Sprintf("key %s not found in context", key)
					}
				}
			case "gte":
				subParts := strings.SplitN(args, ",", 2)
				if len(subParts) == 2 {
					key := strings.TrimSpace(subParts[0])
					thresholdStr := strings.TrimSpace(subParts[1])
					if actual, exists := resolveContextPath(ctx, key); exists {
						check, ok := numericCheck(thresholdStr, func(a, t float64) bool { return a >= t })
						comparable, held, offending := false, false, ""
						if ok {
							comparable, held, offending = everyNumericValue(actual, check)
						}
						switch {
						case !ok || !comparable:
							passed = false
							message = fmt.Sprintf("cannot compare non-finite or non-numeric values: %s=%s", key, offending)
						case !held:
							passed = false
							message = fmt.Sprintf("expected %s >= %s, got %s", key, thresholdStr, offending)
						default:
							message = fmt.Sprintf("condition met: %s >= %s", key, thresholdStr)
						}
					} else {
						passed = false
						message = fmt.Sprintf("key %s not found in context", key)
					}
				}
			case "lte":
				subParts := strings.SplitN(args, ",", 2)
				if len(subParts) == 2 {
					key := strings.TrimSpace(subParts[0])
					thresholdStr := strings.TrimSpace(subParts[1])
					if actual, exists := resolveContextPath(ctx, key); exists {
						check, ok := numericCheck(thresholdStr, func(a, t float64) bool { return a <= t })
						comparable, held, offending := false, false, ""
						if ok {
							comparable, held, offending = everyNumericValue(actual, check)
						}
						switch {
						case !ok || !comparable:
							passed = false
							message = fmt.Sprintf("cannot compare non-finite or non-numeric values: %s=%s", key, offending)
						case !held:
							passed = false
							message = fmt.Sprintf("expected %s <= %s, got %s", key, thresholdStr, offending)
						default:
							message = fmt.Sprintf("condition met: %s <= %s", key, thresholdStr)
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
					if actual, exists := resolveContextPath(ctx, key); exists {
						allowed := make(map[string]bool, len(options))
						for _, opt := range options {
							allowed[strings.TrimSpace(opt)] = true
						}
						if held, offending := everyStringValue(actual, func(s string) bool {
							return allowed[s]
						}); !held {
							passed = false
							message = fmt.Sprintf("expected %s to be one of [%s], got %s", key, optionsStr, offending)
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

// DefaultRules returns the built-in governance rules.
//
// The condition keys must resolve against the fields of PolicyContext, which
// are the top-level NEIR keys. Nested values are reached with a dotted path, and
// a path through a slice applies to every element.
func DefaultRules() []Rule {
	return []Rule{
		{RuleID: "project-required", Condition: "exists:project", Priority: 1, Action: "block", Scope: "spec", Enabled: true},
		{RuleID: "modules-required", Condition: "exists:modules", Priority: 1, Action: "block", Scope: "spec", Enabled: true},
		{RuleID: "architecture-pattern-valid", Condition: "in:architecture.pattern,hexagonal,layered,clean,event-driven,cqrs,microkernel,monolith,monolithic,microservices,serverless", Priority: 2, Action: "warn", Scope: "spec", Enabled: true},
		{RuleID: "deployment-strategy-valid", Condition: "in:deployment.strategy,rolling,blue-green,canary,recreate", Priority: 2, Action: "warn", Scope: "spec", Enabled: true},
		{RuleID: "service-port-positive", Condition: "gt:services.port,0", Priority: 3, Action: "warn", Scope: "spec", Enabled: true},
	}
}