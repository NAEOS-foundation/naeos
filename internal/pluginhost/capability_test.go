// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pluginhost

import (
	"context"
	"strings"
	"testing"
)

func TestCapabilityBoundaryFailsClosed(t *testing.T) {
	b, err := NewCapabilityBoundary(map[string][]string{
		"reporter": {"fs.read"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Authorize("reporter", "read", []string{"fs.read"}); err != nil {
		t.Fatalf("expected granted capability: %v", err)
	}
	err = b.Authorize("reporter", "write", []string{"fs.write"})
	if err == nil || !strings.Contains(err.Error(), "ungranted") {
		t.Fatalf("expected ungranted capability denial, got %v", err)
	}
}

func TestCapabilityBoundaryRejectsWildcards(t *testing.T) {
	if _, err := NewCapabilityBoundary(map[string][]string{"p": {"fs.*"}}); err == nil {
		t.Fatal("expected wildcard capability to be rejected")
	}
}

func TestManagerDeniesPrivilegedPluginActionWithoutGrant(t *testing.T) {
	m := NewManager(t.TempDir())
	p := newStubPlugin("privileged")
	if err := m.Register(p); err != nil {
		t.Fatal(err)
	}
	m.config.Plugins = []PluginInfo{{
		Name: "privileged",
		ActionCapabilities: map[string][]string{
			"read-secret": {"secret.read"},
		},
	}}
	if _, err := m.Execute(context.Background(), "privileged", "read-secret", nil); err == nil {
		t.Fatal("expected privileged action to be denied without a grant")
	}
}

func TestManagerAllowsPrivilegedPluginActionWithExplicitGrant(t *testing.T) {
	m := NewManager(t.TempDir())
	p := newStubPlugin("privileged")
	if err := m.Register(p); err != nil {
		t.Fatal(err)
	}
	m.config.Plugins = []PluginInfo{{
		Name: "privileged",
		ActionCapabilities: map[string][]string{
			"read-secret": {"secret.read"},
		},
	}}
	b, err := NewCapabilityBoundary(map[string][]string{
		"privileged": {"secret.read"},
	})
	if err != nil {
		t.Fatal(err)
	}
	m.SetCapabilityBoundary(b)
	if _, err := m.Execute(context.Background(), "privileged", "read-secret", nil); err != nil {
		t.Fatalf("expected explicitly granted action to execute: %v", err)
	}
}
