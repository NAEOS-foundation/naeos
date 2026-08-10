package gateway

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewReverseProxyNilBackend(t *testing.T) {
	if _, err := NewReverseProxy(nil, &ProxyConfig{}); err == nil {
		t.Fatal("expected error for nil backend")
	}
}

func TestNewReverseProxyEmptyURL(t *testing.T) {
	if _, err := NewReverseProxy(&Backend{Name: "a"}, nil); err == nil {
		t.Fatal("expected error for empty backend URL")
	}
}

func TestNewReverseProxyInvalidURL(t *testing.T) {
	if _, err := NewReverseProxy(&Backend{Name: "a", URL: "://bad"}, nil); err == nil {
		t.Fatal("expected error for invalid backend URL")
	}
}

func TestNewReverseProxyWithConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	rp, err := NewReverseProxy(&Backend{Name: "svc", URL: ts.URL}, &ProxyConfig{Timeout: time.Second})
	if err != nil {
		t.Fatalf("new proxy with config: %v", err)
	}
	if rp.proxy == nil {
		t.Fatal("expected configured proxy")
	}

	rec := httptest.NewRecorder()
	rp.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 through proxy, got %d", rec.Code)
	}
}

func TestNewReverseProxyWithoutConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	rp, err := NewReverseProxy(&Backend{Name: "svc", URL: ts.URL}, nil)
	if err != nil {
		t.Fatalf("new proxy without config: %v", err)
	}
	if rp.config != nil {
		t.Error("expected nil config stored")
	}

	rec := httptest.NewRecorder()
	rp.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 through proxy, got %d", rec.Code)
	}
}

func TestHealthCheckerStartStop(t *testing.T) {
	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer okServer.Close()
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer badServer.Close()

	lb := NewLoadBalancer()
	ok := &Backend{Name: "ok", URL: okServer.URL}
	bad := &Backend{Name: "bad", URL: badServer.URL}
	lb.AddBackend(ok)
	lb.AddBackend(bad)

	hc := NewHealthChecker(20*time.Millisecond, time.Second)
	hc.Start(lb)
	// Starting twice is a no-op (must not panic on later Stop).
	hc.Start(lb)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		lb.mu.RLock()
		okHealthy := ok.Healthy
		badHealthy := bad.Healthy
		lb.mu.RUnlock()
		if okHealthy && !badHealthy {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	lb.mu.RLock()
	okHealthy := ok.Healthy
	badHealthy := bad.Healthy
	lb.mu.RUnlock()
	if !okHealthy {
		t.Error("expected healthy backend to be marked healthy by checker")
	}
	if badHealthy {
		t.Error("expected unhealthy backend to be marked unhealthy by checker")
	}
	if !ok.LastCheck.IsZero() {
		t.Logf("LastCheck set: %v", ok.LastCheck)
	}

	hc.Stop()
	hc.Stop() // idempotent, must not panic
}

func TestHealthCheckerStopNotRunning(t *testing.T) {
	hc := NewHealthChecker(time.Minute, time.Second)
	hc.Stop() // must not panic
}

func TestHealthCheckerStartingWithNoBackends(t *testing.T) {
	hc := NewHealthChecker(10*time.Millisecond, time.Second)
	hc.Start(NewLoadBalancer())
	defer hc.Stop()
	time.Sleep(40 * time.Millisecond)
}

func TestResponseRecorderWriteAndUnwrap(t *testing.T) {
	rec := httptest.NewRecorder()
	rr := &responseRecorder{ResponseWriter: rec, statusCode: http.StatusOK}

	n, err := rr.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes, got %d", n)
	}
	if !rr.written {
		t.Error("expected written flag after Write")
	}
	if rr.Unwrap() != rec {
		t.Error("expected Unwrap to return underlying ResponseWriter")
	}
}

func TestCopyBodyNil(t *testing.T) {
	data, err := CopyBody(nil)
	if err != nil {
		t.Fatalf("copy nil: %v", err)
	}
	if data != nil {
		t.Error("expected nil body for nil reader")
	}
}

func TestCopyBodyReader(t *testing.T) {
	data, err := CopyBody(strings.NewReader("payload"))
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if !bytes.Equal(data, []byte("payload")) {
		t.Errorf("expected payload, got %q", data)
	}
}
