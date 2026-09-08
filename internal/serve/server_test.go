package serve

import (
	"context"
	"fmt"
	"net"
	"net/http"
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
