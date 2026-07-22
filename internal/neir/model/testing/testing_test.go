package testing

import "testing"

func TestTestingStrategyConstants(t *testing.T) {
	tests := []struct {
		s    TestingStrategy
		want string
	}{
		{StrategyUnit, "unit"},
		{StrategyIntegration, "integration"},
		{StrategyE2E, "e2e"},
		{StrategyContract, "contract"},
	}
	for _, tt := range tests {
		if string(tt.s) != tt.want {
			t.Errorf("TestingStrategy(%s) = %s, want %s", tt.want, string(tt.s), tt.want)
		}
	}
}

func TestTesting_ZeroValue(t *testing.T) {
	var tt Testing
	if tt.Strategy != "" {
		t.Error("expected empty Strategy")
	}
	if tt.Frameworks != nil {
		t.Error("expected nil Frameworks")
	}
	if tt.Coverage != nil {
		t.Error("expected nil Coverage")
	}
}

func TestTesting_Full(t *testing.T) {
	tt := Testing{
		Strategy:   StrategyUnit,
		Frameworks: []string{"go-test", "jest"},
		Coverage:   &Coverage{MinPercent: 80.0},
		Fixtures:   []Fixture{{Name: "users", Kind: "json", Path: "./fixtures/users.json"}},
		Attributes: map[string]string{"key": "val"},
	}
	if tt.Strategy != StrategyUnit {
		t.Errorf("expected unit, got %s", tt.Strategy)
	}
	if len(tt.Frameworks) != 2 {
		t.Errorf("expected 2 frameworks, got %d", len(tt.Frameworks))
	}
	if tt.Coverage.MinPercent != 80.0 {
		t.Errorf("expected 80.0, got %f", tt.Coverage.MinPercent)
	}
	if len(tt.Fixtures) != 1 {
		t.Errorf("expected 1 fixture, got %d", len(tt.Fixtures))
	}
	if tt.Fixtures[0].Kind != "json" {
		t.Errorf("expected json, got %s", tt.Fixtures[0].Kind)
	}
}
