package builder

import (
	"sync"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/api"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/component"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/generation"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/infrastructure"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/metadata"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/security"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/storage"
	"github.com/NAEOS-foundation/naeos/internal/specification/resolver"
	testingmodel "github.com/NAEOS-foundation/naeos/internal/neir/model/testing"
)

type LazyNEIR struct {
	raw    *resolver.ResolvedSpec
	neir   *model.NEIR
	loaded bool
	mu     sync.Mutex
}

func newLazyNEIR(raw *resolver.ResolvedSpec) *LazyNEIR {
	return &LazyNEIR{
		raw: raw,
		neir: &model.NEIR{
			Metadata: &metadata.Metadata{
				NEIRVersion:   "0.1.0",
				SchemaVersion: "1.0",
			},
		},
	}
}

func (l *LazyNEIR) loadBasic() {
	if l.neir.Project == nil {
		if rawProject, exists := l.raw.Context["project"]; exists {
			l.neir.Project = &project.Project{Name: rawProject.(string)}
		}
	}
	if profile, ok := l.raw.Context["active_profile"].(string); ok {
		l.neir.ActiveProfile = profile
	}
	if inherits, ok := l.raw.Context["inherits"].(string); ok {
		l.neir.Inherits = inherits
	}
}

func (l *LazyNEIR) Modules() []module.Module {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.loaded {
		l.loadBasic()
	}
	if l.neir.Modules != nil {
		return l.neir.Modules
	}
	if rawModules, exists := l.raw.Context["modules"]; exists {
		switch mods := rawModules.(type) {
		case []map[string]any:
			for _, m := range mods {
				l.neir.Modules = append(l.neir.Modules, extractModule(m))
			}
		case []any:
			for _, raw := range mods {
				if m, ok := raw.(map[string]any); ok {
					l.neir.Modules = append(l.neir.Modules, extractModule(m))
				}
			}
		}
	}
	return l.neir.Modules
}

func (l *LazyNEIR) Services() []service.Service {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.loaded {
		l.loadBasic()
	}
	if l.neir.Services != nil {
		return l.neir.Services
	}
	if rawServices, exists := l.raw.Context["services"]; exists {
		switch svcs := rawServices.(type) {
		case []map[string]any:
			for _, s := range svcs {
				l.neir.Services = append(l.neir.Services, extractService(s))
			}
		case []any:
			for _, raw := range svcs {
				if s, ok := raw.(map[string]any); ok {
					l.neir.Services = append(l.neir.Services, extractService(s))
				}
			}
		}
	}
	return l.neir.Services
}

func (l *LazyNEIR) Architecture() *architecture.Architecture {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.loaded {
		l.loadBasic()
	}
	if l.neir.Architecture != nil {
		return l.neir.Architecture
	}
	if rawArch, exists := l.raw.Context["architecture"]; exists {
		if archMap, ok := rawArch.(map[string]any); ok {
			l.neir.Architecture = extractArchitecture(archMap)
		}
	}
	return l.neir.Architecture
}

func (l *LazyNEIR) Generation() *generation.GenerationConfig {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.loaded {
		l.loadBasic()
	}
	if l.neir.Generation != nil {
		return l.neir.Generation
	}
	if rawGen, exists := l.raw.Context["generation"]; exists {
		if genMap, ok := rawGen.(map[string]any); ok {
			l.neir.Generation = extractGeneration(genMap)
		}
	}
	return l.neir.Generation
}

func (l *LazyNEIR) Deployment() *deployment.Deployment {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.loaded {
		l.loadBasic()
	}
	if l.neir.Deployment != nil {
		return l.neir.Deployment
	}
	if rawDeploy, exists := l.raw.Context["deployment"]; exists {
		if deployMap, ok := rawDeploy.(map[string]any); ok {
			l.neir.Deployment = extractDeployment(deployMap)
		}
	}
	return l.neir.Deployment
}

func (l *LazyNEIR) Testing() *testingmodel.Testing {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.loaded {
		l.loadBasic()
	}
	if l.neir.Testing != nil {
		return l.neir.Testing
	}
	if rawTest, exists := l.raw.Context["testing"]; exists {
		if testMap, ok := rawTest.(map[string]any); ok {
			l.neir.Testing = extractTesting(testMap)
		}
	}
	return l.neir.Testing
}

func (l *LazyNEIR) Infrastructure() *infrastructure.Infrastructure {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.loaded {
		l.loadBasic()
	}
	if l.neir.Infrastructure != nil {
		return l.neir.Infrastructure
	}
	if rawCloud, exists := l.raw.Context["cloud"]; exists {
		if cloudMap, ok := rawCloud.(map[string]any); ok {
			l.neir.Infrastructure = extractCloud(cloudMap)
		}
	}
	return l.neir.Infrastructure
}

func (l *LazyNEIR) Components() []component.Component {
	return l.loadAll().Components
}

func (l *LazyNEIR) APIs() []api.API {
	return l.loadAll().APIs
}

func (l *LazyNEIR) Storage() []storage.Storage {
	return l.loadAll().Storage
}

func (l *LazyNEIR) Security() *security.Security {
	return l.loadAll().Security
}

func (l *LazyNEIR) Project() *project.Project {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.loaded {
		l.loadBasic()
	}
	return l.neir.Project
}

func (l *LazyNEIR) ActiveProfile() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.loaded {
		l.loadBasic()
	}
	return l.neir.ActiveProfile
}

func (l *LazyNEIR) Inherits() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.loaded {
		l.loadBasic()
	}
	return l.neir.Inherits
}

func (l *LazyNEIR) Metadata() *metadata.Metadata {
	return l.neir.Metadata
}

func (l *LazyNEIR) loadAll() *model.NEIR {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.loaded {
		return l.neir
	}
	l.loadBasic()

	if l.neir.Modules == nil {
		if rawModules, exists := l.raw.Context["modules"]; exists {
			switch mods := rawModules.(type) {
			case []map[string]any:
				for _, m := range mods {
					l.neir.Modules = append(l.neir.Modules, extractModule(m))
				}
			case []any:
				for _, raw := range mods {
					if m, ok := raw.(map[string]any); ok {
						l.neir.Modules = append(l.neir.Modules, extractModule(m))
					}
				}
			}
		}
	}
	if l.neir.Modules == nil {
		l.neir.Modules = []module.Module{}
	}

	if l.neir.Services == nil {
		if rawServices, exists := l.raw.Context["services"]; exists {
			switch svcs := rawServices.(type) {
			case []map[string]any:
				for _, s := range svcs {
					l.neir.Services = append(l.neir.Services, extractService(s))
				}
			case []any:
				for _, raw := range svcs {
					if s, ok := raw.(map[string]any); ok {
						l.neir.Services = append(l.neir.Services, extractService(s))
					}
				}
			}
		}
	}
	if l.neir.Services == nil {
		l.neir.Services = []service.Service{}
	}

	if l.neir.Architecture == nil {
		if rawArch, exists := l.raw.Context["architecture"]; exists {
			if archMap, ok := rawArch.(map[string]any); ok {
				l.neir.Architecture = extractArchitecture(archMap)
			}
		}
	}

	if l.neir.Generation == nil {
		if rawGen, exists := l.raw.Context["generation"]; exists {
			if genMap, ok := rawGen.(map[string]any); ok {
				l.neir.Generation = extractGeneration(genMap)
			}
		}
	}

	if l.neir.Deployment == nil {
		if rawDeploy, exists := l.raw.Context["deployment"]; exists {
			if deployMap, ok := rawDeploy.(map[string]any); ok {
				l.neir.Deployment = extractDeployment(deployMap)
			}
		}
	}

	if l.neir.Testing == nil {
		if rawTest, exists := l.raw.Context["testing"]; exists {
			if testMap, ok := rawTest.(map[string]any); ok {
				l.neir.Testing = extractTesting(testMap)
			}
		}
	}

	if l.neir.Infrastructure == nil {
		if rawCloud, exists := l.raw.Context["cloud"]; exists {
			if cloudMap, ok := rawCloud.(map[string]any); ok {
				l.neir.Infrastructure = extractCloud(cloudMap)
			}
		}
	}

	if l.neir.Components == nil {
		l.neir.Components = []component.Component{}
	}
	if l.neir.APIs == nil {
		l.neir.APIs = []api.API{}
	}
	if l.neir.Storage == nil {
		l.neir.Storage = []storage.Storage{}
	}

	l.loaded = true
	return l.neir
}

func (l *LazyNEIR) ToNEIR() *model.NEIR {
	return l.loadAll()
}
