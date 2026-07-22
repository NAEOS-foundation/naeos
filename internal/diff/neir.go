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
	OldValue any
	NewValue any
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

	if epChanges := diffEndpoints(old.Endpoints, new.Endpoints); len(epChanges) > 0 {
		changes = append(changes, epChanges...)
	}
	if mwChanges := diffMiddleware(old.Middleware, new.Middleware); len(mwChanges) > 0 {
		changes = append(changes, mwChanges...)
	}
	if attrChanges := diffAttributes(old.Attributes, new.Attributes); len(attrChanges) > 0 {
		changes = append(changes, attrChanges...)
	}

	return changes
}

func diffEndpoints(old, new []service.Endpoint) []FieldChange {
	var changes []FieldChange
	oldMap := make(map[string]service.Endpoint)
	for _, ep := range old {
		key := ep.Method + ":" + ep.Path
		oldMap[key] = ep
	}
	newMap := make(map[string]service.Endpoint)
	for _, ep := range new {
		key := ep.Method + ":" + ep.Path
		newMap[key] = ep
	}
	for key, oldEp := range oldMap {
		if newEp, exists := newMap[key]; exists {
			if oldEp.Action != newEp.Action {
				changes = append(changes, FieldChange{
					Field:    fmt.Sprintf("endpoint[%s].action", key),
					OldValue: oldEp.Action,
					NewValue: newEp.Action,
				})
			}
		} else {
			changes = append(changes, FieldChange{
				Field:    fmt.Sprintf("endpoint[%s]", key),
				OldValue: "present",
				NewValue: "removed",
			})
		}
	}
	for key := range newMap {
		if _, exists := oldMap[key]; !exists {
			changes = append(changes, FieldChange{
				Field:    fmt.Sprintf("endpoint[%s]", key),
				OldValue: "absent",
				NewValue: "added",
			})
		}
	}
	return changes
}

func diffMiddleware(old, new []string) []FieldChange {
	var changes []FieldChange
	oldSet := make(map[string]bool)
	for _, mw := range old {
		oldSet[mw] = true
	}
	newSet := make(map[string]bool)
	for _, mw := range new {
		newSet[mw] = true
	}
	for _, mw := range old {
		if !newSet[mw] {
			changes = append(changes, FieldChange{
				Field:    fmt.Sprintf("middleware[%s]", mw),
				OldValue: "present",
				NewValue: "removed",
			})
		}
	}
	for _, mw := range new {
		if !oldSet[mw] {
			changes = append(changes, FieldChange{
				Field:    fmt.Sprintf("middleware[%s]", mw),
				OldValue: "absent",
				NewValue: "added",
			})
		}
	}
	return changes
}

func diffAttributes(old, new map[string]string) []FieldChange {
	var changes []FieldChange
	if len(old) != len(new) {
		changes = append(changes, FieldChange{
			Field:    "attributes",
			OldValue: len(old),
			NewValue: len(new),
		})
	}
	for key, oldVal := range old {
		if newVal, exists := new[key]; !exists {
			changes = append(changes, FieldChange{
				Field:    fmt.Sprintf("attributes[%s]", key),
				OldValue: oldVal,
				NewValue: "removed",
			})
		} else if oldVal != newVal {
			changes = append(changes, FieldChange{
				Field:    fmt.Sprintf("attributes[%s]", key),
				OldValue: oldVal,
				NewValue: newVal,
			})
		}
	}
	for key, newVal := range new {
		if _, exists := old[key]; !exists {
			changes = append(changes, FieldChange{
				Field:    fmt.Sprintf("attributes[%s]", key),
				OldValue: "absent",
				NewValue: newVal,
			})
		}
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

	fmt.Fprintf(&sb, "NEIR Diff: %s\n", diff.Summary)
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	if diff.ProjectDiff != nil && len(diff.ProjectDiff.FieldsModified) > 0 {
		sb.WriteString("Project:\n")
		if diff.ProjectDiff.NameChanged {
			if diff.ProjectDiff.OldName != "" && diff.ProjectDiff.NewName != "" {
				fmt.Fprintf(&sb, "  \033[31m-%s\033[0m\n", diff.ProjectDiff.OldName)
				fmt.Fprintf(&sb, "  \033[32m+%s\033[0m\n", diff.ProjectDiff.NewName)
			}
		}
		sb.WriteString("\n")
	}

	if diff.ServicesDiff != nil {
		sd := diff.ServicesDiff
		if len(sd.Added) > 0 {
			fmt.Fprintf(&sb, "\033[32mAdded services (%d):\033[0m\n", len(sd.Added))
			for _, s := range sd.Added {
				fmt.Fprintf(&sb, "  \033[32m+ %s (port=%d)\033[0m\n", s.Name, s.Port)
			}
			sb.WriteString("\n")
		}
		if len(sd.Removed) > 0 {
			fmt.Fprintf(&sb, "\033[31mRemoved services (%d):\033[0m\n", len(sd.Removed))
			for _, s := range sd.Removed {
				fmt.Fprintf(&sb, "  \033[31m- %s (port=%d)\033[0m\n", s.Name, s.Port)
			}
			sb.WriteString("\n")
		}
		if len(sd.Modified) > 0 {
			fmt.Fprintf(&sb, "\033[33mModified services (%d):\033[0m\n", len(sd.Modified))
			for _, m := range sd.Modified {
				fmt.Fprintf(&sb, "  \033[33m~ %s:\033[0m\n", m.Name)
				for _, c := range m.Changes {
					fmt.Fprintf(&sb, "    %s: %v -> %v\n", c.Field, c.OldValue, c.NewValue)
				}
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// RenderVisualDiff generates an HTML page with a visual side-by-side comparison of two NEIR specs.
// The output is a self-contained HTML page with inline CSS.
func RenderVisualDiff(old, new *model.NEIR) string {
	diff := ComputeNEIRDiff(old, new)

	var addedRows, removedRows, modifiedRows string

	if diff.ServicesDiff != nil {
		for _, s := range diff.ServicesDiff.Added {
			addedRows += fmt.Sprintf(`<tr class="added"><td>%s</td><td>%d</td><td>%s</td></tr>`, s.Name, s.Port, s.Kind)
		}
		for _, s := range diff.ServicesDiff.Removed {
			removedRows += fmt.Sprintf(`<tr class="removed"><td>%s</td><td>%d</td><td>%s</td></tr>`, s.Name, s.Port, s.Kind)
		}
		for _, m := range diff.ServicesDiff.Modified {
			var changes string
			for _, c := range m.Changes {
				changes += fmt.Sprintf(`<li><strong>%s</strong>: <span class="old">%v</span> → <span class="new">%v</span></li>`, c.Field, c.OldValue, c.NewValue)
			}
			modifiedRows += fmt.Sprintf(`<tr class="modified"><td>%s</td><td colspan="2"><ul>%s</ul></td></tr>`, m.Name, changes)
		}
	}

	projectInfo := ""
	if diff.ProjectDiff != nil {
		if diff.ProjectDiff.NameChanged {
			projectInfo = fmt.Sprintf(`<div class="project-name"><span class="old">%s</span> → <span class="new">%s</span></div>`, diff.ProjectDiff.OldName, diff.ProjectDiff.NewName)
		}
		if len(diff.ProjectDiff.FieldsModified) > 0 {
			projectInfo += `<div class="project-fields"><strong>Modified fields:</strong> ` + strings.Join(diff.ProjectDiff.FieldsModified, ", ") + `</div>`
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>NEIR Architecture Diff</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; color: #333; }
  h1 { font-size: 1.4em; margin-bottom: 5px; }
  .summary { color: #666; font-size: 0.9em; margin-bottom: 20px; }
  .project-name { margin: 10px 0; padding: 8px; background: #fff; border-left: 3px solid #2196F3; border-radius: 3px; }
  .project-fields { margin: 5px 0 15px 0; font-size: 0.9em; color: #555; }
  .old { color: #d32f2f; text-decoration: line-through; }
  .new { color: #388e3c; font-weight: bold; }
  table { width: 100%%; border-collapse: collapse; margin-bottom: 20px; background: #fff; box-shadow: 0 1px 3px rgba(0,0,0,0.1); border-radius: 4px; overflow: hidden; }
  th { background: #f0f0f0; padding: 8px 12px; text-align: left; font-weight: 600; border-bottom: 2px solid #ddd; }
  td { padding: 8px 12px; border-bottom: 1px solid #eee; }
  .added td { background: #e8f5e9; }
  .added td:first-child::before { content: "+ "; color: #388e3c; font-weight: bold; }
  .removed td { background: #ffebee; }
  .removed td:first-child::before { content: "- "; color: #d32f2f; font-weight: bold; }
  .modified td { background: #fff8e1; }
  .modified td:first-child::before { content: "~ "; color: #f57c00; font-weight: bold; }
  ul { margin: 0; padding-left: 16px; }
  .section-title { font-weight: 600; margin: 15px 0 5px 0; padding: 4px 8px; border-radius: 3px; display: inline-block; }
  .section-title.added-title { background: #e8f5e9; color: #388e3c; }
  .section-title.removed-title { background: #ffebee; color: #d32f2f; }
  .section-title.modified-title { background: #fff8e1; color: #f57c00; }
  .graph { display: flex; gap: 10px; margin: 15px 0; flex-wrap: wrap; }
  .graph-node { padding: 8px 12px; border-radius: 4px; font-size: 0.85em; border: 1px solid #ddd; background: #fff; }
  .graph-node.added { border-color: #388e3c; background: #e8f5e9; }
  .graph-node.removed { border-color: #d32f2f; background: #ffebee; }
  .graph-node.modified { border-color: #f57c00; background: #fff8e1; }
  .graph-arrow { color: #999; font-size: 1.2em; align-self: center; }
</style>
</head><body>
<h1>NEIR Architecture Diff</h1>
<div class="summary">%s</div>
%s
<div class="graph">
%s
</div>
<div class="section-title added-title">Added (%d)</div>
<table><tr><th>Service</th><th>Port</th><th>Type</th></tr>%s</table>
<div class="section-title removed-title">Removed (%d)</div>
<table><tr><th>Service</th><th>Port</th><th>Type</th></tr>%s</table>
<div class="section-title modified-title">Modified (%d)</div>
<table><tr><th>Service</th><th>Changes</th></tr>%s</table>
</body></html>`,
		diff.Summary,
		projectInfo,
		renderGraphNodes(diff),
		len(diff.ServicesDiff.Added), addedRows,
		len(diff.ServicesDiff.Removed), removedRows,
		len(diff.ServicesDiff.Modified), modifiedRows,
	)
}

func renderGraphNodes(diff *NEIRDiff) string {
	if diff.ServicesDiff == nil {
		return ""
	}
	var nodes []string
	for _, s := range diff.ServicesDiff.Added {
		nodes = append(nodes, fmt.Sprintf(`<div class="graph-node added">+ %s</div>`, s.Name))
		nodes = append(nodes, `<div class="graph-arrow">→</div>`)
	}
	for _, s := range diff.ServicesDiff.Removed {
		nodes = append(nodes, fmt.Sprintf(`<div class="graph-node removed">- %s</div>`, s.Name))
		nodes = append(nodes, `<div class="graph-arrow">→</div>`)
	}
	for _, m := range diff.ServicesDiff.Modified {
		nodes = append(nodes, fmt.Sprintf(`<div class="graph-node modified">~ %s</div>`, m.Name))
		nodes = append(nodes, `<div class="graph-arrow">→</div>`)
	}
	if len(nodes) > 0 {
		nodes = nodes[:len(nodes)-1]
	}
	return strings.Join(nodes, "\n")
}
