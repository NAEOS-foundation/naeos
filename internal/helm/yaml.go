package helm

import (
	"strings"
)

// chartMeta is the minimal parsed Chart.yaml structure.
type chartMeta struct {
	APIVersion  string       `yaml:"apiVersion"`
	Name        string       `yaml:"name"`
	Version     string       `yaml:"version"`
	Description string       `yaml:"description"`
	Type        string       `yaml:"type"`
	KubeVersion string       `yaml:"kubeVersion"`
	AppVersion  string       `yaml:"appVersion"`
	Keywords    []string     `yaml:"keywords"`
	Maintainers []Maintainer `yaml:"maintainers"`
}

// maintainer is a local alias to avoid ambiguity with the exported type.
type maintainer struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
	URL   string `yaml:"url"`
}

// parseChartYAML parses Chart.yaml content.
func parseChartYAML(data []byte) (*chartMeta, error) {
	var meta struct {
		APIVersion  string       `yaml:"apiVersion"`
		Name        string       `yaml:"name"`
		Version     string       `yaml:"version"`
		Description string       `yaml:"description"`
		Type        string       `yaml:"type"`
		KubeVersion string       `yaml:"kubeVersion"`
		AppVersion  string       `yaml:"appVersion"`
		Keywords    []string     `yaml:"keywords"`
		Maintainers []maintainer `yaml:"maintainers"`
	}
	if err := unmarshalYAML(data, &meta); err != nil {
		return nil, err
	}

	result := &chartMeta{
		APIVersion:  meta.APIVersion,
		Name:        meta.Name,
		Version:     meta.Version,
		Description: meta.Description,
		Type:        meta.Type,
		KubeVersion: meta.KubeVersion,
		AppVersion:  meta.AppVersion,
		Keywords:    meta.Keywords,
	}
	for _, m := range meta.Maintainers {
		result.Maintainers = append(result.Maintainers, Maintainer{
			Name:  m.Name,
			Email: m.Email,
			URL:   m.URL,
		})
	}
	return result, nil
}

// parseValuesYAML parses values.yaml into a simple key->Value map.
func parseValuesYAML(data []byte) map[string]Value {
	var raw map[string]any
	if err := unmarshalYAML(data, &raw); err != nil {
		return nil
	}
	return flattenValues(raw, "")
}

func flattenValues(raw map[string]any, prefix string) map[string]Value {
	out := make(map[string]Value)
	for k, v := range raw {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]any:
			for k2, v2 := range flattenValues(val, key) {
				out[k2] = v2
			}
		default:
			out[key] = Value{Default: val, Type: typeName(val)}
		}
	}
	return out
}

func typeName(v any) string {
	switch v.(type) {
	case bool:
		return "bool"
	case int, int64, float64:
		return "int"
	case string:
		return "string"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return "unknown"
	}
}

// RenderChartYAML renders a Chart object back to YAML for testing.
func RenderChartYAML(c *Chart) string {
	var b strings.Builder
	if c.APIVersion != "" {
		b.WriteString("apiVersion: " + c.APIVersion + "\n")
	}
	if c.Name != "" {
		b.WriteString("name: " + c.Name + "\n")
	}
	if c.Description != "" {
		b.WriteString("description: " + c.Description + "\n")
	}
	if c.Type != "" {
		b.WriteString("type: " + c.Type + "\n")
	}
	if c.KubeVersion != "" {
		b.WriteString("kubeVersion: " + c.KubeVersion + "\n")
	}
	b.WriteString("version: " + c.Version + "\n")
	if c.AppVersion != "" {
		b.WriteString("appVersion: " + c.AppVersion + "\n")
	}
	return b.String()
}
