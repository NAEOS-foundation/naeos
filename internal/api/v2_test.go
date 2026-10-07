// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/database"
)

func TestV2Version(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	req := httptest.NewRequest(http.MethodGet, "/api/v2/version", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("API-Version"); got != "2" {
		t.Fatalf("API-Version=%q", got)
	}
}

func TestV2PipelinesCursorPagination(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	s.pipelines = []pipelineRun{
		{ID: "1", Project: "one"},
		{ID: "2", Project: "two"},
		{ID: "3", Project: "three"},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v2/pipelines?limit=2", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var first struct {
		Data       []pipelineRun `json:"data"`
		NextCursor string        `json:"next_cursor"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	if len(first.Data) != 2 || first.NextCursor == "" {
		t.Fatalf("unexpected first page: %+v", first)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v2/pipelines?limit=2&cursor="+first.NextCursor, nil)
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	var second struct {
		Data       []pipelineRun `json:"data"`
		NextCursor string        `json:"next_cursor"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &second); err != nil {
		t.Fatal(err)
	}
	if len(second.Data) != 1 || second.NextCursor != "" {
		t.Fatalf("unexpected second page: %+v", second)
	}
}

func TestV2InvalidCursorUsesRFC7807(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	req := httptest.NewRequest(http.MethodGet, "/api/v2/pipelines?cursor=bad!", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("content-type=%q", got)
	}
	var problem RFC7807Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatal(err)
	}
	if problem.Status != http.StatusBadRequest || problem.Title != "Bad Request" {
		t.Fatalf("problem=%+v", problem)
	}
}

func TestV2IdempotencyRequiresKeyForMutation(t *testing.T) {
	s := NewServer(":0", &AuthConfig{Enabled: false})
	req := httptest.NewRequest(http.MethodPost, "/api/v2/test-mutation", nil)
	rec := httptest.NewRecorder()
	s.v2IdempotencyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("mutation handler must not run without Idempotency-Key")
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("content-type=%q", rec.Header().Get("Content-Type"))
	}
}

func TestV2IdempotencyReplaysIdenticalMutationAndRejectsKeyReuse(t *testing.T) {
	v2Idempotency.Lock()
	v2Idempotency.items = make(map[string]idempotencyEntry)
	v2Idempotency.Unlock()

	var calls atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"accepted":true}`))
	})
	middleware := (&Server{}).v2IdempotencyMiddleware(handler)

	first := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v2/test-mutation", nil)
	req.Header.Set("Idempotency-Key", "release-v3-7-test")
	middleware.ServeHTTP(first, req)

	second := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v2/test-mutation", nil)
	req.Header.Set("Idempotency-Key", "release-v3-7-test")
	middleware.ServeHTTP(second, req)

	if first.Code != http.StatusAccepted || second.Code != http.StatusAccepted {
		t.Fatalf("statuses=%d,%d", first.Code, second.Code)
	}
	if first.Body.String() != second.Body.String() || first.Header().Get("Content-Type") != second.Header().Get("Content-Type") {
		t.Fatalf("replay differs: first=%q second=%q", first.Body.String(), second.Body.String())
	}
	if calls.Load() != 1 {
		t.Fatalf("handler calls=%d, want 1", calls.Load())
	}

	third := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v2/test-mutation", nil)
	req.Header.Set("Idempotency-Key", "release-v3-7-test")
	req.Body = http.NoBody
	req.URL.RawQuery = "different=true"
	middleware.ServeHTTP(third, req)
	if third.Code != http.StatusConflict {
		t.Fatalf("key reuse status=%d body=%s", third.Code, third.Body.String())
	}
	if calls.Load() != 1 {
		t.Fatalf("handler calls after conflicting reuse=%d, want 1", calls.Load())
	}
}

func TestV2IdempotencyPersistsAcrossDatabaseReconnect(t *testing.T) {
	dbPath := t.TempDir() + "/idempotency.db"
	db := database.NewRealSQLite()
	if err := db.Connect(&database.Config{Database: dbPath}); err != nil {
		t.Fatal(err)
	}

	var firstCalls atomic.Int32
	s1 := NewServer(":0", &AuthConfig{Enabled: false})
	s1.SetDatabase(db)
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firstCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"created":true}`))
	})

	first := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v2/test-mutation", nil)
	req.Header.Set("Idempotency-Key", "durable-reconnect-test")
	s1.v2IdempotencyMiddleware(handler1).ServeHTTP(first, req)
	if first.Code != http.StatusCreated || firstCalls.Load() != 1 {
		t.Fatalf("first status=%d calls=%d body=%s", first.Code, firstCalls.Load(), first.Body.String())
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db2 := database.NewRealSQLite()
	if err := db2.Connect(&database.Config{Database: dbPath}); err != nil {
		t.Fatal(err)
	}
	defer db2.Close()

	var replayCalls atomic.Int32
	s2 := NewServer(":0", &AuthConfig{Enabled: false})
	s2.SetDatabase(db2)
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		replayCalls.Add(1)
		t.Fatal("durable replay must not execute mutation handler")
	})
	second := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v2/test-mutation", nil)
	req.Header.Set("Idempotency-Key", "durable-reconnect-test")
	s2.v2IdempotencyMiddleware(handler2).ServeHTTP(second, req)

	if second.Code != http.StatusCreated {
		t.Fatalf("replay status=%d body=%s", second.Code, second.Body.String())
	}
	if second.Body.String() != first.Body.String() {
		t.Fatalf("replay body=%q first=%q", second.Body.String(), first.Body.String())
	}
	if replayCalls.Load() != 0 {
		t.Fatalf("replay handler calls=%d, want 0", replayCalls.Load())
	}
}
