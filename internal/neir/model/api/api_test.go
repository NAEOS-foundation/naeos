package api

import (
	"testing"
)

func TestProtocolConstants(t *testing.T) {
	tests := []struct {
		protocol Protocol
		want     string
	}{
		{ProtocolHTTP, "http"},
		{ProtocolGRPC, "grpc"},
		{ProtocolGraphQL, "graphql"},
		{ProtocolWS, "websocket"},
	}
	for _, tt := range tests {
		if string(tt.protocol) != tt.want {
			t.Errorf("Protocol(%s) = %s, want %s", tt.want, string(tt.protocol), tt.want)
		}
	}
}

func TestAPI_ZeroValue(t *testing.T) {
	var a API
	if a.Name != "" {
		t.Error("expected empty Name")
	}
	if a.Endpoints != nil {
		t.Error("expected nil Endpoints")
	}
	if a.Attributes != nil {
		t.Error("expected nil Attributes")
	}
}

func TestAPI_Full(t *testing.T) {
	a := API{
		Name:     "users",
		Version:  "1.0",
		Protocol: ProtocolHTTP,
		Endpoints: []APIEndpoint{
			{Method: "GET", Path: "/users", Summary: "List users"},
		},
		Schemas: []Schema{
			{Name: "User", Fields: map[string]string{"id": "string"}},
		},
		Attributes: map[string]string{"versioned": "true"},
	}
	if a.Name != "users" {
		t.Errorf("expected users, got %s", a.Name)
	}
	if a.Protocol != ProtocolHTTP {
		t.Errorf("expected http, got %s", a.Protocol)
	}
	if len(a.Endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(a.Endpoints))
	}
	if a.Endpoints[0].Method != "GET" {
		t.Errorf("expected GET, got %s", a.Endpoints[0].Method)
	}
	if len(a.Schemas) != 1 {
		t.Errorf("expected 1 schema, got %d", len(a.Schemas))
	}
	if a.Schemas[0].Fields["id"] != "string" {
		t.Errorf("expected string, got %s", a.Schemas[0].Fields["id"])
	}
}
