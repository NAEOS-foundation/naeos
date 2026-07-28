package supabase

import (
	"os"
	"testing"
)

func TestSaveLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	SetConfigDir(tmpDir)
	t.Cleanup(func() { SetConfigDir(".naeos/supabase") })

	cfg := &Config{
		ProjectRef:     "abc123",
		URL:            "https://abc123.supabase.co",
		AnonKey:        "eyjanonkey",
		ServiceRoleKey: "eyjservicekey",
	}

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if loaded.ProjectRef != cfg.ProjectRef {
		t.Errorf("ProjectRef: got %s, want %s", loaded.ProjectRef, cfg.ProjectRef)
	}
	if loaded.URL != cfg.URL {
		t.Errorf("URL: got %s, want %s", loaded.URL, cfg.URL)
	}
	if loaded.AnonKey != cfg.AnonKey {
		t.Errorf("AnonKey: got %s, want %s", loaded.AnonKey, cfg.AnonKey)
	}
	if loaded.ServiceRoleKey != cfg.ServiceRoleKey {
		t.Errorf("ServiceRoleKey: got %s, want %s", loaded.ServiceRoleKey, cfg.ServiceRoleKey)
	}
}

func TestLoadConfigNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	SetConfigDir(tmpDir)
	t.Cleanup(func() { SetConfigDir(".naeos/supabase") })

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error when config does not exist")
	}
}

func TestNewClient(t *testing.T) {
	cfg := &Config{
		ProjectRef: "test",
		URL:        "https://test.supabase.co",
		AnonKey:    "test-key",
	}

	client := NewClient(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.Config().ProjectRef != "test" {
		t.Errorf("expected ProjectRef 'test', got %s", client.Config().ProjectRef)
	}
}

func TestAuthToken(t *testing.T) {
	client := NewClient(&Config{URL: "https://test.supabase.co"})

	token := client.AuthToken()
	if token != "" {
		t.Errorf("expected empty token, got %s", token)
	}

	client.SetAuthToken("test-token")
	token = client.AuthToken()
	if token != "test-token" {
		t.Errorf("expected 'test-token', got %s", token)
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"abc12345", "abc12345"},
		{"abcdefghijklmnop", "abcd...mnop"},
	}

	for _, tt := range tests {
		got := MaskKey(tt.input)
		if got != tt.want {
			t.Errorf("maskKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Error("expected non-empty config path")
	}
}

func TestConfigFilePath(t *testing.T) {
	_ = configFilePath()
}

func TestSaveConfigPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	SetConfigDir(tmpDir)
	t.Cleanup(func() { SetConfigDir(".naeos/supabase") })

	cfg := &Config{
		ProjectRef: "perm-test",
		URL:        "https://test.supabase.co",
		AnonKey:    "test-key",
	}

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	info, err := os.Stat(configFilePath())
	if err != nil {
		t.Fatalf("Stat config file: %v", err)
	}
	if info.Mode()&0o077 != 0 {
		t.Errorf("expected restricted permissions, got %o", info.Mode()&0o777)
	}
}
