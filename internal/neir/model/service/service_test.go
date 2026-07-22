package service

import "testing"

func TestServiceKindConstants(t *testing.T) {
	tests := []struct {
		k    ServiceKind
		want string
	}{
		{KindHTTP, "http"},
		{KindGRPC, "grpc"},
		{KindWorker, "worker"},
		{KindCLI, "cli"},
		{KindJob, "job"},
	}
	for _, tt := range tests {
		if string(tt.k) != tt.want {
			t.Errorf("ServiceKind(%s) = %s, want %s", tt.want, string(tt.k), tt.want)
		}
	}
}

func TestService_ZeroValue(t *testing.T) {
	var s Service
	if s.Name != "" {
		t.Error("expected empty Name")
	}
	if s.Kind != "" {
		t.Error("expected empty Kind")
	}
	if s.Endpoints != nil {
		t.Error("expected nil Endpoints")
	}
}

func TestService_Full(t *testing.T) {
	s := Service{
		Name:        "user-api",
		Kind:        KindHTTP,
		Port:        8080,
		Description: "User service",
		Endpoints: []Endpoint{
			{Method: "GET", Path: "/users", Action: "list"},
		},
		Middleware: []string{"auth", "logging"},
		Attributes: map[string]string{"key": "val"},
	}
	if s.Name != "user-api" {
		t.Errorf("expected user-api, got %s", s.Name)
	}
	if s.Kind != KindHTTP {
		t.Errorf("expected http, got %s", s.Kind)
	}
	if s.Port != 8080 {
		t.Errorf("expected 8080, got %d", s.Port)
	}
	if len(s.Endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(s.Endpoints))
	}
	if s.Endpoints[0].Action != "list" {
		t.Errorf("expected list, got %s", s.Endpoints[0].Action)
	}
	if len(s.Middleware) != 2 {
		t.Errorf("expected 2 middleware, got %d", len(s.Middleware))
	}
}
