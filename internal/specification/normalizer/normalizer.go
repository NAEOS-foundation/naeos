package normalizer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

const normalizerVersion = "1.0.0"

type Normalizer interface {
	Normalize(doc any) (*NormalizedSpec, error)
}

type NormalizedSpec struct {
	Values map[string]any
}

type DefaultNormalizer struct{}

type Difference struct {
	Path string
	Type string
	Old  any
	New  any
}

func NewNormalizer() Normalizer {
	return DefaultNormalizer{}
}

func Version() string {
	return normalizerVersion
}

func (DefaultNormalizer) Normalize(doc any) (*NormalizedSpec, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	specDoc, ok := doc.(*parser.SpecDocument)
	if !ok {
		return &NormalizedSpec{Values: map[string]any{"source": doc}}, nil
	}

	modules := normalizeModules(specDoc.Modules)
	services := normalizeServices(specDoc.Services)

	result := map[string]any{
		"project":  specDoc.Project,
		"modules":  modules,
		"services": services,
		"source":   specDoc,
	}

	if specDoc.Architecture != nil {
		result["architecture"] = normalizeArchitecture(specDoc.Architecture)
	}
	if specDoc.Deployment != nil {
		result["deployment"] = normalizeDeployment(specDoc.Deployment)
	}
	if specDoc.Testing != nil {
		result["testing"] = normalizeTesting(specDoc.Testing)
	}
	if specDoc.Generation != nil {
		result["generation"] = normalizeGeneration(specDoc.Generation)
	}

	return &NormalizedSpec{Values: result}, nil
}

func NormalizeRaw(data map[string]any) (*NormalizedSpec, error) {
	if data == nil {
		return nil, fmt.Errorf("input data is nil")
	}

	result := make(map[string]any, len(data))
	for k, v := range data {
		result[k] = v
	}

	if project, ok := data["project"]; ok {
		result["project"] = normalizeRawString(project)
	}

	if modules, ok := data["modules"]; ok {
		normalized, err := normalizeRawModules(modules)
		if err != nil {
			return nil, fmt.Errorf("normalizing modules: %w", err)
		}
		result["modules"] = normalized
	}

	if services, ok := data["services"]; ok {
		normalized, err := normalizeRawServices(services)
		if err != nil {
			return nil, fmt.Errorf("normalizing services: %w", err)
		}
		result["services"] = normalized
	}

	if arch, ok := data["architecture"]; ok {
		normalized, err := normalizeRawMap(arch)
		if err != nil {
			return nil, fmt.Errorf("normalizing architecture: %w", err)
		}
		result["architecture"] = normalized
	}

	if deploy, ok := data["deployment"]; ok {
		normalized, err := normalizeRawMap(deploy)
		if err != nil {
			return nil, fmt.Errorf("normalizing deployment: %w", err)
		}
		result["deployment"] = normalized
	}

	if testing, ok := data["testing"]; ok {
		normalized, err := normalizeRawMap(testing)
		if err != nil {
			return nil, fmt.Errorf("normalizing testing: %w", err)
		}
		result["testing"] = normalized
	}

	if gen, ok := data["generation"]; ok {
		normalized, err := normalizeRawMap(gen)
		if err != nil {
			return nil, fmt.Errorf("normalizing generation: %w", err)
		}
		result["generation"] = normalized
	}

	return &NormalizedSpec{Values: result}, nil
}

func normalizeRawString(v any) string {
	s, _ := v.(string)
	return s
}

func normalizeRawModules(v any) (any, error) {
	switch m := v.(type) {
	case []any:
		result := make([]map[string]any, 0, len(m))
		for _, item := range m {
			entry, err := normalizeRawMap(item)
			if err != nil {
				return nil, err
			}
			result = append(result, entry)
		}
		return result, nil
	case []map[string]any:
		result := make([]map[string]any, 0, len(m))
		for _, entry := range m {
			result = append(result, entry)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("expected array of modules, got %T", v)
	}
}

func normalizeRawServices(v any) (any, error) {
	switch s := v.(type) {
	case []any:
		result := make([]map[string]any, 0, len(s))
		for _, item := range s {
			entry, err := normalizeRawMap(item)
			if err != nil {
				return nil, err
			}
			result = append(result, entry)
		}
		return result, nil
	case []map[string]any:
		result := make([]map[string]any, 0, len(s))
		for _, entry := range s {
			result = append(result, entry)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("expected array of services, got %T", v)
	}
}

func normalizeRawMap(v any) (map[string]any, error) {
	switch m := v.(type) {
	case map[string]any:
		result := make(map[string]any, len(m))
		for k, val := range m {
			result[k] = val
		}
		return result, nil
	case map[string]string:
		result := make(map[string]any, len(m))
		for k, val := range m {
			result[k] = val
		}
		return result, nil
	default:
		return nil, fmt.Errorf("expected map, got %T", v)
	}
}

func Flatten(values map[string]any) map[string]any {
	result := make(map[string]any)
	flattenWalk(values, "", result)
	return result
}

func flattenWalk(m map[string]any, prefix string, result map[string]any) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]any:
			flattenWalk(val, key, result)
		default:
			result[key] = val
		}
	}
}

func Unflatten(flat map[string]any) map[string]any {
	result := make(map[string]any)

	for key, val := range flat {
		parts := strings.Split(key, ".")
		current := result

		for i := 0; i < len(parts)-1; i++ {
			part := parts[i]
			if next, ok := current[part]; ok {
				if nextMap, ok := next.(map[string]any); ok {
					current = nextMap
				} else {
					newMap := make(map[string]any)
					current[part] = newMap
					current = newMap
				}
			} else {
				newMap := make(map[string]any)
				current[part] = newMap
				current = newMap
			}
		}

		current[parts[len(parts)-1]] = val
	}

	return result
}

func MergeNormalized(a, b *NormalizedSpec) *NormalizedSpec {
	if a == nil && b == nil {
		return &NormalizedSpec{Values: make(map[string]any)}
	}
	if a == nil {
		return copySpec(b)
	}
	if b == nil {
		return copySpec(a)
	}

	result := deepCopyMap(a.Values)
	deepMerge(result, b.Values)

	return &NormalizedSpec{Values: result}
}

func deepCopyMap(m map[string]any) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case map[string]any:
			result[k] = deepCopyMap(val)
		default:
			result[k] = val
		}
	}
	return result
}

func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if srcMap, ok := v.(map[string]any); ok {
			if dstMap, ok := dst[k].(map[string]any); ok {
				deepMerge(dstMap, srcMap)
				continue
			}
		}
		dst[k] = v
	}
}

func copySpec(spec *NormalizedSpec) *NormalizedSpec {
	return &NormalizedSpec{Values: deepCopyMap(spec.Values)}
}

func DiffNormalized(a, b *NormalizedSpec) []Difference {
	if a == nil && b == nil {
		return nil
	}
	if a == nil {
		a = &NormalizedSpec{Values: make(map[string]any)}
	}
	if b == nil {
		b = &NormalizedSpec{Values: make(map[string]any)}
	}

	var diffs []Difference
	diffs = diffMaps(a.Values, b.Values, "", diffs)
	return diffs
}

func diffMaps(a, b map[string]any, prefix string, diffs []Difference) []Difference {
	allKeys := make(map[string]bool)
	for k := range a {
		allKeys[k] = true
	}
	for k := range b {
		allKeys[k] = true
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}

		valA, inA := a[k]
		valB, inB := b[k]

		if inA && !inB {
			diffs = append(diffs, Difference{Path: path, Type: "removed", Old: valA, New: nil})
			continue
		}
		if !inA && inB {
			diffs = append(diffs, Difference{Path: path, Type: "added", Old: nil, New: valB})
			continue
		}

		mapA, isMapA := valA.(map[string]any)
		mapB, isMapB := valB.(map[string]any)

		if isMapA && isMapB {
			diffs = diffMaps(mapA, mapB, path, diffs)
			continue
		}

		if fmt.Sprintf("%v", valA) != fmt.Sprintf("%v", valB) {
			diffs = append(diffs, Difference{Path: path, Type: "changed", Old: valA, New: valB})
		}
	}

	return diffs
}

func InferTypes(values map[string]any) map[string]any {
	result := make(map[string]any, len(values))
	for k, v := range values {
		switch val := v.(type) {
		case map[string]any:
			result[k] = InferTypes(val)
		case []any:
			result[k] = inferArrayType(val)
		default:
			result[k] = annotateType(v)
		}
	}
	return result
}

func inferArrayType(arr []any) map[string]any {
	if len(arr) == 0 {
		return map[string]any{"_type": "array", "_items": "empty"}
	}

	itemTypes := make(map[string]bool)
	for _, item := range arr {
		itemTypes[describeType(item)] = true
	}

	typeList := make([]string, 0, len(itemTypes))
	for t := range itemTypes {
		typeList = append(typeList, t)
	}
	sort.Strings(typeList)

	return map[string]any{
		"_type":  "array",
		"_items": strings.Join(typeList, "|"),
		"_count": len(arr),
	}
}

func annotateType(v any) map[string]any {
	return map[string]any{
		"_type":  describeType(v),
		"_value": v,
	}
}

func describeType(v any) string {
	switch v.(type) {
	case bool:
		return "bool"
	case int, int8, int16, int32, int64:
		return "int"
	case uint, uint8, uint16, uint32, uint64:
		return "uint"
	case float32, float64:
		return "float"
	case string:
		return "string"
	case nil:
		return "null"
	case map[string]any:
		return "map"
	case []any:
		return "array"
	default:
		return "unknown"
	}
}

func ExtractSchema(normalized *NormalizedSpec) map[string]string {
	if normalized == nil {
		return map[string]string{}
	}

	result := make(map[string]string)
	extractSchemaWalk(normalized.Values, "", result)
	return result
}

func extractSchemaWalk(m map[string]any, prefix string, result map[string]string) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}

		switch val := m[k].(type) {
		case map[string]any:
			extractSchemaWalk(val, path, result)
		case []any:
			result[path] = "array"
		default:
			result[path] = describeType(val)
		}
	}
}

func normalizeModules(modules []parser.Module) []map[string]any {
	result := make([]map[string]any, 0, len(modules))
	for _, m := range modules {
		entry := map[string]any{
			"name": m.Name,
			"path": m.Path,
		}
		if m.Description != "" {
			entry["description"] = m.Description
		}
		if len(m.Dependencies) > 0 {
			entry["dependencies"] = m.Dependencies
		}
		result = append(result, entry)
	}
	return result
}

func normalizeServices(services []parser.Service) []map[string]any {
	result := make([]map[string]any, 0, len(services))
	for _, s := range services {
		entry := map[string]any{
			"name": s.Name,
			"kind": s.Kind,
			"port": s.Port,
		}
		if s.Description != "" {
			entry["description"] = s.Description
		}
		if len(s.Endpoints) > 0 {
			eps := make([]map[string]any, 0, len(s.Endpoints))
			for _, ep := range s.Endpoints {
				eps = append(eps, map[string]any{
					"method": ep.Method,
					"path":   ep.Path,
					"action": ep.Action,
				})
			}
			entry["endpoints"] = eps
		}
		result = append(result, entry)
	}
	return result
}

func normalizeArchitecture(arch *parser.Architecture) map[string]any {
	result := map[string]any{
		"pattern":     arch.Pattern,
		"description": arch.Description,
	}
	if len(arch.Principles) > 0 {
		result["principles"] = arch.Principles
	}
	return result
}

func normalizeDeployment(deploy *parser.Deployment) map[string]any {
	result := map[string]any{
		"strategy": deploy.Strategy,
	}
	if len(deploy.Environments) > 0 {
		envs := make([]map[string]any, 0, len(deploy.Environments))
		for _, env := range deploy.Environments {
			envs = append(envs, map[string]any{"name": env})
		}
		result["environments"] = envs
	}
	return result
}

func normalizeGeneration(gen *parser.Generation) map[string]any {
	result := map[string]any{}
	if len(gen.Languages) > 0 {
		result["languages"] = gen.Languages
	}
	if gen.OutputDir != "" {
		result["output_dir"] = gen.OutputDir
	}
	if gen.ModuleDir != "" {
		result["module_dir"] = gen.ModuleDir
	}
	return result
}

func normalizeTesting(test *parser.Testing) map[string]any {
	result := map[string]any{
		"strategy": test.Strategy,
	}
	if test.Coverage != "" {
		result["coverage"] = test.Coverage
	}
	return result
}
