// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

type SpecDiff struct {
	Project     *FieldDiff
	Modules     []ModuleDiff
	Services    []ServiceDiff
	ArchDiff    *FieldDiff
	DeployDiff  *FieldDiff
	TestingDiff *FieldDiff
	Summary     string
}

type FieldDiff struct {
	OldValue any
	NewValue any
	Type     ChangeType
}

type ModuleDiff struct {
	Name      string
	Type      ChangeType
	OldModule *parser.Module
	NewModule *parser.Module
}

type ServiceDiff struct {
	Name       string
	Type       ChangeType
	OldService *parser.Service
	NewService *parser.Service
}

func CompareSpecs(oldDoc, newDoc *parser.SpecDocument) *SpecDiff {
	diff := &SpecDiff{}

	if oldDoc.Project != newDoc.Project {
		diff.Project = &FieldDiff{
			OldValue: oldDoc.Project,
			NewValue: newDoc.Project,
			Type:     ChangeModified,
		}
	}

	diff.Modules = compareModules(oldDoc.Modules, newDoc.Modules)
	diff.Services = compareServices(oldDoc.Services, newDoc.Services)

	diff.Summary = formatSpecDiffSummary(diff)
	return diff
}

func compareModules(oldMods, newMods []parser.Module) []ModuleDiff {
	var diffs []ModuleDiff

	oldMap := make(map[string]*parser.Module)
	for i := range oldMods {
		oldMap[oldMods[i].Name] = &oldMods[i]
	}

	newMap := make(map[string]*parser.Module)
	for i := range newMods {
		newMap[newMods[i].Name] = &newMods[i]
	}

	for name, oldMod := range oldMap {
		if newMod, exists := newMap[name]; exists {
			if modulesEqual(oldMod, newMod) {
				diffs = append(diffs, ModuleDiff{
					Name:      name,
					Type:      ChangeUnchanged,
					OldModule: oldMod,
					NewModule: newMod,
				})
			} else {
				diffs = append(diffs, ModuleDiff{
					Name:      name,
					Type:      ChangeModified,
					OldModule: oldMod,
					NewModule: newMod,
				})
			}
		} else {
			diffs = append(diffs, ModuleDiff{
				Name:      name,
				Type:      ChangeRemoved,
				OldModule: oldMod,
			})
		}
	}

	for name, newMod := range newMap {
		if _, exists := oldMap[name]; !exists {
			diffs = append(diffs, ModuleDiff{
				Name:      name,
				Type:      ChangeAdded,
				NewModule: newMod,
			})
		}
	}

	return diffs
}

func compareServices(oldSvcs, newSvcs []parser.Service) []ServiceDiff {
	var diffs []ServiceDiff

	oldMap := make(map[string]*parser.Service)
	for i := range oldSvcs {
		oldMap[oldSvcs[i].Name] = &oldSvcs[i]
	}

	newMap := make(map[string]*parser.Service)
	for i := range newSvcs {
		newMap[newSvcs[i].Name] = &newSvcs[i]
	}

	for name, oldSvc := range oldMap {
		if newSvc, exists := newMap[name]; exists {
			if servicesEqual(oldSvc, newSvc) {
				diffs = append(diffs, ServiceDiff{
					Name:       name,
					Type:       ChangeUnchanged,
					OldService: oldSvc,
					NewService: newSvc,
				})
			} else {
				diffs = append(diffs, ServiceDiff{
					Name:       name,
					Type:       ChangeModified,
					OldService: oldSvc,
					NewService: newSvc,
				})
			}
		} else {
			diffs = append(diffs, ServiceDiff{
				Name:       name,
				Type:       ChangeRemoved,
				OldService: oldSvc,
			})
		}
	}

	for name, newSvc := range newMap {
		if _, exists := oldMap[name]; !exists {
			diffs = append(diffs, ServiceDiff{
				Name:       name,
				Type:       ChangeAdded,
				NewService: newSvc,
			})
		}
	}

	return diffs
}

func modulesEqual(a, b *parser.Module) bool {
	if a.Name != b.Name || a.Path != b.Path || a.Description != b.Description {
		return false
	}
	if len(a.Dependencies) != len(b.Dependencies) {
		return false
	}
	for i := range a.Dependencies {
		if a.Dependencies[i] != b.Dependencies[i] {
			return false
		}
	}
	return true
}

func servicesEqual(a, b *parser.Service) bool {
	if a.Name != b.Name || a.Kind != b.Kind || a.Port != b.Port || a.Description != b.Description {
		return false
	}
	if len(a.Endpoints) != len(b.Endpoints) {
		return false
	}
	for i := range a.Endpoints {
		if a.Endpoints[i].Method != b.Endpoints[i].Method ||
			a.Endpoints[i].Path != b.Endpoints[i].Path ||
			a.Endpoints[i].Action != b.Endpoints[i].Action {
			return false
		}
	}
	return true
}

func formatSpecDiffSummary(diff *SpecDiff) string {
	var parts []string

	if diff.Project != nil {
		parts = append(parts, fmt.Sprintf("Project: %v → %v", diff.Project.OldValue, diff.Project.NewValue))
	}

	added, removed, modified := 0, 0, 0
	for _, m := range diff.Modules {
		switch m.Type {
		case ChangeAdded:
			added++
		case ChangeRemoved:
			removed++
		case ChangeModified:
			modified++
		}
	}

	if added+removed+modified > 0 {
		parts = append(parts, fmt.Sprintf("Modules: +%d, -%d, ~%d", added, removed, modified))
	}

	added, removed, modified = 0, 0, 0
	for _, s := range diff.Services {
		switch s.Type {
		case ChangeAdded:
			added++
		case ChangeRemoved:
			removed++
		case ChangeModified:
			modified++
		}
	}

	if added+removed+modified > 0 {
		parts = append(parts, fmt.Sprintf("Services: +%d, -%d, ~%d", added, removed, modified))
	}

	return strings.Join(parts, "; ")
}

func FormatSpecDiff(diff *SpecDiff) string {
	var sb strings.Builder

	if diff.Project != nil {
		fmt.Fprintf(&sb, "Project: %v → %v\n", diff.Project.OldValue, diff.Project.NewValue)
	}

	for _, m := range diff.Modules {
		switch m.Type {
		case ChangeAdded:
			fmt.Fprintf(&sb, "\033[32m+ Module: %s (%s)\033[0m\n", m.Name, m.NewModule.Path)
		case ChangeRemoved:
			fmt.Fprintf(&sb, "\033[31m- Module: %s (%s)\033[0m\n", m.Name, m.OldModule.Path)
		case ChangeModified:
			fmt.Fprintf(&sb, "\033[33m~ Module: %s\033[0m\n", m.Name)
			if m.OldModule.Path != m.NewModule.Path {
				fmt.Fprintf(&sb, "  Path: %s → %s\n", m.OldModule.Path, m.NewModule.Path)
			}
			if m.OldModule.Description != m.NewModule.Description {
				fmt.Fprintf(&sb, "  Description: %s → %s\n", m.OldModule.Description, m.NewModule.Description)
			}
		}
	}

	for _, s := range diff.Services {
		switch s.Type {
		case ChangeAdded:
			fmt.Fprintf(&sb, "\033[32m+ Service: %s (kind=%s, port=%d)\033[0m\n", s.Name, s.NewService.Kind, s.NewService.Port)
		case ChangeRemoved:
			fmt.Fprintf(&sb, "\033[31m- Service: %s (kind=%s, port=%d)\033[0m\n", s.Name, s.OldService.Kind, s.OldService.Port)
		case ChangeModified:
			fmt.Fprintf(&sb, "\033[33m~ Service: %s\033[0m\n", s.Name)
			if s.OldService.Port != s.NewService.Port {
				fmt.Fprintf(&sb, "  Port: %d → %d\n", s.OldService.Port, s.NewService.Port)
			}
			if s.OldService.Kind != s.NewService.Kind {
				fmt.Fprintf(&sb, "  Kind: %s → %s\n", s.OldService.Kind, s.NewService.Kind)
			}
		}
	}

	return sb.String()
}
