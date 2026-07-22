package promptlib

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// FuncMap provides custom template functions for prompt rendering.
var FuncMap = template.FuncMap{
	"join":      strings.Join,
	"upper":     strings.ToUpper,
	"lower":     strings.ToLower,
	"title":     titleFunc,
	"hasPrefix": strings.HasPrefix,
	"hasSuffix": strings.HasSuffix,
	"contains":  strings.Contains,
	"trim":      strings.TrimSpace,
	"replace":   strings.ReplaceAll,
	"json":      toJSONFunc,
	"yaml":      toYAMLFunc,
	"default":   defaultFunc,
	"len":       lenFunc,
	"rangeSeq":  rangeSeqFunc,
	"bt":        backtickFunc,
	"code":      codeFunc,
}

func toJSONFunc(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}
	return string(b), nil
}

func toYAMLFunc(v any) (string, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("yaml marshal: %w", err)
	}
	return string(b), nil
}

func defaultFunc(def, val any) any {
	if val == nil {
		return def
	}
	switch v := val.(type) {
	case string:
		if v == "" {
			return def
		}
	case []any:
		if len(v) == 0 {
			return def
		}
	case []string:
		if len(v) == 0 {
			return def
		}
	}
	return val
}

func lenFunc(v any) int {
	switch val := v.(type) {
	case string:
		return len(val)
	case []any:
		return len(val)
	case []string:
		return len(val)
	case map[string]any:
		return len(val)
	default:
		return 0
	}
}

func rangeSeqFunc(n int) []int {
	result := make([]int, n)
	for i := range result {
		result[i] = i + 1
	}
	return result
}

func backtickFunc() string {
	return "`"
}

func codeFunc(s string) string {
	return "`" + s + "`"
}

func titleFunc(s string) string {
	if s == "" {
		return s
	}
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
