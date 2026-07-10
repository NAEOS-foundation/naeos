package parser

import (
	"fmt"
	"strconv"
	"gopkg.in/yaml.v3"
	"regexp"
	"strings"
)

type Parser interface {
	Parse(input string) (*SpecDocument, error)
}

type ParserFunc func(input string) (*SpecDocument, error)

func (f ParserFunc) Parse(input string) (*SpecDocument, error) {
	return f(input)
}

type Module struct {
	Name string
	Path string
}

type Service struct {
	Name string
	Kind string
	Port int
}

type SpecDocument struct {
  main
	Raw  string
	Data any

	Raw     string
	Project string
	Modules []Module
	Services []Service
}

func NewParser() Parser {
	return ParserFunc(func(input string) (*SpecDocument, error) {
		if input == "" {
			return nil, fmt.Errorf("input cannot be empty")
		}

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

		return &SpecDocument{Raw: input, Data: value}, nil
	})
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
		return nil, nil
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
			return nil, nil
		}
		if i, err := strconv.ParseInt(node.Value, 10, 64); err == nil {
			return i, nil
		}
		if f, err := strconv.ParseFloat(node.Value, 64); err == nil {
			return f, nil
		}
		return node.Value, nil
	}

		doc := &SpecDocument{Raw: input}
		var currentSection string
		for _, line := range strings.Split(input, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}

			switch {
			case strings.HasPrefix(trimmed, "project:"):
				doc.Project = strings.TrimSpace(strings.TrimPrefix(trimmed, "project:"))
				currentSection = ""
			case strings.HasPrefix(trimmed, "modules:"):
				currentSection = "modules"
			case strings.HasPrefix(trimmed, "services:"):
				currentSection = "services"
			case strings.HasPrefix(trimmed, "- name:"):
				name := strings.TrimSpace(strings.TrimPrefix(trimmed, "- name:"))
				switch currentSection {
				case "modules":
					doc.Modules = append(doc.Modules, Module{Name: name})
				case "services":
					doc.Services = append(doc.Services, Service{Name: name})
				}
			case strings.HasPrefix(trimmed, "path:"):
				path := strings.TrimSpace(strings.TrimPrefix(trimmed, "path:"))
				if len(doc.Modules) > 0 {
					doc.Modules[len(doc.Modules)-1].Path = path
				}
			case strings.HasPrefix(trimmed, "kind:"):
				kind := strings.TrimSpace(strings.TrimPrefix(trimmed, "kind:"))
				if len(doc.Services) > 0 {
					doc.Services[len(doc.Services)-1].Kind = kind
				}
			case strings.HasPrefix(trimmed, "port:"):
				port, err := parsePort(trimmed)
				if err == nil && len(doc.Services) > 0 {
					doc.Services[len(doc.Services)-1].Port = port
				}
			}
		}

		applyDefaults(doc, input)
		return doc, nil
	})
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
