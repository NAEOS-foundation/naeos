package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/NAEOS-foundation/naeos/internal/securityext"
)

var includePattern = regexp.MustCompile(`\$include\{([^}]+)\}`)
var importPattern = regexp.MustCompile(`\$import\{([^}]+(?:::[\w.]+)?)\}`)
var fnPattern = regexp.MustCompile(`\$fn\{([a-zA-Z_][a-zA-Z0-9_]*)\(([^)]*)\)\}`)
var ifPattern = regexp.MustCompile(`^\$if\{([^}]+)\}\s*$`)

type IncludeResolver struct {
	baseDir  string
	loaded   map[string]string
	maxDepth int
}

func NewIncludeResolver(baseDir string) *IncludeResolver {
	return &IncludeResolver{
		baseDir:  baseDir,
		loaded:   make(map[string]string),
		maxDepth: 10,
	}
}

func (r *IncludeResolver) ResolveIncludes(input string) (string, error) {
	return r.resolveWithDepth(input, 0)
}

func (r *IncludeResolver) resolveWithDepth(input string, depth int) (string, error) {
	if depth > r.maxDepth {
		return "", fmt.Errorf("include depth exceeded maximum (%d)", r.maxDepth)
	}

	result := input
	for {
		matches := includePattern.FindStringSubmatch(result)
		if matches == nil {
			break
		}

		filePath := strings.TrimSpace(matches[1])
		if r.baseDir != "" {
			filePath = filepath.Join(r.baseDir, filePath)
		}
		filePath = filepath.Clean(filePath)
		if r.baseDir != "" {
			var err error
			filePath, err = securityext.ValidateFilePath(filePath, r.baseDir)
			if err != nil {
				return "", fmt.Errorf("path traversal in include: %w", err)
			}
		} else if strings.Contains(filePath, "..") {
			return "", fmt.Errorf("path traversal detected in include: %s", matches[1])
		}

		if cached, ok := r.loaded[filePath]; ok {
			result = strings.Replace(result, matches[0], cached, 1)
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("include %s: %w", matches[1], err)
		}

		r.loaded[filePath] = string(content)

		resolved, err := r.resolveWithDepth(string(content), depth+1)
		if err != nil {
			return "", err
		}

		result = strings.Replace(result, matches[0], resolved, 1)
	}

	return result, nil
}

type ImportResolver struct {
	baseDir  string
	loaded   map[string]string
	maxDepth int
}

func NewImportResolver(baseDir string) *ImportResolver {
	return &ImportResolver{
		baseDir:  baseDir,
		loaded:   make(map[string]string),
		maxDepth: 10,
	}
}

func (r *ImportResolver) ResolveImports(input string) (string, error) {
	return r.resolveWithDepth(input, 0)
}

func (r *ImportResolver) resolveWithDepth(input string, depth int) (string, error) {
	if depth > r.maxDepth {
		return "", fmt.Errorf("import depth exceeded maximum (%d)", r.maxDepth)
	}

	result := input
	for {
		matches := importPattern.FindStringSubmatch(result)
		if matches == nil {
			break
		}

		raw := strings.TrimSpace(matches[1])
		var filePath, section string
		if idx := strings.Index(raw, "::"); idx >= 0 {
			filePath = strings.TrimSpace(raw[:idx])
			section = strings.TrimSpace(raw[idx+2:])
		} else {
			filePath = raw
		}

		if r.baseDir != "" && !filepath.IsAbs(filePath) {
			filePath = filepath.Join(r.baseDir, filePath)
		}
		filePath = filepath.Clean(filePath)
		if r.baseDir != "" {
			var err error
			filePath, err = securityext.ValidateFilePath(filePath, r.baseDir)
			if err != nil {
				return "", fmt.Errorf("path traversal in import: %w", err)
			}
		} else if strings.Contains(filePath, "..") {
			return "", fmt.Errorf("path traversal detected in import: %s", matches[1])
		}

		cacheKey := filePath + "::" + section
		if cached, ok := r.loaded[cacheKey]; ok {
			result = strings.Replace(result, matches[0], cached, 1)
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("import %s: %w", matches[1], err)
		}

		var resolved string
		if section != "" {
			var root yaml.Node
			if err := yaml.Unmarshal(content, &root); err != nil {
				return "", fmt.Errorf("import %s: parse error: %w", matches[1], err)
			}
			data, err := parseYAMLNode(root.Content[0])
			if err != nil {
				return "", fmt.Errorf("import %s: %w", matches[1], err)
			}
			if m, ok := data.(map[string]any); ok {
				parts := strings.Split(section, ".")
				current := any(m)
				for _, part := range parts {
					if cm, ok := current.(map[string]any); ok {
						current = cm[part]
					} else {
						current = nil
						break
					}
				}
				if current != nil {
					out, err := yaml.Marshal(current)
					if err != nil {
						return "", fmt.Errorf("import %s: marshal error: %w", matches[1], err)
					}
					resolved = strings.TrimSpace(string(out))
				} else {
					return "", fmt.Errorf("import %s: section %q not found", matches[1], section)
				}
			} else {
				resolved = strings.TrimSpace(string(content))
			}
		} else {
			resolved = strings.TrimSpace(string(content))
		}

		recursed, err := r.resolveWithDepth(resolved, depth+1)
		if err != nil {
			return "", err
		}

		r.loaded[cacheKey] = recursed
		result = strings.Replace(result, matches[0], recursed, 1)
	}

	return result, nil
}

type FuncRegistry struct {
	funcs map[string]func(args string) string
}

func NewFuncRegistry() *FuncRegistry {
	r := &FuncRegistry{
		funcs: make(map[string]func(args string) string),
	}
	r.registerBuiltin()
	return r
}

func (r *FuncRegistry) Register(name string, fn func(args string) string) {
	r.funcs[name] = fn
}

func (r *FuncRegistry) Resolve(input string) string {
	result := input
	for {
		matches := fnPattern.FindStringSubmatch(result)
		if matches == nil {
			break
		}

		name := matches[1]
		args := matches[2]

		if fn, ok := r.funcs[name]; ok {
			replacement := fn(args)
			result = strings.Replace(result, matches[0], replacement, 1)
		} else {
			break
		}
	}
	return result
}

func (r *FuncRegistry) registerBuiltin() {
	r.Register("upper", func(args string) string {
		return strings.ToUpper(args)
	})
	r.Register("lower", func(args string) string {
		return strings.ToLower(args)
	})
	r.Register("slug", func(args string) string {
		slug := strings.ToLower(strings.TrimSpace(args))
		slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
		return strings.Trim(slug, "-")
	})
	r.Register("default", func(args string) string {
		parts := strings.SplitN(args, ",", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
			return strings.TrimSpace(parts[0])
		}
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
		return args
	})
	r.Register("len", func(args string) string {
		return fmt.Sprintf("%d", len(strings.TrimSpace(args)))
	})
	r.Register("coalesce", func(args string) string {
		parts := strings.Split(args, ",")
		for _, p := range parts {
			v := strings.TrimSpace(p)
			if v != "" {
				return v
			}
		}
		return ""
	})
}

type ConditionalResolver struct {
	env map[string]string
}

func NewConditionalResolver() *ConditionalResolver {
	return &ConditionalResolver{
		env: make(map[string]string),
	}
}

func (r *ConditionalResolver) SetEnv(key, value string) {
	r.env[key] = value
}

func (r *ConditionalResolver) SetEnvs(envs map[string]string) {
	for k, v := range envs {
		r.env[k] = v
	}
}

func (r *ConditionalResolver) Resolve(input string) string {
	lines := strings.Split(input, "\n")
	var result []string

	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if match := ifPattern.FindStringSubmatch(trimmed); match != nil {
			condition := strings.TrimSpace(match[1])
			if r.evaluateCondition(condition) {
				i++
				for i < len(lines) && strings.TrimSpace(lines[i]) != "$endif" {
					result = append(result, lines[i])
					i++
				}
			} else {
				i++
				for i < len(lines) && strings.TrimSpace(lines[i]) != "$endif" {
					i++
				}
			}
			if i < len(lines) {
				i++
			}
			continue
		}

		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n")
}

func (r *ConditionalResolver) evaluateCondition(cond string) bool {
	cond = strings.TrimSpace(cond)

	if strings.HasPrefix(cond, "!") {
		return !r.evaluateCondition(cond[1:])
	}

	if strings.Contains(cond, "==") {
		parts := strings.SplitN(cond, "==", 2)
		key := strings.TrimSpace(parts[0])
		expected := strings.TrimSpace(parts[1])
		expected = strings.Trim(expected, "\"'")
		actual, ok := r.env[key]
		if !ok {
			return false
		}
		return actual == expected
	}

	if strings.Contains(cond, "!=") {
		parts := strings.SplitN(cond, "!=", 2)
		key := strings.TrimSpace(parts[0])
		expected := strings.TrimSpace(parts[1])
		expected = strings.Trim(expected, "\"'")
		actual, ok := r.env[key]
		if !ok {
			return true
		}
		return actual != expected
	}

	if strings.HasPrefix(cond, "defined:") {
		key := strings.TrimSpace(cond[8:])
		_, ok := r.env[key]
		return ok
	}

	if val, ok := r.env[cond]; ok {
		return val == "true" || val == "1" || val != ""
	}

	return false
}
