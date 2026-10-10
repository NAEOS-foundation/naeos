// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package serve

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func TestServerStartShutdownLifecycle(t *testing.T) {
	addr := freeAddr(t)
	cfg := DefaultConfig()
	cfg.Listeners = []Listener{{Addr: addr, Name: "api", API: true}}

	srv, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- srv.StartWithContext(ctx)
	}()

	// Poll the /healthz endpoint until the server is up.
	base := "http://" + addr
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(base + "/healthz")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 500 {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	resp, err := http.Get(base + "/healthz")
	if err != nil {
		cancel()
		t.Fatalf("server did not come up: %v", err)
	}
	_ = resp.Body.Close()

	// Cancel triggers graceful shutdown and StartWithContext returns.
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected shutdown error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Listeners = []Listener{{Addr: ":0", TLSCert: "cert.pem"}} // missing key
	if _, err := New(cfg); err == nil {
		t.Fatal("expected New to reject partial TLS config")
	}
}

func TestServerServicesHealthEndpoint(t *testing.T) {
	addr := freeAddr(t)
	srv, err := New(&Config{
		Listeners: []Listener{{Addr: addr, Name: "api", API: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.StartWithContext(ctx) }()

	base := "http://" + addr
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(base + "/healthz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("expected /healthz to return 200 via %s", fmt.Sprintf("%s/healthz", base))
}

func TestServerWiresConfiguredSandBaseAuthorizationEndpoint(t *testing.T) {
	dir := t.TempDir()
	configPath := dir + "/sandbase-authz.json"
	ledgerPath := dir + "/authorization-ledger.json"
	config := `{
		"policy": {
			"id": "sandbase-policy",
			"version": 1,
			"status": "active",
			"created_at": "2026-10-10T00:00:00Z",
			"updated_at": "2026-10-10T00:00:00Z",
			"allowed_capabilities": ["tool.execute"]
		},
		"grant": {
			"grant_id": "grant-sandbase",
			"agent_id": "trusted-agent",
			"policy_id": "sandbase-policy",
			"policy_version": 1,
			"capabilities": ["tool.execute"],
			"created_at": "2026-10-10T00:00:00Z",
			"expires_at": "2099-01-01T00:00:00Z",
			"status": "active"
		}
	}`
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NAEOS_SANDBASE_AUTHZ_CONFIG", configPath)
	t.Setenv("NAEOS_SANDBASE_AUTHZ_TOKEN", "integration-test-token")
	t.Setenv("NAEOS_SANDBASE_AUTHZ_LEDGER", ledgerPath)

	cfg := DefaultConfig()
	cfg.Listeners = []Listener{{Addr: freeAddr(t), Name: "api", API: true}}
	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("expected configured SandBase handler to initialize: %v", err)
	}

	body := `{
		"schema":"sandbase.authz/v1",
		"session_id":"session-1",
		"invocation_id":"call-1",
		"capability":"tool.execute",
		"target":"repository.write",
		"arguments_digest":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"policy_context_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"digest_schema":"sandbase.digest/v1"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/sandbase/authorize", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer integration-test-token")
	w := httptest.NewRecorder()
	srv.api.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected configured route to authorize request, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"decision":"allow"`) {
		t.Fatalf("expected allow response from configured route, got %s", w.Body.String())
	}
	if _, err := os.Stat(ledgerPath); err != nil {
		t.Fatalf("expected configured authorization ledger to be persisted: %v", err)
	}
}

func TestServerRejectsSandBaseConfigWithoutBearerToken(t *testing.T) {
	dir := t.TempDir()
	configPath := dir + "/sandbase-authz.json"
	if err := os.WriteFile(configPath, []byte(`{"policy":{"id":"p","version":1,"status":"active"},"grant":{"grant_id":"g","agent_id":"a","policy_id":"p","policy_version":1,"status":"active","capabilities":["tool.execute"]}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NAEOS_SANDBASE_AUTHZ_CONFIG", configPath)
	t.Setenv("NAEOS_SANDBASE_AUTHZ_TOKEN", "")
	cfg := DefaultConfig()
	if _, err := New(cfg); err == nil {
		t.Fatal("expected startup to fail closed when the SandBase bearer token is missing")
	}
}
