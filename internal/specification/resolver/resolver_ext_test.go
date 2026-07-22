package resolver

import (
	"testing"
)

func TestDetectModulePathConflicts(t *testing.T) {
	conflicts := ConflictDetection(map[string]any{
		"modules": []map[string]any{
			{"name": "auth", "path": "./internal/auth"},
			{"name": "user", "path": "./internal/auth"},
		},
	})
	if len(conflicts) == 0 {
		t.Error("expected module path conflict")
	}
}

func TestValidateEndpointPaths(t *testing.T) {
	result := ValidateSpec(&ResolvedSpec{
		Context: map[string]any{
			"services": []map[string]any{
				{
					"name": "api",
					"endpoints": []map[string]any{
						{"path": "users"},
					},
				},
			},
		},
	})
	if result.Valid {
		t.Error("expected invalid for endpoint path without leading /")
	}
}

func TestToIntFloat64(t *testing.T) {
	result := ValidateSpec(&ResolvedSpec{
		Context: map[string]any{
			"services": []map[string]any{
				{"name": "api", "port": float64(99999)},
			},
		},
	})
	if result.Valid {
		t.Error("expected invalid for out-of-range port with float64")
	}
}
