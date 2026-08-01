package auth

import (
	"strings"
	"testing"
)

func TestSSORegistryNew(t *testing.T) {
	r := NewSSORegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestSSORegistryRegister(t *testing.T) {
	r := NewSSORegistry()
	p := NewOIDCProvider(SSOConfig{
		Name:         "test-oidc",
		Issuer:       "https://example.com",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})

	err := r.Register(p)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	got, ok := r.Get("test-oidc")
	if !ok {
		t.Fatal("expected provider to be found")
	}
	if got.Name() != "test-oidc" {
		t.Errorf("expected name 'test-oidc', got %q", got.Name())
	}
}

func TestSSORegistryRegisterInvalid(t *testing.T) {
	r := NewSSORegistry()
	p := NewOIDCProvider(SSOConfig{})
	err := r.Register(p)
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestSSORegistryList(t *testing.T) {
	r := NewSSORegistry()
	r.Register(NewOIDCProvider(SSOConfig{
		Name:         "p1",
		Issuer:       "https://example.com",
		ClientID:     "id",
		ClientSecret: "secret",
	}))
	r.Register(NewOIDCProvider(SSOConfig{
		Name:         "p2",
		Issuer:       "https://example2.com",
		ClientID:     "id2",
		ClientSecret: "secret2",
	}))

	providers := r.List()
	if len(providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(providers))
	}
}

func TestSSORegistryRemove(t *testing.T) {
	r := NewSSORegistry()
	r.Register(NewOIDCProvider(SSOConfig{
		Name:         "test",
		Issuer:       "https://example.com",
		ClientID:     "id",
		ClientSecret: "secret",
	}))

	r.Remove("test")
	_, ok := r.Get("test")
	if ok {
		t.Error("expected provider to be removed")
	}
}

func TestSSORegistryGetNotFound(t *testing.T) {
	r := NewSSORegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("expected false for nonexistent")
	}
}

func TestOIDCProviderValidate(t *testing.T) {
	tests := []struct {
		name string
		cfg  SSOConfig
		err  bool
	}{
		{"valid", SSOConfig{Name: "o", Issuer: "https://ex.com", ClientID: "id", ClientSecret: "secret"}, false},
		{"no-name", SSOConfig{Issuer: "https://ex.com", ClientID: "id", ClientSecret: "secret"}, true},
		{"no-issuer", SSOConfig{Name: "o", ClientID: "id", ClientSecret: "secret"}, true},
		{"no-client-id", SSOConfig{Name: "o", Issuer: "https://ex.com", ClientSecret: "secret"}, true},
		{"no-secret", SSOConfig{Name: "o", Issuer: "https://ex.com", ClientID: "id"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewOIDCProvider(tt.cfg)
			err := p.Validate()
			if tt.err && err == nil {
				t.Error("expected error")
			}
			if !tt.err && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestOIDCProviderTypeName(t *testing.T) {
	p := NewOIDCProvider(SSOConfig{Name: "test", Issuer: "https://ex.com", ClientID: "id", ClientSecret: "secret"})
	if p.Type() != ProviderOIDC {
		t.Errorf("expected OIDC type, got %s", p.Type())
	}
	if p.Name() != "test" {
		t.Errorf("expected name 'test', got %q", p.Name())
	}
}

func TestOIDCProviderGetAuthorizationURL(t *testing.T) {
	p := NewOIDCProvider(SSOConfig{
		Name:         "test",
		Issuer:       "https://accounts.example.com",
		ClientID:     "client-123",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"openid", "email"},
	})
	p.discovery = &OIDCDiscoveryDocument{
		AuthorizationEndpoint: "https://accounts.example.com/auth",
	}

	url := p.GetAuthorizationURL("state-abc")
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
	if !strings.Contains(url, "client_id=client-123") {
		t.Errorf("expected client_id in URL, got: %s", url)
	}
	if !strings.Contains(url, "state=state-abc") {
		t.Errorf("expected state in URL, got: %s", url)
	}
}

func TestOIDCProviderDiscoverNoServer(t *testing.T) {
	p := NewOIDCProvider(SSOConfig{
		Name:         "test",
		Issuer:       "http://localhost:1", // no server here
		ClientID:     "id",
		ClientSecret: "secret",
	})
	err := p.Discover()
	if err == nil {
		t.Error("expected error (no server)")
	}
}

func TestSAMLProviderValidate(t *testing.T) {
	tests := []struct {
		name string
		cfg  SSOConfig
		err  bool
	}{
		{"no-name", SSOConfig{SSOURL: "https://ex.com/saml", EntityID: "https://ex.com"}, true},
		{"no-sso-url", SSOConfig{Name: "saml", EntityID: "https://ex.com"}, true},
		{"no-entity-id", SSOConfig{Name: "saml", SSOURL: "https://ex.com/saml"}, true},
		{"valid", SSOConfig{Name: "saml", SSOURL: "https://ex.com/saml", EntityID: "https://ex.com"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewSAMLProvider(tt.cfg)
			err := p.Validate()
			if tt.err && err == nil {
				t.Error("expected error")
			}
			if !tt.err && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSAMLProviderTypeName(t *testing.T) {
	p := NewSAMLProvider(SSOConfig{Name: "okta-saml", SSOURL: "https://ex.com/saml", EntityID: "https://ex.com"})
	if p.Type() != ProviderSAML {
		t.Errorf("expected SAML type, got %s", p.Type())
	}
	if p.Name() != "okta-saml" {
		t.Errorf("expected name 'okta-saml', got %q", p.Name())
	}
}

func TestSAMLProviderUnsupportedMethods(t *testing.T) {
	p := NewSAMLProvider(SSOConfig{Name: "saml", SSOURL: "https://ex.com/saml", EntityID: "https://ex.com"})
	_, err := p.ExchangeCode("code")
	if err == nil {
		t.Error("expected error for ExchangeCode")
	}
	_, err = p.GetUser(nil)
	if err == nil {
		t.Error("expected error for GetUser")
	}
}

func TestSAMLProviderParseResponseSuccess(t *testing.T) {
	p := NewSAMLProvider(SSOConfig{Name: "saml", SSOURL: "https://ex.com/saml", EntityID: "https://ex.com"})

	samlResp := `<?xml version="1.0" encoding="UTF-8"?>
<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">
  <samlp:Status>
    <samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/>
  </samlp:Status>
  <saml:Assertion ID="assertion-1" IssueInstant="2024-01-01T00:00:00Z">
    <saml:Issuer>https://ex.com</saml:Issuer>
    <saml:Subject>
      <saml:NameID>user@example.com</saml:NameID>
    </saml:Subject>
    <saml:AttributeStatement>
      <saml:Attribute Name="email">
        <saml:AttributeValue>user@example.com</saml:AttributeValue>
      </saml:Attribute>
      <saml:Attribute Name="displayName">
        <saml:AttributeValue>John Doe</saml:AttributeValue>
      </saml:Attribute>
    </saml:AttributeStatement>
  </saml:Assertion>
</samlp:Response>`

	user, err := p.ParseResponse(samlResp)
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	if user.ID != "user@example.com" {
		t.Errorf("expected NameID 'user@example.com', got %q", user.ID)
	}
	if user.Email != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got %q", user.Email)
	}
	if user.Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got %q", user.Name)
	}
}

func TestSAMLProviderParseResponseFailure(t *testing.T) {
	p := NewSAMLProvider(SSOConfig{Name: "saml", SSOURL: "https://ex.com/saml", EntityID: "https://ex.com"})

	samlResp := `<?xml version="1.0" encoding="UTF-8"?>
<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol">
  <samlp:Status>
    <samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Responder"/>
  </samlp:Status>
</samlp:Response>`

	_, err := p.ParseResponse(samlResp)
	if err == nil {
		t.Error("expected error for failed SAML response")
	}
}

func TestSAMLProviderParseResponseNoAssertion(t *testing.T) {
	p := NewSAMLProvider(SSOConfig{Name: "saml", SSOURL: "https://ex.com/saml", EntityID: "https://ex.com"})

	samlResp := `<?xml version="1.0" encoding="UTF-8"?>
<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol">
  <samlp:Status>
    <samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/>
  </samlp:Status>
</samlp:Response>`

	_, err := p.ParseResponse(samlResp)
	if err == nil {
		t.Error("expected error for missing assertion")
	}
}

func TestLDAPProviderValidate(t *testing.T) {
	tests := []struct {
		name string
		cfg  SSOConfig
		err  bool
	}{
		{"no-name", SSOConfig{Host: "ldap.example.com", BaseDN: "dc=example,dc=com"}, true},
		{"no-host", SSOConfig{Name: "ldap", BaseDN: "dc=example,dc=com"}, true},
		{"no-base-dn", SSOConfig{Name: "ldap", Host: "ldap.example.com"}, true},
		{"valid", SSOConfig{Name: "ldap", Host: "ldap.example.com", BaseDN: "dc=example,dc=com"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLDAPProvider(tt.cfg)
			err := p.Validate()
			if tt.err && err == nil {
				t.Error("expected error")
			}
			if !tt.err && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLDAPProviderDefaults(t *testing.T) {
	p := NewLDAPProvider(SSOConfig{Name: "ldap", Host: "ldap.example.com", BaseDN: "dc=example,dc=com"})
	if err := p.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	cfg := p.Config()
	if cfg.Port != 389 {
		t.Errorf("expected default port 389, got %d", cfg.Port)
	}
	if cfg.UserFilter != "(uid=%s)" {
		t.Errorf("expected default filter, got %q", cfg.UserFilter)
	}
	if cfg.AttrMap == nil {
		t.Error("expected default attr map")
	}
}

func TestLDAPProviderGetAuthorizationURL(t *testing.T) {
	p := NewLDAPProvider(SSOConfig{Name: "ldap", Host: "ldap.example.com", Port: 389, BaseDN: "dc=example,dc=com"})
	url := p.GetAuthorizationURL("")
	if url != "ldap://ldap.example.com:389" {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestLDAPProviderUnsupportedMethods(t *testing.T) {
	p := NewLDAPProvider(SSOConfig{Name: "ldap", Host: "ldap.example.com", BaseDN: "dc=example,dc=com"})
	_, err := p.ExchangeCode("code")
	if err == nil {
		t.Error("expected error for ExchangeCode")
	}
	_, err = p.GetUser(nil)
	if err == nil {
		t.Error("expected error for GetUser")
	}
}

func TestLDAPProviderAuthNoServer(t *testing.T) {
	p := NewLDAPProvider(SSOConfig{Name: "ldap", Host: "127.0.0.1", Port: 1, BaseDN: "dc=example,dc=com"})
	_, err := p.Authenticate("user", "pass")
	if err == nil {
		t.Error("expected error (no server)")
	}
}

func TestManagerSSO(t *testing.T) {
	m := NewManager()
	sso := m.SSO()
	if sso == nil {
		t.Fatal("expected non-nil SSO registry")
	}
}

func TestOIDCProviderConfig(t *testing.T) {
	cfg := SSOConfig{Name: "test", Issuer: "https://ex.com", ClientID: "id", ClientSecret: "secret", Enabled: true}
	p := NewOIDCProvider(cfg)
	got := p.Config()
	if got.Name != "test" {
		t.Errorf("expected name 'test', got %q", got.Name)
	}
	if got.ClientSecret != "secret" {
		t.Errorf("expected client secret preserved")
	}
}

func TestSAMLProviderGetAuthorizationURL(t *testing.T) {
	p := NewSAMLProvider(SSOConfig{Name: "saml", SSOURL: "https://ex.com/saml/sso", EntityID: "https://ex.com"})
	url := p.GetAuthorizationURL("state")
	if url != "https://ex.com/saml/sso" {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestJWKToPublicKey(t *testing.T) {
	// Test with RSA key values
	key := JWK{
		Kty: "RSA",
		N:   "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMsE64v4y1jYd0F8b7Vr9kDYs5S5qO7J0l0e4WQcHvhE2E1lQhA9gO6cC8UqJxPqGkI2d4Mq8bC5s5qQMyFRAkRwXw6LIDPJYfpwC1q6Gg7c8oX4s8jKxF8pM9GQz8oFJn8MnQW5Lj8S6QoQqM5oB8n8Q6b5F5vF5v",
		E:   "AQAB",
	}
	pub, err := jwkToPublicKey(key)
	if err != nil {
		t.Fatalf("jwkToPublicKey: %v", err)
	}
	if pub == nil {
		t.Error("expected non-nil public key")
	}
}

func TestEncLen(t *testing.T) {
	tests := []struct {
		input int
		want  byte
	}{
		{0, 0},
		{1, 1},
		{255, 255},
		{256, 255},
		{-1, 255},
	}
	for _, tt := range tests {
		got := encLen(tt.input)
		if got != tt.want {
			t.Errorf("encLen(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestEncodeOctetString(t *testing.T) {
	tests := []struct {
		input []byte
		want  []byte
	}{
		{nil, []byte{0x04, 0}},
		{[]byte{}, []byte{0x04, 0}},
		{[]byte("hello"), []byte{0x04, 5, 'h', 'e', 'l', 'l', 'o'}},
		{[]byte{0x01, 0x02}, []byte{0x04, 2, 0x01, 0x02}},
	}
	for _, tt := range tests {
		got := encodeOctetString(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("encodeOctetString(%v) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("encodeOctetString(%v)[%d] = %d, want %d", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestParseLDAPSet(t *testing.T) {
	data := []byte{
		0x31, 0x05, // SET of length 5
		0x04, 0x03, 'a', 'b', 'c', // OCTET STRING "abc"
		0x31, 0x06, // SET of length 6
		0x04, 0x04, 'd', 'e', 'f', 'g', // OCTET STRING "defg"
	}
	_, items := parseLDAPSet(data)
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestParseLDAPSequence(t *testing.T) {
	data := []byte{
		0x30, 0x05, // SEQUENCE of length 5
		0x04, 0x03, 'a', 'b', 'c', // OCTET STRING "abc"
	}
	pos, items := parseLDAPSequence(data)
	if pos != len(data) {
		t.Errorf("expected pos=%d, got %d", len(data), pos)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestParseLDAPString(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []byte
	}{
		{"valid", []byte{0x04, 0x03, 'a', 'b', 'c'}, []byte("abc")},
		{"too short", []byte{0x04}, nil},
		{"truncated", []byte{0x04, 0x05, 'a', 'b'}, nil},
		{"empty", []byte{0x04, 0x00}, []byte{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLDAPString(tt.data)
			if len(got) != len(tt.want) {
				t.Errorf("parseLDAPString(%v) = %v, want %v", tt.data, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseLDAPString(%v)[%d] = %d, want %d", tt.data, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSAMLProviderConfig(t *testing.T) {
	cfg := SSOConfig{Name: "saml-test", SSOURL: "https://ex.com/saml", EntityID: "https://ex.com", Enabled: true}
	p := NewSAMLProvider(cfg)
	got := p.Config()
	if got.Name != "saml-test" {
		t.Errorf("expected name 'saml-test', got %q", got.Name)
	}
	if got.EntityID != "https://ex.com" {
		t.Errorf("expected entity id preserved")
	}
}

func TestLDAPProviderTypeName(t *testing.T) {
	p := NewLDAPProvider(SSOConfig{Name: "my-ldap", Host: "ldap.example.com", BaseDN: "dc=example,dc=com"})
	if p.Type() != ProviderLDAP {
		t.Errorf("expected LDAP type, got %s", p.Type())
	}
	if p.Name() != "my-ldap" {
		t.Errorf("expected name 'my-ldap', got %q", p.Name())
	}
}

func TestOIDCProviderExchangeCodeNoDiscovery(t *testing.T) {
	p := NewOIDCProvider(SSOConfig{
		Name:         "test",
		Issuer:       "https://ex.com",
		ClientID:     "id",
		ClientSecret: "secret",
	})
	_, err := p.ExchangeCode("code")
	if err == nil {
		t.Error("expected error because discovery not performed")
	}
}

func TestOIDCProviderGetUserNoDiscovery(t *testing.T) {
	p := NewOIDCProvider(SSOConfig{
		Name:         "test",
		Issuer:       "https://ex.com",
		ClientID:     "id",
		ClientSecret: "secret",
	})
	_, err := p.GetUser(&OAuth2Token{AccessToken: "tok"})
	if err == nil {
		t.Error("expected error because discovery not performed")
	}
}
