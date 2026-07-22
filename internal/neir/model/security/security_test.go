package security

import "testing"

func TestSecurity_ZeroValue(t *testing.T) {
	var s Security
	if s.Authentication != nil {
		t.Error("expected nil Authentication")
	}
	if s.Secrets != nil {
		t.Error("expected nil Secrets")
	}
}

func TestSecurity_Full(t *testing.T) {
	s := Security{
		Authentication: &Authentication{Method: "oauth2", Provider: "google"},
		Authorization:  &Authorization{Model: "rbac", Roles: []string{"admin", "viewer"}},
		Encryption:     &Encryption{InTransit: true, AtRest: true, Algorithm: "aes-256"},
		Secrets: []Secret{
			{Name: "db-password", Kind: "env"},
		},
		Attributes: map[string]string{"key": "val"},
	}
	if s.Authentication.Method != "oauth2" {
		t.Errorf("expected oauth2, got %s", s.Authentication.Method)
	}
	if s.Authentication.Provider != "google" {
		t.Errorf("expected google, got %s", s.Authentication.Provider)
	}
	if s.Authorization.Model != "rbac" {
		t.Errorf("expected rbac, got %s", s.Authorization.Model)
	}
	if len(s.Authorization.Roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(s.Authorization.Roles))
	}
	if !s.Encryption.InTransit || !s.Encryption.AtRest {
		t.Error("expected encryption enabled")
	}
	if len(s.Secrets) != 1 {
		t.Errorf("expected 1 secret, got %d", len(s.Secrets))
	}
	if s.Secrets[0].Name != "db-password" {
		t.Errorf("expected db-password, got %s", s.Secrets[0].Name)
	}
}
