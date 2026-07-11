package dashboard

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates/*
var templatesFS embed.FS

type Dashboard struct {
	templates *template.Template
}

func New() (*Dashboard, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Dashboard{
		templates: tmpl,
	}, nil
}

func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	d.templates.ExecuteTemplate(w, "index.html", nil)
}

type Stats struct {
	Projects    int `json:"projects"`
	Artifacts   int `json:"artifacts"`
	Pipelines   int `json:"pipelines"`
	LastRun     string `json:"last_run"`
}

func GetStats() *Stats {
	return &Stats{
		Projects:  1,
		Artifacts: 0,
		Pipelines: 0,
		LastRun:   "never",
	}
}
