package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSSOListEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "sso", "list")
	if err != nil {
		t.Fatalf("sso list failed: %v", err)
	}
	if !strings.Contains(output, "No SSO providers configured.") {
		t.Fatalf("expected empty-state message, got %q", output)
	}
}

func TestSSOConfigureLDAPLifecycle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "sso", "configure", "ldap", "--name", "corp-ad", "--host", "ad.example.com", "--base-dn", "dc=example,dc=com")
	if err != nil {
		t.Fatalf("sso configure ldap failed: %v", err)
	}
	if !strings.Contains(output, "configured successfully") {
		t.Fatalf("expected success message, got %q", output)
	}

	output, err = executeCommand(root, "auth", "sso", "list")
	if err != nil {
		t.Fatalf("sso list failed: %v", err)
	}
	if !strings.Contains(output, "corp-ad") || !strings.Contains(output, "ldap") {
		t.Fatalf("expected configured provider in list, got %q", output)
	}

	output, err = executeCommand(root, "auth", "sso", "remove", "corp-ad")
	if err != nil {
		t.Fatalf("sso remove failed: %v", err)
	}
	if !strings.Contains(output, "Removed SSO provider \"corp-ad\"") {
		t.Fatalf("expected remove message, got %q", output)
	}

	output, err = executeCommand(root, "auth", "sso", "list")
	if err != nil {
		t.Fatalf("sso list after remove failed: %v", err)
	}
	if !strings.Contains(output, "No SSO providers configured.") {
		t.Fatalf("expected empty list after remove, got %q", output)
	}
}

func TestSSOConfigureSAML(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "sso", "configure", "saml", "--name", "okta", "--sso-url", "https://idp.example.com/sso", "--entity-id", "https://app.example.com")
	if err != nil {
		t.Fatalf("sso configure saml failed: %v", err)
	}
	if !strings.Contains(output, "configured successfully") {
		t.Fatalf("expected success message, got %q", output)
	}
}

func TestSSOConfigureSAMLBadCert(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "sso", "configure", "saml", "--name", "okta", "--sso-url", "https://idp.example.com/sso", "--entity-id", "https://app.example.com", "--cert-file", filepath.Join(t.TempDir(), "missing.pem"))
	if err == nil {
		t.Fatal("expected error for missing cert file")
	}
}

func TestSSOConfigureOIDC(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "sso", "configure", "oidc", "--name", "azure", "--issuer", "http://127.0.0.1:1/tenant", "--client-id", "cid", "--client-secret", "secret")
	if err != nil {
		t.Fatalf("sso configure oidc should succeed despite failed discovery: %v", err)
	}
	if !strings.Contains(output, "configured successfully") {
		t.Fatalf("expected success message, got %q", output)
	}
}

func TestSSOConfigureOIDCMissingIssuer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "sso", "configure", "oidc", "--name", "azure", "--client-id", "cid", "--client-secret", "secret")
	if err == nil || !strings.Contains(err.Error(), "--issuer is required") {
		t.Fatalf("expected issuer required error, got %v", err)
	}
}

func TestSSOConfigureLDAPMissingHost(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "sso", "configure", "ldap", "--name", "corp-ad", "--base-dn", "dc=example,dc=com")
	if err == nil || !strings.Contains(err.Error(), "--host is required") {
		t.Fatalf("expected host required error, got %v", err)
	}
}

func TestSSOConfigureLDAPMissingBaseDN(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "sso", "configure", "ldap", "--name", "corp-ad", "--host", "ad.example.com")
	if err == nil || !strings.Contains(err.Error(), "--base-dn is required") {
		t.Fatalf("expected base-dn required error, got %v", err)
	}
}

func TestSSOConfigureSAMLMissingSSOURL(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "sso", "configure", "saml", "--name", "okta", "--entity-id", "https://app.example.com")
	if err == nil || !strings.Contains(err.Error(), "--sso-url is required") {
		t.Fatalf("expected sso-url required error, got %v", err)
	}
}

func TestSSOConfigureUnsupportedType(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "sso", "configure", "kerberos", "--name", "k5")
	if err == nil || !strings.Contains(err.Error(), "unsupported provider type") {
		t.Fatalf("expected unsupported type error, got %v", err)
	}
}

func TestSSOConfigureMissingName(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "sso", "configure", "ldap", "--host", "ad.example.com", "--base-dn", "dc=example,dc=com")
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
}

func TestSSORemoveNotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	root := NewRootCommand()
	_, err := executeCommand(root, "auth", "sso", "remove", "ghost")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestSSOListWithSeededConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "naeos")
	writeTestFile(t, configDir, "sso.json", `[
		{"name":"prod-idp","type":"oidc","issuer":"https://issuer.example.com","enabled":true},
		{"name":"legacy-ldap","type":"ldap","host":"ldap.example.com","port":389,"enabled":false}
	]`)

	root := NewRootCommand()
	output, err := executeCommand(root, "auth", "sso", "list")
	if err != nil {
		t.Fatalf("sso list failed: %v", err)
	}
	if !strings.Contains(output, "prod-idp") || !strings.Contains(output, "legacy-ldap") {
		t.Fatalf("expected seeded providers in list, got %q", output)
	}
	if !strings.Contains(output, "yes") || !strings.Contains(output, "no") {
		t.Fatalf("expected enabled flags in list, got %q", output)
	}

	output, err = executeCommand(root, "auth", "sso", "remove", "prod-idp")
	if err != nil {
		t.Fatalf("sso remove failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(configDir, "sso.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "prod-idp") {
		t.Fatalf("expected prod-idp removed from config, got %q", string(data))
	}
}
