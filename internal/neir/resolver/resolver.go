package resolver

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	psr "github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

type ProfileResolver struct{}

func NewProfileResolver() *ProfileResolver {
	return &ProfileResolver{}
}

type ResolveResult struct {
	NEIR            *model.NEIR
	RemovedModules  []string
	RemovedServices []string
	Warnings        []string
}

func (r *ProfileResolver) Resolve(neir *model.NEIR, envVars map[string]string) (*ResolveResult, error) {
	if neir == nil {
		return nil, fmt.Errorf("neir model is nil")
	}

	result := &ResolveResult{
		NEIR:            neir,
		RemovedModules:  nil,
		RemovedServices: nil,
		Warnings:        nil,
	}

	if envVars == nil {
		envVars = make(map[string]string)
	}

	if neir.ActiveProfile != "" {
		envVars["profile"] = neir.ActiveProfile
	}

	if neir.Deployment != nil {
		for _, env := range neir.Deployment.Environments {
			if env.Name == neir.ActiveProfile || env.Name == envVars["env"] {
				for k, v := range env.Variables {
					envVars[k] = v
				}
				break
			}
		}
	}

	var keptModules []module.Module
	for _, m := range neir.Modules {
		if m.Condition == "" {
			keptModules = append(keptModules, m)
			continue
		}
		if psr.EvaluateCondition(m.Condition, envVars) {
			keptModules = append(keptModules, m)
		} else {
			result.RemovedModules = append(result.RemovedModules, m.Name)
		}
	}
	neir.Modules = keptModules

	var keptServices []service.Service
	for _, s := range neir.Services {
		if s.Condition == "" {
			keptServices = append(keptServices, s)
			continue
		}
		if psr.EvaluateCondition(s.Condition, envVars) {
			keptServices = append(keptServices, s)
		} else {
			result.RemovedServices = append(result.RemovedServices, s.Name)
		}
	}
	neir.Services = keptServices

	removedSet := make(map[string]bool, len(result.RemovedModules))
	for _, n := range result.RemovedModules {
		removedSet[n] = true
	}
	for i := range neir.Modules {
		var deps []string
		for _, d := range neir.Modules[i].Dependencies {
			if removedSet[d] {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("module %q depends on removed module %q", neir.Modules[i].Name, d))
				continue
			}
			deps = append(deps, d)
		}
		neir.Modules[i].Dependencies = deps
	}

	if neir.ActiveProfile != "" {
		var matched bool
		for _, env := range neir.Deployment.Environments {
			if env.Name == neir.ActiveProfile {
				matched = true
				break
			}
		}
		if !matched && len(neir.Deployment.Environments) > 0 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("active profile %q not found in deployment environments", neir.ActiveProfile))
		}
	}

	return result, nil
}
