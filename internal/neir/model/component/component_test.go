package component

import "testing"

func TestComponentKindConstants(t *testing.T) {
	tests := []struct {
		k    ComponentKind
		want string
	}{
		{KindHandler, "handler"},
		{KindService, "service"},
		{KindRepository, "repository"},
		{KindMiddleware, "middleware"},
		{KindModel, "model"},
		{KindConfig, "config"},
		{KindWorker, "worker"},
		{KindScheduler, "scheduler"},
	}
	for _, tt := range tests {
		if string(tt.k) != tt.want {
			t.Errorf("ComponentKind(%s) = %s, want %s", tt.want, string(tt.k), tt.want)
		}
	}
}

func TestComponent_ZeroValue(t *testing.T) {
	var c Component
	if c.Name != "" {
		t.Error("expected empty Name")
	}
	if c.Kind != "" {
		t.Error("expected empty Kind")
	}
}

func TestComponent_Full(t *testing.T) {
	c := Component{
		Name:         "user-handler",
		Kind:         KindHandler,
		Module:       "users",
		Description:  "Handles user requests",
		Dependencies: []string{"user-service"},
		Attributes:   map[string]string{"key": "val"},
	}
	if c.Name != "user-handler" {
		t.Errorf("expected user-handler, got %s", c.Name)
	}
	if c.Kind != KindHandler {
		t.Errorf("expected handler, got %s", c.Kind)
	}
	if c.Module != "users" {
		t.Errorf("expected users, got %s", c.Module)
	}
	if len(c.Dependencies) != 1 || c.Dependencies[0] != "user-service" {
		t.Errorf("expected [user-service], got %v", c.Dependencies)
	}
}
