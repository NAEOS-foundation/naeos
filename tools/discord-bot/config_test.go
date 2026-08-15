package main

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	base := DefaultConfig()
	base.Token = "token"
	base.AppID = "app"

	tt := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{name: "valid", mutate: func(c *Config) {}, wantErr: false},
		{name: "missing token", mutate: func(c *Config) { c.Token = "" }, wantErr: true},
		{name: "missing app id", mutate: func(c *Config) { c.AppID = "" }, wantErr: true},
		{name: "empty repo", mutate: func(c *Config) { c.Repo = "" }, wantErr: true},
		{name: "zero poll interval", mutate: func(c *Config) { c.ReleasePollInterval = 0 }, wantErr: true},
		{name: "negative poll interval", mutate: func(c *Config) { c.ReleasePollInterval = -time.Second }, wantErr: true},
		{name: "empty state file", mutate: func(c *Config) { c.StateFile = "" }, wantErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg := base
			tc.mutate(&cfg)
			err := cfg.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Repo != "NAEOS-foundation/naeos" {
		t.Fatalf("unexpected default repo: %q", cfg.Repo)
	}
	if cfg.ReleasePollInterval <= 0 {
		t.Fatalf("expected positive poll interval, got %v", cfg.ReleasePollInterval)
	}
}
