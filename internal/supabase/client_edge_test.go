package supabase

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDirOverride(t *testing.T) {
	dir := t.TempDir()
	SetConfigDir(dir)
	t.Cleanup(func() { SetConfigDir("") })

	if got := configDir(); got != dir {
		t.Errorf("expected override dir %q, got %q", dir, got)
	}
	if got := DefaultConfigPath(); got != dir {
		t.Errorf("expected DefaultConfigPath %q, got %q", dir, got)
	}
	if got := configFilePath(); got != filepath.Join(dir, configFile) {
		t.Errorf("expected configFilePath under override, got %q", got)
	}
}

func TestConfigDirHome(t *testing.T) {
	SetConfigDir("")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("user home dir: %v", err)
	}
	want := filepath.Join(home, ".naeos/supabase")
	if got := configDir(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestManagementURLCustom(t *testing.T) {
	c := NewClient(&Config{ManagementURL: "https://mgmt.example.com"})
	if got := c.managementURL(); got != "https://mgmt.example.com" {
		t.Errorf("expected custom management URL, got %q", got)
	}
}

func TestManagementURLDefault(t *testing.T) {
	c := NewClient(&Config{})
	if got := c.managementURL(); got != "https://api.supabase.com" {
		t.Errorf("expected default management URL, got %q", got)
	}
}
