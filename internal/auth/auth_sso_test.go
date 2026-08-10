package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestJWKToPublicKeyInvalid(t *testing.T) {
	t.Parallel()

	_, err := jwkToPublicKey(JWK{Kty: "RSA", N: "!!!invalid-base64!!!", E: "AQAB"})
	if err == nil {
		t.Error("expected error for invalid base64 n")
	}

	_, err = jwkToPublicKey(JWK{Kty: "RSA", N: "dGVzdA==", E: "!!!invalid-base64!!!"})
	if err == nil {
		t.Error("expected error for invalid base64 e")
	}
}

func TestLDAPProviderAccessors(t *testing.T) {
	t.Parallel()

	p := NewLDAPProvider(SSOConfig{Name: "my-ldap", Host: "ldap.example.com", BaseDN: "dc=example,dc=com"})
	if p.Type() != ProviderLDAP {
		t.Errorf("expected ProviderLDAP, got %s", p.Type())
	}
	if p.Name() != "my-ldap" {
		t.Errorf("expected name 'my-ldap', got %q", p.Name())
	}
}

func TestParseBER(t *testing.T) {
	data := []byte{
		0x30, 0x05, // SEQUENCE of length 5
		0x04, 0x03, 'a', 'b', 'c', // OCTET STRING "abc"
		0x31, 0x06, // SET of length 6
		0x04, 0x04, 'd', 'e', 'f', 'g', // OCTET STRING "defg"
		0x30, 0x08, // SEQUENCE of length 8
		0x04, 0x01, 'x', 0x31, 0x03, 0x04, 0x01, 'y', // nested
	}
	items := parseBER(data)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	first := parseBER(items[0].content)
	if items[0].tag != 0x30 || len(first) != 1 || string(first[0].content) != "abc" {
		t.Errorf("unexpected first item: %+v", items[0])
	}
	second := parseBER(items[1].content)
	if items[1].tag != 0x31 || len(second) != 1 || string(second[0].content) != "defg" {
		t.Errorf("unexpected second item: %+v", items[1])
	}
	if items[2].tag != 0x30 {
		t.Errorf("unexpected third tag: %d", items[2].tag)
	}
	nested := parseBER(items[2].content)
	if len(nested) != 2 || nested[0].tag != 0x04 || string(nested[0].content) != "x" || nested[1].tag != 0x31 {
		t.Errorf("unexpected nested parse: %+v", nested)
	}
	inner := parseBER(nested[1].content)
	if len(inner) != 1 || inner[0].tag != 0x04 || string(inner[0].content) != "y" {
		t.Errorf("unexpected inner parse: %+v", inner)
	}
}

func TestParseBERTruncated(t *testing.T) {
	items := parseBER([]byte{0x04, 0x05, 'a', 'b'})
	if len(items) != 0 {
		t.Errorf("expected no items for truncated data, got %d", len(items))
	}
	items = parseBER([]byte{0x04})
	if len(items) != 0 {
		t.Errorf("expected no items for lone tag, got %d", len(items))
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

func TestSAMLProviderConfig(t *testing.T) {
	t.Parallel()

	cfg := SSOConfig{Name: "test-saml", SSOURL: "https://ex.com/saml", EntityID: "https://ex.com", Enabled: true}
	p := NewSAMLProvider(cfg)
	got := p.Config()
	if got.Name != "test-saml" {
		t.Errorf("expected name 'test-saml', got %q", got.Name)
	}
	if !got.Enabled {
		t.Error("expected Enabled to be preserved")
	}
}

func TestSAMLProviderValidateWithCertFile(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	os.WriteFile(certPath, []byte("not a real cert"), 0o600)

	p := NewSAMLProvider(SSOConfig{
		Name:     "saml-with-cert",
		SSOURL:   "https://ex.com/saml",
		EntityID: "https://ex.com",
		CertFile: certPath,
	})
	err := p.Validate()
	if err == nil {
		t.Error("expected error loading invalid cert")
	}
}

func TestOIDCProviderExtractUserFromIDToken(t *testing.T) {
	t.Parallel()

	claims := `{"sub":"user1","name":"Test User","email":"test@example.com"}`
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	idToken := "header." + payload + ".signature"

	p := NewOIDCProvider(SSOConfig{Name: "test", Issuer: "https://ex.com", ClientID: "id", ClientSecret: "secret"})
	user, err := p.extractUserFromIDToken(&OAuth2Token{IDToken: idToken})
	if err != nil {
		t.Fatalf("extractUserFromIDToken: %v", err)
	}
	if user.ID != "user1" {
		t.Errorf("expected ID 'user1', got %q", user.ID)
	}
	if user.Name != "Test User" {
		t.Errorf("expected Name 'Test User', got %q", user.Name)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got %q", user.Email)
	}
}

func TestOIDCProviderExtractUserFromIDTokenErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		token *OAuth2Token
	}{
		{"empty id_token", &OAuth2Token{IDToken: ""}},
		{"invalid format", &OAuth2Token{IDToken: "only.two"}},
		{"bad base64 payload", &OAuth2Token{IDToken: "header.!!!.sig"}},
		{"invalid JSON payload", &OAuth2Token{IDToken: "header." + base64.RawURLEncoding.EncodeToString([]byte("not-json")) + ".sig"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewOIDCProvider(SSOConfig{Name: "test", Issuer: "https://ex.com", ClientID: "id", ClientSecret: "secret"})
			_, err := p.extractUserFromIDToken(tt.token)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestOIDCProviderVerifyIDToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	n := base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes())

	p := NewOIDCProvider(SSOConfig{
		Name:         "test",
		Issuer:       "https://ex.com",
		ClientID:     "id",
		ClientSecret: "secret",
	})
	p.jwks = &JWKSResponse{
		Keys: []JWK{{Kty: "RSA", N: n, E: e}},
	}

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user1","name":"Test User","email":"test@ex.com"}`))
	message := header + "." + payload
	h := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, h[:])
	if err != nil {
		t.Fatal(err)
	}
	signature := base64.RawURLEncoding.EncodeToString(sig)
	idToken := message + "." + signature

	userInfo, err := p.verifyIDToken(idToken)
	if err != nil {
		t.Fatalf("verifyIDToken: %v", err)
	}
	if userInfo.Sub != "user1" {
		t.Errorf("expected sub 'user1', got %q", userInfo.Sub)
	}
	if userInfo.Email != "test@ex.com" {
		t.Errorf("expected email 'test@ex.com', got %q", userInfo.Email)
	}
}

func TestOIDCProviderVerifyIDTokenErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		idToken string
	}{
		{"invalid format", "header.payload"},
		{"bad signature base64", "header.payload.!!!bad-sig"},
		{"no jwks and no discovery", "header.payload.c2ln"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewOIDCProvider(SSOConfig{Name: "test", Issuer: "https://ex.com", ClientID: "id", ClientSecret: "secret"})
			_, err := p.verifyIDToken(tt.idToken)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestOIDCProviderDiscoverWithServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(OIDCDiscoveryDocument{
				Issuer:                "http://" + r.Host,
				JWKSUri:               "http://" + r.Host + "/jwks",
				AuthorizationEndpoint: "http://" + r.Host + "/auth",
				TokenEndpoint:         "http://" + r.Host + "/token",
				UserinfoEndpoint:      "http://" + r.Host + "/userinfo",
			})
		case "/jwks":
			json.NewEncoder(w).Encode(JWKSResponse{Keys: []JWK{}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := NewOIDCProvider(SSOConfig{
		Name:         "test",
		Issuer:       srv.URL,
		ClientID:     "id",
		ClientSecret: "secret",
	})

	err := p.Discover()
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if p.discovery == nil {
		t.Fatal("expected discovery document")
	}
	if p.discovery.TokenEndpoint != srv.URL+"/token" {
		t.Errorf("unexpected token endpoint: %s", p.discovery.TokenEndpoint)
	}
}

func TestOIDCProviderExchangeCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			json.NewEncoder(w).Encode(OIDCTokenResponse{
				AccessToken: "access-token-123",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := NewOIDCProvider(SSOConfig{
		Name:         "test",
		Issuer:       srv.URL,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  srv.URL + "/cb",
	})
	p.discovery = &OIDCDiscoveryDocument{
		TokenEndpoint: srv.URL + "/token",
	}

	token, err := p.ExchangeCode("auth-code-456")
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	if token.AccessToken != "access-token-123" {
		t.Errorf("expected 'access-token-123', got %q", token.AccessToken)
	}
	if token.TokenType != "Bearer" {
		t.Errorf("expected 'Bearer', got %q", token.TokenType)
	}
}

func TestOIDCProviderGetUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(OIDCUserInfo{
			Sub:   "oidc-user-1",
			Email: "oidc@example.com",
			Name:  "OIDC User",
		})
	}))
	defer srv.Close()

	p := NewOIDCProvider(SSOConfig{Name: "test", Issuer: srv.URL, ClientID: "id", ClientSecret: "secret"})
	p.discovery = &OIDCDiscoveryDocument{
		UserinfoEndpoint: srv.URL + "/userinfo",
	}

	user, err := p.GetUser(&OAuth2Token{AccessToken: "tok"})
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if user.Email != "oidc@example.com" {
		t.Errorf("expected 'oidc@example.com', got %q", user.Email)
	}
	if user.ID != "oidc-user-1" {
		t.Errorf("expected 'oidc-user-1', got %q", user.ID)
	}
	if user.Name != "OIDC User" {
		t.Errorf("expected 'OIDC User', got %q", user.Name)
	}
}

func TestOIDCProviderGetUserFromIDToken(t *testing.T) {
	claims := `{"sub":"sub-from-idtoken","name":"From Token","email":"from@token.com"}`
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	idToken := "header." + payload + ".sig"

	p := NewOIDCProvider(SSOConfig{Name: "test", Issuer: "https://ex.com", ClientID: "id", ClientSecret: "secret"})

	user, err := p.GetUser(&OAuth2Token{IDToken: idToken})
	if err != nil {
		t.Fatalf("GetUser (fallback): %v", err)
	}
	if user.ID != "sub-from-idtoken" {
		t.Errorf("expected 'sub-from-idtoken', got %q", user.ID)
	}
}

func TestOIDCProviderGetUserNoIDToken(t *testing.T) {
	p := NewOIDCProvider(SSOConfig{Name: "test", Issuer: "https://ex.com", ClientID: "id", ClientSecret: "secret"})
	_, err := p.GetUser(&OAuth2Token{})
	if err == nil {
		t.Error("expected error when no id_token and no userinfo endpoint")
	}
}

func TestFetchOIDCDiscovery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(OIDCDiscoveryDocument{
			Issuer:                "http://" + r.Host,
			JWKSUri:               "http://" + r.Host + "/jwks",
			AuthorizationEndpoint: "http://" + r.Host + "/auth",
			TokenEndpoint:         "http://" + r.Host + "/token",
			UserinfoEndpoint:      "http://" + r.Host + "/userinfo",
		})
	}))
	defer srv.Close()

	doc, err := fetchOIDCDiscovery(srv.URL, srv.Client())
	if err != nil {
		t.Fatalf("fetchOIDCDiscovery: %v", err)
	}
	if doc.Issuer != srv.URL {
		t.Errorf("expected issuer %q, got %q", srv.URL, doc.Issuer)
	}
	if doc.TokenEndpoint != srv.URL+"/token" {
		t.Errorf("expected token endpoint %q, got %q", srv.URL+"/token", doc.TokenEndpoint)
	}
}

func TestFetchJWKS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(JWKSResponse{
			Keys: []JWK{{Kty: "RSA", N: "dGVzdA==", E: "AQAB"}},
		})
	}))
	defer srv.Close()

	jwks, err := fetchJWKS(srv.URL, srv.Client())
	if err != nil {
		t.Fatalf("fetchJWKS: %v", err)
	}
	if len(jwks.Keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(jwks.Keys))
	}
	if jwks.Keys[0].Kty != "RSA" {
		t.Errorf("expected 'RSA', got %q", jwks.Keys[0].Kty)
	}
}

func TestFetchJWKSError(t *testing.T) {
	client := &http.Client{}
	_, err := fetchJWKS("http://127.0.0.1:1/jwks", client)
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestFetchOIDCDiscoveryError(t *testing.T) {
	client := &http.Client{}
	_, err := fetchOIDCDiscovery("http://127.0.0.1:1", client)
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestOIDCProviderExchangeCodeWithIDToken(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	n := base64.RawURLEncoding.EncodeToString(privKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privKey.E)).Bytes())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(OIDCDiscoveryDocument{
				Issuer:                "http://" + r.Host,
				JWKSUri:               "http://" + r.Host + "/jwks",
				TokenEndpoint:         "http://" + r.Host + "/token",
				AuthorizationEndpoint: "http://" + r.Host + "/auth",
			})
		case "/jwks":
			json.NewEncoder(w).Encode(JWKSResponse{
				Keys: []JWK{{Kty: "RSA", N: n, E: e}},
			})
		case "/token":
			header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
			payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user1","name":"Test","email":"test@ex.com"}`))
			msg := header + "." + payload
			h := sha256.Sum256([]byte(msg))
			sig, _ := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, h[:])
			sigEnc := base64.RawURLEncoding.EncodeToString(sig)
			idToken := msg + "." + sigEnc

			json.NewEncoder(w).Encode(OIDCTokenResponse{
				AccessToken: "tok",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				IDToken:     idToken,
			})
		}
	}))
	defer srv.Close()

	p := NewOIDCProvider(SSOConfig{
		Name:         "test",
		Issuer:       srv.URL,
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURL:  srv.URL + "/cb",
	})

	token, err := p.ExchangeCode("code")
	if err != nil {
		t.Fatalf("ExchangeCode with ID token: %v", err)
	}
	if token.IDToken == "" {
		t.Error("expected ID token to be set")
	}
	if token.UserID != "user1" {
		t.Errorf("expected UserID 'user1', got %q", token.UserID)
	}
}

func TestOIDCProviderGetUserWithUserinfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/userinfo" {
			if r.Header.Get("Authorization") != "Bearer tok" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			json.NewEncoder(w).Encode(OIDCUserInfo{
				Sub:   "ui-user",
				Email: "ui@example.com",
				Name:  "UI User",
			})
		}
	}))
	defer srv.Close()

	p := NewOIDCProvider(SSOConfig{Name: "test", Issuer: srv.URL, ClientID: "id", ClientSecret: "secret"})
	p.discovery = &OIDCDiscoveryDocument{
		UserinfoEndpoint: srv.URL + "/userinfo",
	}

	user, err := p.GetUser(&OAuth2Token{AccessToken: "tok"})
	if err != nil {
		t.Fatalf("GetUser with userinfo: %v", err)
	}
	if user.Email != "ui@example.com" {
		t.Errorf("expected 'ui@example.com', got %q", user.Email)
	}
}

func TestOIDCProviderExchangeCodeAutoDiscover(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(OIDCDiscoveryDocument{
				Issuer:                "http://" + r.Host,
				JWKSUri:               "http://" + r.Host + "/jwks",
				TokenEndpoint:         "http://" + r.Host + "/token",
				AuthorizationEndpoint: "http://" + r.Host + "/auth",
			})
		case "/token":
			json.NewEncoder(w).Encode(OIDCTokenResponse{
				AccessToken: "auto-discovered-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			})
		case "/jwks":
			json.NewEncoder(w).Encode(JWKSResponse{Keys: []JWK{}})
		}
	}))
	defer srv.Close()

	p := NewOIDCProvider(SSOConfig{
		Name:         "test",
		Issuer:       srv.URL,
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURL:  srv.URL + "/cb",
	})

	token, err := p.ExchangeCode("code789")
	if err != nil {
		t.Fatalf("ExchangeCode with auto-discover: %v", err)
	}
	if token.AccessToken != "auto-discovered-token" {
		t.Errorf("expected 'auto-discovered-token', got %q", token.AccessToken)
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
