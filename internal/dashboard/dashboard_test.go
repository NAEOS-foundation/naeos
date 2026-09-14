// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- Existing tests (kept unchanged) ---

func TestNew(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("failed to create dashboard: %v", err)
	}
	if d == nil {
		t.Error("expected non-nil dashboard")
	}
}

func TestServeHTTP(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("failed to create dashboard: %v", err)
	}

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
	w := httptest.NewRecorder()

	d.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type text/html, got %s", ct)
	}
}

func TestGetStatsDefaults(t *testing.T) {
	statsMu.Lock()
	globalStats = Stats{}
	statsMu.Unlock()

	s := GetStats()
	if s.Projects != 0 {
		t.Errorf("expected 0 projects, got %d", s.Projects)
	}
	if s.Artifacts != 0 {
		t.Errorf("expected 0 artifacts, got %d", s.Artifacts)
	}
	if s.Pipelines != 0 {
		t.Errorf("expected 0 pipelines, got %d", s.Pipelines)
	}
}

func TestRecordPipelineRun(t *testing.T) {
	statsMu.Lock()
	globalStats = Stats{}
	statsMu.Unlock()

	RecordPipelineRun()
	RecordPipelineRun()

	s := GetStats()
	if s.Pipelines != 2 {
		t.Errorf("expected 2 pipelines, got %d", s.Pipelines)
	}
	if s.LastRun == "" {
		t.Error("expected non-empty last run")
	}
}

func TestSetProjects(t *testing.T) {
	SetProjects(5)
	s := GetStats()
	if s.Projects != 5 {
		t.Errorf("expected 5 projects, got %d", s.Projects)
	}
}

func TestSetArtifacts(t *testing.T) {
	SetArtifacts(10)
	s := GetStats()
	if s.Artifacts != 10 {
		t.Errorf("expected 10 artifacts, got %d", s.Artifacts)
	}
}

func TestStatsPersistence(t *testing.T) {
	statsMu.Lock()
	globalStats = Stats{}
	statsFile = ""
	statsMu.Unlock()

	dir := t.TempDir()
	path := filepath.Join(dir, "stats.json")

	SetStatsFile(path)

	SetProjects(3)
	SetArtifacts(7)
	RecordPipelineRun()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read stats file: %v", err)
	}

	var loaded Stats
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal stats: %v", err)
	}
	if loaded.Projects != 3 {
		t.Errorf("expected 3 projects in file, got %d", loaded.Projects)
	}
	if loaded.Artifacts != 7 {
		t.Errorf("expected 7 artifacts in file, got %d", loaded.Artifacts)
	}
	if loaded.Pipelines != 1 {
		t.Errorf("expected 1 pipeline in file, got %d", loaded.Pipelines)
	}

	statsMu.Lock()
	globalStats = Stats{}
	statsFile = ""
	statsMu.Unlock()
}

func TestStatsLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stats.json")

	data := Stats{Projects: 11, Artifacts: 22, Pipelines: 5, LastRun: "2026-01-01T00:00:00Z"}
	b, _ := json.Marshal(data)
	os.WriteFile(path, b, 0o644)

	SetStatsFile(path)
	s := GetStats()
	if s.Projects != 11 {
		t.Errorf("expected 11 projects from file, got %d", s.Projects)
	}
	if s.Artifacts != 22 {
		t.Errorf("expected 22 artifacts from file, got %d", s.Artifacts)
	}

	statsMu.Lock()
	globalStats = Stats{}
	statsFile = ""
	statsMu.Unlock()
}

// --- DashboardConfig tests ---

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.RefreshInterval != 5*time.Second {
		t.Errorf("expected 5s refresh, got %v", cfg.RefreshInterval)
	}
	if cfg.MaxLogEntries != 500 {
		t.Errorf("expected 500 max entries, got %d", cfg.MaxLogEntries)
	}
	if cfg.MaxSubscribers != 64 {
		t.Errorf("expected 64 max subscribers, got %d", cfg.MaxSubscribers)
	}
}

func TestDefaultConfigJSON(t *testing.T) {
	cfg := DefaultConfig()
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	var decoded DashboardConfig
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}
	if decoded.RefreshInterval != cfg.RefreshInterval {
		t.Errorf("roundtrip mismatch for RefreshInterval")
	}
}

// --- ActivityLog tests ---

func TestActivityLogAddAndEntries(t *testing.T) {
	al := NewActivityLog(10)
	entry := al.Add(LevelInfo, "test message")

	if entry.ID != 1 {
		t.Errorf("expected ID 1, got %d", entry.ID)
	}
	if entry.Level != LevelInfo {
		t.Errorf("expected info level, got %s", entry.Level)
	}
	if entry.Message != "test message" {
		t.Errorf("unexpected message: %s", entry.Message)
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if al.Len() != 1 {
		t.Errorf("expected len 1, got %d", al.Len())
	}
}

func TestActivityLogMultipleEntries(t *testing.T) {
	al := NewActivityLog(50)
	al.Add(LevelInfo, "first")
	al.Add(LevelWarn, "second")
	al.Add(LevelError, "third")

	entries := al.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Message != "first" {
		t.Errorf("expected first, got %s", entries[0].Message)
	}
	if entries[2].Level != LevelError {
		t.Errorf("expected error level on third entry")
	}
}

func TestActivityLogMaxLen(t *testing.T) {
	al := NewActivityLog(3)
	for i := 0; i < 5; i++ {
		al.Add(LevelInfo, "msg")
	}

	if al.Len() != 3 {
		t.Errorf("expected 3 entries after overflow, got %d", al.Len())
	}

	entries := al.Entries()
	if entries[0].ID != 3 {
		t.Errorf("expected oldest retained ID to be 3, got %d", entries[0].ID)
	}
}

func TestActivityLogZeroDefault(t *testing.T) {
	al := NewActivityLog(0)
	if al.maxLen != 500 {
		t.Errorf("expected default 500, got %d", al.maxLen)
	}
}

func TestActivityLogFilterByLevel(t *testing.T) {
	al := NewActivityLog(50)
	al.Add(LevelInfo, "a")
	al.Add(LevelWarn, "b")
	al.Add(LevelError, "c")
	al.Add(LevelInfo, "d")

	filtered := al.Filter(LogFilter{Level: LevelWarn})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 warn entry, got %d", len(filtered))
	}
	if filtered[0].Message != "b" {
		t.Errorf("expected 'b', got %s", filtered[0].Message)
	}
}

func TestActivityLogFilterBySubstring(t *testing.T) {
	al := NewActivityLog(50)
	al.Add(LevelInfo, "build started")
	al.Add(LevelInfo, "test passed")
	al.Add(LevelInfo, "build finished")

	filtered := al.Filter(LogFilter{Substring: "build"})
	if len(filtered) != 2 {
		t.Fatalf("expected 2 build entries, got %d", len(filtered))
	}
}

func TestActivityLogFilterByTime(t *testing.T) {
	al := NewActivityLog(50)
	al.Add(LevelInfo, "old")
	time.Sleep(5 * time.Millisecond)
	cutoff := time.Now()
	time.Sleep(5 * time.Millisecond)
	al.Add(LevelInfo, "new")

	filtered := al.Filter(LogFilter{Since: cutoff})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 entry since cutoff, got %d", len(filtered))
	}
	if filtered[0].Message != "new" {
		t.Errorf("expected 'new', got %s", filtered[0].Message)
	}
}

func TestActivityLogFilterLimit(t *testing.T) {
	al := NewActivityLog(50)
	for i := 0; i < 10; i++ {
		al.Add(LevelInfo, "msg")
	}

	filtered := al.Filter(LogFilter{Limit: 3})
	if len(filtered) != 3 {
		t.Fatalf("expected 3 limited entries, got %d", len(filtered))
	}
}

func TestActivityLogClear(t *testing.T) {
	al := NewActivityLog(50)
	al.Add(LevelInfo, "one")
	al.Add(LevelError, "two")
	al.Clear()

	if al.Len() != 0 {
		t.Errorf("expected 0 after clear, got %d", al.Len())
	}
	entries := al.Entries()
	if len(entries) != 0 {
		t.Errorf("expected empty slice after clear")
	}
}

// --- ComponentHealth tests ---

func TestComponentHealthSetAndGet(t *testing.T) {
	ch := NewComponentHealth()
	ch.Set("api", Healthy, "running")

	info, ok := ch.Get("api")
	if !ok {
		t.Fatal("expected to find api component")
	}
	if info.Status != Healthy {
		t.Errorf("expected healthy, got %s", info.Status)
	}
	if info.Message != "running" {
		t.Errorf("expected 'running', got %s", info.Message)
	}
	if info.CheckCount != 1 {
		t.Errorf("expected check count 1, got %d", info.CheckCount)
	}
}

func TestComponentHealthSetUpdates(t *testing.T) {
	ch := NewComponentHealth()
	ch.Set("worker", Healthy, "ok")
	ch.Set("worker", Degraded, "slow")
	ch.Set("worker", Unhealthy, "down")

	info, _ := ch.Get("worker")
	if info.Status != Unhealthy {
		t.Errorf("expected unhealthy, got %s", info.Status)
	}
	if info.CheckCount != 3 {
		t.Errorf("expected check count 3, got %d", info.CheckCount)
	}
	if info.Message != "down" {
		t.Errorf("expected 'down', got %s", info.Message)
	}
}

func TestComponentHealthGetMissing(t *testing.T) {
	ch := NewComponentHealth()
	_, ok := ch.Get("nonexistent")
	if ok {
		t.Error("expected false for missing component")
	}
}

func TestComponentHealthAllSorted(t *testing.T) {
	ch := NewComponentHealth()
	ch.Set("z-comp", Healthy, "")
	ch.Set("a-comp", Healthy, "")
	ch.Set("m-comp", Healthy, "")

	all := ch.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 components, got %d", len(all))
	}
	if all[0].Name != "a-comp" || all[1].Name != "m-comp" || all[2].Name != "z-comp" {
		t.Error("expected alphabetical sort")
	}
}

func TestComponentHealthRemove(t *testing.T) {
	ch := NewComponentHealth()
	ch.Set("db", Healthy, "ok")

	removed := ch.Remove("db")
	if !removed {
		t.Error("expected true for existing removal")
	}
	_, ok := ch.Get("db")
	if ok {
		t.Error("expected false after removal")
	}

	removed = ch.Remove("db")
	if removed {
		t.Error("expected false for double removal")
	}
}

func TestComponentHealthSummary(t *testing.T) {
	ch := NewComponentHealth()
	ch.Set("a", Healthy, "")
	ch.Set("b", Degraded, "")
	ch.Set("c", Unhealthy, "")
	ch.Set("d", Healthy, "")

	total, h, d, u := ch.Summary()
	if total != 4 || h != 2 || d != 1 || u != 1 {
		t.Errorf("unexpected summary: total=%d h=%d d=%d u=%d", total, h, d, u)
	}
}

func TestComponentHealthEmpty(t *testing.T) {
	ch := NewComponentHealth()
	all := ch.All()
	if len(all) != 0 {
		t.Errorf("expected 0, got %d", len(all))
	}
	total, h, d, u := ch.Summary()
	if total != 0 || h != 0 || d != 0 || u != 0 {
		t.Error("expected all-zero summary")
	}
}

// --- EventNotifier tests ---

func TestEventNotifierSubscribe(t *testing.T) {
	en := NewEventNotifier(8)
	id, ch := en.Subscribe()
	if id == "" {
		t.Error("expected non-empty subscriber ID")
	}
	if ch == nil {
		t.Error("expected non-nil channel")
	}
	if en.SubscriberCount() != 1 {
		t.Errorf("expected 1 subscriber, got %d", en.SubscriberCount())
	}
	en.Close()
}

func TestEventNotifierUnsubscribe(t *testing.T) {
	en := NewEventNotifier(8)
	id, _ := en.Subscribe()
	en.Unsubscribe(id)

	if en.SubscriberCount() != 0 {
		t.Errorf("expected 0 after unsub, got %d", en.SubscriberCount())
	}
	en.Close()
}

func TestEventNotifierUnsubscribeInvalid(t *testing.T) {
	en := NewEventNotifier(8)
	en.Unsubscribe("no-such-id")
	if en.SubscriberCount() != 0 {
		t.Errorf("expected 0, got %d", en.SubscriberCount())
	}
	en.Close()
}

func TestEventNotifierBroadcast(t *testing.T) {
	en := NewEventNotifier(8)
	_, ch := en.Subscribe()

	evt := DashboardEvent{
		Type:      EventStatsUpdate,
		Payload:   "test",
		Timestamp: time.Now(),
	}
	en.Broadcast(evt)

	select {
	case received := <-ch:
		if received.Type != EventStatsUpdate {
			t.Errorf("expected stats_update, got %s", received.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timed out waiting for event")
	}
	en.Close()
}

func TestEventNotifierBufferFull(t *testing.T) {
	en := NewEventNotifier(2)
	_, ch := en.Subscribe()

	for i := 0; i < 5; i++ {
		en.Broadcast(DashboardEvent{Type: EventLogEntry, Timestamp: time.Now()})
	}

	if len(ch) != 2 {
		t.Errorf("expected buffer to cap at 2, got %d", len(ch))
	}
	en.Close()
}

func TestEventNotifierDefaultBuffer(t *testing.T) {
	en := NewEventNotifier(0)
	_, ch := en.Subscribe()
	if ch == nil {
		t.Error("expected non-nil channel with default buffer")
	}
	en.Close()
}

func TestEventNotifierClose(t *testing.T) {
	en := NewEventNotifier(8)
	en.Subscribe()
	en.Subscribe()

	en.Close()
	if en.SubscriberCount() != 0 {
		t.Errorf("expected 0 after close, got %d", en.SubscriberCount())
	}
}

// --- APIHandler tests ---

func TestAPIHandlerStats(t *testing.T) {
	statsMu.Lock()
	globalStats = Stats{Projects: 3, Artifacts: 7, Pipelines: 2}
	statsMu.Unlock()
	defer func() {
		statsMu.Lock()
		globalStats = Stats{}
		statsMu.Unlock()
	}()

	d, _ := New()
	ah := NewAPIHandler(d, NewActivityLog(10), NewComponentHealth(), DefaultConfig())

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/stats", nil)
	w := httptest.NewRecorder()
	ah.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.OK {
		t.Error("expected OK true")
	}
	if resp.Data == nil {
		t.Error("expected data")
	}
}

func TestAPIHandlerActivity(t *testing.T) {
	d, _ := New()
	al := NewActivityLog(50)
	al.Add(LevelInfo, "boot")
	al.Add(LevelWarn, "low mem")
	al.Add(LevelError, "crash")

	ah := NewAPIHandler(d, al, NewComponentHealth(), DefaultConfig())

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/activity", nil)
	w := httptest.NewRecorder()
	ah.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.OK {
		t.Error("expected OK true")
	}

	data, _ := json.Marshal(resp.Data)
	var entries []LogEntry
	json.Unmarshal(data, &entries)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestAPIHandlerActivityFilter(t *testing.T) {
	d, _ := New()
	al := NewActivityLog(50)
	al.Add(LevelInfo, "a")
	al.Add(LevelWarn, "b")
	al.Add(LevelError, "c")

	ah := NewAPIHandler(d, al, NewComponentHealth(), DefaultConfig())

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/activity?level=error", nil)
	w := httptest.NewRecorder()
	ah.ServeHTTP(w, req)

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)

	data, _ := json.Marshal(resp.Data)
	var entries []LogEntry
	json.Unmarshal(data, &entries)
	if len(entries) != 1 {
		t.Errorf("expected 1 error entry, got %d", len(entries))
	}
}

func TestAPIHandlerHealth(t *testing.T) {
	d, _ := New()
	ch := NewComponentHealth()
	ch.Set("api", Healthy, "up")
	ch.Set("db", Degraded, "slow")

	ah := NewAPIHandler(d, NewActivityLog(10), ch, DefaultConfig())

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/health", nil)
	w := httptest.NewRecorder()
	ah.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.OK {
		t.Error("expected OK true")
	}
}

func TestAPIHandlerNotFound(t *testing.T) {
	d, _ := New()
	ah := NewAPIHandler(d, NewActivityLog(10), NewComponentHealth(), DefaultConfig())

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/unknown", nil)
	w := httptest.NewRecorder()
	ah.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.OK {
		t.Error("expected OK false")
	}
	if resp.Error != "not found" {
		t.Errorf("expected 'not found' error, got %s", resp.Error)
	}
}

func TestAPIHandlerContentType(t *testing.T) {
	d, _ := New()
	ah := NewAPIHandler(d, NewActivityLog(10), NewComponentHealth(), DefaultConfig())

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/stats", nil)
	w := httptest.NewRecorder()
	ah.ServeHTTP(w, req)

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestAPIResponseTimestamp(t *testing.T) {
	d, _ := New()
	ah := NewAPIHandler(d, NewActivityLog(10), NewComponentHealth(), DefaultConfig())

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/stats", nil)
	w := httptest.NewRecorder()
	ah.ServeHTTP(w, req)

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

// --- Broadcast helpers tests ---

func TestBroadcastStatsUpdate(t *testing.T) {
	statsMu.Lock()
	globalStats = Stats{Projects: 5, Pipelines: 3}
	statsMu.Unlock()
	defer func() {
		statsMu.Lock()
		globalStats = Stats{}
		statsMu.Unlock()
	}()

	en := NewEventNotifier(8)
	_, ch := en.Subscribe()

	BroadcastStatsUpdate(en)

	select {
	case evt := <-ch:
		if evt.Type != EventStatsUpdate {
			t.Errorf("expected stats_update, got %s", evt.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timed out")
	}
	en.Close()
}

func TestBroadcastLogEntry(t *testing.T) {
	en := NewEventNotifier(8)
	_, ch := en.Subscribe()

	entry := LogEntry{ID: 42, Level: LevelInfo, Message: "hello"}
	BroadcastLogEntry(en, entry)

	select {
	case evt := <-ch:
		if evt.Type != EventLogEntry {
			t.Errorf("expected log_entry, got %s", evt.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timed out")
	}
	en.Close()
}

func TestBroadcastHealthChange(t *testing.T) {
	en := NewEventNotifier(8)
	_, ch := en.Subscribe()

	info := ComponentInfo{Name: "api", Status: Unhealthy, Message: "down"}
	BroadcastHealthChange(en, info)

	select {
	case evt := <-ch:
		if evt.Type != EventHealthChange {
			t.Errorf("expected health_change, got %s", evt.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timed out")
	}
	en.Close()
}

// --- Integration: API with activity filter query params ---

func TestAPIActivityContainsFilter(t *testing.T) {
	d, _ := New()
	al := NewActivityLog(50)
	al.Add(LevelInfo, "deploy started")
	al.Add(LevelInfo, "test passed")
	al.Add(LevelInfo, "deploy complete")

	ah := NewAPIHandler(d, al, NewComponentHealth(), DefaultConfig())

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/activity?contains=deploy", nil)
	w := httptest.NewRecorder()
	ah.ServeHTTP(w, req)

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)

	data, _ := json.Marshal(resp.Data)
	var entries []LogEntry
	json.Unmarshal(data, &entries)
	if len(entries) != 2 {
		t.Errorf("expected 2 deploy entries, got %d", len(entries))
	}
}

func TestAPIActivityLimitParam(t *testing.T) {
	d, _ := New()
	al := NewActivityLog(50)
	for i := 0; i < 20; i++ {
		al.Add(LevelInfo, "msg")
	}

	ah := NewAPIHandler(d, al, NewComponentHealth(), DefaultConfig())

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/activity?limit=5", nil)
	w := httptest.NewRecorder()
	ah.ServeHTTP(w, req)

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)

	data, _ := json.Marshal(resp.Data)
	var entries []LogEntry
	json.Unmarshal(data, &entries)
	if len(entries) != 5 {
		t.Errorf("expected 5 entries with limit, got %d", len(entries))
	}
}

func TestAPIHealthSummaryJSON(t *testing.T) {
	d, _ := New()
	ch := NewComponentHealth()
	ch.Set("a", Healthy, "")
	ch.Set("b", Unhealthy, "")

	ah := NewAPIHandler(d, NewActivityLog(10), ch, DefaultConfig())

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/health", nil)
	w := httptest.NewRecorder()
	ah.ServeHTTP(w, req)

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)

	body, _ := json.Marshal(resp.Data)
	if !strings.Contains(string(body), `"total":2`) {
		t.Errorf("expected total:2 in response, got %s", string(body))
	}
}

func TestEventNotifierMultipleSubscribers(t *testing.T) {
	en := NewEventNotifier(8)
	_, ch1 := en.Subscribe()
	_, ch2 := en.Subscribe()

	evt := DashboardEvent{Type: EventStatsUpdate, Timestamp: time.Now()}
	en.Broadcast(evt)

	for _, ch := range []<-chan DashboardEvent{ch1, ch2} {
		select {
		case <-ch:
		case <-time.After(100 * time.Millisecond):
			t.Error("timed out waiting on subscriber")
		}
	}
	en.Close()
}
