package configprovider

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseReferencePlain(t *testing.T) {
	t.Parallel()
	ref, err := ParseReference("plain-value")
	if err != nil {
		t.Fatalf("ParseReference: %v", err)
	}
	if ref.Scheme != "" || ref.Value != "plain-value" {
		t.Errorf("unexpected reference: %+v", ref)
	}
}

func TestParseReferenceScheme(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		scheme Scheme
		value string
	}{
		{"env:API_KEY", SchemeEnv, "API_KEY"},
		{"file:/etc/app/key", SchemeFile, "/etc/app/key"},
		{"secret:default/myapp/token", SchemeSecret, "default/myapp/token"},
		{"vault:secret/api#key", SchemeVault, "secret/api#key"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			ref, err := ParseReference(tt.input)
			if err != nil {
				t.Fatalf("ParseReference: %v", err)
			}
			if ref.Scheme != tt.scheme || ref.Value != tt.value {
				t.Errorf("ParseReference(%q) = %+v", tt.input, ref)
			}
		})
	}
}

func TestParseReferenceUnknownScheme(t *testing.T) {
	t.Parallel()
	ref, err := ParseReference("unknown:value")
	if err != nil {
		t.Fatalf("ParseReference: %v", err)
	}
	if ref.Scheme != "" || ref.Value != "unknown:value" {
		t.Errorf("expected plain value, got %+v", ref)
	}
}

func TestReferenceString(t *testing.T) {
	t.Parallel()
	ref := Reference{Scheme: SchemeEnv, Value: "FOO"}
	if ref.String() != "env:FOO" {
		t.Errorf("unexpected: %s", ref.String())
	}
	plain := Reference{Value: "plain"}
	if plain.String() != "plain" {
		t.Errorf("unexpected: %s", plain.String())
	}
}

func TestEnvProviderResolve(t *testing.T) {
	t.Parallel()
	p := NewEnvProvider()
	p.getenv = func(key string) string {
		if key == "MY_VAR" {
			return "my-value"
		}
		return ""
	}
	val, err := p.Resolve(Reference{Scheme: SchemeEnv, Value: "MY_VAR"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if val != "my-value" {
		t.Errorf("expected my-value, got %s", val)
	}
}

func TestEnvProviderMissing(t *testing.T) {
	t.Parallel()
	p := NewEnvProvider()
	p.getenv = func(string) string { return "" }
	_, err := p.Resolve(Reference{Scheme: SchemeEnv, Value: "MISSING"})
	if err == nil {
		t.Error("expected error for missing env var")
	}
}

func TestEnvProviderNameScheme(t *testing.T) {
	t.Parallel()
	p := NewEnvProvider()
	if p.Name() != "env" {
		t.Errorf("expected name env, got %s", p.Name())
	}
	if p.Scheme() != SchemeEnv {
		t.Errorf("unexpected scheme")
	}
	if !p.Connected() {
		t.Error("expected env connected")
	}
}

func TestFileProviderResolve(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(path, []byte("  file-secret\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	p := NewFileProvider()
	val, err := p.Resolve(Reference{Scheme: SchemeFile, Value: path})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if val != "file-secret" {
		t.Errorf("expected file-secret (trimmed), got %q", val)
	}
}

func TestFileProviderMissing(t *testing.T) {
	t.Parallel()
	p := NewFileProvider()
	_, err := p.Resolve(Reference{Scheme: SchemeFile, Value: "/nonexistent/file"})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestK8sSecretProvider(t *testing.T) {
	t.Parallel()
	p := NewK8sSecretProvider()
	p.AddSecret("default", "myapp", "token", "super-secret")
	if !p.Connected() {
		t.Error("expected connected after adding secret")
	}
	val, err := p.Resolve(Reference{Scheme: SchemeSecret, Value: "default/myapp/token"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if val != "super-secret" {
		t.Errorf("expected super-secret, got %s", val)
	}
}

func TestK8sSecretProviderInvalidRef(t *testing.T) {
	t.Parallel()
	p := NewK8sSecretProvider()
	_, err := p.Resolve(Reference{Scheme: SchemeSecret, Value: "toofew"})
	if err == nil {
		t.Error("expected error for invalid ref")
	}
}

func TestK8sSecretProviderMissing(t *testing.T) {
	t.Parallel()
	p := NewK8sSecretProvider()
	_, err := p.Resolve(Reference{Scheme: SchemeSecret, Value: "default/myapp/token"})
	if err == nil {
		t.Error("expected error for missing secret")
	}
}

func TestK8sSecretProviderRemove(t *testing.T) {
	t.Parallel()
	p := NewK8sSecretProvider()
	p.AddSecret("default", "myapp", "token", "v")
	p.RemoveSecret("default", "myapp", "token")
	if p.Connected() {
		t.Error("expected disconnected after remove")
	}
}

func TestVaultProvider(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider()
	p.AddSecret("secret/api", "key", "vault-secret")
	if !p.Connected() {
		t.Error("expected connected")
	}
	val, err := p.Resolve(Reference{Scheme: SchemeVault, Value: "secret/api#key"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if val != "vault-secret" {
		t.Errorf("expected vault-secret, got %s", val)
	}
}

func TestVaultProviderInvalidRef(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider()
	_, err := p.Resolve(Reference{Scheme: SchemeVault, Value: "nohash"})
	if err == nil {
		t.Error("expected error for invalid ref")
	}
}

func TestVaultProviderMissing(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider()
	_, err := p.Resolve(Reference{Scheme: SchemeVault, Value: "secret/x#y"})
	if err == nil {
		t.Error("expected error for missing secret")
	}
}

func TestChainResolve(t *testing.T) {
	t.Parallel()
	env := NewEnvProvider()
	env.getenv = func(key string) string {
		if key == "FOO" {
			return "bar"
		}
		return ""
	}
	chain := NewChain(env, NewFileProvider())
	result, err := chain.Resolve(Reference{Scheme: SchemeEnv, Value: "FOO"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if result.Value != "bar" {
		t.Errorf("expected bar, got %s", result.Value)
	}
	if result.Provider != "env" {
		t.Errorf("expected provider env, got %s", result.Provider)
	}
}

func TestChainResolveNoProvider(t *testing.T) {
	t.Parallel()
	chain := NewChain(NewEnvProvider())
	_, err := chain.Resolve(Reference{Scheme: SchemeVault, Value: "secret/x#y"})
	if err == nil {
		t.Error("expected error for unhandled scheme")
	}
}

func TestChainAddProviders(t *testing.T) {
	t.Parallel()
	chain := NewChain(NewEnvProvider())
	chain.Add(NewFileProvider())
	names := chain.Providers()
	if len(names) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(names))
	}
	if names[0] != "env" || names[1] != "file" {
		t.Errorf("unexpected provider order: %v", names)
	}
}

func TestResolverPlainValue(t *testing.T) {
	t.Parallel()
	chain := NewChain(NewEnvProvider())
	r := NewResolver(chain)
	cfg, err := r.ResolveKey("host", "localhost:8080")
	if err != nil {
		t.Fatalf("ResolveKey: %v", err)
	}
	if cfg.Value != "localhost:8080" || cfg.Provider != "" {
		t.Errorf("unexpected config: %+v", cfg)
	}
}

func TestResolverEnvKey(t *testing.T) {
	t.Parallel()
	env := NewEnvProvider()
	env.getenv = func(key string) string {
		if key == "DB_PASSWORD" {
			return "pw"
		}
		return ""
	}
	chain := NewChain(env, NewFileProvider())
	r := NewResolver(chain)
	cfg, err := r.ResolveKey("dbPassword", "env:DB_PASSWORD")
	if err != nil {
		t.Fatalf("ResolveKey: %v", err)
	}
	if cfg.Value != "pw" || cfg.Provider != "env" || cfg.Scheme != SchemeEnv {
		t.Errorf("unexpected config: %+v", cfg)
	}
}

func TestResolverUnknownSchemePassesThrough(t *testing.T) {
	t.Parallel()
	chain := NewChain(NewEnvProvider())
	r := NewResolver(chain)
	cfg, err := r.ResolveKey("k", "mystery:value")
	if err != nil {
		t.Fatalf("ResolveKey: %v", err)
	}
	if cfg.Value != "mystery:value" {
		t.Errorf("expected pass-through, got %+v", cfg)
	}
}

func TestResolverColonValue(t *testing.T) {
	t.Parallel()
	chain := NewChain(NewEnvProvider())
	r := NewResolver(chain)
	cfg, err := r.ResolveKey("host", "localhost:8080")
	if err != nil {
		t.Fatalf("ResolveKey: %v", err)
	}
	if cfg.Value != "localhost:8080" || cfg.Provider != "" {
		t.Errorf("unexpected config: %+v", cfg)
	}
}

func TestResolverResolveMap(t *testing.T) {
	t.Parallel()
	env := NewEnvProvider()
	env.getenv = func(key string) string {
		if key == "DB_PASSWORD" {
			return "pw"
		}
		return ""
	}
	chain := NewChain(env, NewFileProvider())
	r := NewResolver(chain)

	out, err := r.ResolveMap(map[string]string{
		"host":       "localhost",
		"dbPassword": "env:DB_PASSWORD",
	})
	if err != nil {
		t.Fatalf("ResolveMap: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(out))
	}
	if out[0].Key != "dbPassword" || out[0].Provider != "env" {
		t.Errorf("unexpected first config: %+v", out[0])
	}
	if out[1].Key != "host" || out[1].Value != "localhost" {
		t.Errorf("unexpected second config: %+v", out[1])
	}
}

func TestResolverResolveMapError(t *testing.T) {
	t.Parallel()
	chain := NewChain(NewEnvProvider())
	r := NewResolver(chain)
	_, err := r.ResolveMap(map[string]string{
		"good": "plain",
		"bad":  "vault:secret/x#y",
	})
	if err == nil {
		t.Error("expected error for unresolvable value")
	}
}

func TestConcurrentChainResolve(t *testing.T) {
	env := NewEnvProvider()
	env.getenv = func(key string) string { return "v" }
	chain := NewChain(env)
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				if _, err := chain.Resolve(Reference{Scheme: SchemeEnv, Value: "K"}); err != nil {
					t.Errorf("Resolve: %v", err)
				}
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestConcurrentSecretProvider(t *testing.T) {
	p := NewK8sSecretProvider()
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			p.AddSecret("ns", "app", "key", "value")
			p.Resolve(Reference{Scheme: SchemeSecret, Value: "ns/app/key"})
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}