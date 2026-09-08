package policy

import (
	"testing"
)

func TestRegistryRegisterAndGetActive(t *testing.T) {
	reg := NewRegistry()
	p := &Policy{
		ID:      "prod-deploy",
		Name:    "Production Deployment",
		Version: "1.0.0",
		Scope:   Scope{Resource: "deploy", Action: "run", Environment: "production"},
		Default: DecisionRequireApproval,
	}
	if err := reg.Register(p); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	got, ok := reg.GetActive("prod-deploy")
	if !ok {
		t.Fatal("expected active policy")
	}
	if got.Version != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %s", got.Version)
	}
	if !got.Active {
		t.Fatal("expected newly registered policy to be active")
	}
}

func TestRegistryVersioning(t *testing.T) {
	reg := NewRegistry()
	v1 := &Policy{ID: "p", Version: "1.0.0", Default: DecisionAllow}
	v2 := &Policy{ID: "p", Version: "2.0.0", Default: DecisionDeny}
	if err := reg.Register(v1); err != nil {
		t.Fatalf("register v1: %v", err)
	}
	if err := reg.Register(v2); err != nil {
		t.Fatalf("register v2: %v", err)
	}

	active, _ := reg.GetActive("p")
	if active.Version != "2.0.0" {
		t.Fatalf("expected newest to be active, got %s", active.Version)
	}
	if v1.Active {
		t.Fatal("expected old version to be deactivated")
	}
	if !v2.Active {
		t.Fatal("expected new version to be active")
	}

	if err := reg.SetActive("p", "1.0.0"); err != nil {
		t.Fatalf("set active: %v", err)
	}
	active, _ = reg.GetActive("p")
	if active.Version != "1.0.0" {
		t.Fatalf("expected 1.0.0 active, got %s", active.Version)
	}

	if err := reg.SetActive("p", "9.9.9"); err == nil {
		t.Fatal("expected error for unknown version")
	}

	if v := reg.Versions("p"); len(v) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(v))
	}
}

func TestRegistryRegisterDuplicateVersion(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(&Policy{ID: "p", Version: "1.0.0", Default: DecisionAllow}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := reg.Register(&Policy{ID: "p", Version: "1.0.0", Default: DecisionDeny}); err == nil {
		t.Fatal("expected conflict for duplicate version")
	}
}

func TestRegistryValidate(t *testing.T) {
	tt := []struct {
		name  string
		p     *Policy
		valid bool
	}{
		{name: "valid", p: &Policy{ID: "x", Version: "1.0.0", Default: DecisionAllow}, valid: true},
		{name: "nil", p: nil, valid: false},
		{name: "missing id", p: &Policy{Version: "1.0.0", Default: DecisionAllow}, valid: false},
		{name: "missing version", p: &Policy{ID: "x", Default: DecisionAllow}, valid: false},
		{name: "bad default", p: &Policy{ID: "x", Version: "1.0.0", Default: Decision("MAYBE")}, valid: false},
		{name: "bad rule decision", p: &Policy{ID: "x", Version: "1.0.0", Default: DecisionAllow, Rules: []PolicyRule{{RuleID: "r", Decision: Decision("NOPE")}}}, valid: false},
		{name: "missing rule id", p: &Policy{ID: "x", Version: "1.0.0", Default: DecisionAllow, Rules: []PolicyRule{{Decision: DecisionAllow}}}, valid: false},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.p)
			if tc.valid && err != nil {
				t.Fatalf("expected valid, got %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("expected invalid")
			}
		})
	}
}

func TestRegistryActiveForScopeMatching(t *testing.T) {
	reg := NewRegistry()
	// precise scope
	if err := reg.Register(&Policy{ID: "p1", Version: "1", Scope: Scope{Resource: "deploy", Action: "run", Environment: "production"}, Default: DecisionAllow}); err != nil {
		t.Fatal(err)
	}
	// wildcard scope
	if err := reg.Register(&Policy{ID: "p2", Version: "1", Default: DecisionDeny}); err != nil {
		t.Fatal(err)
	}

	if got := reg.ActiveFor("deploy", "run", "production"); len(got) != 2 {
		t.Fatalf("expected 2 matching policies, got %d", len(got))
	}
	if got := reg.ActiveFor("other", "run", "production"); len(got) != 1 || got[0].ID != "p2" {
		t.Fatalf("expected only wildcard p2, got %+v", got)
	}
}

func TestRegistryDelete(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(&Policy{ID: "p", Version: "1", Default: DecisionAllow}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Delete("p"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := reg.Delete("p"); err == nil {
		t.Fatal("expected error deleting missing policy")
	}
	if _, ok := reg.GetActive("p"); ok {
		t.Fatal("expected policy gone")
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(&Policy{ID: "a", Version: "1", Default: DecisionAllow}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(&Policy{ID: "b", Version: "1", Default: DecisionAllow}); err != nil {
		t.Fatal(err)
	}
	if got := reg.List(); len(got) != 2 {
		t.Fatalf("expected 2 listed policies, got %d", len(got))
	}
}
