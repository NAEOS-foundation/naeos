package deployment

import "testing"

func TestStrategyConstants(t *testing.T) {
	tests := []struct {
		s    Strategy
		want string
	}{
		{StrategyRolling, "rolling"},
		{StrategyBlueGreen, "blue-green"},
		{StrategyCanary, "canary"},
		{StrategyRecreate, "recreate"},
	}
	for _, tt := range tests {
		if string(tt.s) != tt.want {
			t.Errorf("Strategy(%s) = %s, want %s", tt.want, string(tt.s), tt.want)
		}
	}
}

func TestDeployment_ZeroValue(t *testing.T) {
	var d Deployment
	if d.Target != "" {
		t.Error("expected empty Target")
	}
	if d.Scaling != nil {
		t.Error("expected nil Scaling")
	}
}

func TestDeployment_Full(t *testing.T) {
	d := Deployment{
		Target:   "production",
		Strategy: StrategyBlueGreen,
		Environments: []Environment{
			{Name: "staging", Kind: "k8s", Config: map[string]string{"namespace": "staging"}},
		},
		Scaling:    &Scaling{Min: 1, Max: 10, Replicas: 3},
		Attributes: map[string]string{"key": "val"},
	}
	if d.Target != "production" {
		t.Errorf("expected production, got %s", d.Target)
	}
	if d.Strategy != StrategyBlueGreen {
		t.Errorf("expected blue-green, got %s", d.Strategy)
	}
	if len(d.Environments) != 1 {
		t.Errorf("expected 1 env, got %d", len(d.Environments))
	}
	if d.Scaling.Min != 1 || d.Scaling.Max != 10 || d.Scaling.Replicas != 3 {
		t.Errorf("Scaling = %+v, want Min=1 Max=10 Replicas=3", d.Scaling)
	}
}

func TestScaling_ZeroValue(t *testing.T) {
	var s Scaling
	if s.Min != 0 {
		t.Errorf("expected 0 Min, got %d", s.Min)
	}
}
