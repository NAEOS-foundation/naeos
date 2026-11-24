// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package config

import "testing"

func TestLoadDefault(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load default: %v", err)
	}
	if cfg.Addr != DefaultAddr {
		t.Fatalf("Addr = %q, want %q", cfg.Addr, DefaultAddr)
	}
}

func TestLoadValidAddress(t *testing.T) {
	cfg, err := Load("127.0.0.1:9090")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Addr != "127.0.0.1:9090" {
		t.Fatalf("Addr = %q", cfg.Addr)
	}
}

func TestLoadRejectsInvalidAddress(t *testing.T) {
	for _, addr := range []string{"", " :8088", "host", "127.0.0.1:0", "127.0.0.1:65536", "127.0.0.1:abc", " 127.0.0.1:8088"} {
		if addr == "" {
			continue // Empty string intentionally selects the default.
		}
		if _, err := Load(addr); err == nil {
			t.Errorf("Load(%q) expected error", addr)
		}
	}
}

func TestLoadValidHostnameAndIPv6(t *testing.T) {
	for _, addr := range []string{"localhost:8088", "[::1]:8088"} {
		cfg, err := Load(addr)
		if err != nil {
			t.Errorf("Load(%q): %v", addr, err)
			continue
		}
		if cfg.Addr != addr {
			t.Errorf("Load(%q) Addr = %q", addr, cfg.Addr)
		}
	}
}
