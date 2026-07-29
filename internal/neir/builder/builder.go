package builder

import (
	"fmt"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/generation"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/infrastructure"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/metadata"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	testingmodel "github.com/NAEOS-foundation/naeos/internal/neir/model/testing"
	"github.com/NAEOS-foundation/naeos/internal/specification/resolver"
)

type Builder interface {
	Build(resolved any) (*model.NEIR, error)
}

type LazyBuilder interface {
	Builder
	BuildLazy(resolved any) (*LazyNEIR, error)
}

type DefaultBuilder struct{}

func NewBuilder() Builder {
	return DefaultBuilder{}
}

func NewLazyBuilder() LazyBuilder {
	return DefaultBuilder{}
}

func (b DefaultBuilder) Build(resolved any) (*model.NEIR, error) {
	return b.build(resolved)
}

func (b DefaultBuilder) BuildLazy(resolved any) (*LazyNEIR, error) {
	if resolved == nil {
		return nil, naeoserr.New(naeoserr.ErrValidation, "resolved spec is nil")
	}
	resolvedSpec, ok := resolved.(*resolver.ResolvedSpec)
	if !ok {
		return nil, naeoserr.New(naeoserr.ErrInternal, fmt.Sprintf("expected *resolver.ResolvedSpec, got %T", resolved))
	}
	return newLazyNEIR(resolvedSpec), nil
}

func (b DefaultBuilder) build(resolved any) (*model.NEIR, error) {
	if resolved == nil {
		return nil, naeoserr.New(naeoserr.ErrValidation, "resolved spec is nil")
	}

	resolvedSpec, ok := resolved.(*resolver.ResolvedSpec)
	if !ok {
		return nil, naeoserr.New(naeoserr.ErrInternal, fmt.Sprintf("expected *resolver.ResolvedSpec, got %T", resolved))
	}

	neir := &model.NEIR{
		Metadata: &metadata.Metadata{
			NEIRVersion:   "0.1.0",
			SchemaVersion: "1.0",
		},
	}

	if rawProject, exists := resolvedSpec.Context["project"]; exists {
		neir.Project = &project.Project{Name: fmt.Sprint(rawProject)}
	}

	if profile, ok := resolvedSpec.Context["active_profile"].(string); ok {
		neir.ActiveProfile = profile
	}
	if inherits, ok := resolvedSpec.Context["inherits"].(string); ok {
		neir.Inherits = inherits
	}

	if rawModules, exists := resolvedSpec.Context["modules"]; exists {
		switch mods := rawModules.(type) {
		case []map[string]any:
			for _, m := range mods {
				neir.Modules = append(neir.Modules, extractModule(m))
			}
		case []any:
			for _, raw := range mods {
				if m, ok := raw.(map[string]any); ok {
					neir.Modules = append(neir.Modules, extractModule(m))
				}
			}
		}
	}

	if rawServices, exists := resolvedSpec.Context["services"]; exists {
		switch svcs := rawServices.(type) {
		case []map[string]any:
			for _, s := range svcs {
				neir.Services = append(neir.Services, extractService(s))
			}
		case []any:
			for _, raw := range svcs {
				if s, ok := raw.(map[string]any); ok {
					neir.Services = append(neir.Services, extractService(s))
				}
			}
		}
	}

	if rawArch, exists := resolvedSpec.Context["architecture"]; exists {
		if archMap, ok := rawArch.(map[string]any); ok {
			neir.Architecture = extractArchitecture(archMap)
		}
	}

	if rawGen, exists := resolvedSpec.Context["generation"]; exists {
		if genMap, ok := rawGen.(map[string]any); ok {
			neir.Generation = extractGeneration(genMap)
		}
	}

	if rawDeploy, exists := resolvedSpec.Context["deployment"]; exists {
		if deployMap, ok := rawDeploy.(map[string]any); ok {
			neir.Deployment = extractDeployment(deployMap)
		}
	}

	if rawTest, exists := resolvedSpec.Context["testing"]; exists {
		if testMap, ok := rawTest.(map[string]any); ok {
			neir.Testing = extractTesting(testMap)
		}
	}

	if rawCloud, exists := resolvedSpec.Context["cloud"]; exists {
		if cloudMap, ok := rawCloud.(map[string]any); ok {
			neir.Infrastructure = extractCloud(cloudMap)
		}
	}

	return neir, nil
}

func extractModule(m map[string]any) module.Module {
	mod := module.Module{}
	if name, ok := m["name"].(string); ok {
		mod.Name = name
	}
	if path, ok := m["path"].(string); ok {
		mod.Path = path
	}
	if desc, ok := m["description"].(string); ok {
		mod.Description = desc
	}
	if cond, ok := m["condition"].(string); ok {
		mod.Condition = cond
	}
	if deps, ok := m["dependencies"].([]any); ok {
		for _, d := range deps {
			if s, ok := d.(string); ok {
				mod.Dependencies = append(mod.Dependencies, s)
			}
		}
	}
	return mod
}

func extractService(s map[string]any) service.Service {
	svc := service.Service{}
	if name, ok := s["name"].(string); ok {
		svc.Name = name
	}
	if kind, ok := s["kind"].(string); ok {
		svc.Kind = service.ServiceKind(kind)
	}
	if port, ok := s["port"].(int); ok {
		svc.Port = port
	}
	if desc, ok := s["description"].(string); ok {
		svc.Description = desc
	}
	if cond, ok := s["condition"].(string); ok {
		svc.Condition = cond
	}
	if endpoints, ok := s["endpoints"].([]any); ok {
		for _, e := range endpoints {
			if epMap, ok := e.(map[string]any); ok {
				ep := service.Endpoint{}
				if method, ok := epMap["method"].(string); ok {
					ep.Method = method
				}
				if path, ok := epMap["path"].(string); ok {
					ep.Path = path
				}
				if action, ok := epMap["action"].(string); ok {
					ep.Action = action
				}
				svc.Endpoints = append(svc.Endpoints, ep)
			}
		}
	}
	return svc
}

func extractArchitecture(m map[string]any) *architecture.Architecture {
	arch := &architecture.Architecture{}
	if pattern, ok := m["pattern"].(string); ok {
		arch.Pattern = architecture.Pattern(pattern)
	}
	if desc, ok := m["description"].(string); ok {
		arch.Description = desc
	}
	return arch
}

func extractGeneration(m map[string]any) *generation.GenerationConfig {
	gen := &generation.GenerationConfig{}
	if langs, ok := m["languages"].([]any); ok {
		for _, l := range langs {
			if s, ok := l.(string); ok {
				gen.Languages = append(gen.Languages, language.Language(s))
			}
		}
	} else if langs, ok := m["languages"].([]string); ok {
		for _, l := range langs {
			gen.Languages = append(gen.Languages, language.Language(l))
		}
	}
	if outputDir, ok := m["output_dir"].(string); ok {
		gen.OutputDir = outputDir
	}
	if moduleDir, ok := m["module_dir"].(string); ok {
		gen.ModuleDir = moduleDir
	}
	return gen
}

func extractDeployment(m map[string]any) *deployment.Deployment {
	deploy := &deployment.Deployment{}
	if strategy, ok := m["strategy"].(string); ok {
		deploy.Strategy = deployment.Strategy(strategy)
	}
	if envs, ok := m["environments"].([]any); ok {
		for _, e := range envs {
			if envMap, ok := e.(map[string]any); ok {
				env := deployment.Environment{}
				if name, ok := envMap["name"].(string); ok {
					env.Name = name
				}
				if kind, ok := envMap["kind"].(string); ok {
					env.Kind = kind
				}
				if vars, ok := envMap["variables"].(map[string]any); ok {
					env.Variables = make(map[string]string, len(vars))
					for k, v := range vars {
						env.Variables[k] = fmt.Sprint(v)
					}
				}
				deploy.Environments = append(deploy.Environments, env)
			} else if name, ok := e.(string); ok {
				deploy.Environments = append(deploy.Environments, deployment.Environment{Name: name})
			}
		}
	}
	return deploy
}

func extractTesting(m map[string]any) *testingmodel.Testing {
	test := &testingmodel.Testing{}
	if strategy, ok := m["strategy"].(string); ok {
		test.Strategy = testingmodel.TestingStrategy(strategy)
	}
	if coverage, ok := m["coverage"].(string); ok {
		minPercent := 0.0
		switch coverage {
		case "high":
			minPercent = 80.0
		case "medium":
			minPercent = 60.0
		case "low":
			minPercent = 40.0
		}
		test.Coverage = &testingmodel.Coverage{MinPercent: minPercent}
	}
	return test
}

func extractCloud(m map[string]any) *infrastructure.Infrastructure {
	infra := &infrastructure.Infrastructure{}

	if provider, ok := m["provider"].(string); ok {
		infra.Provider = infrastructure.Provider(provider)
	}
	if region, ok := m["region"].(string); ok {
		infra.Region = region
	}
	if project, ok := m["project"].(string); ok {
		infra.Project = project
	}
	if env, ok := m["environment"].(string); ok {
		infra.Environment = env
	}

	if rawResources, ok := m["resources"].([]any); ok {
		for _, raw := range rawResources {
			if resMap, ok := raw.(map[string]any); ok {
				res := infrastructure.Resource{}
				if name, ok := resMap["name"].(string); ok {
					res.Name = name
				}
				if kind, ok := resMap["kind"].(string); ok {
					res.Kind = kind
				}
				if resType, ok := resMap["type"].(string); ok {
					res.Type = resType
				}
				if spec, ok := resMap["spec"].(map[string]any); ok {
					res.Spec = make(map[string]string)
					for k, v := range spec {
						res.Spec[k] = fmt.Sprint(v)
					}
				}
				infra.Resources = append(infra.Resources, res)
			}
		}
	}

	if attrs, ok := m["attributes"].(map[string]any); ok {
		infra.Attributes = make(map[string]string)
		for k, v := range attrs {
			infra.Attributes[k] = fmt.Sprint(v)
		}
	}

	return infra
}
