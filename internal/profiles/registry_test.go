package profiles

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewRemoteClient(t *testing.T) {
	reg := RemoteRegistry{URL: "https://example.com", APIKey: "test-key"}
	rc := NewRemoteClient(reg)
	if rc == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestRemoteClientPublish(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/profiles" {
			t.Errorf("expected /profiles, got %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %s", auth)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	rc := NewRemoteClient(RemoteRegistry{URL: srv.URL, APIKey: "test-key"})
	err := rc.Publish([]Profile{{ID: "test", Name: "Test Profile"}})
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}
}

func TestRemoteClientPublishError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid profile"}`))
	}))
	defer srv.Close()

	rc := NewRemoteClient(RemoteRegistry{URL: srv.URL})
	err := rc.Publish([]Profile{{ID: "bad"}})
	if err == nil {
		t.Fatal("expected error for bad request")
	}
}

func TestRemoteClientSubscribe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"remote-1","name":"Remote Profile","industry":"tech"}]`))
	}))
	defer srv.Close()

	rc := NewRemoteClient(RemoteRegistry{URL: srv.URL})
	profiles, err := rc.Subscribe()
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].ID != "remote-1" {
		t.Errorf("expected remote-1, got %s", profiles[0].ID)
	}
}

func TestRemoteClientSubscribeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	rc := NewRemoteClient(RemoteRegistry{URL: srv.URL})
	_, err := rc.Subscribe()
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestRegistrySubscribeAndStop(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"sub-profile","name":"Subscribed Profile","industry":"tech"}]`))
	}))
	defer srv.Close()

	reg := NewRegistry()
	sub := reg.Subscribe(RemoteRegistry{URL: srv.URL}, 50*time.Millisecond)
	if sub == nil {
		t.Fatal("expected non-nil subscription")
	}

	time.Sleep(120 * time.Millisecond)

	sub.Stop()

	p, ok := reg.Get("sub-profile")
	if !ok {
		t.Fatal("expected to find subscribed profile")
	}
	if p.Name != "Subscribed Profile" {
		t.Errorf("expected 'Subscribed Profile', got %q", p.Name)
	}
}

func TestSubscriptionStopIdempotent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	reg := NewRegistry()
	sub := reg.Subscribe(RemoteRegistry{URL: srv.URL}, 1*time.Hour)

	sub.Stop()
	sub.Stop()

	// Should not panic
}

func TestRemoteClientPublishServerURL(t *testing.T) {
	rc := NewRemoteClient(RemoteRegistry{URL: "http://invalid.local:1"})
	err := rc.Publish([]Profile{{ID: "x"}})
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestRemoteClientSubscribeServerURL(t *testing.T) {
	rc := NewRemoteClient(RemoteRegistry{URL: "http://invalid.local:1"})
	_, err := rc.Subscribe()
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestRegistryLoadBuiltin(t *testing.T) {
	reg := NewRegistry()
	list := reg.List()
	if len(list) < 5 {
		t.Errorf("expected at least 5 builtin profiles, got %d", len(list))
	}
}

func TestRegistryGet(t *testing.T) {
	reg := NewRegistry()

	p, ok := reg.Get("saas")
	if !ok {
		t.Fatal("expected to find saas profile")
	}
	if p.Name != "SaaS Application" {
		t.Errorf("expected SaaS Application, got %q", p.Name)
	}
	if p.Industry != "technology" {
		t.Errorf("expected technology industry, got %q", p.Industry)
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	reg := NewRegistry()
	_, ok := reg.Get("nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestRegistryRegister(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&Profile{
		ID:          "custom",
		Name:        "Custom Profile",
		Description: "A custom profile",
		Industry:    "custom",
	})

	p, ok := reg.Get("custom")
	if !ok {
		t.Fatal("expected to find custom profile")
	}
	if p.Name != "Custom Profile" {
		t.Errorf("expected Custom Profile, got %q", p.Name)
	}
}

func TestRegistrySearch(t *testing.T) {
	reg := NewRegistry()
	results := reg.Search("saas")
	if len(results) == 0 {
		t.Fatal("expected to find saas profile")
	}

	found := false
	for _, p := range results {
		if strings.Contains(strings.ToLower(p.Name), "saas") {
			found = true
		}
	}
	if !found {
		t.Error("expected saas in search results")
	}
}

func TestRegistryByIndustry(t *testing.T) {
	reg := NewRegistry()
	aiProfiles := reg.ByIndustry("artificial-intelligence")
	if len(aiProfiles) == 0 {
		t.Fatal("expected at least 1 AI profile")
	}
	for _, p := range aiProfiles {
		if p.Industry != "artificial-intelligence" {
			t.Errorf("expected ai industry, got %q", p.Industry)
		}
	}
}

func TestRegistryByIndustryEmpty(t *testing.T) {
	reg := NewRegistry()
	results := reg.ByIndustry("nonexistent")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestToSpecYAML(t *testing.T) {
	reg := NewRegistry()
	p, ok := reg.Get("saas")
	if !ok {
		t.Fatal("expected saas profile")
	}

	spec := reg.ToSpecYAML(p)
	if spec == "" {
		t.Error("expected non-empty spec")
	}
	if !strings.Contains(spec, "modules:") {
		t.Error("expected modules section")
	}
	if !strings.Contains(spec, "services:") {
		t.Error("expected services section")
	}
	if !strings.Contains(spec, "architecture:") {
		t.Error("expected architecture section")
	}
	if !strings.Contains(spec, "testing:") {
		t.Error("expected testing section")
	}
}

func TestProfileFields(t *testing.T) {
	reg := NewRegistry()

	tests := []struct {
		id       string
		modules  int
		services int
	}{
		{"saas", 6, 2},
		{"ai-agent", 7, 3},
		{"fintech", 7, 3},
		{"healthcare", 7, 2},
		{"government", 7, 3},
	}

	for _, tt := range tests {
		p, ok := reg.Get(tt.id)
		if !ok {
			t.Errorf("profile %q not found", tt.id)
			continue
		}
		if len(p.Modules) != tt.modules {
			t.Errorf("profile %s: expected %d modules, got %d", tt.id, tt.modules, len(p.Modules))
		}
		if len(p.Services) != tt.services {
			t.Errorf("profile %s: expected %d services, got %d", tt.id, tt.services, len(p.Services))
		}
	}
}
