package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func buildSearchEntry(name, value string) []byte {
	attr := encodeOctetString([]byte(name))
	valBytes := encodeOctetString([]byte(value))
	vals := append([]byte{0x31, byte(len(valBytes))}, valBytes...)
	attrSeq := append(attr, vals...)
	attrSeq = append([]byte{0x30, byte(len(attrSeq))}, attrSeq...)
	return attrSeq
}

func TestLDAPProviderAuthenticateSuccess(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, 4096)
		if _, err := conn.Read(buf); err != nil {
			return
		}
		conn.Write([]byte{0x30, 0x08, 0x02, 0x01, 0x01, 0x61, 0x05, 0x0a, 0x01, 0x00})

		if _, err := conn.Read(buf); err != nil {
			return
		}
		dn := []byte("uid=alice,dc=example,dc=com")
		attrSeq := buildSearchEntry("uid", "alice")
		attrSeq = append(attrSeq, buildSearchEntry("cn", "Alice Smith")...)
		attrSeq = append(attrSeq, buildSearchEntry("mail", "alice@example.com")...)
		attributes := append([]byte{0x30, byte(len(attrSeq))}, attrSeq...)
		entry := append(encodeOctetString(dn), attributes...)
		entry = append([]byte{0x64, byte(len(entry))}, entry...)
		msg := append([]byte{0x02, 0x01, 0x02}, entry...)
		msg = append([]byte{0x30, byte(len(msg))}, msg...)
		conn.Write(msg)
	}()

	p := NewLDAPProvider(SSOConfig{
		Name:   "ldap",
		Host:   "127.0.0.1",
		Port:   ln.Addr().(*net.TCPAddr).Port,
		BaseDN: "dc=example,dc=com",
	})
	if err := p.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	user, err := p.Authenticate("alice", "secret")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if user.ID != "alice" {
		t.Errorf("expected id alice, got %q", user.ID)
	}
	if user.Name != "Alice Smith" {
		t.Errorf("expected name 'Alice Smith', got %q", user.Name)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("expected email, got %q", user.Email)
	}
}

func TestLDAPProviderAuthenticateBindFailure(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 4096)
		if _, err := conn.Read(buf); err != nil {
			return
		}
		conn.Write([]byte{0x30, 0x08, 0x02, 0x01, 0x01, 0x61, 0x05, 0x0a, 0x01, 49})
	}()

	p := NewLDAPProvider(SSOConfig{
		Name:   "ldap",
		Host:   "127.0.0.1",
		Port:   ln.Addr().(*net.TCPAddr).Port,
		BaseDN: "dc=example,dc=com",
	})
	_, err = p.Authenticate("alice", "wrong")
	if err == nil {
		t.Fatal("expected bind failure error")
	}
}

func TestLDAPProviderAuthenticateUserBindPath(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 4096)
		for i := 0; i < 2; i++ {
			if _, err := conn.Read(buf); err != nil {
				return
			}
			conn.Write([]byte{0x30, 0x08, 0x02, 0x01, 0x01, 0x61, 0x05, 0x0a, 0x01, 0x00})
		}
		if _, err := conn.Read(buf); err != nil {
			return
		}
		conn.Write([]byte{0x30, 0x03, 0x02, 0x01, 0x02})
	}()

	p := NewLDAPProvider(SSOConfig{
		Name:         "ldap",
		Host:         "127.0.0.1",
		Port:         ln.Addr().(*net.TCPAddr).Port,
		BaseDN:       "dc=example,dc=com",
		BindDN:       "cn=admin,dc=example,dc=com",
		BindPassword: "adminpass",
	})

	user, err := p.Authenticate("alice", "secret")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if user.ID != "alice" {
		t.Errorf("expected id alice, got %q", user.ID)
	}
}

func TestLDAPProviderDialTLSFailure(t *testing.T) {
	p := NewLDAPProvider(SSOConfig{
		Name:   "ldap",
		Host:   "127.0.0.1",
		Port:   636,
		BaseDN: "dc=example,dc=com",
	})
	_, err := p.Authenticate("alice", "secret")
	if err == nil {
		t.Fatal("expected TLS dial failure")
	}
}

func TestLDAPConnClose(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 4096)
		conn.Read(buf)
		conn.Write([]byte{0x30, 0x08, 0x02, 0x01, 0x01, 0x61, 0x05, 0x0a, 0x01, 0x00})
		conn.Read(buf)
		conn.Write([]byte{0x30, 0x03, 0x02, 0x01, 0x02})
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	c := &ldapConn{conn: conn}
	if err := c.bind("cn=admin", "pw"); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if _, err := c.search("uid=x", "(uid=x)", "dc=example,dc=com"); err != nil {
		t.Fatalf("search: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestLDAPConnBindWriteFailure(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		conn.Close()
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	c := &ldapConn{conn: conn}
	err = c.bind("", "")
	if err == nil {
		t.Fatal("expected error when server closes connection")
	}
}

func signJWT(t *testing.T, priv *rsa.PrivateKey, claims map[string]any, kid string) string {
	t.Helper()
	header := map[string]string{"alg": "RS256", "typ": "JWT", "kid": kid}
	hb, _ := json.Marshal(header)
	cb, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(hb) + "." + base64.RawURLEncoding.EncodeToString(cb)
	hash := sha256.Sum256([]byte(payload))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return payload + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func jwkForPub(pub *rsa.PublicKey, kid string) JWK {
	return JWK{
		Kty: "RSA",
		Alg: "RS256",
		Kid: kid,
		Use: "sig",
		N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01}),
	}
}

func TestOIDCProviderDiscoverSuccess(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(OIDCDiscoveryDocument{
				Issuer:                ts.URL,
				JWKSUri:               ts.URL + "/.well-known/jwks.json",
				AuthorizationEndpoint: ts.URL + "/authorize",
				TokenEndpoint:         ts.URL + "/token",
				UserinfoEndpoint:      ts.URL + "/userinfo",
			})
		case "/.well-known/jwks.json":
			json.NewEncoder(w).Encode(JWKSResponse{Keys: []JWK{jwkForPub(&priv.PublicKey, "test-key")}})
		}
	}))
	defer ts.Close()

	p := NewOIDCProvider(SSOConfig{Name: "oidc", Issuer: ts.URL, ClientID: "id", ClientSecret: "secret"})
	if err := p.Discover(); err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if p.discovery == nil || p.discovery.TokenEndpoint == "" {
		t.Fatal("expected discovery document")
	}
	if p.jwks == nil || len(p.jwks.Keys) != 1 {
		t.Fatal("expected JWKS")
	}
}

func TestOIDCProviderExchangeCodeSuccess(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	idToken := signJWT(t, priv, map[string]any{"sub": "user-123", "name": "Jane Doe", "email": "jane@example.com"}, "test-key")

	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(OIDCDiscoveryDocument{
				Issuer:           ts.URL,
				JWKSUri:          ts.URL + "/.well-known/jwks.json",
				TokenEndpoint:    ts.URL + "/token",
				UserinfoEndpoint: ts.URL + "/userinfo",
			})
		case "/.well-known/jwks.json":
			json.NewEncoder(w).Encode(JWKSResponse{Keys: []JWK{jwkForPub(&priv.PublicKey, "test-key")}})
		case "/token":
			json.NewEncoder(w).Encode(OIDCTokenResponse{
				AccessToken: "access-1",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				IDToken:     idToken,
			})
		}
	}))
	defer ts.Close()

	p := NewOIDCProvider(SSOConfig{
		Name:         "oidc",
		Issuer:       ts.URL,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/cb",
	})

	token, err := p.ExchangeCode("auth-code")
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	if token.AccessToken != "access-1" {
		t.Errorf("expected access token, got %q", token.AccessToken)
	}
	if token.UserID != "user-123" {
		t.Errorf("expected UserID from ID token, got %q", token.UserID)
	}
	if token.IDToken == "" {
		t.Error("expected ID token stored")
	}
	if token.ExpiresAt.Before(time.Now()) {
		t.Error("expected future expiry")
	}
}

func TestOIDCProviderExchangeCodeInvalidIDToken(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(OIDCDiscoveryDocument{
				Issuer:        ts.URL,
				JWKSUri:       ts.URL + "/.well-known/jwks.json",
				TokenEndpoint: ts.URL + "/token",
			})
		case "/.well-known/jwks.json":
			json.NewEncoder(w).Encode(JWKSResponse{Keys: []JWK{jwkForPub(&priv.PublicKey, "other-key")}})
		case "/token":
			json.NewEncoder(w).Encode(OIDCTokenResponse{
				AccessToken: "access-1",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				IDToken:     "header.payload.tampered-signature",
			})
		}
	}))
	defer ts.Close()

	p := NewOIDCProvider(SSOConfig{Name: "oidc", Issuer: ts.URL, ClientID: "id", ClientSecret: "secret"})
	token, err := p.ExchangeCode("code")
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	if token.UserID != "" {
		t.Errorf("expected no UserID for invalid signature, got %q", token.UserID)
	}
}

func TestOIDCProviderGetUserInfo(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(OIDCDiscoveryDocument{
				Issuer:           ts.URL,
				JWKSUri:          ts.URL + "/jwks",
				UserinfoEndpoint: ts.URL + "/userinfo",
			})
		case "/jwks":
			json.NewEncoder(w).Encode(JWKSResponse{Keys: []JWK{}})
		case "/userinfo":
			json.NewEncoder(w).Encode(OIDCUserInfo{Sub: "sub-1", Name: "John", Email: "john@example.com"})
		}
	}))
	defer ts.Close()

	p := NewOIDCProvider(SSOConfig{Name: "oidc", Issuer: ts.URL, ClientID: "id", ClientSecret: "secret"})
	if err := p.Discover(); err != nil {
		t.Fatalf("Discover: %v", err)
	}
	user, err := p.GetUser(&OAuth2Token{AccessToken: "tok"})
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if user.ID != "sub-1" || user.Email != "john@example.com" || user.Name != "John" {
		t.Errorf("unexpected user: %+v", user)
	}
}

func TestOIDCProviderDiscoverErrors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid json"))
	}))
	defer ts.Close()

	p := NewOIDCProvider(SSOConfig{Name: "oidc", Issuer: ts.URL, ClientID: "id", ClientSecret: "secret"})
	if err := p.Discover(); err == nil {
		t.Error("expected discovery error for bad JSON")
	}
}

func TestOIDCProviderDiscoverJWKSFailure(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/.well-known/openid-configuration" {
			json.NewEncoder(w).Encode(OIDCDiscoveryDocument{
				Issuer:  ts.URL,
				JWKSUri: ts.URL + "/jwks",
			})
			return
		}
		w.Write([]byte("{invalid"))
	}))
	defer ts.Close()

	p := NewOIDCProvider(SSOConfig{Name: "oidc", Issuer: ts.URL, ClientID: "id", ClientSecret: "secret"})
	if err := p.Discover(); err == nil {
		t.Error("expected error when JWKS fetch fails")
	}
}

func TestSAMLProviderLoadCertFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cert.pem")

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "saml"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err := os.WriteFile(path, pemData, 0o600); err != nil {
		t.Fatal(err)
	}

	cert, err := loadCertFile(path)
	if err != nil {
		t.Fatalf("loadCertFile: %v", err)
	}
	if cert.Subject.CommonName != "saml" {
		t.Errorf("expected CN saml, got %q", cert.Subject.CommonName)
	}

	if _, err := loadCertFile(filepath.Join(dir, "missing.pem")); err == nil {
		t.Error("expected error for missing file")
	}

	badPath := filepath.Join(dir, "bad.pem")
	os.WriteFile(badPath, []byte("not pem"), 0o600)
	if _, err := loadCertFile(badPath); err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestSAMLProviderValidateResponseWithCertFile(t *testing.T) {
	p := NewSAMLProvider(SSOConfig{
		Name:     "saml",
		SSOURL:   "https://ex.com/saml",
		EntityID: "https://ex.com",
	})
	cfg := p.Config()
	if cfg.SSOURL != "https://ex.com/saml" {
		t.Errorf("unexpected SSOURL: %s", cfg.SSOURL)
	}
}
