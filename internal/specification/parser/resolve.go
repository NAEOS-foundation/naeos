package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var varPattern = regexp.MustCompile(`\$\{([^}]+)\}`)
var envPattern = regexp.MustCompile(`\$env\{([^}]+)\}`)
var refPattern = regexp.MustCompile(`\$ref\{([^}]+)\}`)

type VariableResolver struct {
	vars  map[string]string
	refs  map[string]any
	envs  map[string]string
}

func NewVariableResolver() *VariableResolver {
	return &VariableResolver{
		vars: make(map[string]string),
		refs: make(map[string]any),
		envs: make(map[string]string),
	}
}

func (r *VariableResolver) SetVar(key, value string) {
	r.vars[key] = value
}

func (r *VariableResolver) SetRef(key string, value any) {
	r.refs[key] = value
}

func (r *VariableResolver) SetVars(vars map[string]string) {
	for k, v := range vars {
		r.vars[k] = v
	}
}

func (r *VariableResolver) Resolve(input string) (string, error) {
	result := input

	// Resolve ${var} — custom variables
	result = varPattern.ReplaceAllStringFunc(result, func(match string) string {
		key := varPattern.FindStringSubmatch(match)[1]
		if val, ok := r.vars[key]; ok {
			return val
		}
		return match
	})

	// Resolve $env{VAR} — environment variables
	result = envPattern.ReplaceAllStringFunc(result, func(match string) string {
		key := envPattern.FindStringSubmatch(match)[1]
		if val, ok := r.envs[key]; ok {
			return val
		}
		if val := os.Getenv(key); val != "" {
			return val
		}
		return match
	})

	// Resolve $ref{path} — references to other parts of the spec
	result = refPattern.ReplaceAllStringFunc(result, func(match string) string {
		key := refPattern.FindStringSubmatch(match)[1]
		if val, ok := r.refs[key]; ok {
			return fmt.Sprintf("%v", val)
		}
		return match
	})

	return result, nil
}

func (r *VariableResolver) ResolveMap(m map[string]any) (map[string]any, error) {
	result := make(map[string]any, len(m))
	for k, v := range m {
		resolved, err := r.resolveValue(v)
		if err != nil {
			return nil, fmt.Errorf("resolve key %q: %w", k, err)
		}
		result[k] = resolved
	}
	return result, nil
}

func (r *VariableResolver) resolveValue(v any) (any, error) {
	switch val := v.(type) {
	case string:
		return r.Resolve(val)
	case map[string]any:
		return r.ResolveMap(val)
	case []any:
		result := make([]any, len(val))
		for i, item := range val {
			resolved, err := r.resolveValue(item)
			if err != nil {
				return nil, err
			}
			result[i] = resolved
		}
		return result, nil
	default:
		return v, nil
	}
}

type ValidationIssue struct {
	Severity string
	Rule     string
	Message  string
	Line     int
}

type ValidationResult struct {
	Valid      bool
	Issues     []ValidationIssue
	Warnings   []ValidationIssue
	ModuleCount int
	ServiceCount int
}

type SpecValidator struct {
	resolver *VariableResolver
}

func NewSpecValidator() *SpecValidator {
	return &SpecValidator{
		resolver: NewVariableResolver(),
	}
}

func (v *SpecValidator) Validate(data any) *ValidationResult {
	result := &ValidationResult{Valid: true}
	v.validateNode(data, "", result)
	result.Valid = len(result.Issues) == 0
	return result
}

func (v *SpecValidator) validateNode(data any, path string, result *ValidationResult) {
	switch val := data.(type) {
	case map[string]any:
		v.validateMap(val, path, result)
	case []any:
		for i, item := range val {
			v.validateNode(item, fmt.Sprintf("%s[%d]", path, i), result)
		}
	}
}

func (v *SpecValidator) validateMap(m map[string]any, path string, result *ValidationResult) {
	// Check for circular $ref
	for key, val := range m {
		if str, ok := val.(string); ok {
			if matches := refPattern.FindStringSubmatch(str); len(matches) > 1 {
				ref := matches[1]
				if _, exists := v.resolver.refs[ref]; !exists {
					result.Issues = append(result.Issues, ValidationIssue{
						Severity: "error",
						Rule:     "ref-not-found",
						Message:  fmt.Sprintf("reference $ref{%s} not found", ref),
					})
				}
			}
		}
		v.validateNode(val, path+"."+key, result)
	}
}

func (v *SpecValidator) ValidateModules(modules []Module) []ValidationIssue {
	var issues []ValidationIssue

	seen := make(map[string]int)
	for i, m := range modules {
		if m.Name == "" {
			issues = append(issues, ValidationIssue{
				Severity: "error",
				Rule:     "module-name-required",
				Message:  fmt.Sprintf("module[%d] missing name", i),
			})
			continue
		}

		if prev, dup := seen[m.Name]; dup {
			issues = append(issues, ValidationIssue{
				Severity: "error",
				Rule:     "module-duplicate",
				Message:  fmt.Sprintf("duplicate module %q at indices %d and %d", m.Name, prev, i),
			})
		}
		seen[m.Name] = i
	}

	// Check for circular dependencies
	depGraph := make(map[string][]string)
	for _, m := range modules {
		depGraph[m.Name] = m.Dependencies
	}
	if cycles := detectCycles(depGraph); len(cycles) > 0 {
		for _, cycle := range cycles {
			issues = append(issues, ValidationIssue{
				Severity: "error",
				Rule:     "circular-dependency",
				Message:  fmt.Sprintf("circular dependency detected: %s", strings.Join(cycle, " -> ")),
			})
		}
	}

	// Check dangling dependencies
	moduleNames := make(map[string]bool)
	for _, m := range modules {
		moduleNames[m.Name] = true
	}
	for _, m := range modules {
		for _, dep := range m.Dependencies {
			if !moduleNames[dep] {
				issues = append(issues, ValidationIssue{
					Severity: "error",
					Rule:     "dependency-not-found",
					Message:  fmt.Sprintf("module %q depends on %q which does not exist", m.Name, dep),
				})
			}
		}
	}

	return issues
}

func (v *SpecValidator) ValidateServices(services []Service) []ValidationIssue {
	var issues []ValidationIssue

	seenPorts := make(map[int]string)
	for i, s := range services {
		if s.Name == "" {
			issues = append(issues, ValidationIssue{
				Severity: "error",
				Rule:     "service-name-required",
				Message:  fmt.Sprintf("service[%d] missing name", i),
			})
		}
		if s.Port < 0 || s.Port > 65535 {
			issues = append(issues, ValidationIssue{
				Severity: "error",
				Rule:     "service-port-range",
				Message:  fmt.Sprintf("service %q port %d out of range (0-65535)", s.Name, s.Port),
			})
		}
		if s.Port > 0 {
			if prev, exists := seenPorts[s.Port]; exists {
				issues = append(issues, ValidationIssue{
					Severity: "warning",
					Rule:     "service-port-conflict",
					Message:  fmt.Sprintf("services %q and %q share port %d", prev, s.Name, s.Port),
				})
			}
			seenPorts[s.Port] = s.Name
		}
	}

	return issues
}

func detectCycles(graph map[string][]string) [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, dep := range graph[node] {
			if !visited[dep] {
				dfs(dep)
			} else if recStack[dep] {
				// Found cycle
				cycleStart := -1
				for i, n := range path {
					if n == dep {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]string, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycle = append(cycle, dep)
					cycles = append(cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
	}

	for node := range graph {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}
