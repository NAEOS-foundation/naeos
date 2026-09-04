package helm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewChart(t *testing.T) {
	t.Parallel()
	chart, err := NewChart(ChartConfig{Name: "myapp", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("NewChart: %v", err)
	}
	if chart.APIVersion != "v2" {
		t.Errorf("expected apiVersion v2, got %s", chart.APIVersion)
	}
	if chart.Name != "myapp" {
		t.Errorf("expected name myapp, got %s", chart.Name)
	}
	if chart.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", chart.Version)
	}
	if chart.Type != "application" {
		t.Errorf("expected type application, got %s", chart.Type)
	}
	if chart.AppVersion != "1.0.0" {
		t.Errorf("expected default appVersion 1.0.0, got %s", chart.AppVersion)
	}
	if len(chart.Templates) == 0 {
		t.Error("expected templates")
	}
	if len(chart.Values) == 0 {
		t.Error("expected values")
	}
}

func TestNewChartDefaultsVersion(t *testing.T) {
	t.Parallel()
	chart, err := NewChart(ChartConfig{Name: "app"})
	if err != nil {
		t.Fatalf("NewChart: %v", err)
	}
	if chart.Version != "0.1.0" {
		t.Errorf("expected default version 0.1.0, got %s", chart.Version)
	}
}

func TestNewChartEmptyName(t *testing.T) {
	t.Parallel()
	_, err := NewChart(ChartConfig{Name: ""})
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestValidateValid(t *testing.T) {
	t.Parallel()
	chart, _ := NewChart(ChartConfig{Name: "valid"})
	errs := chart.Validate()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateEmpty(t *testing.T) {
	t.Parallel()
	chart := &Chart{}
	errs := chart.Validate()
	if len(errs) == 0 {
		t.Error("expected validation errors")
	}
}

func TestRenderChartMetadata(t *testing.T) {
	t.Parallel()
	chart, _ := NewChart(ChartConfig{Name: "meta", Version: "2.0.0", Description: "test chart"})
	meta := chart.RenderChartMetadata()
	if !strings.Contains(meta, "apiVersion: v2") {
		t.Error("missing apiVersion")
	}
	if !strings.Contains(meta, "name: meta") {
		t.Error("missing name")
	}
	if !strings.Contains(meta, "version: 2.0.0") {
		t.Error("missing version")
	}
	if !strings.Contains(meta, "description: test chart") {
		t.Error("missing description")
	}
}

func TestRenderValues(t *testing.T) {
	t.Parallel()
	chart, _ := NewChart(ChartConfig{Name: "valtest", AppVersion: "3.0.0"})
	values := chart.RenderValues()
	if !strings.Contains(values, "replicaCount: 1") {
		t.Error("missing replicaCount")
	}
	if !strings.Contains(values, "ghcr.io/naeos-foundation/valtest") {
		t.Error("missing image repository")
	}
	if !strings.Contains(values, "tag: \"3.0.0\"") {
		t.Error("missing image tag")
	}
}

func TestDefaultImageRepo(t *testing.T) {
	t.Parallel()
	got := DefaultImageRepo("test-app")
	if got != "ghcr.io/naeos-foundation/test-app" {
		t.Errorf("unexpected repo: %s", got)
	}
}

func TestTemplatesPresent(t *testing.T) {
	t.Parallel()
	chart, _ := NewChart(ChartConfig{Name: "tpl"})
	for _, name := range []string{"deployment.yaml", "service.yaml", "ingress.yaml", "hpa.yaml", "_helpers.tpl", "serviceaccount.yaml", "NOTES.txt"} {
		content, ok := chart.GetTemplate(name)
		if !ok {
			t.Errorf("missing template %s", name)
		}
		if content == "" {
			t.Errorf("empty template %s", name)
		}
	}
}

func TestSortedTemplateNames(t *testing.T) {
	t.Parallel()
	chart, _ := NewChart(ChartConfig{Name: "sort"})
	names := chart.SortedTemplateNames()
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("not sorted: %v", names)
		}
	}
}

func TestGetTemplateMissing(t *testing.T) {
	t.Parallel()
	chart, _ := NewChart(ChartConfig{Name: "missing"})
	_, ok := chart.GetTemplate("nonexistent.yaml")
	if ok {
		t.Error("expected not found")
	}
}

func TestWriteToDisk(t *testing.T) {
	t.Parallel()
	chart, _ := NewChart(ChartConfig{Name: "writetest"})
	dir := t.TempDir()

	if err := chart.WriteToDisk(dir); err != nil {
		t.Fatalf("WriteToDisk: %v", err)
	}

	for _, file := range []string{"Chart.yaml", "values.yaml", "templates/deployment.yaml", "templates/service.yaml", "templates/ingress.yaml", "templates/hpa.yaml", "templates/serviceaccount.yaml", "templates/NOTES.txt", "templates/_helpers.tpl"} {
		if _, err := os.Stat(filepath.Join(dir, file)); err != nil {
			t.Errorf("missing %s: %v", file, err)
		}
	}
}

func TestLoadFromDisk(t *testing.T) {
	t.Parallel()
	chart, _ := NewChart(ChartConfig{Name: "roundtrip", Version: "1.2.3", AppVersion: "4.0.0"})
	dir := t.TempDir()
	if err := chart.WriteToDisk(dir); err != nil {
		t.Fatalf("WriteToDisk: %v", err)
	}

	loaded, err := LoadFromDisk(dir)
	if err != nil {
		t.Fatalf("LoadFromDisk: %v", err)
	}
	if loaded.Name != "roundtrip" {
		t.Errorf("expected name roundtrip, got %s", loaded.Name)
	}
	if loaded.Version != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %s", loaded.Version)
	}
	if len(loaded.Templates) == 0 {
		t.Error("expected loaded templates")
	}
}

func TestLoadFromDiskMissing(t *testing.T) {
	t.Parallel()
	_, err := LoadFromDisk("/nonexistent/dir")
	if err == nil {
		t.Error("expected error for missing dir")
	}
}

func TestTemplateDeploymentContainsKind(t *testing.T) {
	t.Parallel()
	content := TemplateDeployment("myapp")
	if !strings.Contains(content, "kind: Deployment") {
		t.Error("expected Deployment kind")
	}
	if !strings.Contains(content, "myapp.fullname") {
		t.Error("expected fullname helper")
	}
}

func TestTemplateServiceContainsKind(t *testing.T) {
	t.Parallel()
	content := TemplateService("myapp")
	if !strings.Contains(content, "kind: Service") {
		t.Error("expected Service kind")
	}
}

func TestTemplateIngressConditional(t *testing.T) {
	t.Parallel()
	content := TemplateIngress("myapp")
	if !strings.Contains(content, "if .Values.ingress.enabled") {
		t.Error("expected ingress conditional")
	}
	if !strings.Contains(content, "kind: Ingress") {
		t.Error("expected Ingress kind")
	}
}

func TestTemplateHelpersDefines(t *testing.T) {
	t.Parallel()
	content := TemplateHelpers("myapp")
	for _, def := range []string{"myapp.name", "myapp.fullname", "myapp.labels", "myapp.selectorLabels", "myapp.serviceAccountName"} {
		if !strings.Contains(content, def) {
			t.Errorf("missing helper definition %s", def)
		}
	}
}

func TestMaintainersRendered(t *testing.T) {
	t.Parallel()
	chart, _ := NewChart(ChartConfig{Name: "maint"})
	chart.Maintainers = []Maintainer{{Name: "alice", Email: "a@example.com"}}
	meta := chart.RenderChartMetadata()
	if !strings.Contains(meta, "alice") {
		t.Error("missing maintainer name")
	}
	if !strings.Contains(meta, "a@example.com") {
		t.Error("missing maintainer email")
	}
}

func TestKeywordsRendered(t *testing.T) {
	t.Parallel()
	chart, _ := NewChart(ChartConfig{Name: "kw"})
	chart.Keywords = []string{"web", "api"}
	meta := chart.RenderChartMetadata()
	if !strings.Contains(meta, "- web") {
		t.Error("missing keyword")
	}
}

func TestValidateMissingTemplates(t *testing.T) {
	t.Parallel()
	chart := &Chart{APIVersion: "v2", Name: "x", Version: "1.0.0"}
	errs := chart.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "templates") {
			found = true
		}
	}
	if !found {
		t.Error("expected template validation error")
	}
}

func TestParseChartYAML(t *testing.T) {
	t.Parallel()
	data := []byte("apiVersion: v2\nname: test\nversion: 1.0.0\n")
	meta, err := parseChartYAML(data)
	if err != nil {
		t.Fatalf("parseChartYAML: %v", err)
	}
	if meta.Name != "test" {
		t.Errorf("expected name test, got %s", meta.Name)
	}
	if meta.APIVersion != "v2" {
		t.Errorf("expected apiVersion v2, got %s", meta.APIVersion)
	}
}

func TestParseValuesYAML(t *testing.T) {
	t.Parallel()
	v := parseValuesYAML([]byte("replicaCount: 2\nimage:\n  repository: repo\n"))
	if v["replicaCount"].Default != 2 {
		t.Errorf("expected replicaCount 2, got %v", v["replicaCount"].Default)
	}
	if v["image.repository"].Default != "repo" {
		t.Errorf("expected image.repository repo, got %v", v["image.repository"].Default)
	}
}
