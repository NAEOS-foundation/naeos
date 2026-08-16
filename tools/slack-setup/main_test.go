package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newTestServer returns an httptest server that records Slack-style calls and
// replies with the given JSON envelopes in order.
func newTestServer(t *testing.T, responses []string) (*httptest.Server, *[]string) {
	t.Helper()
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		method := r.URL.Path
		calls = append(calls, method)
		if len(responses) == 0 {
			http.Error(w, "no response", http.StatusInternalServerError)
			return
		}
		body := responses[0]
		responses = responses[1:]
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	return srv, &calls
}

// withAPIBase overrides apiBase for the duration of the test.
func withAPIBase(t *testing.T, base string) {
	t.Helper()
	orig := apiBase
	apiBase = base
	t.Cleanup(func() { apiBase = orig })
}

func TestApiCallOK(t *testing.T) {
	srv, _ := newTestServer(t, []string{`{"ok":true,"channel":{"id":"C123","name":"general"}}`})
	defer srv.Close()
	withAPIBase(t, srv.URL+"/")

	var out struct {
		Channel struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"channel"`
	}
	err := apiCall(context.Background(), "tok", "conversations.create", url.Values{"name": {"general"}}, &out)
	if err != nil {
		t.Fatalf("apiCall: %v", err)
	}
	if out.Channel.ID != "C123" {
		t.Errorf("got channel %q, want C123", out.Channel.ID)
	}
}

func TestApiCallError(t *testing.T) {
	srv, _ := newTestServer(t, []string{`{"ok":false,"error":"invalid_auth"}`})
	defer srv.Close()
	withAPIBase(t, srv.URL+"/")

	err := apiCall(context.Background(), "bad", "conversations.list", url.Values{}, nil)
	if err == nil {
		t.Fatal("expected error for ok:false")
	}
	if !strings.Contains(err.Error(), "invalid_auth") {
		t.Errorf("got error %q, want it to mention invalid_auth", err)
	}
}

func TestFindChannelID(t *testing.T) {
	srv, _ := newTestServer(t, []string{
		`{"ok":true,"channels":[{"id":"C1","name":"general"},{"id":"C2","name":"help"}]}`,
		`{"ok":true,"channels":[{"id":"C1","name":"general"},{"id":"C2","name":"help"}]}`,
	})
	defer srv.Close()
	withAPIBase(t, srv.URL+"/")

	ctx := context.Background()
	id, err := findChannelID(ctx, "tok", "help")
	if err != nil {
		t.Fatalf("findChannelID: %v", err)
	}
	if id != "C2" {
		t.Errorf("got %q, want C2", id)
	}

	id, err = findChannelID(ctx, "tok", "missing")
	if err != nil {
		t.Fatalf("findChannelID missing: %v", err)
	}
	if id != "" {
		t.Errorf("got %q, want empty", id)
	}
}
