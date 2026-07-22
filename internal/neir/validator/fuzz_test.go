package validator

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/module"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/service"
)

func FuzzValidateDetailed(f *testing.F) {
	f.Add("test", "1.0.0", "module-a", "module-b", "http", uint(8080))
	f.Add("", "", "", "", "", uint(0))
	f.Add("my-app", "2.0.0", "core", "api", "grpc", uint(443))

	f.Fuzz(func(t *testing.T, projName, version, modName, modDep, svcKind string, port uint) {
		neir := &model.NEIR{
			Project: &project.Project{
				Name:    projName,
				Version: version,
			},
			Modules: []module.Module{
				{Name: modName, Dependencies: []string{modDep}},
			},
			Services: []service.Service{
				{Name: "svc", Kind: service.ServiceKind(svcKind), Port: int(port)},
			},
		}

		result := ValidateDetailed(neir)

		if !result.Valid {
			if len(result.Errors) == 0 && len(result.Warns) == 0 {
				t.Error("expected errors or warns when invalid")
			}
		}
		for _, e := range result.Errors {
			if e == "" {
				t.Error("error message should not be empty")
			}
		}
		for _, w := range result.Warns {
			if w == "" {
				t.Error("warning message should not be empty")
			}
		}
	})
}

func FuzzValidateDetailedNil(f *testing.F) {
	f.Fuzz(func(t *testing.T, _ byte) {
		result := ValidateDetailed(nil)
		if result.Valid {
			t.Error("expected invalid result for nil input")
		}
	})
}
