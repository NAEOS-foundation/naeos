package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
)

type NEIRDiff struct {
	ProjectDiff  *ProjectDiff
	ServicesDiff *ServicesDiff
	Summary      string
}

type ProjectDiff struct {
	NameChanged    bool
	OldName        string
	NewName        string
	FieldsModified []string
}

type ServicesDiff struct {
	Added    []service.Service
	Removed  []service.Service
	Modified []ServiceModification
}

type ServiceModification struct {
	Name    string
	Changes []FieldChange
}

type FieldChange struct {
	Field    string
	OldValue interface{}
	NewValue interface{}
}

func ComputeNEIRDiff(old, new *model.NEIR) *NEIRDiff {
	diff := &NEIRDiff{}

	if old == nil && new == nil {
		return diff
	}

	if old == nil {
		diff.ProjectDiff = &ProjectDiff{
			NameChanged: true,
			NewName:     new.Project.Name,
		}
		if new.Services != nil {
			diff.ServicesDiff = &ServicesDiff{Added: new.Services}
		}
		diff.Summary = fmt.Sprintf("new spec: %d services", len(new.Services))
		return diff
	}

	if new == nil {
		diff.ProjectDiff = &ProjectDiff{
			NameChanged: true,
			OldName:     old.Project.Name,
		}
		if old.Services != nil {
			diff.ServicesDiff = &ServicesDiff{Removed: old.Services}
		}
		diff.Summary = fmt.Sprintf("spec removed: %d services", len(old.Services))
		return diff
	}

	diff.ProjectDiff = diffProject(old, new)
	diff.ServicesDiff = diffServices(old.Services, new.Services)
	diff.Summary = buildSummary(diff.ProjectDiff, diff.ServicesDiff)

	return diff
}

func diffProject(old, new *model.NEIR) *ProjectDiff {
	pd := &ProjectDiff{}

	if old.Project == nil && new.Project == nil {
		return pd
	}
	if old.Project == nil {
		pd.NameChanged = true
		pd.NewName = new.Project.Name
		pd.FieldsModified = append(pd.FieldsModified, "project")
		return pd
	}
	if new.Project == nil {
		pd.NameChanged = true
		pd.OldName = old.Project.Name
		pd.FieldsModified = append(pd.FieldsModified, "project")
		return pd
	}

	if old.Project.Name != new.Project.Name {
		pd.NameChanged = true
		pd.OldName = old.Project.Name
		pd.NewName = new.Project.Name
		pd.FieldsModified = append(pd.FieldsModified, "name")
	}

	if old.Project.Version != new.Project.Version {
		pd.FieldsModified = append(pd.FieldsModified, "version")
	}

	return pd
}

func diffServices(oldServices, newServices []service.Service) *ServicesDiff {
	sd := &ServicesDiff{}

	oldMap := make(map[string]service.Service)
	for _, s := range oldServices {
		oldMap[s.Name] = s
	}
	newMap := make(map[string]service.Service)
	for _, s := range newServices {
		newMap[s.Name] = s
	}

	for name, s := range newMap {
		if _, exists := oldMap[name]; !exists {
			sd.Added = append(sd.Added, s)
		}
	}

	for name, s := range oldMap {
		if _, exists := newMap[name]; !exists {
			sd.Removed = append(sd.Removed, s)
		}
	}

	for name, oldSvc := range oldMap {
		newSvc, exists := newMap[name]
		if !exists {
			continue
		}
		changes := diffServiceFields(oldSvc, newSvc)
		if len(changes) > 0 {
			sd.Modified = append(sd.Modified, ServiceModification{
				Name:    name,
				Changes: changes,
			})
		}
	}

	sort.Slice(sd.Added, func(i, j int) bool { return sd.Added[i].Name < sd.Added[j].Name })
	sort.Slice(sd.Removed, func(i, j int) bool { return sd.Removed[i].Name < sd.Removed[j].Name })
	sort.Slice(sd.Modified, func(i, j int) bool { return sd.Modified[i].Name < sd.Modified[j].Name })

	return sd
}

func diffServiceFields(old, new service.Service) []FieldChange {
	var changes []FieldChange

	if old.Port != new.Port {
		changes = append(changes, FieldChange{Field: "port", OldValue: old.Port, NewValue: new.Port})
	}
	if old.Description != new.Description {
		changes = append(changes, FieldChange{Field: "description", OldValue: old.Description, NewValue: new.Description})
	}
	if old.Kind != new.Kind {
		changes = append(changes, FieldChange{Field: "kind", OldValue: old.Kind, NewValue: new.Kind})
	}

	oldEndp := fmt.Sprintf("%v", old.Endpoints)
	newEndp := fmt.Sprintf("%v", new.Endpoints)
	if oldEndp != newEndp {
		changes = append(changes, FieldChange{Field: "endpoints", OldValue: oldEndp, NewValue: newEndp})
	}

	oldMid := fmt.Sprintf("%v", old.Middleware)
	newMid := fmt.Sprintf("%v", new.Middleware)
	if oldMid != newMid {
		changes = append(changes, FieldChange{Field: "middleware", OldValue: oldMid, NewValue: newMid})
	}

	oldAttr := fmt.Sprintf("%v", old.Attributes)
	newAttr := fmt.Sprintf("%v", new.Attributes)
	if oldAttr != newAttr {
		changes = append(changes, FieldChange{Field: "attributes", OldValue: oldAttr, NewValue: newAttr})
	}

	return changes
}

func buildSummary(pd *ProjectDiff, sd *ServicesDiff) string {
	var parts []string

	if pd != nil && pd.NameChanged {
		if pd.OldName != "" && pd.NewName != "" {
			parts = append(parts, fmt.Sprintf("project %s -> %s", pd.OldName, pd.NewName))
		} else if pd.NewName != "" {
			parts = append(parts, fmt.Sprintf("project added: %s", pd.NewName))
		} else {
			parts = append(parts, "project removed")
		}
	}

	if sd != nil {
		if len(sd.Added) > 0 {
			names := make([]string, len(sd.Added))
			for i, s := range sd.Added {
				names[i] = s.Name
			}
			parts = append(parts, fmt.Sprintf("+%d services (%s)", len(sd.Added), strings.Join(names, ", ")))
		}
		if len(sd.Removed) > 0 {
			names := make([]string, len(sd.Removed))
			for i, s := range sd.Removed {
				names[i] = s.Name
			}
			parts = append(parts, fmt.Sprintf("-%d services (%s)", len(sd.Removed), strings.Join(names, ", ")))
		}
		if len(sd.Modified) > 0 {
			names := make([]string, len(sd.Modified))
			for i, m := range sd.Modified {
				names[i] = m.Name
			}
			parts = append(parts, fmt.Sprintf("~%d services modified (%s)", len(sd.Modified), strings.Join(names, ", ")))
		}
	}

	if len(parts) == 0 {
		return "no changes"
	}
	return strings.Join(parts, "; ")
}

func FormatNEIRDiff(diff *NEIRDiff) string {
	if diff == nil {
		return ""
	}
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("NEIR Diff: %s\n", diff.Summary))
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	if diff.ProjectDiff != nil && len(diff.ProjectDiff.FieldsModified) > 0 {
		sb.WriteString("Project:\n")
		if diff.ProjectDiff.NameChanged {
			if diff.ProjectDiff.OldName != "" && diff.ProjectDiff.NewName != "" {
				sb.WriteString(fmt.Sprintf("  \033[31m-%s\033[0m\n", diff.ProjectDiff.OldName))
				sb.WriteString(fmt.Sprintf("  \033[32m+%s\033[0m\n", diff.ProjectDiff.NewName))
			}
		}
		sb.WriteString("\n")
	}

	if diff.ServicesDiff != nil {
		sd := diff.ServicesDiff
		if len(sd.Added) > 0 {
			sb.WriteString(fmt.Sprintf("\033[32mAdded services (%d):\033[0m\n", len(sd.Added)))
			for _, s := range sd.Added {
				sb.WriteString(fmt.Sprintf("  \033[32m+ %s (port=%d)\033[0m\n", s.Name, s.Port))
			}
			sb.WriteString("\n")
		}
		if len(sd.Removed) > 0 {
			sb.WriteString(fmt.Sprintf("\033[31mRemoved services (%d):\033[0m\n", len(sd.Removed)))
			for _, s := range sd.Removed {
				sb.WriteString(fmt.Sprintf("  \033[31m- %s (port=%d)\033[0m\n", s.Name, s.Port))
			}
			sb.WriteString("\n")
		}
		if len(sd.Modified) > 0 {
			sb.WriteString(fmt.Sprintf("\033[33mModified services (%d):\033[0m\n", len(sd.Modified)))
			for _, m := range sd.Modified {
				sb.WriteString(fmt.Sprintf("  \033[33m~ %s:\033[0m\n", m.Name))
				for _, c := range m.Changes {
					sb.WriteString(fmt.Sprintf("    %s: %v -> %v\n", c.Field, c.OldValue, c.NewValue))
				}
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
