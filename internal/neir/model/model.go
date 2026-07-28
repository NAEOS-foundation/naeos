package model

import (
	"github.com/NAEOS-foundation/naeos/internal/neir/model/ai"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/api"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/architecture"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/component"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/docs"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/domain"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/generation"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/infrastructure"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/metadata"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/security"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/storage"
	testingmodel "github.com/NAEOS-foundation/naeos/internal/neir/model/testing"
)

type NEIR struct {
	Project        *project.Project               `json:"project,omitempty"`
	Architecture   *architecture.Architecture     `json:"architecture,omitempty"`
	Domain         *domain.Domain                 `json:"domain,omitempty"`
	Modules        []module.Module                `json:"modules,omitempty"`
	Components     []component.Component          `json:"components,omitempty"`
	Services       []service.Service              `json:"services,omitempty"`
	APIs           []api.API                      `json:"apis,omitempty"`
	Storage        []storage.Storage              `json:"storage,omitempty"`
	Infrastructure *infrastructure.Infrastructure `json:"infrastructure,omitempty"`
	Security       *security.Security             `json:"security,omitempty"`
	AI             *ai.AI                         `json:"ai,omitempty"`
	Documentation  *docs.Documentation            `json:"documentation,omitempty"`
	Deployment     *deployment.Deployment         `json:"deployment,omitempty"`
	Testing        *testingmodel.Testing          `json:"testing,omitempty"`
	Metadata       *metadata.Metadata             `json:"metadata,omitempty"`
	Generation     *generation.GenerationConfig   `json:"generation,omitempty"`
	ActiveProfile  string                         `json:"active_profile,omitempty"`
	Inherits       string                         `json:"inherits,omitempty"`
}
