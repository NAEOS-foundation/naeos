package helm

import (
	"fmt"
	"sort"
	"strings"
)

// Chart is an in-memory Helm chart with Chart.yaml metadata and templates.
type Chart struct {
	APIVersion  string            `json:"apiVersion"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description,omitempty"`
	Type        string            `json:"type,omitempty"`
	KubeVersion string            `json:"kubeVersion,omitempty"`
	AppVersion  string            `json:"appVersion,omitempty"`
	Keywords    []string          `json:"keywords,omitempty"`
	Maintainers []Maintainer      `json:"maintainers,omitempty"`
	Values      map[string]Value  `json:"values,omitempty"`
	Templates   []Template        `json:"templates,omitempty"`
}

// Maintainer is a chart maintainer.
type Maintainer struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// Value describes a chart value with metadata.
type Value struct {
	Type        string `json:"type,omitempty"`
	Default     any    `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// Template is a Helm chart template file (Go template syntax).
type Template struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// ChartConfig configures chart creation.
type ChartConfig struct {
	Name        string
	Version     string
	Description string
	AppVersion  string
	KubeVersion string
}

// NewChart creates a new Helm chart with sensible defaults.
func NewChart(cfg ChartConfig) (*Chart, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("chart name is required")
	}
	if cfg.Version == "" {
		cfg.Version = "0.1.0"
	}

	chart := &Chart{
		APIVersion:  "v2",
		Name:        cfg.Name,
		Version:     cfg.Version,
		Description: cfg.Description,
		Type:        "application",
		KubeVersion: cfg.KubeVersion,
		AppVersion:  cfg.AppVersion,
	}

	if chart.AppVersion == "" {
		chart.AppVersion = "1.0.0"
	}

	// Sensible default values.
	chart.Values = map[string]Value{
		"replicaCount": {Type: "int", Default: 1, Description: "Number of replicas"},
		"image.repository": {Type: "string", Default: DefaultImageRepo(chart.Name), Description: "Container image repository"},
		"image.tag":       {Type: "string", Default: chart.AppVersion, Description: "Container image tag"},
		"image.pullPolicy": {Type: "string", Default: "IfNotPresent", Description: "Image pull policy"},
		"service.type":    {Type: "string", Default: "ClusterIP", Description: "Kubernetes service type"},
		"service.port":    {Type: "int", Default: 80, Description: "Service port"},
		"service.targetPort": {Type: "int", Default: 8080, Description: "Container target port"},
		"resources.limits.cpu":    {Type: "string", Default: "500m", Description: "CPU limit"},
		"resources.limits.memory": {Type: "string", Default: "128Mi", Description: "Memory limit"},
		"resources.requests.cpu":  {Type: "string", Default: "100m", Description: "CPU request"},
		"resources.requests.memory": {Type: "string", Default: "64Mi", Description: "Memory request"},
		"ingress.enabled":  {Type: "bool", Default: false, Description: "Enable ingress"},
		"ingress.host":     {Type: "string", Default: chart.Name + ".example.com", Description: "Ingress host"},
		"autoscaling.enabled": {Type: "bool", Default: false, Description: "Enable autoscaling"},
		"autoscaling.minReplicas": {Type: "int", Default: 1, Description: "Min replicas"},
		"autoscaling.maxReplicas": {Type: "int", Default: 10, Description: "Max replicas"},
		"autoscaling.targetCPUUtilizationPercentage": {Type: "int", Default: 80, Description: "CPU target utilization"},
		"nodeSelector": {Type: "object", Default: map[string]any{}, Description: "Node selector"},
		"tolerations":  {Type: "array", Default: []any{}, Description: "Tolerations"},
		"affinity":     {Type: "object", Default: map[string]any{}, Description: "Affinity"},
		"fullnameOverride": {Type: "string", Default: chart.Name, Description: "Override full name"},
	}

	chart.Templates = []Template{
		{Name: "deployment.yaml", Content: TemplateDeployment(chart.Name)},
		{Name: "service.yaml", Content: TemplateService(chart.Name)},
		{Name: "ingress.yaml", Content: TemplateIngress(chart.Name)},
		{Name: "hpa.yaml", Content: TemplateHPA(chart.Name)},
		{Name: "_helpers.tpl", Content: TemplateHelpers(chart.Name)},
		{Name: "serviceaccount.yaml", Content: TemplateServiceAccount(chart.Name)},
		{Name: "NOTES.txt", Content: TemplateNotes(chart.Name)},
	}

	return chart, nil
}

// DefaultImageRepo returns a default image repository for a chart name.
func DefaultImageRepo(name string) string {
	return "ghcr.io/naeos-foundation/" + name
}

// Validate checks that a chart is well-formed.
func (c *Chart) Validate() []error {
	var errs []error
	if c.APIVersion == "" {
		errs = append(errs, fmt.Errorf("apiVersion is required"))
	}
	if c.Name == "" {
		errs = append(errs, fmt.Errorf("name is required"))
	}
	if c.Version == "" {
		errs = append(errs, fmt.Errorf("version is required"))
	}
	for _, t := range c.Templates {
		if t.Name == "" {
			errs = append(errs, fmt.Errorf("template name is required"))
		}
	}
	if len(c.Templates) == 0 {
		errs = append(errs, fmt.Errorf("chart must have templates"))
	}
	return errs
}

// Render produces the Chart.yaml content.
func (c *Chart) RenderChartMetadata() string {
	var b strings.Builder
	fmt.Fprintf(&b, "apiVersion: %s\n", c.APIVersion)
	fmt.Fprintf(&b, "name: %s\n", c.Name)
	fmt.Fprintf(&b, "description: %s\n", c.Description)
	fmt.Fprintf(&b, "type: %s\n", c.Type)
	if c.KubeVersion != "" {
		fmt.Fprintf(&b, "kubeVersion: %s\n", c.KubeVersion)
	}
	fmt.Fprintf(&b, "version: %s\n", c.Version)
	fmt.Fprintf(&b, "appVersion: %q\n", c.AppVersion)
	if len(c.Keywords) > 0 {
		fmt.Fprintf(&b, "keywords:\n")
		for _, k := range c.Keywords {
			fmt.Fprintf(&b, "  - %s\n", k)
		}
	}
	if len(c.Maintainers) > 0 {
		fmt.Fprintf(&b, "maintainers:\n")
		for _, m := range c.Maintainers {
			fmt.Fprintf(&b, "  - name: %s\n", m.Name)
			if m.Email != "" {
				fmt.Fprintf(&b, "    email: %s\n", m.Email)
			}
			if m.URL != "" {
				fmt.Fprintf(&b, "    url: %s\n", m.URL)
			}
		}
	}
	return b.String()
}

// RenderValues produces the values.yaml content.
func (c *Chart) RenderValues() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Default values for %s.\n", c.Name)
	fmt.Fprintf(&b, "# This is a YAML-formatted file.\n")
	fmt.Fprintf(&b, "# Declare variables to be passed into your templates.\n\n")
	fmt.Fprintf(&b, "replicaCount: 1\n\n")
	fmt.Fprintf(&b, "image:\n")
	fmt.Fprintf(&b, "  repository: %s\n", DefaultImageRepo(c.Name))
	fmt.Fprintf(&b, "  tag: %q\n", c.AppVersion)
	fmt.Fprintf(&b, "  pullPolicy: IfNotPresent\n\n")
	fmt.Fprintf(&b, "imagePullSecrets: []\n")
	fmt.Fprintf(&b, "nameOverride: \"\"\n")
	fmt.Fprintf(&b, "fullnameOverride: %q\n", c.Name)
	fmt.Fprintf(&b, "\nserviceAccount:\n")
	fmt.Fprintf(&b, "  create: true\n")
	fmt.Fprintf(&b, "  annotations: {}\n")
	fmt.Fprintf(&b, "  name: \"\"\n")
	fmt.Fprintf(&b, "\nservice:\n")
	fmt.Fprintf(&b, "  type: ClusterIP\n")
	fmt.Fprintf(&b, "  port: 80\n")
	fmt.Fprintf(&b, "  targetPort: 8080\n")
	fmt.Fprintf(&b, "\ningress:\n")
	fmt.Fprintf(&b, "  enabled: false\n")
	fmt.Fprintf(&b, "  className: \"\"\n")
	fmt.Fprintf(&b, "  annotations: {}\n")
	fmt.Fprintf(&b, "  host: %s.example.com\n", c.Name)
	fmt.Fprintf(&b, "  tls: []\n")
	fmt.Fprintf(&b, "\nautoscaling:\n")
	fmt.Fprintf(&b, "  enabled: false\n")
	fmt.Fprintf(&b, "  minReplicas: 1\n")
	fmt.Fprintf(&b, "  maxReplicas: 10\n")
	fmt.Fprintf(&b, "  targetCPUUtilizationPercentage: 80\n")
	fmt.Fprintf(&b, "\nresources:\n")
	fmt.Fprintf(&b, "  limits:\n")
	fmt.Fprintf(&b, "    cpu: 500m\n")
	fmt.Fprintf(&b, "    memory: 128Mi\n")
	fmt.Fprintf(&b, "  requests:\n")
	fmt.Fprintf(&b, "    cpu: 100m\n")
	fmt.Fprintf(&b, "    memory: 64Mi\n")
	fmt.Fprintf(&b, "\nnodeSelector: {}\n")
	fmt.Fprintf(&b, "tolerations: []\n")
	fmt.Fprintf(&b, "affinity: {}\n")
	return b.String()
}

// SortedTemplateNames returns template names sorted alphabetically.
func (c *Chart) SortedTemplateNames() []string {
	names := make([]string, 0, len(c.Templates))
	for _, t := range c.Templates {
		names = append(names, t.Name)
	}
	sort.Strings(names)
	return names
}

// GetTemplate returns the template with the given name.
func (c *Chart) GetTemplate(name string) (string, bool) {
	for _, t := range c.Templates {
		if t.Name == name {
			return t.Content, true
		}
	}
	return "", false
}
