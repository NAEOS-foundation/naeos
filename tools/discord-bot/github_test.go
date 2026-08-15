package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/repos/NAEOS-foundation/naeos":
			_, _ = w.Write([]byte(`{
				"html_url": "https://github.com/NAEOS-foundation/naeos",
				"description": "Declarative engineering platform",
				"stargazers_count": 42,
				"forks_count": 7,
				"open_issues_count": 3,
				"archived": false,
				"updated_at": "2026-08-01T00:00:00Z"
			}`))
		case "/repos/NAEOS-foundation/naeos/license":
			_, _ = w.Write([]byte(`{"license":{"spdx_id":"Apache-2.0"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := NewGitHubClient("NAEOS-foundation/naeos", "")
	c.baseURL = srv.URL

	st, err := c.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if st.Stars != 42 || st.Forks != 7 || st.OpenIssues != 3 {
		t.Fatalf("unexpected status fields: %+v", st)
	}
	if st.License != "Apache-2.0" {
		t.Fatalf("unexpected license: %q", st.License)
	}
	if st.Description != "Declarative engineering platform" {
		t.Fatalf("unexpected description: %q", st.Description)
	}
}

func TestGitHubLatestReleaseSkipsDraftsAndPrereleases(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/repos/o/r/releases" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[
			{"tag_name":"v0.1.0","name":"pre","draft":true,"published_at":"2026-01-01T00:00:00Z"},
			{"tag_name":"v1.0.0","name":"rc","draft":false,"prerelease":true,"published_at":"2026-02-01T00:00:00Z"},
			{"tag_name":"v3.0.0","name":"Stable","draft":false,"prerelease":false,"published_at":"2026-08-01T00:00:00Z","html_url":"https://github.com/o/r/releases/tag/v3.0.0","body":"pipeline profiling"}
		]`))
	}))
	defer srv.Close()

	c := NewGitHubClient("o/r", "")
	c.baseURL = srv.URL

	rel, err := c.LatestRelease(context.Background())
	if err != nil {
		t.Fatalf("LatestRelease() error: %v", err)
	}
	if rel.TagName != "v3.0.0" {
		t.Fatalf("expected v3.0.0, got %q", rel.TagName)
	}
	if rel.Body != "pipeline profiling" {
		t.Fatalf("unexpected body: %q", rel.Body)
	}
}

func TestGitHubErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewGitHubClient("o/r", "")
	c.baseURL = srv.URL

	if _, err := c.Status(context.Background()); err == nil {
		t.Fatal("expected error for 404 response")
	}
}
