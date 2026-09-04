package helm

import (
	"os"
	"path/filepath"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// WriteToDisk writes the chart structure to a directory on disk.
// Layout:
//
//	<dir>/
//	  Chart.yaml
//	  values.yaml
//	  templates/
//	    deployment.yaml
//	    service.yaml
//	    ...
func (c *Chart) WriteToDisk(dir string) error {
	if err := os.MkdirAll(filepath.Join(dir, "templates"), 0o755); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "create templates dir")
	}

	files := map[string]string{
		"Chart.yaml":  c.RenderChartMetadata(),
		"values.yaml": c.RenderValues(),
	}
	for _, t := range c.Templates {
		files[filepath.Join("templates", t.Name)] = t.Content
	}

	for rel, content := range files {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrInternal, "create dir")
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrInternal, "write %s", rel)
		}
	}
	return nil
}

// LoadFromDisk reads a chart structure from disk.
func LoadFromDisk(dir string) (*Chart, error) {
	chartPath := filepath.Join(dir, "Chart.yaml")
	data, err := os.ReadFile(chartPath)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "read Chart.yaml")
	}

	meta, err := parseChartYAML(data)
	if err != nil {
		return nil, err
	}

	chart := &Chart{
		APIVersion:  meta.APIVersion,
		Name:        meta.Name,
		Version:     meta.Version,
		Description: meta.Description,
		Type:        meta.Type,
		KubeVersion: meta.KubeVersion,
		AppVersion:  meta.AppVersion,
		Keywords:    meta.Keywords,
		Maintainers: meta.Maintainers,
	}

	valuesPath := filepath.Join(dir, "values.yaml")
	if valuesData, err := os.ReadFile(valuesPath); err == nil {
		chart.Values = parseValuesYAML(valuesData)
	}

	templatesDir := filepath.Join(dir, "templates")
	if files, err := os.ReadDir(templatesDir); err == nil {
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			tplData, err := os.ReadFile(filepath.Join(templatesDir, f.Name()))
			if err != nil {
				continue
			}
			chart.Templates = append(chart.Templates, Template{
				Name:    f.Name(),
				Content: string(tplData),
			})
		}
	}

	return chart, nil
}
