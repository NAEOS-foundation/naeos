package diff

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
)

func TestNEIRDiffNoChanges(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
		},
	}

	diff := ComputeNEIRDiff(neir, neir)
	if diff.Summary != "no changes" {
		t.Errorf("expected 'no changes', got %q", diff.Summary)
	}
}

func TestNEIRDiffAddedServices(t *testing.T) {
	old := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
		},
	}
	new := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
			{Name: "worker", Port: 9090},
		},
	}

	diff := ComputeNEIRDiff(old, new)
	if len(diff.ServicesDiff.Added) != 1 {
		t.Errorf("expected 1 added, got %d", len(diff.ServicesDiff.Added))
	}
	if diff.ServicesDiff.Added[0].Name != "worker" {
		t.Errorf("expected 'worker' added, got %s", diff.ServicesDiff.Added[0].Name)
	}
	if !strings.Contains(diff.Summary, "+1 services") {
		t.Errorf("expected summary to mention +1 services, got %q", diff.Summary)
	}
}

func TestNEIRDiffRemovedServices(t *testing.T) {
	old := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
			{Name: "worker", Port: 9090},
		},
	}
	new := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
		},
	}

	diff := ComputeNEIRDiff(old, new)
	if len(diff.ServicesDiff.Removed) != 1 {
		t.Errorf("expected 1 removed, got %d", len(diff.ServicesDiff.Removed))
	}
	if diff.ServicesDiff.Removed[0].Name != "worker" {
		t.Errorf("expected 'worker' removed, got %s", diff.ServicesDiff.Removed[0].Name)
	}
}

func TestNEIRDiffModifiedPort(t *testing.T) {
	old := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
		},
	}
	new := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 3000},
		},
	}

	diff := ComputeNEIRDiff(old, new)
	if len(diff.ServicesDiff.Modified) != 1 {
		t.Errorf("expected 1 modified, got %d", len(diff.ServicesDiff.Modified))
	}
	mod := diff.ServicesDiff.Modified[0]
	if mod.Name != "api" {
		t.Errorf("expected 'api' modified, got %s", mod.Name)
	}
	if len(mod.Changes) != 1 || mod.Changes[0].Field != "port" {
		t.Errorf("expected port change, got %v", mod.Changes)
	}
}

func TestNEIRDiffProjectRenamed(t *testing.T) {
	old := &model.NEIR{
		Project: &project.Project{Name: "old-name"},
	}
	new := &model.NEIR{
		Project: &project.Project{Name: "new-name"},
	}

	diff := ComputeNEIRDiff(old, new)
	if !diff.ProjectDiff.NameChanged {
		t.Error("expected name changed")
	}
	if diff.ProjectDiff.OldName != "old-name" || diff.ProjectDiff.NewName != "new-name" {
		t.Errorf("expected old-name -> new-name, got %s -> %s", diff.ProjectDiff.OldName, diff.ProjectDiff.NewName)
	}
}

func TestNEIRDiffNilOld(t *testing.T) {
	new := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
		},
	}

	diff := ComputeNEIRDiff(nil, new)
	if len(diff.ServicesDiff.Added) != 1 {
		t.Errorf("expected 1 added, got %d", len(diff.ServicesDiff.Added))
	}
}

func TestNEIRDiffNilNew(t *testing.T) {
	old := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
	}

	diff := ComputeNEIRDiff(old, nil)
	if !strings.Contains(diff.Summary, "spec removed") {
		t.Errorf("expected 'spec removed', got %q", diff.Summary)
	}
}

func TestNEIRDiffBothNil(t *testing.T) {
	diff := ComputeNEIRDiff(nil, nil)
	if diff.Summary != "" {
		t.Errorf("expected empty summary for both nil, got %q", diff.Summary)
	}
}

func TestFormatNEIRDiff(t *testing.T) {
	old := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 8080},
		},
	}
	new := &model.NEIR{
		Project: &project.Project{Name: "myapp"},
		Services: []service.Service{
			{Name: "api", Port: 3000},
			{Name: "worker", Port: 9090},
		},
	}

	diff := ComputeNEIRDiff(old, new)
	formatted := FormatNEIRDiff(diff)

	if !strings.Contains(formatted, "NEIR Diff") {
		t.Error("expected NEIR Diff header")
	}
	if !strings.Contains(formatted, "+1 services") {
		t.Error("expected +1 services in formatted output")
	}
}

func TestFormatNEIRDiffNil(t *testing.T) {
	if FormatNEIRDiff(nil) != "" {
		t.Error("expected empty string for nil diff")
	}
}
