package migration

import (
	"fmt"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

type TransformStep struct {
	FromVersion string
	ToVersion   string
	Description string
	Transform   func(data map[string]any) (map[string]any, error)
}

type MigrationEngine struct {
	steps map[string]TransformStep
}

func NewMigrationEngine() *MigrationEngine {
	e := &MigrationEngine{
		steps: make(map[string]TransformStep),
	}
	e.registerBuiltinSteps()
	return e
}

func (e *MigrationEngine) Register(step TransformStep) {
	key := fmt.Sprintf("%s->%s", step.FromVersion, step.ToVersion)
	e.steps[key] = step
}

func (e *MigrationEngine) Migrate(data map[string]any, fromVersion, toVersion string) (map[string]any, error) {
	current := fromVersion
	result := make(map[string]any)
	for k, v := range data {
		result[k] = v
	}

	for current != toVersion {
		next := e.nextVersion(current)
		if next == "" {
			return nil, fmt.Errorf("no migration path from %s to %s", current, toVersion)
		}

		key := fmt.Sprintf("%s->%s", current, next)
		step, ok := e.steps[key]
		if !ok {
			return nil, fmt.Errorf("missing migration step: %s", key)
		}

		var err error
		result, err = step.Transform(result)
		if err != nil {
			return nil, fmt.Errorf("migration %s failed: %w", key, err)
		}

		current = next
	}

	return result, nil
}

func (e *MigrationEngine) nextVersion(current string) string {
	versions := []string{"0.1.0", "0.2.0", "0.3.0"}
	for i, v := range versions {
		if v == current && i+1 < len(versions) {
			return versions[i+1]
		}
	}
	return ""
}

func (e *MigrationEngine) Plan(fromVersion, toVersion string) []TransformStep {
	var plan []TransformStep
	current := fromVersion
	for current != toVersion {
		next := e.nextVersion(current)
		if next == "" {
			break
		}
		key := fmt.Sprintf("%s->%s", current, next)
		if step, ok := e.steps[key]; ok {
			plan = append(plan, step)
		}
		current = next
	}
	return plan
}

func (e *MigrationEngine) registerBuiltinSteps() {
	e.Register(TransformStep{
		FromVersion: "0.1.0",
		ToVersion:   "0.2.0",
		Description: "Add generation config and normalize module structure",
		Transform: func(data map[string]any) (map[string]any, error) {
			if _, ok := data["version"]; !ok {
				data["version"] = "0.2.0"
			}

			if modules, ok := data["modules"].([]any); ok {
				for i, raw := range modules {
					if modMap, ok := raw.(map[string]any); ok {
						if _, ok := modMap["path"]; !ok {
							if name, ok := modMap["name"].(string); ok {
								modMap["path"] = fmt.Sprintf("./%s", parser.Slugify(name))
							}
						}
						if _, ok := modMap["dependencies"]; !ok {
							modMap["dependencies"] = []any{}
						}
					}
					modules[i] = raw
				}
				data["modules"] = modules
			}

			if _, ok := data["generation"]; !ok {
				data["generation"] = map[string]any{
					"languages": []any{"go"},
				}
			}

			return data, nil
		},
	})

	e.Register(TransformStep{
		FromVersion: "0.2.0",
		ToVersion:   "0.3.0",
		Description: "Add architecture defaults, security context, and normalize services",
		Transform: func(data map[string]any) (map[string]any, error) {
			data["version"] = "0.3.0"

			if _, ok := data["architecture"]; !ok {
				data["architecture"] = map[string]any{
					"pattern":     "hexagonal",
					"principles": []any{"loose-coupling", "high-cohesion"},
				}
			}

			if services, ok := data["services"].([]any); ok {
				for i, raw := range services {
					if svcMap, ok := raw.(map[string]any); ok {
						if _, ok := svcMap["kind"]; !ok {
							svcMap["kind"] = "http"
						}
						if _, ok := svcMap["endpoints"]; !ok {
							svcMap["endpoints"] = []any{}
						}
					}
					services[i] = raw
				}
				data["services"] = services
			}

			if _, ok := data["security"]; !ok {
				data["security"] = map[string]any{
					"audit_logging": true,
					"encryption":    "tls",
				}
			}

			if _, ok := data["testing"]; !ok {
				data["testing"] = map[string]any{
					"strategy": "unit",
					"coverage": "80",
				}
			}

			return data, nil
		},
	})
}

func (e *MigrationEngine) AvailableVersions() []string {
	return []string{"0.1.0", "0.2.0", "0.3.0"}
}

func (e *MigrationEngine) VersionBetween(from, to string) bool {
	versions := e.AvailableVersions()
	fromIdx, toIdx := -1, -1
	for i, v := range versions {
		if v == from {
			fromIdx = i
		}
		if v == to {
			toIdx = i
		}
	}
	return fromIdx >= 0 && toIdx >= 0 && fromIdx < toIdx
}

func FormatMigrationPlan(plan []TransformStep) string {
	if len(plan) == 0 {
		return "No migrations needed."
	}

	var sb strings.Builder
	sb.WriteString("Migration Plan:\n")
	for i, step := range plan {
		sb.WriteString(fmt.Sprintf("  %d. %s → %s: %s\n", i+1, step.FromVersion, step.ToVersion, step.Description))
	}
	return sb.String()
}
