package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Parser interface {
	Parse(input string) (*SpecDocument, error)
}

type ParserFunc func(input string) (*SpecDocument, error)

func (f ParserFunc) Parse(input string) (*SpecDocument, error) {
	return f(input)
}

type Module struct {
	Name         string
	Path         string
	Description  string
	Dependencies []string
}

type Service struct {
	Name        string
	Kind        string
	Port        int
	Description string
	Endpoints   []Endpoint
}

type Endpoint struct {
	Method string
	Path   string
	Action string
}

type Architecture struct {
	Pattern     string
	Description string
	Principles  []string
}

type Deployment struct {
	Strategy     string
	Environments []string
}

type Testing struct {
	Strategy string
	Coverage string
}

type Generation struct {
	Languages []string
	OutputDir string
	ModuleDir string
}

type SpecDocument struct {
	Raw          string
	Data         any
	Project      string
	Modules      []Module
	Services     []Service
	Architecture *Architecture
	Deployment   *Deployment
	Testing      *Testing
	Generation   *Generation
}

func NewParser(baseDir string) Parser {
	return ParserFunc(func(input string) (*SpecDocument, error) {
		if input == "" {
			return nil, fmt.Errorf("input cannot be empty")
		}

		resolver := NewImportResolver(baseDir)
		resolved, err := resolver.ResolveImports(input)
		if err != nil {
			return nil, fmt.Errorf("resolve imports: %w", err)
		}
		input = resolved

		var root yaml.Node
		if err := yaml.Unmarshal([]byte(input), &root); err != nil {
			return nil, fmt.Errorf("parse spec: %w", err)
		}

		if len(root.Content) == 0 {
			return nil, fmt.Errorf("empty specification document")
		}

		value, err := parseYAMLNode(root.Content[0])
		if err != nil {
			return nil, err
		}

		doc := &SpecDocument{Raw: input, Data: value}

		if m, ok := value.(map[string]any); ok {
			if project, ok := m["project"].(string); ok {
				doc.Project = project
			}
			if rawModules, ok := m["modules"].([]any); ok {
				for _, raw := range rawModules {
					if mod, ok := raw.(map[string]any); ok {
						doc.Modules = append(doc.Modules, extractModule(mod))
					}
				}
			}
			if rawServices, ok := m["services"].([]any); ok {
				for _, raw := range rawServices {
					if svc, ok := raw.(map[string]any); ok {
						doc.Services = append(doc.Services, extractService(svc))
					}
				}
			}
			if rawArch, ok := m["architecture"].(map[string]any); ok {
				doc.Architecture = extractArchitecture(rawArch)
			}
			if rawDeploy, ok := m["deployment"].(map[string]any); ok {
				doc.Deployment = extractDeployment(rawDeploy)
			}
			if rawTest, ok := m["testing"].(map[string]any); ok {
				doc.Testing = extractTesting(rawTest)
			}
			if rawGen, ok := m["generation"].(map[string]any); ok {
				doc.Generation = extractGeneration(rawGen)
			}

			if versionStr := ExtractVersionFromData(m); versionStr != "" {
				result := CheckSpecVersion(versionStr)
				if !result.Valid {
					return nil, fmt.Errorf("spec version check: %s", result.Message)
				}
			}
		}

		return doc, nil
	})
}

func extractModule(m map[string]any) Module {
	mod := Module{}
	if name, ok := m["name"].(string); ok {
		mod.Name = name
	}
	if path, ok := m["path"].(string); ok {
		mod.Path = path
	}
	if desc, ok := m["description"].(string); ok {
		mod.Description = desc
	}
	if deps, ok := m["dependencies"].([]any); ok {
		for _, d := range deps {
			if s, ok := d.(string); ok {
				mod.Dependencies = append(mod.Dependencies, s)
			}
		}
	}
	return mod
}

func extractService(s map[string]any) Service {
	svc := Service{}
	if name, ok := s["name"].(string); ok {
		svc.Name = name
	}
	if kind, ok := s["kind"].(string); ok {
		svc.Kind = kind
	}
	if port, ok := s["port"].(int); ok {
		svc.Port = port
	}
	if desc, ok := s["description"].(string); ok {
		svc.Description = desc
	}
	if rawEndpoints, ok := s["endpoints"].([]any); ok {
		for _, raw := range rawEndpoints {
			if ep, ok := raw.(map[string]any); ok {
				svc.Endpoints = append(svc.Endpoints, extractEndpoint(ep))
			}
		}
	}
	return svc
}

func extractEndpoint(m map[string]any) Endpoint {
	ep := Endpoint{}
	if method, ok := m["method"].(string); ok {
		ep.Method = method
	}
	if path, ok := m["path"].(string); ok {
		ep.Path = path
	}
	if action, ok := m["action"].(string); ok {
		ep.Action = action
	}
	return ep
}

func extractArchitecture(m map[string]any) *Architecture {
	arch := &Architecture{}
	if pattern, ok := m["pattern"].(string); ok {
		arch.Pattern = pattern
	}
	if desc, ok := m["description"].(string); ok {
		arch.Description = desc
	}
	if principles, ok := m["principles"].([]any); ok {
		for _, p := range principles {
			if s, ok := p.(string); ok {
				arch.Principles = append(arch.Principles, s)
			}
		}
	}
	return arch
}

func extractDeployment(m map[string]any) *Deployment {
	deploy := &Deployment{}
	if strategy, ok := m["strategy"].(string); ok {
		deploy.Strategy = strategy
	}
	if envs, ok := m["environments"].([]any); ok {
		for _, e := range envs {
			if s, ok := e.(string); ok {
				deploy.Environments = append(deploy.Environments, s)
			}
		}
	}
	return deploy
}

func extractGeneration(m map[string]any) *Generation {
	gen := &Generation{}
	if langs, ok := m["languages"].([]any); ok {
		for _, l := range langs {
			if s, ok := l.(string); ok {
				gen.Languages = append(gen.Languages, s)
			}
		}
	}
	if outputDir, ok := m["output_dir"].(string); ok {
		gen.OutputDir = outputDir
	}
	if moduleDir, ok := m["module_dir"].(string); ok {
		gen.ModuleDir = moduleDir
	}
	return gen
}

func extractTesting(m map[string]any) *Testing {
	test := &Testing{}
	if strategy, ok := m["strategy"].(string); ok {
		test.Strategy = strategy
	}
	if coverage, ok := m["coverage"].(string); ok {
		test.Coverage = coverage
	}
	return test
}

func parseYAMLNode(node *yaml.Node) (any, error) {
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, fmt.Errorf("empty document")
		}
		return parseYAMLNode(node.Content[0])
	case yaml.MappingNode:
		result := map[string]any{}
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			if keyNode.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("map keys must be scalar")
			}
			value, err := parseYAMLNode(valueNode)
			if err != nil {
				return nil, err
			}
			result[keyNode.Value] = value
		}
		return result, nil
	case yaml.SequenceNode:
		result := make([]any, len(node.Content))
		for i, child := range node.Content {
			value, err := parseYAMLNode(child)
			if err != nil {
				return nil, err
			}
			result[i] = value
		}
		return result, nil
	case yaml.ScalarNode:
		return parseYAMLScalar(node)
	case yaml.AliasNode:
		if node.Alias == nil {
			return nil, fmt.Errorf("invalid alias node")
		}
		return parseYAMLNode(node.Alias)
	default:
		return nil, fmt.Errorf("unsupported YAML node kind %d", node.Kind)
	}
}

func parseYAMLScalar(node *yaml.Node) (any, error) {
	if node.Tag == "!!null" {
		return nil, nil // YAML null literal — represent as Go nil
	}

	switch node.Tag {
	case "!!bool":
		return strconv.ParseBool(node.Value)
	case "!!int":
		return strconv.ParseInt(node.Value, 10, 64)
	case "!!float":
		return strconv.ParseFloat(node.Value, 64)
	case "!!str":
		return node.Value, nil
	default:
		if node.Value == "true" || node.Value == "false" {
			return strconv.ParseBool(node.Value)
		}
		if node.Value == "null" || node.Value == "~" {
			return nil, nil // Explicit null/tilde value — represent as Go nil
		}
		if i, err := strconv.ParseInt(node.Value, 10, 64); err == nil {
			return i, nil
		}
		if f, err := strconv.ParseFloat(node.Value, 64); err == nil {
			return f, nil
		}
		return node.Value, nil
	}
}

func applyDefaults(doc *SpecDocument, input string) {
	if doc.Project == "" {
		doc.Project = defaultProjectName(input)
	}
	if len(doc.Modules) == 0 {
		moduleName := defaultModuleName(doc.Project)
		doc.Modules = []Module{{Name: moduleName, Path: fmt.Sprintf("./%s", slugify(moduleName))}}
	}
}

func defaultProjectName(input string) string {
	value := strings.TrimSpace(input)
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.TrimSpace(value)
	if value == "" {
		return "default-project"
	}
	candidate := slugify(value)
	if candidate == "" {
		return "default-project"
	}
	return candidate
}

func defaultModuleName(project string) string {
	value := strings.TrimSpace(project)
	if value == "" {
		return "default-module"
	}
	return slugify(value)
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "default"
	}
	return value
}
func DefaultProjectNameForInput(input string) string {
	return defaultProjectName(input)
}
func DefaultModuleNameForProject(project string) string {
	return defaultModuleName(project)
}
func Slugify(value string) string {
	return slugify(value)
}
func parsePort(line string) (int, error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid port line")
	}
	var port int
	_, err := fmt.Sscanf(parts[1], "%d", &port)
	if err != nil {
		return 0, err
	}
	return port, nil
}

func MergeSpecs(docs ...*SpecDocument) *SpecDocument {
	if len(docs) == 0 {
		return nil
	}
	if len(docs) == 1 {
		return docs[0]
	}

	merged := &SpecDocument{
		Raw:  docs[0].Raw,
		Data: docs[0].Data,
	}

	if docs[0].Project != "" {
		merged.Project = docs[0].Project
	}

	seenModules := make(map[string]bool)
	for _, doc := range docs {
		if doc.Project != "" && merged.Project == "" {
			merged.Project = doc.Project
		}
		for _, m := range doc.Modules {
			if !seenModules[m.Name] {
				seenModules[m.Name] = true
				merged.Modules = append(merged.Modules, m)
			}
		}
		merged.Services = append(merged.Services, doc.Services...)
		if doc.Architecture != nil && merged.Architecture == nil {
			merged.Architecture = doc.Architecture
		}
		if doc.Deployment != nil && merged.Deployment == nil {
			merged.Deployment = doc.Deployment
		}
		if doc.Testing != nil && merged.Testing == nil {
			merged.Testing = doc.Testing
		}
		if doc.Generation != nil && merged.Generation == nil {
			merged.Generation = doc.Generation
		}
	}

	return merged
}
