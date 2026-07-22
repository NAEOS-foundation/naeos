package diff

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
)

func TestRenderVisualDiffNoChanges(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project:  &project.Project{Name: "myapp"},
		Services: []service.Service{{Name: "api", Port: 8080}},
	}
	html := RenderVisualDiff(neir, neir)
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected HTML doctype")
	}
	if !strings.Contains(html, "NEIR Architecture Diff") {
		t.Error("expected title")
	}
	if !strings.Contains(html, "no changes") {
		t.Error("expected no changes summary")
	}
}

func TestRenderVisualDiffWithAdded(t *testing.T) {
	t.Parallel()
	old := &model.NEIR{
		Project:  &project.Project{Name: "myapp"},
		Services: []service.Service{{Name: "api", Port: 8080}},
	}
	new := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
			{Name: "worker", Port: 9090, Kind: service.KindWorker},
		},
	}
	html := RenderVisualDiff(old, new)
	if !strings.Contains(html, "worker") {
		t.Error("expected added service name in HTML")
	}
	if !strings.Contains(html, "9090") {
		t.Error("expected added service port in HTML")
	}
	if !strings.Contains(html, "graph-node added") {
		t.Error("expected graph-node added class")
	}
}

func TestRenderVisualDiffWithRemoved(t *testing.T) {
	t.Parallel()
	old := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
			{Name: "old-svc", Port: 3000},
		},
	}
	new := &model.NEIR{
		Project:  &project.Project{Name: "myapp"},
		Services: []service.Service{{Name: "api", Port: 8080}},
	}
	html := RenderVisualDiff(old, new)
	if !strings.Contains(html, "old-svc") {
		t.Error("expected removed service name in HTML")
	}
	if !strings.Contains(html, "graph-node removed") {
		t.Error("expected graph-node removed class")
	}
}

func TestRenderVisualDiffWithModified(t *testing.T) {
	t.Parallel()
	old := &model.NEIR{
		Project:  &project.Project{Name: "myapp"},
		Services: []service.Service{{Name: "api", Port: 8080}},
	}
	new := &model.NEIR{
		Project:  &project.Project{Name: "myapp"},
		Services: []service.Service{{Name: "api", Port: 3000}},
	}
	html := RenderVisualDiff(old, new)
	if !strings.Contains(html, "graph-node modified") {
		t.Error("expected graph-node modified class")
	}
}

func TestRenderVisualDiffProjectRenamed(t *testing.T) {
	t.Parallel()
	old := &model.NEIR{Project: &project.Project{Name: "old-name"}}
	new := &model.NEIR{Project: &project.Project{Name: "new-name"}}
	html := RenderVisualDiff(old, new)
	if !strings.Contains(html, "old-name") || !strings.Contains(html, "new-name") {
		t.Error("expected project name change in HTML")
	}
	if !strings.Contains(html, "project-name") {
		t.Error("expected project-name class")
	}
}

func TestRenderVisualDiffNilOld(t *testing.T) {
	t.Parallel()
	new := &model.NEIR{
		Project:  &project.Project{Name: "myapp"},
		Services: []service.Service{{Name: "api", Port: 8080}},
	}
	html := RenderVisualDiff(nil, new)
	if !strings.Contains(html, "api") {
		t.Error("expected service in HTML")
	}
}

func TestRenderVisualDiffNilNew(t *testing.T) {
	t.Parallel()
	old := &model.NEIR{
		Project:  &project.Project{Name: "myapp"},
		Services: []service.Service{{Name: "api", Port: 8080}},
	}
	html := RenderVisualDiff(old, nil)
	if !strings.Contains(html, "api") {
		t.Error("expected service in HTML")
	}
}

func TestRenderVisualDiffBothEmpty(t *testing.T) {
	t.Parallel()
	html := RenderVisualDiff(&model.NEIR{}, &model.NEIR{})
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected valid HTML")
	}
}

func TestRenderGraphNodesEmpty(t *testing.T) {
	t.Parallel()
	diff := &NEIRDiff{ServicesDiff: nil}
	result := renderGraphNodes(diff)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestRenderGraphNodesMixed(t *testing.T) {
	t.Parallel()
	diff := &NEIRDiff{
		ServicesDiff: &ServicesDiff{
			Added:   []service.Service{{Name: "new-svc"}},
			Removed: []service.Service{{Name: "old-svc"}},
			Modified: []ServiceModification{
				{Name: "mod-svc"},
			},
		},
	}
	result := renderGraphNodes(diff)
	if !strings.Contains(result, "new-svc") {
		t.Error("expected new-svc in graph")
	}
	if !strings.Contains(result, "old-svc") {
		t.Error("expected old-svc in graph")
	}
	if !strings.Contains(result, "mod-svc") {
		t.Error("expected mod-svc in graph")
	}
	if strings.HasSuffix(strings.TrimSpace(result), "→") {
		t.Error("graph should not end with arrow")
	}
}

func TestRenderVisualDiffFieldsModified(t *testing.T) {
	t.Parallel()
	old := &model.NEIR{
		Project: &project.Project{Name: "myapp", Version: "1.0"},
	}
	new := &model.NEIR{
		Project: &project.Project{Name: "myapp", Version: "2.0"},
	}
	html := RenderVisualDiff(old, new)
	if !strings.Contains(html, "project-fields") {
		t.Error("expected project-fields section for modified fields")
	}
}
