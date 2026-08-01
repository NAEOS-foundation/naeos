package validator

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/ai"
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
)

func TestValidatorAcceptsValidNEIR(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "acme-api"},
		Modules: []module.Module{{Name: "auth", Path: "./internal/auth"}},
	}
	v := NewValidator()
	if err := v.Validate(neir); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidatorRejectsIncompleteNEIR(t *testing.T) {
	neir := &model.NEIR{Project: &project.Project{}, Modules: []module.Module{}}
	v := NewValidator()
	if err := v.Validate(neir); err == nil {
		t.Fatalf("expected validation error for incomplete NEIR")
	}
}

func TestValidatorRejectsNilInput(t *testing.T) {
	v := NewValidator()
	if err := v.Validate(nil); err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestValidatorRejectsDuplicateModuleNames(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{
			{Name: "auth", Path: "./internal/auth"},
			{Name: "auth", Path: "./internal/auth2"},
		},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected validation to fail for duplicate module names")
	}
}

func TestValidatorRejectsEmptyModuleName(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "", Path: "./internal/empty"}},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected validation to fail for empty module name")
	}
}

func TestValidatorRejectsEmptyModulePath(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: ""}},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected validation to fail for empty module path")
	}
}

func TestValidatorRejectsInvalidServicePort(t *testing.T) {
	neir := &model.NEIR{
		Project:  &project.Project{Name: "test"},
		Modules:  []module.Module{{Name: "core", Path: "./internal/core"}},
		Services: []service.Service{{Name: "api", Port: 99999}},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected validation to fail for invalid service port")
	}
}

func TestValidatorReportsMultipleErrors(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{},
		Modules: []module.Module{},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected validation to fail")
	}
	if len(result.Errors) < 2 {
		t.Fatalf("expected multiple errors, got %d", len(result.Errors))
	}
}

func TestValidatorRejectsWrongType(t *testing.T) {
	result := ValidateDetailed("not a NEIR")
	if result.Valid {
		t.Fatal("expected validation to fail for wrong type")
	}
}

func TestValidatorRejectsNilProject(t *testing.T) {
	neir := &model.NEIR{
		Project: nil,
		Modules: []module.Module{{Name: "auth", Path: "./internal/auth"}},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected validation to fail for nil project")
	}
}

func TestValidatorRejectsServiceEmptyName(t *testing.T) {
	neir := &model.NEIR{
		Project:  &project.Project{Name: "test"},
		Modules:  []module.Module{{Name: "core", Path: "./internal/core"}},
		Services: []service.Service{{Name: "", Port: 8080}},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected validation to fail for empty service name")
	}
}

func TestValidatorWarnsDuplicatePorts(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core"}},
		Services: []service.Service{
			{Name: "api", Port: 8080},
			{Name: "admin", Port: 8080},
		},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatal("expected duplicate port to be a warning, not error")
	}
	if len(result.Warns) == 0 {
		t.Fatal("expected warning for duplicate ports")
	}
}

func TestValidatorAcceptsValidArchitecture(t *testing.T) {
	neir := &model.NEIR{
		Project:      &project.Project{Name: "test"},
		Modules:      []module.Module{{Name: "core", Path: "./internal/core"}},
		Architecture: &architecture.Architecture{Pattern: "hexagonal"},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid architecture, got errors: %v", result.Errors)
	}
}

func TestValidatorRejectsArchitectureEmptyPattern(t *testing.T) {
	neir := &model.NEIR{
		Project:      &project.Project{Name: "test"},
		Modules:      []module.Module{{Name: "core", Path: "./internal/core"}},
		Architecture: &architecture.Architecture{Pattern: ""},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for empty architecture pattern")
	}
}

func TestValidatorRejectsArchitectureUnsupportedPattern(t *testing.T) {
	neir := &model.NEIR{
		Project:      &project.Project{Name: "test"},
		Modules:      []module.Module{{Name: "core", Path: "./internal/core"}},
		Architecture: &architecture.Architecture{Pattern: "microservices"},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for unsupported architecture pattern")
	}
}

func TestValidatorAcceptsAllArchitecturePatterns(t *testing.T) {
	patterns := []architecture.Pattern{
		"layered", "clean", "hexagonal", "microkernel", "event-driven", "cqrs", "monolith",
	}
	for _, p := range patterns {
		neir := &model.NEIR{
			Project:      &project.Project{Name: "test"},
			Modules:      []module.Module{{Name: "core", Path: "./internal/core"}},
			Architecture: &architecture.Architecture{Pattern: p},
		}
		result := ValidateDetailed(neir)
		if !result.Valid {
			t.Fatalf("expected pattern %q to be valid, got error: %v", p, result.Errors)
		}
	}
}

func TestValidatorRejectsDeploymentEmptyStrategy(t *testing.T) {
	neir := &model.NEIR{
		Project:    &project.Project{Name: "test"},
		Modules:    []module.Module{{Name: "core", Path: "./internal/core"}},
		Deployment: &deployment.Deployment{Strategy: ""},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for empty deployment strategy")
	}
}

func TestValidatorRejectsDeploymentUnsupportedStrategy(t *testing.T) {
	neir := &model.NEIR{
		Project:    &project.Project{Name: "test"},
		Modules:    []module.Module{{Name: "core", Path: "./internal/core"}},
		Deployment: &deployment.Deployment{Strategy: "custom"},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for unsupported deployment strategy")
	}
}

func TestValidatorWarnsDeploymentEmptyEnvironments(t *testing.T) {
	neir := &model.NEIR{
		Project:    &project.Project{Name: "test"},
		Modules:    []module.Module{{Name: "core", Path: "./internal/core"}},
		Deployment: &deployment.Deployment{Strategy: "rolling", Environments: []deployment.Environment{}},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Warns) == 0 {
		t.Fatal("expected warning for empty environments")
	}
}

func TestValidatorRejectsTestingEmptyStrategy(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core"}},
		Testing: &testingmodel.Testing{Strategy: ""},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for empty testing strategy")
	}
}

func TestValidatorRejectsTestingUnsupportedStrategy(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core"}},
		Testing: &testingmodel.Testing{Strategy: "fuzz"},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for unsupported testing strategy")
	}
}

func TestValidatorRejectsTestingCoverageOutOfRange(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core"}},
		Testing: &testingmodel.Testing{Strategy: "unit", Coverage: &testingmodel.Coverage{MinPercent: 150}},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for coverage > 100")
	}
}

func TestValidatorAcceptsTestingCoverageBoundaries(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core"}},
		Testing: &testingmodel.Testing{Strategy: "unit", Coverage: &testingmodel.Coverage{MinPercent: 0}},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected coverage 0 to be valid, got: %v", result.Errors)
	}
}

func TestValidatorRejectsInfrastructureEmptyProvider(t *testing.T) {
	neir := &model.NEIR{
		Project:        &project.Project{Name: "test"},
		Modules:        []module.Module{{Name: "core", Path: "./internal/core"}},
		Infrastructure: &infrastructure.Infrastructure{Provider: ""},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for empty infrastructure provider")
	}
}

func TestValidatorRejectsInfrastructureUnsupportedProvider(t *testing.T) {
	neir := &model.NEIR{
		Project:        &project.Project{Name: "test"},
		Modules:        []module.Module{{Name: "core", Path: "./internal/core"}},
		Infrastructure: &infrastructure.Infrastructure{Provider: "digitalocean"},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for unsupported infrastructure provider")
	}
}

func TestValidatorWarnsInfrastructureEmptyRegion(t *testing.T) {
	neir := &model.NEIR{
		Project:        &project.Project{Name: "test"},
		Modules:        []module.Module{{Name: "core", Path: "./internal/core"}},
		Infrastructure: &infrastructure.Infrastructure{Provider: "aws", Region: ""},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Warns) == 0 {
		t.Fatal("expected warning for empty region")
	}
}

func TestValidatorRejectsAIMissingModelName(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core"}},
		AI:      &ai.AI{Models: []ai.Model{{Name: ""}}},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for empty AI model name")
	}
}

func TestValidatorRejectsAIMissingPromptName(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core"}},
		AI:      &ai.AI{Prompts: []ai.Prompt{{Name: ""}}},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for empty AI prompt name")
	}
}

func TestValidatorWarnsMetadataMissingVersions(t *testing.T) {
	neir := &model.NEIR{
		Project:  &project.Project{Name: "test"},
		Modules:  []module.Module{{Name: "core", Path: "./internal/core"}},
		Metadata: &metadata.Metadata{},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Warns) < 2 {
		t.Fatal("expected warnings for missing metadata versions")
	}
}

func TestValidatorWarnsGenerationEmptyLanguages(t *testing.T) {
	neir := &model.NEIR{
		Project:    &project.Project{Name: "test"},
		Modules:    []module.Module{{Name: "core", Path: "./internal/core"}},
		Generation: &generation.GenerationConfig{Languages: []language.Language{}},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Warns) == 0 {
		t.Fatal("expected warning for empty languages")
	}
}

func TestValidatorRejectsGenerationUnsupportedLanguage(t *testing.T) {
	neir := &model.NEIR{
		Project:    &project.Project{Name: "test"},
		Modules:    []module.Module{{Name: "core", Path: "./internal/core"}},
		Generation: &generation.GenerationConfig{Languages: []language.Language{"cobol"}},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for unsupported language")
	}
}

func TestValidatorWarnsGenerationEmptyOutputDir(t *testing.T) {
	neir := &model.NEIR{
		Project:    &project.Project{Name: "test"},
		Modules:    []module.Module{{Name: "core", Path: "./internal/core"}},
		Generation: &generation.GenerationConfig{Languages: []language.Language{"go"}, OutputDir: ""},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Warns) == 0 {
		t.Fatal("expected warning for empty output dir")
	}
}

func TestValidatorAcceptsValidWithAllSections(t *testing.T) {
	neir := &model.NEIR{
		Project:        &project.Project{Name: "myapp"},
		Modules:        []module.Module{{Name: "core", Path: "./core"}},
		Services:       []service.Service{{Name: "api", Port: 8080}},
		Architecture:   &architecture.Architecture{Pattern: "clean"},
		Deployment:     &deployment.Deployment{Strategy: "rolling", Environments: []deployment.Environment{{Name: "prod"}}},
		Testing:        &testingmodel.Testing{Strategy: "unit", Coverage: &testingmodel.Coverage{MinPercent: 80}},
		Infrastructure: &infrastructure.Infrastructure{Provider: "aws", Region: "us-east-1"},
		AI:             &ai.AI{Models: []ai.Model{{Name: "gpt-4"}}, Prompts: []ai.Prompt{{Name: "review"}}},
		Generation:     &generation.GenerationConfig{Languages: []language.Language{"go", "typescript"}, OutputDir: "./dist"},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid NEIR with all sections, got errors: %v", result.Errors)
	}
}

func TestValidatorRejectsWrongTypeInput(t *testing.T) {
	result := ValidateDetailed(42)
	if result.Valid {
		t.Fatal("expected validation to fail for int input")
	}
}

func TestValidatorWarnsModuleConditionInvalid(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{
			{Name: "core", Path: "./internal/core", Condition: "invalid op"},
		},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected condition warning not error, got: %v", result.Errors)
	}
	if len(result.Warns) == 0 {
		t.Fatal("expected warning for invalid module condition")
	}
}

func TestValidatorWarnsServiceConditionInvalid(t *testing.T) {
	neir := &model.NEIR{
		Project:  &project.Project{Name: "test"},
		Modules:  []module.Module{{Name: "core", Path: "./internal/core"}},
		Services: []service.Service{{Name: "api", Port: 8080, Condition: "bad cond"}},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected condition warning not error, got: %v", result.Errors)
	}
	if len(result.Warns) == 0 {
		t.Fatal("expected warning for invalid service condition")
	}
}

func TestValidatorWarnsActiveProfileNotFound(t *testing.T) {
	neir := &model.NEIR{
		Project:       &project.Project{Name: "test"},
		Modules:       []module.Module{{Name: "core", Path: "./internal/core"}},
		ActiveProfile: "staging",
		Deployment: &deployment.Deployment{
			Strategy:     "rolling",
			Environments: []deployment.Environment{{Name: "prod"}, {Name: "dev"}},
		},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid with warning, got errors: %v", result.Errors)
	}
	if len(result.Warns) == 0 {
		t.Fatal("expected warning for active_profile not found")
	}
}

func TestValidatorNoDuplicatePortWarningForZero(t *testing.T) {
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core"}},
		Services: []service.Service{
			{Name: "api", Port: 0},
			{Name: "admin", Port: 0},
		},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
	for _, w := range result.Warns {
		if w != "" {
			t.Fatalf("expected no warning for port=0 duplicates, got: %v", result.Warns)
		}
	}
}

func TestValidateConditionExpr(t *testing.T) {
	tests := []struct {
		name    string
		cond    string
		wantErr bool
	}{
		{"empty string", "", false},
		{"equals operator valid", "env == prod", false},
		{"not equals operator valid", "env != dev", false},
		{"defined prefix valid", "defined:FEATURE_X", false},
		{"not operator valid", "!disabled", false},
		{"missing key", "== prod", true},
		{"missing value", "env ==", true},
		{"unknown operator", "env ~= prod", true},
		{"just bang", "!", true},
		{"equals with spaces", "region == us-east-1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConditionExpr(tt.cond)
			if tt.wantErr && err == "" {
				t.Errorf("validateConditionExpr(%q) expected error, got none", tt.cond)
			}
			if !tt.wantErr && err != "" {
				t.Errorf("validateConditionExpr(%q) unexpected error: %s", tt.cond, err)
			}
		})
	}
}

func TestFindOp(t *testing.T) {
	tests := []struct {
		s, op string
		want  int
	}{
		{"env == prod", "==", 4},
		{"env != dev", "!=", 4},
		{"defined:FEATURE_X", "defined:", 0},
		{"no operator here", "==", -1},
		{"a == b == c", "==", 2},
		{"hello!", "!", 5},
		{"short", "==", -1},
	}
	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.op, func(t *testing.T) {
			got := findOp(tt.s, tt.op)
			if got != tt.want {
				t.Errorf("findOp(%q, %q) = %d, want %d", tt.s, tt.op, got, tt.want)
			}
		})
	}
}
