package validator

import (
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
	testingmodel "github.com/NAEOS-foundation/naeos/internal/neir/model/testing"
)

func TestFindOp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		s    string
		op   string
		want int
	}{
		{s: "env == prod", op: "==", want: 4},
		{s: "key != val", op: "!=", want: 4},
		{s: "no match", op: "==", want: -1},
		{s: "a == b", op: "==", want: 2},
		{s: "empty", op: "x", want: -1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.s, func(t *testing.T) {
			t.Parallel()
			got := findOp(tt.s, tt.op)
			if got != tt.want {
				t.Errorf("findOp(%q, %q) = %d, want %d", tt.s, tt.op, got, tt.want)
			}
		})
	}
}

func TestValidateConditionExpr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cond string
		want string
	}{
		{name: "empty", cond: "", want: ""},
		{name: "eq valid", cond: "env == prod", want: ""},
		{name: "neq valid", cond: "env != dev", want: ""},
		{name: "defined valid", cond: "defined:FEATURE_X", want: ""},
		{name: "not valid", cond: "!disabled", want: ""},
		{name: "missing key", cond: "== value", want: "unknown operator"},
		{name: "missing value", cond: "key ==", want: "invalid syntax: missing key or value"},
		{name: "invalid operator", cond: "foo bar", want: "unknown operator"},
		{name: "not short", cond: "!x", want: ""},
		{name: "not exactly", cond: "x!", want: "unknown operator"},
		{name: "eq missing key", cond: "== value", want: "unknown operator"},
		{name: "eq missing key alt", cond: "==value", want: "unknown operator"},
		{name: "neq missing value", cond: "key !=", want: "invalid syntax: missing key or value"},
		{name: "no operator", cond: "= missing", want: "unknown operator"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := validateConditionExpr(tt.cond)
			if tt.want == "" && got != "" {
				t.Errorf("validateConditionExpr(%q) = %q, want empty", tt.cond, got)
			}
			if tt.want != "" && !strings.Contains(got, tt.want) {
				t.Errorf("validateConditionExpr(%q) = %q, want containing %q", tt.cond, got, tt.want)
			}
		})
	}
}

func TestValidateDetailedModuleConditionValid(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core", Condition: "env == prod"}},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateDetailedModuleConditionInvalid(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core", Condition: "bad expr here"}},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid with warning, got errors: %v", result.Errors)
	}
	var found bool
	for _, w := range result.Warns {
		if strings.Contains(w, "condition") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected warning for invalid module condition")
	}
}

func TestValidateDetailedServiceConditionValid(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project:  &project.Project{Name: "test"},
		Modules:  []module.Module{{Name: "core", Path: "./internal/core"}},
		Services: []service.Service{{Name: "api", Port: 8080, Condition: "env == prod"}},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateDetailedServiceConditionInvalid(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project:  &project.Project{Name: "test"},
		Modules:  []module.Module{{Name: "core", Path: "./internal/core"}},
		Services: []service.Service{{Name: "api", Port: 8080, Condition: "?"}},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid with warning, got errors: %v", result.Errors)
	}
	var found bool
	for _, w := range result.Warns {
		if strings.Contains(w, "condition") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected warning for invalid service condition")
	}
}

func TestValidateDetailedActiveProfileNotFound(t *testing.T) {
	t.Parallel()
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
	var found bool
	for _, w := range result.Warns {
		if strings.Contains(w, "active_profile") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected warning about active_profile not found")
	}
}

func TestValidateDetailedActiveProfileFound(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project:       &project.Project{Name: "test"},
		Modules:       []module.Module{{Name: "core", Path: "./internal/core"}},
		ActiveProfile: "prod",
		Deployment: &deployment.Deployment{
			Strategy:     "rolling",
			Environments: []deployment.Environment{{Name: "prod"}, {Name: "dev"}},
		},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
	for _, w := range result.Warns {
		if strings.Contains(w, "active_profile") {
			t.Fatalf("unexpected warning about active_profile: %s", w)
		}
	}
}

func TestValidateDetailedActiveProfileNoDeployment(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project:       &project.Project{Name: "test"},
		Modules:       []module.Module{{Name: "core", Path: "./internal/core"}},
		ActiveProfile: "staging",
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateDetailedTestingCoverage100(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core"}},
		Testing: &testingmodel.Testing{Strategy: "unit", Coverage: &testingmodel.Coverage{MinPercent: 100}},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected coverage 100 to be valid, got: %v", result.Errors)
	}
}

func TestValidateDetailedTestingCoverageNegative(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project: &project.Project{Name: "test"},
		Modules: []module.Module{{Name: "core", Path: "./internal/core"}},
		Testing: &testingmodel.Testing{Strategy: "unit", Coverage: &testingmodel.Coverage{MinPercent: -1}},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for negative coverage")
	}
}

func TestValidateDetailedServicePortZero(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project:  &project.Project{Name: "test"},
		Modules:  []module.Module{{Name: "core", Path: "./internal/core"}},
		Services: []service.Service{{Name: "api", Port: 0}},
	}
	result := ValidateDetailed(neir)
	if !result.Valid {
		t.Fatalf("expected port 0 to be valid (not set), got: %v", result.Errors)
	}
}

func TestValidateDetailedServicePortNegative(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project:  &project.Project{Name: "test"},
		Modules:  []module.Module{{Name: "core", Path: "./internal/core"}},
		Services: []service.Service{{Name: "api", Port: -1}},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for negative port")
	}
}

func TestValidateDetailedEmptyModulesNoServices(t *testing.T) {
	t.Parallel()
	neir := &model.NEIR{
		Project:  &project.Project{Name: "test"},
		Modules:  []module.Module{},
		Services: []service.Service{},
	}
	result := ValidateDetailed(neir)
	if result.Valid {
		t.Fatal("expected error for empty modules")
	}
}
