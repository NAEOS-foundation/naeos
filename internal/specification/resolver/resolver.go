package resolver

import (
	"fmt"
	"regexp"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/specification/normalizer"
)

type Resolver interface {
	Resolve(spec any) (*ResolvedSpec, error)
}

type ResolvedSpec struct {
	Context map[string]any
}

type DefaultResolver struct{}

func NewResolver() Resolver {
	return DefaultResolver{}
}

func (DefaultResolver) Resolve(spec any) (*ResolvedSpec, error) {
	if spec == nil {
		return nil, naeoserr.New(naeoserr.ErrValidation, "spec is nil")
	}

	normalized, ok := spec.(*normalizer.NormalizedSpec)
	if !ok {
		return &ResolvedSpec{Context: map[string]any{"resolved": true}}, nil
	}

	context := map[string]any{}
	for key, value := range normalized.Values {
		context[key] = value
	}

	resolveModuleDependencies(context)
	resolveServiceEndpoints(context)
	populateDefaults(context)

	return &ResolvedSpec{Context: context}, nil
}

func resolveModuleDependencies(context map[string]any) {
	rawModules, exists := context["modules"]
	if !exists {
		return
	}

	modules, ok := rawModules.([]map[string]any)
	if !ok {
		return
	}

	moduleNames := make(map[string]bool, len(modules))
	for _, m := range modules {
		if name, ok := m["name"].(string); ok {
			moduleNames[name] = true
		}
	}

	for _, m := range modules {
		if deps, ok := m["dependencies"].([]any); ok {
			validDeps := make([]any, 0, len(deps))
			for _, d := range deps {
				if depName, ok := d.(string); ok {
					if moduleNames[depName] {
						validDeps = append(validDeps, d)
					}
				}
			}
			m["dependencies"] = validDeps
		}
	}
}

func resolveServiceEndpoints(context map[string]any) {
	rawServices, exists := context["services"]
	if !exists {
		return
	}

	services, ok := rawServices.([]map[string]any)
	if !ok {
		return
	}

	for _, svc := range services {
		rawEndpoints, ok := svc["endpoints"]
		if !ok {
			continue
		}
		endpoints, ok := rawEndpoints.([]map[string]any)
		if !ok {
			continue
		}
		for _, ep := range endpoints {
			if method, ok := ep["method"].(string); ok {
				ep["method"] = method
			}
			if path, ok := ep["path"].(string); ok && path != "" {
				if path[0] != '/' {
					ep["path"] = "/" + path
				}
			}
		}
	}
}

func populateDefaults(context map[string]any) {
	if _, exists := context["architecture"]; !exists {
		context["architecture"] = map[string]any{
			"pattern":     "layered",
			"description": "default layered architecture",
		}
	}

	rawModules, exists := context["modules"]
	if exists {
		if modules, ok := rawModules.([]map[string]any); ok {
			for _, m := range modules {
				if _, hasPath := m["path"]; !hasPath {
					if name, ok := m["name"].(string); ok {
						m["path"] = fmt.Sprintf("./internal/%s", name)
					}
				}
				if _, hasDeps := m["dependencies"]; !hasDeps {
					m["dependencies"] = []any{}
				}
			}
		}
	}
}

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type ValidationError struct {
	Field    string
	Message  string
	Severity Severity
}

type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

func ValidateSpec(spec *ResolvedSpec) ValidationResult {
	var errs []ValidationError

	if spec == nil || spec.Context == nil {
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Field: "spec", Message: "resolved spec is nil", Severity: SeverityError}},
		}
	}

	errs = append(errs, validateModuleCycles(spec.Context)...)
	errs = append(errs, validateModuleDuplicates(spec.Context)...)
	errs = append(errs, validatePortRanges(spec.Context)...)
	errs = append(errs, validateEndpointPaths(spec.Context)...)
	errs = append(errs, validateArchitecturePattern(spec.Context)...)

	valid := true
	for _, e := range errs {
		if e.Severity == SeverityError {
			valid = false
			break
		}
	}

	return ValidationResult{Valid: valid, Errors: errs}
}

func validateModuleCycles(context map[string]any) []ValidationError {
	var errs []ValidationError

	rawModules, exists := context["modules"]
	if !exists {
		return nil
	}

	modules, ok := rawModules.([]map[string]any)
	if !ok {
		return nil
	}

	nameToIdx := make(map[string]int, len(modules))
	for i, m := range modules {
		if name, ok := m["name"].(string); ok {
			nameToIdx[name] = i
		}
	}

	adj := make(map[string][]string, len(modules))
	for _, m := range modules {
		name, _ := m["name"].(string)
		deps, _ := m["dependencies"].([]any)
		for _, d := range deps {
			if dep, ok := d.(string); ok {
				adj[name] = append(adj[name], dep)
			}
		}
	}

	const (
		white = 0
		gray  = 1
		black = 2
	)
	colors := make(map[string]int, len(modules))
	for name := range nameToIdx {
		colors[name] = white
	}

	var cycle []string
	var dfs func(string)
	dfs = func(node string) {
		if cycle != nil {
			return
		}
		colors[node] = gray
		for _, nb := range adj[node] {
			if colors[nb] == gray {
				cycle = []string{nb, node}
				return
			}
			if colors[nb] == white {
				dfs(nb)
			}
		}
		colors[node] = black
	}

	for name := range nameToIdx {
		if colors[name] == white {
			dfs(name)
			if cycle != nil {
				break
			}
		}
	}

	if cycle != nil {
		errs = append(errs, ValidationError{
			Field:    "modules.dependencies",
			Message:  fmt.Sprintf("cycle detected involving modules %q", cycle),
			Severity: SeverityError,
		})
	}

	return errs
}

func validateModuleDuplicates(context map[string]any) []ValidationError {
	var errs []ValidationError

	rawModules, exists := context["modules"]
	if !exists {
		return nil
	}

	modules, ok := rawModules.([]map[string]any)
	if !ok {
		return nil
	}

	seen := make(map[string]bool, len(modules))
	for _, m := range modules {
		if name, ok := m["name"].(string); ok {
			if seen[name] {
				errs = append(errs, ValidationError{
					Field:    "modules",
					Message:  fmt.Sprintf("duplicate module name %q", name),
					Severity: SeverityError,
				})
			}
			seen[name] = true
		}
	}

	return errs
}

func validatePortRanges(context map[string]any) []ValidationError {
	var errs []ValidationError

	rawServices, exists := context["services"]
	if !exists {
		return nil
	}

	services, ok := rawServices.([]map[string]any)
	if !ok {
		return nil
	}

	for _, svc := range services {
		name, _ := svc["name"].(string)
		portRaw, exists := svc["port"]
		if !exists {
			continue
		}
		port, ok := toInt(portRaw)
		if !ok {
			errs = append(errs, ValidationError{
				Field:    fmt.Sprintf("services.%s.port", name),
				Message:  "port is not a valid integer",
				Severity: SeverityError,
			})
			continue
		}
		if port < 1 || port > 65535 {
			errs = append(errs, ValidationError{
				Field:    fmt.Sprintf("services.%s.port", name),
				Message:  fmt.Sprintf("port %d is out of range 1-65535", port),
				Severity: SeverityError,
			})
		}
	}

	return errs
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

func validateEndpointPaths(context map[string]any) []ValidationError {
	var errs []ValidationError

	rawServices, exists := context["services"]
	if !exists {
		return nil
	}

	services, ok := rawServices.([]map[string]any)
	if !ok {
		return nil
	}

	for _, svc := range services {
		svcName, _ := svc["name"].(string)
		rawEndpoints, exists := svc["endpoints"]
		if !exists {
			continue
		}
		endpoints, ok := rawEndpoints.([]map[string]any)
		if !ok {
			continue
		}
		for _, ep := range endpoints {
			path, ok := ep["path"].(string)
			if !ok || path == "" {
				continue
			}
			if path[0] != '/' {
				errs = append(errs, ValidationError{
					Field:    fmt.Sprintf("services.%s.endpoints.path", svcName),
					Message:  fmt.Sprintf("endpoint path %q must start with /", path),
					Severity: SeverityError,
				})
			}
		}
	}

	return errs
}

func validateArchitecturePattern(context map[string]any) []ValidationError {
	var errs []ValidationError

	rawArch, exists := context["architecture"]
	if !exists {
		return nil
	}
	arch, ok := rawArch.(map[string]any)
	if !ok {
		return nil
	}
	pattern, ok := arch["pattern"].(string)
	if !ok || strings.TrimSpace(pattern) == "" {
		errs = append(errs, ValidationError{
			Field:    "architecture.pattern",
			Message:  "architecture pattern must be non-empty",
			Severity: SeverityError,
		})
	}

	return errs
}

var envVarRe = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

func ResolveEnvironmentVariables(context map[string]any) map[string]any {
	env := make(map[string]any, len(context))
	for k, v := range context {
		env[k] = resolveEnvValue(v)
	}
	return env
}

func resolveEnvValue(v any) any {
	switch val := v.(type) {
	case string:
		return envVarRe.ReplaceAllStringFunc(val, func(match string) string {
			varName := match[2 : len(match)-1]
			if envVal, ok := lookupEnv(varName); ok {
				return envVal
			}
			return match
		})
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = resolveEnvValue(item)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(val))
		for k2, v2 := range val {
			out[k2] = resolveEnvValue(v2)
		}
		return out
	default:
		return v
	}
}

func lookupEnv(name string) (string, bool) {
	val, exists := envStore[name]
	return val, exists
}

var envStore = map[string]string{}

func SetEnvForTest(name, value string) {
	envStore[name] = value
}

func ClearEnvForTest() {
	envStore = map[string]string{}
}

var refRe = regexp.MustCompile(`\$\{ref:([A-Za-z0-9_.]+)\}`)

func ResolveReferences(context map[string]any) map[string]any {
	resolved := make(map[string]any, len(context))
	for k, v := range context {
		resolved[k] = resolveRefValue(v, context)
	}
	return resolved
}

func resolveRefValue(v any, root map[string]any) any {
	switch val := v.(type) {
	case string:
		return refRe.ReplaceAllStringFunc(val, func(match string) string {
			refPath := match[5 : len(match)-1]
			parts := strings.SplitN(refPath, ".", 2)
			if len(parts) != 2 {
				return match
			}
			target := parts[0]
			field := parts[1]
			targetVal, exists := root[target]
			if !exists {
				return match
			}
			if targetMap, ok := targetVal.(map[string]any); ok {
				if fieldVal, ok := targetMap[field]; ok {
					return fmt.Sprintf("%v", fieldVal)
				}
			}
			return match
		})
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = resolveRefValue(item, root)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(val))
		for k2, v2 := range val {
			out[k2] = resolveRefValue(v2, root)
		}
		return out
	default:
		return v
	}
}

func CrossValidateServices(context map[string]any) []ValidationError {
	var errs []ValidationError

	rawServices, exists := context["services"]
	if !exists {
		return nil
	}
	services, ok := rawServices.([]map[string]any)
	if !ok {
		return nil
	}

	moduleNames := make(map[string]bool)
	rawModules, exists := context["modules"]
	if exists {
		if modules, ok := rawModules.([]map[string]any); ok {
			for _, m := range modules {
				if name, ok := m["name"].(string); ok {
					moduleNames[name] = true
				}
			}
		}
	}

	type endpointKey struct {
		service string
		method  string
		path    string
	}
	endpointSeen := make(map[endpointKey]bool)

	for _, svc := range services {
		svcName, _ := svc["name"].(string)

		if modRef, ok := svc["module"].(string); ok {
			if !moduleNames[modRef] {
				errs = append(errs, ValidationError{
					Field:    fmt.Sprintf("services.%s.module", svcName),
					Message:  fmt.Sprintf("service %q references non-existent module %q", svcName, modRef),
					Severity: SeverityError,
				})
			}
		}

		rawEndpoints, exists := svc["endpoints"]
		if !exists {
			continue
		}
		endpoints, ok := rawEndpoints.([]map[string]any)
		if !ok {
			continue
		}
		for _, ep := range endpoints {
			method, _ := ep["method"].(string)
			path, _ := ep["path"].(string)
			ek := endpointKey{service: svcName, method: method, path: path}
			if endpointSeen[ek] {
				errs = append(errs, ValidationError{
					Field:    fmt.Sprintf("services.%s.endpoints", svcName),
					Message:  fmt.Sprintf("duplicate endpoint %s %s in service %q", method, path, svcName),
					Severity: SeverityWarning,
				})
			}
			endpointSeen[ek] = true
		}
	}

	return errs
}

type Conflict struct {
	Path    string
	Left    any
	Right   any
	Message string
}

func ConflictDetection(context map[string]any) []Conflict {
	var conflicts []Conflict

	rawModules, exists := context["modules"]
	if exists {
		if modules, ok := rawModules.([]map[string]any); ok {
			conflicts = append(conflicts, detectModulePathConflicts(modules)...)
		}
	}

	rawServices, exists := context["services"]
	if exists {
		if services, ok := rawServices.([]map[string]any); ok {
			conflicts = append(conflicts, detectPortConflicts(services)...)
		}
	}

	rawArch, exists := context["architecture"]
	if exists {
		if arch, ok := rawArch.(map[string]any); ok {
			conflicts = append(conflicts, detectArchitectureConflicts(arch)...)
		}
	}

	return conflicts
}

func detectModulePathConflicts(modules []map[string]any) []Conflict {
	var conflicts []Conflict
	pathToModules := make(map[string][]string)

	for _, m := range modules {
		name, _ := m["name"].(string)
		path, _ := m["path"].(string)
		if path == "" {
			continue
		}
		pathToModules[path] = append(pathToModules[path], name)
	}

	for path, names := range pathToModules {
		if len(names) > 1 {
			conflicts = append(conflicts, Conflict{
				Path:    path,
				Left:    names[0],
				Right:   names[1],
				Message: fmt.Sprintf("modules %q and %q share the same path %q", names[0], names[1], path),
			})
		}
	}

	return conflicts
}

func detectPortConflicts(services []map[string]any) []Conflict {
	var conflicts []Conflict
	portToServices := make(map[int][]string)

	for _, svc := range services {
		name, _ := svc["name"].(string)
		portRaw, exists := svc["port"]
		if !exists {
			continue
		}
		port, ok := toInt(portRaw)
		if !ok {
			continue
		}
		portToServices[port] = append(portToServices[port], name)
	}

	for port, names := range portToServices {
		if len(names) > 1 {
			conflicts = append(conflicts, Conflict{
				Path:    fmt.Sprintf("port:%d", port),
				Left:    names[0],
				Right:   names[1],
				Message: fmt.Sprintf("services %q and %q use the same port %d", names[0], names[1], port),
			})
		}
	}

	return conflicts
}

func detectArchitectureConflicts(arch map[string]any) []Conflict {
	var conflicts []Conflict

	pattern, _ := arch["pattern"].(string)
	principles, _ := arch["principles"].([]any)
	if pattern != "" && len(principles) > 0 {
		patternLower := strings.ToLower(pattern)
		for _, p := range principles {
			if pStr, ok := p.(string); ok {
				pLower := strings.ToLower(pStr)
				if patternLower == "microservices" && pLower == "monolith" {
					conflicts = append(conflicts, Conflict{
						Path:    "architecture",
						Left:    pattern,
						Right:   pStr,
						Message: fmt.Sprintf("architecture pattern %q conflicts with principle %q", pattern, pStr),
					})
				}
				if patternLower == "monolith" && pLower == "microservices" {
					conflicts = append(conflicts, Conflict{
						Path:    "architecture",
						Left:    pattern,
						Right:   pStr,
						Message: fmt.Sprintf("architecture pattern %q conflicts with principle %q", pattern, pStr),
					})
				}
			}
		}
	}

	return conflicts
}

type ResolutionStep struct {
	Function    string
	Description string
}

type ResolutionContext struct {
	Steps    []ResolutionStep
	Warnings []string
}

func (rc *ResolutionContext) AddStep(fn, desc string) {
	rc.Steps = append(rc.Steps, ResolutionStep{Function: fn, Description: desc})
}

func (rc *ResolutionContext) AddWarning(msg string) {
	rc.Warnings = append(rc.Warnings, msg)
}

func ResolveWithTrace(spec any) (*ResolvedSpec, *ResolutionContext, error) {
	ctx := &ResolutionContext{}

	if spec == nil {
		return nil, ctx, naeoserr.New(naeoserr.ErrValidation, "spec is nil")
	}

	normalized, ok := spec.(*normalizer.NormalizedSpec)
	if !ok {
		ctx.AddStep("ResolveWithTrace", "non-normalized spec passed through")
		return &ResolvedSpec{Context: map[string]any{"resolved": true}}, ctx, nil
	}

	resolved := map[string]any{}
	for key, value := range normalized.Values {
		resolved[key] = value
	}

	ctx.AddStep("ResolveWithTrace", "copied normalized values into context")

	resolveModuleDependencies(resolved)
	ctx.AddStep("resolveModuleDependencies", "resolved module dependency references")

	resolveServiceEndpoints(resolved)
	ctx.AddStep("resolveServiceEndpoints", "normalized service endpoint paths and methods")

	populateDefaults(resolved)
	ctx.AddStep("populateDefaults", "filled in default architecture and module paths")

	if rawModules, exists := resolved["modules"]; exists {
		if modules, ok := rawModules.([]map[string]any); ok {
			for _, m := range modules {
				if name, ok := m["name"].(string); ok {
					if deps, ok := m["dependencies"].([]any); ok && len(deps) == 0 {
						ctx.AddWarning(fmt.Sprintf("module %q has no dependencies", name))
					}
				}
			}
		}
	}

	if rawServices, exists := resolved["services"]; exists {
		if services, ok := rawServices.([]map[string]any); ok {
			for _, svc := range services {
				name, _ := svc["name"].(string)
				portRaw, exists := svc["port"]
				if !exists {
					ctx.AddWarning(fmt.Sprintf("service %q has no port defined", name))
				} else if port, ok := toInt(portRaw); ok {
					if port < 1024 {
						ctx.AddWarning(fmt.Sprintf("service %q uses privileged port %d", name, port))
					}
				}
			}
		}
	}

	return &ResolvedSpec{Context: resolved}, ctx, nil
}
