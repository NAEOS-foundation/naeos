package diff

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

func TestCompareSpecsNoChanges(t *testing.T) {
	t.Parallel()
	doc := &parser.SpecDocument{
		Project: "myapp",
		Modules: []parser.Module{
			{Name: "core", Path: "./core", Description: "Core"},
		},
		Services: []parser.Service{
			{Name: "api", Kind: "http", Port: 8080},
		},
	}
	diff := CompareSpecs(doc, doc)
	if diff.Project != nil {
		t.Error("expected nil project diff when same")
	}
	if len(diff.Modules) != 1 {
		t.Errorf("expected 1 module diff, got %d", len(diff.Modules))
	}
	if diff.Modules[0].Type != ChangeUnchanged {
		t.Errorf("expected ChangeUnchanged, got %s", diff.Modules[0].Type)
	}
	if len(diff.Services) != 1 {
		t.Errorf("expected 1 service diff, got %d", len(diff.Services))
	}
	if diff.Services[0].Type != ChangeUnchanged {
		t.Errorf("expected ChangeUnchanged, got %s", diff.Services[0].Type)
	}
}

func TestCompareSpecsProjectChanged(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{Project: "old-proj"}
	new := &parser.SpecDocument{Project: "new-proj"}
	diff := CompareSpecs(old, new)
	if diff.Project == nil {
		t.Fatal("expected project diff")
	}
	if diff.Project.Type != ChangeModified {
		t.Errorf("expected ChangeModified, got %s", diff.Project.Type)
	}
	if diff.Project.OldValue != "old-proj" {
		t.Errorf("expected old-proj, got %v", diff.Project.OldValue)
	}
	if diff.Project.NewValue != "new-proj" {
		t.Errorf("expected new-proj, got %v", diff.Project.NewValue)
	}
}

func TestCompareSpecsModuleAdded(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "core"}},
	}
	new := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "core"}, {Name: "auth", Path: "./auth"}},
	}
	diff := CompareSpecs(old, new)
	found := false
	for _, m := range diff.Modules {
		if m.Name == "auth" && m.Type == ChangeAdded {
			found = true
		}
	}
	if !found {
		t.Error("expected auth module to be added")
	}
}

func TestCompareSpecsModuleRemoved(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "core"}, {Name: "auth"}},
	}
	new := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "core"}},
	}
	diff := CompareSpecs(old, new)
	found := false
	for _, m := range diff.Modules {
		if m.Name == "auth" && m.Type == ChangeRemoved {
			found = true
		}
	}
	if !found {
		t.Error("expected auth module to be removed")
	}
}

func TestCompareSpecsModuleModified(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "core", Path: "./core", Description: "old desc"}},
	}
	new := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "core", Path: "./core", Description: "new desc"}},
	}
	diff := CompareSpecs(old, new)
	found := false
	for _, m := range diff.Modules {
		if m.Name == "core" && m.Type == ChangeModified {
			found = true
		}
	}
	if !found {
		t.Error("expected core module to be modified")
	}
}

func TestCompareSpecsServiceAdded(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Services: []parser.Service{{Name: "api"}},
	}
	new := &parser.SpecDocument{
		Services: []parser.Service{{Name: "api"}, {Name: "worker", Port: 9090}},
	}
	diff := CompareSpecs(old, new)
	found := false
	for _, s := range diff.Services {
		if s.Name == "worker" && s.Type == ChangeAdded {
			found = true
		}
	}
	if !found {
		t.Error("expected worker service to be added")
	}
}

func TestCompareSpecsServiceRemoved(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Services: []parser.Service{{Name: "api"}, {Name: "worker"}},
	}
	new := &parser.SpecDocument{
		Services: []parser.Service{{Name: "api"}},
	}
	diff := CompareSpecs(old, new)
	found := false
	for _, s := range diff.Services {
		if s.Name == "worker" && s.Type == ChangeRemoved {
			found = true
		}
	}
	if !found {
		t.Error("expected worker service to be removed")
	}
}

func TestCompareSpecsServiceModified(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Services: []parser.Service{{Name: "api", Kind: "http", Port: 8080}},
	}
	new := &parser.SpecDocument{
		Services: []parser.Service{{Name: "api", Kind: "grpc", Port: 9090}},
	}
	diff := CompareSpecs(old, new)
	found := false
	for _, s := range diff.Services {
		if s.Name == "api" && s.Type == ChangeModified {
			found = true
		}
	}
	if !found {
		t.Error("expected api service to be modified")
	}
}

func TestCompareSpecsServiceEqual(t *testing.T) {
	t.Parallel()
	svc := parser.Service{Name: "api", Kind: "http", Port: 8080, Endpoints: []parser.Endpoint{
		{Method: "GET", Path: "/users", Action: "list"},
	}}
	old := &parser.SpecDocument{Services: []parser.Service{svc}}
	new := &parser.SpecDocument{Services: []parser.Service{svc}}
	diff := CompareSpecs(old, new)
	if diff.Services[0].Type != ChangeUnchanged {
		t.Errorf("expected ChangeUnchanged, got %s", diff.Services[0].Type)
	}
}

func TestModulesEqual(t *testing.T) {
	t.Parallel()
	a := &parser.Module{Name: "core", Path: "./core", Description: "desc", Dependencies: []string{"a", "b"}}
	b := &parser.Module{Name: "core", Path: "./core", Description: "desc", Dependencies: []string{"a", "b"}}
	if !modulesEqual(a, b) {
		t.Error("expected modules to be equal")
	}
}

func TestModulesNotEqualName(t *testing.T) {
	t.Parallel()
	a := &parser.Module{Name: "core"}
	b := &parser.Module{Name: "other"}
	if modulesEqual(a, b) {
		t.Error("expected modules to not be equal")
	}
}

func TestModulesNotEqualPath(t *testing.T) {
	t.Parallel()
	a := &parser.Module{Name: "core", Path: "./a"}
	b := &parser.Module{Name: "core", Path: "./b"}
	if modulesEqual(a, b) {
		t.Error("expected modules to not be equal")
	}
}

func TestModulesNotEqualDescription(t *testing.T) {
	t.Parallel()
	a := &parser.Module{Name: "core", Description: "old"}
	b := &parser.Module{Name: "core", Description: "new"}
	if modulesEqual(a, b) {
		t.Error("expected modules to not be equal")
	}
}

func TestModulesNotEqualDepsLength(t *testing.T) {
	t.Parallel()
	a := &parser.Module{Dependencies: []string{"a"}}
	b := &parser.Module{Dependencies: []string{"a", "b"}}
	if modulesEqual(a, b) {
		t.Error("expected modules to not be equal")
	}
}

func TestModulesNotEqualDepsValue(t *testing.T) {
	t.Parallel()
	a := &parser.Module{Dependencies: []string{"a"}}
	b := &parser.Module{Dependencies: []string{"b"}}
	if modulesEqual(a, b) {
		t.Error("expected modules to not be equal")
	}
}

func TestServicesEqual(t *testing.T) {
	t.Parallel()
	svc := parser.Service{Name: "api", Kind: "http", Port: 8080, Description: "desc",
		Endpoints: []parser.Endpoint{{Method: "GET", Path: "/a", Action: "x"}}}
	if !servicesEqual(&svc, &svc) {
		t.Error("expected services to be equal")
	}
}

func TestServicesNotEqualName(t *testing.T) {
	t.Parallel()
	a := &parser.Service{Name: "api"}
	b := &parser.Service{Name: "other"}
	if servicesEqual(a, b) {
		t.Error("expected services to not be equal")
	}
}

func TestServicesNotEqualKind(t *testing.T) {
	t.Parallel()
	a := &parser.Service{Name: "api", Kind: "http"}
	b := &parser.Service{Name: "api", Kind: "grpc"}
	if servicesEqual(a, b) {
		t.Error("expected services to not be equal")
	}
}

func TestServicesNotEqualPort(t *testing.T) {
	t.Parallel()
	a := &parser.Service{Name: "api", Port: 8080}
	b := &parser.Service{Name: "api", Port: 9090}
	if servicesEqual(a, b) {
		t.Error("expected services to not be equal")
	}
}

func TestServicesNotEqualDescription(t *testing.T) {
	t.Parallel()
	a := &parser.Service{Name: "api", Description: "old"}
	b := &parser.Service{Name: "api", Description: "new"}
	if servicesEqual(a, b) {
		t.Error("expected services to not be equal")
	}
}

func TestServicesNotEqualEndpointsLength(t *testing.T) {
	t.Parallel()
	a := &parser.Service{Endpoints: []parser.Endpoint{{}}}
	b := &parser.Service{Endpoints: []parser.Endpoint{{}, {}}}
	if servicesEqual(a, b) {
		t.Error("expected services to not be equal")
	}
}

func TestServicesNotEqualEndpointMethod(t *testing.T) {
	t.Parallel()
	a := &parser.Service{Endpoints: []parser.Endpoint{{Method: "GET"}}}
	b := &parser.Service{Endpoints: []parser.Endpoint{{Method: "POST"}}}
	if servicesEqual(a, b) {
		t.Error("expected services to not be equal")
	}
}

func TestServicesNotEqualEndpointPath(t *testing.T) {
	t.Parallel()
	a := &parser.Service{Endpoints: []parser.Endpoint{{Path: "/a"}}}
	b := &parser.Service{Endpoints: []parser.Endpoint{{Path: "/b"}}}
	if servicesEqual(a, b) {
		t.Error("expected services to not be equal")
	}
}

func TestServicesNotEqualEndpointAction(t *testing.T) {
	t.Parallel()
	a := &parser.Service{Endpoints: []parser.Endpoint{{Action: "x"}}}
	b := &parser.Service{Endpoints: []parser.Endpoint{{Action: "y"}}}
	if servicesEqual(a, b) {
		t.Error("expected services to not be equal")
	}
}

func TestFormatSpecDiff(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Project:  "old-proj",
		Modules:  []parser.Module{{Name: "core", Path: "./core"}},
		Services: []parser.Service{{Name: "api", Kind: "http", Port: 8080}},
	}
	new := &parser.SpecDocument{
		Project:  "new-proj",
		Modules:  []parser.Module{{Name: "core", Path: "./core"}, {Name: "auth", Path: "./auth"}},
		Services: []parser.Service{{Name: "api", Kind: "grpc", Port: 9090}},
	}
	diff := CompareSpecs(old, new)
	formatted := FormatSpecDiff(diff)
	if !strings.Contains(formatted, "old-proj") {
		t.Error("formatted diff should contain old project name")
	}
	if !strings.Contains(formatted, "new-proj") {
		t.Error("formatted diff should contain new project name")
	}
	if !strings.Contains(formatted, "auth") {
		t.Error("formatted diff should contain added module")
	}
	if !strings.Contains(formatted, "9090") {
		t.Error("formatted diff should contain new port")
	}
}

func TestFormatSpecDiffEmpty(t *testing.T) {
	t.Parallel()
	doc := &parser.SpecDocument{Project: "same"}
	diff := CompareSpecs(doc, doc)
	formatted := FormatSpecDiff(diff)
	if formatted != "" {
		t.Errorf("expected empty formatted diff for no changes, got %q", formatted)
	}
}

func TestFormatSpecDiffSummary(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Project:  "p1",
		Modules:  []parser.Module{{Name: "a"}},
		Services: []parser.Service{{Name: "s1"}},
	}
	new := &parser.SpecDocument{
		Project:  "p2",
		Modules:  []parser.Module{{Name: "a"}, {Name: "b"}},
		Services: []parser.Service{{Name: "s1"}, {Name: "s2"}},
	}
	diff := CompareSpecs(old, new)
	if !strings.Contains(diff.Summary, "Project: p1") {
		t.Error("summary should contain old project")
	}
	if !strings.Contains(diff.Summary, "p2") {
		t.Error("summary should contain new project")
	}
	if !strings.Contains(diff.Summary, "+1") {
		t.Error("summary should contain added count")
	}
}

func TestFormatSpecDiffSummaryNoChanges(t *testing.T) {
	t.Parallel()
	doc := &parser.SpecDocument{Project: "p"}
	diff := CompareSpecs(doc, doc)
	if diff.Summary != "" {
		t.Errorf("expected empty summary for no changes, got %q", diff.Summary)
	}
}

func TestFormatSpecDiffModulePathChanged(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "core", Path: "./old"}},
	}
	new := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "core", Path: "./new"}},
	}
	diff := CompareSpecs(old, new)
	formatted := FormatSpecDiff(diff)
	if !strings.Contains(formatted, "./old") || !strings.Contains(formatted, "./new") {
		t.Error("formatted diff should show path change")
	}
}

func TestFormatSpecDiffModuleDescChanged(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "core", Description: "old"}},
	}
	new := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "core", Description: "new"}},
	}
	diff := CompareSpecs(old, new)
	formatted := FormatSpecDiff(diff)
	if !strings.Contains(formatted, "old") || !strings.Contains(formatted, "new") {
		t.Error("formatted diff should show description change")
	}
}

func TestFormatSpecDiffServicePortChanged(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Services: []parser.Service{{Name: "api", Port: 8080}},
	}
	new := &parser.SpecDocument{
		Services: []parser.Service{{Name: "api", Port: 9090}},
	}
	diff := CompareSpecs(old, new)
	formatted := FormatSpecDiff(diff)
	if !strings.Contains(formatted, "8080") || !strings.Contains(formatted, "9090") {
		t.Error("formatted diff should show port change")
	}
}

func TestFormatSpecDiffServiceKindChanged(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Services: []parser.Service{{Name: "api", Kind: "http"}},
	}
	new := &parser.SpecDocument{
		Services: []parser.Service{{Name: "api", Kind: "grpc"}},
	}
	diff := CompareSpecs(old, new)
	formatted := FormatSpecDiff(diff)
	if !strings.Contains(formatted, "http") || !strings.Contains(formatted, "grpc") {
		t.Error("formatted diff should show kind change")
	}
}

func TestFormatSpecDiffServiceRemoved(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Services: []parser.Service{{Name: "api", Kind: "http", Port: 8080}},
	}
	new := &parser.SpecDocument{
		Services: []parser.Service{},
	}
	diff := CompareSpecs(old, new)
	formatted := FormatSpecDiff(diff)
	if !strings.Contains(formatted, "api") {
		t.Error("formatted diff should contain removed service name")
	}
}

func TestFormatSpecDiffServiceAdded(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{}
	new := &parser.SpecDocument{
		Services: []parser.Service{{Name: "worker", Kind: "grpc", Port: 9090}},
	}
	diff := CompareSpecs(old, new)
	formatted := FormatSpecDiff(diff)
	if !strings.Contains(formatted, "worker") {
		t.Error("formatted diff should contain added service name")
	}
}

func TestCompareSpecsBothEmpty(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{}
	new := &parser.SpecDocument{}
	diff := CompareSpecs(old, new)
	if diff.Project != nil {
		t.Error("expected nil project diff for empty docs")
	}
	if len(diff.Modules) != 0 {
		t.Errorf("expected 0 module diffs, got %d", len(diff.Modules))
	}
	if len(diff.Services) != 0 {
		t.Errorf("expected 0 service diffs, got %d", len(diff.Services))
	}
}

func TestCompareSpecsEmptyModulesToPopulated(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{}
	new := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "a"}, {Name: "b"}},
	}
	diff := CompareSpecs(old, new)
	if len(diff.Modules) != 2 {
		t.Errorf("expected 2 module diffs, got %d", len(diff.Modules))
	}
	for _, m := range diff.Modules {
		if m.Type != ChangeAdded {
			t.Errorf("expected ChangeAdded for %s, got %s", m.Name, m.Type)
		}
	}
}

func TestCompareSpecsPopulatedToEmptyModules(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "a"}},
	}
	new := &parser.SpecDocument{}
	diff := CompareSpecs(old, new)
	if len(diff.Modules) != 1 {
		t.Errorf("expected 1 module diff, got %d", len(diff.Modules))
	}
	if diff.Modules[0].Type != ChangeRemoved {
		t.Errorf("expected ChangeRemoved, got %s", diff.Modules[0].Type)
	}
}

func TestCompareSpecsSummaryModulesOnly(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "a"}},
	}
	new := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "a"}, {Name: "b"}},
	}
	diff := CompareSpecs(old, new)
	if !strings.Contains(diff.Summary, "Modules:") {
		t.Error("summary should contain Modules section")
	}
}

func TestCompareSpecsSummaryServicesOnly(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Services: []parser.Service{{Name: "s1"}},
	}
	new := &parser.SpecDocument{
		Services: []parser.Service{{Name: "s1"}, {Name: "s2"}},
	}
	diff := CompareSpecs(old, new)
	if !strings.Contains(diff.Summary, "Services:") {
		t.Error("summary should contain Services section")
	}
}

func TestFormatSpecDiffModuleRemoved(t *testing.T) {
	t.Parallel()
	old := &parser.SpecDocument{
		Modules: []parser.Module{{Name: "auth", Path: "./auth"}},
	}
	new := &parser.SpecDocument{}
	diff := CompareSpecs(old, new)
	formatted := FormatSpecDiff(diff)
	if !strings.Contains(formatted, "auth") {
		t.Error("formatted diff should contain removed module name")
	}
}
