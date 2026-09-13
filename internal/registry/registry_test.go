// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"sync"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r.Count() != 0 {
		t.Fatalf("expected 0 entries, got %d", r.Count())
	}
}

func TestRegister(t *testing.T) {
	r := NewRegistry()
	err := r.Register("test-service", "service-component")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Count() != 1 {
		t.Fatalf("expected 1 entry, got %d", r.Count())
	}
}

func TestRegisterEmptyName(t *testing.T) {
	r := NewRegistry()
	err := r.Register("", "component")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("test", "component")
	err := r.Register("test", "component")
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestRegisterWithMeta(t *testing.T) {
	r := NewRegistry()
	err := r.RegisterWithMeta("svc", "2.0", "core", "comp", map[string]string{"env": "prod"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry, err := r.GetEntry("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Version != "2.0" || entry.Category != "core" || entry.Metadata["env"] != "prod" {
		t.Fatalf("entry fields not set correctly: %+v", entry)
	}
}

func TestResolve(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("test", "component")
	result, err := r.Resolve("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "component" {
		t.Fatalf("expected 'component', got %v", result)
	}
}

func TestResolveNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Resolve("missing")
	if err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestGetEntry(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterWithMeta("test", "1.0.0", "service", "component", map[string]string{"key": "value"})
	entry, err := r.GetEntry("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Name != "test" {
		t.Fatalf("expected name 'test', got %s", entry.Name)
	}
	if entry.Version != "1.0.0" {
		t.Fatalf("expected version '1.0.0', got %s", entry.Version)
	}
	if entry.Category != "service" {
		t.Fatalf("expected category 'service', got %s", entry.Category)
	}
	if entry.Metadata["key"] != "value" {
		t.Fatalf("expected metadata key=value, got %v", entry.Metadata)
	}
}

func TestGetEntryNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.GetEntry("missing")
	if err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestUnregister(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("test", "component")
	err := r.Unregister("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Count() != 0 {
		t.Fatalf("expected 0 entries after unregister, got %d", r.Count())
	}
}

func TestUnregisterNotFound(t *testing.T) {
	r := NewRegistry()
	err := r.Unregister("missing")
	if err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestRegisteredEntries(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("a", "1")
	_ = r.Register("b", "2")
	_ = r.Register("c", "3")
	entries := r.RegisteredEntries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
}

func TestFindByCategory(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterWithMeta("api", "1.0", "service", "api-component", nil)
	_ = r.RegisterWithMeta("db", "1.0", "storage", "db-component", nil)
	_ = r.RegisterWithMeta("web", "2.0", "service", "web-component", nil)

	services := r.FindByCategory("service")
	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}
}

func TestFindByCategoryNone(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterWithMeta("api", "1.0", "service", "comp", nil)
	result := r.FindByCategory("nonexistent")
	if len(result) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result))
	}
}

func TestFindByVersion(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterWithMeta("a", "1.0.0", "", "comp-a", nil)
	_ = r.RegisterWithMeta("b", "2.0.0", "", "comp-b", nil)
	_ = r.RegisterWithMeta("c", "1.0.0", "", "comp-c", nil)

	v1 := r.FindByVersion("1.0.0")
	if len(v1) != 2 {
		t.Fatalf("expected 2 entries with version 1.0.0, got %d", len(v1))
	}
}

func TestFindByVersionNone(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterWithMeta("a", "1.0.0", "", "comp", nil)
	result := r.FindByVersion("9.9.9")
	if len(result) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result))
	}
}

func TestCount(t *testing.T) {
	r := NewRegistry()
	if r.Count() != 0 {
		t.Fatalf("expected 0, got %d", r.Count())
	}
	_ = r.Register("a", "1")
	_ = r.Register("b", "2")
	if r.Count() != 2 {
		t.Fatalf("expected 2, got %d", r.Count())
	}
}

func TestContains(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("test", "component")
	if !r.Contains("test") {
		t.Fatal("expected Contains to return true")
	}
	if r.Contains("missing") {
		t.Fatal("expected Contains to return false")
	}
}

func TestContainsEmpty(t *testing.T) {
	r := NewRegistry()
	if r.Contains("anything") {
		t.Fatal("expected Contains to return false for empty registry")
	}
}

func TestOnRegisterHook(t *testing.T) {
	r := NewRegistry()
	var called bool
	var receivedName string

	r.OnRegister(func(entry *Entry) {
		called = true
		receivedName = entry.Name
	})

	_ = r.Register("hooked", "comp")

	if !called {
		t.Fatal("expected OnRegister hook to be called")
	}
	if receivedName != "hooked" {
		t.Fatalf("expected hook to receive name 'hooked', got %s", receivedName)
	}
}

func TestOnUnregisterHook(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("hooked", "comp")

	var called bool
	var receivedName string

	r.OnUnregister(func(entry *Entry) {
		called = true
		receivedName = entry.Name
	})

	_ = r.Unregister("hooked")

	if !called {
		t.Fatal("expected OnUnregister hook to be called")
	}
	if receivedName != "hooked" {
		t.Fatalf("expected hook to receive name 'hooked', got %s", receivedName)
	}
}

func TestOnResolveHook(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("hooked", "comp")

	var called bool
	var receivedName string

	r.OnResolve(func(entry *Entry) {
		called = true
		receivedName = entry.Name
	})

	_, _ = r.Resolve("hooked")

	if !called {
		t.Fatal("expected OnResolve hook to be called")
	}
	if receivedName != "hooked" {
		t.Fatalf("expected hook to receive name 'hooked', got %s", receivedName)
	}
}

func TestOnResolveHookNotCalledOnMiss(t *testing.T) {
	r := NewRegistry()
	var called bool
	r.OnResolve(func(entry *Entry) {
		called = true
	})
	_, _ = r.Resolve("missing")
	if called {
		t.Fatal("OnResolve should not be called for missing entries")
	}
}

func TestAddTagsToEntry(t *testing.T) {
	e := &Entry{Name: "test"}
	e.AddTags("alpha", "beta")
	if len(e.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(e.Tags))
	}
	if !e.HasTag("alpha") || !e.HasTag("beta") {
		t.Fatal("expected tags alpha and beta")
	}
}

func TestAddTagsToEntryDuplicate(t *testing.T) {
	e := &Entry{Name: "test"}
	e.AddTags("alpha")
	e.AddTags("alpha")
	if len(e.Tags) != 1 {
		t.Fatalf("expected 1 tag after duplicate add, got %d", len(e.Tags))
	}
}

func TestHasTag(t *testing.T) {
	e := &Entry{Name: "test", Tags: []string{"x", "y"}}
	if !e.HasTag("x") {
		t.Fatal("expected HasTag to return true for 'x'")
	}
	if e.HasTag("z") {
		t.Fatal("expected HasTag to return false for 'z'")
	}
}

func TestRemoveTagsFromEntry(t *testing.T) {
	e := &Entry{Name: "test", Tags: []string{"a", "b", "c"}}
	e.RemoveTags("b")
	if len(e.Tags) != 2 {
		t.Fatalf("expected 2 tags after remove, got %d", len(e.Tags))
	}
	if e.HasTag("b") {
		t.Fatal("expected tag 'b' to be removed")
	}
}

func TestRemoveTagsMultiple(t *testing.T) {
	e := &Entry{Name: "test", Tags: []string{"a", "b", "c", "d"}}
	e.RemoveTags("a", "c")
	if len(e.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(e.Tags))
	}
	if e.HasTag("a") || e.HasTag("c") {
		t.Fatal("expected tags a and c to be removed")
	}
}

func TestRegistryAddTags(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("svc", "comp")
	err := r.AddTags("svc", "web", "api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tags, _ := r.GetTags("svc")
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
}

func TestRegistryAddTagsNotFound(t *testing.T) {
	r := NewRegistry()
	err := r.AddTags("missing", "tag")
	if err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestRegistryFindByTag(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("a", "1")
	_ = r.Register("b", "2")
	_ = r.Register("c", "3")
	_ = r.AddTags("a", "web")
	_ = r.AddTags("b", "web")
	_ = r.AddTags("c", "db")

	results := r.FindByTag("web")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestRegistryFindByTagNoMatch(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("a", "1")
	_ = r.AddTags("a", "web")
	results := r.FindByTag("nonexistent")
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestRegistryRemoveTags(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("svc", "comp")
	_ = r.AddTags("svc", "web", "api")
	err := r.RemoveTags("svc", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tags, _ := r.GetTags("svc")
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(tags))
	}
	if tags[0] != "api" {
		t.Fatalf("expected remaining tag 'api', got %s", tags[0])
	}
}

func TestRegistryRemoveTagsNotFound(t *testing.T) {
	r := NewRegistry()
	err := r.RemoveTags("missing", "tag")
	if err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestRegistryGetTags(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("svc", "comp")
	_ = r.AddTags("svc", "a", "b")
	tags, err := r.GetTags("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
}

func TestRegistryGetTagsNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.GetTags("missing")
	if err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestRegistryGetTagsReturnsCopy(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("svc", "comp")
	_ = r.AddTags("svc", "a")
	tags, _ := r.GetTags("svc")
	tags[0] = "modified"
	tags2, _ := r.GetTags("svc")
	if tags2[0] != "a" {
		t.Fatal("GetTags should return a copy, not a reference")
	}
}

func TestFindByPattern(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("service-api", "1")
	_ = r.Register("service-web", "2")
	_ = r.Register("database-pg", "3")

	results := r.FindByPattern("service-*")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestFindByPatternExact(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("exact-name", "1")
	results := r.FindByPattern("exact-name")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestFindByPatternNoMatch(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("service-api", "1")
	results := r.FindByPattern("database-*")
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestFindByPatternQuestionMark(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("svc1", "1")
	_ = r.Register("svc2", "2")
	_ = r.Register("svc10", "3")

	results := r.FindByPattern("svc?")
	if len(results) != 2 {
		t.Fatalf("expected 2 results for svc?, got %d", len(results))
	}
}

func TestSnapshot(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterWithMeta("a", "1.0", "core", "comp-a", map[string]string{"k": "v"})
	_ = r.RegisterWithMeta("b", "2.0", "web", "comp-b", nil)
	_ = r.AddTags("a", "tag1")

	snap := r.Snapshot()
	if len(snap.Entries) != 2 {
		t.Fatalf("expected 2 snapshot entries, got %d", len(snap.Entries))
	}

	var found bool
	for _, se := range snap.Entries {
		if se.Name == "a" {
			found = true
			if se.Version != "1.0" {
				t.Fatalf("expected version 1.0, got %s", se.Version)
			}
			if se.Metadata["k"] != "v" {
				t.Fatalf("expected metadata k=v")
			}
			if len(se.Tags) != 1 || se.Tags[0] != "tag1" {
				t.Fatalf("expected tag1, got %v", se.Tags)
			}
		}
	}
	if !found {
		t.Fatal("expected to find entry 'a' in snapshot")
	}
}

func TestSnapshotEmpty(t *testing.T) {
	r := NewRegistry()
	snap := r.Snapshot()
	if len(snap.Entries) != 0 {
		t.Fatalf("expected 0 snapshot entries, got %d", len(snap.Entries))
	}
}

func TestRestore(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterWithMeta("a", "1.0", "core", "comp-a", map[string]string{"k": "v"})
	_ = r.AddTags("a", "web")

	snap := r.Snapshot()

	r2 := NewRegistry()
	err := r2.Restore(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r2.Count() != 1 {
		t.Fatalf("expected 1 entry after restore, got %d", r2.Count())
	}

	entry, err := r2.GetEntry("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Version != "1.0" || entry.Category != "core" {
		t.Fatalf("entry fields not restored: %+v", entry)
	}
	if entry.Metadata["k"] != "v" {
		t.Fatalf("metadata not restored")
	}

	tags, _ := r2.GetTags("a")
	if len(tags) != 1 || tags[0] != "web" {
		t.Fatalf("tags not restored, got %v", tags)
	}
}

func TestRestoreNil(t *testing.T) {
	r := NewRegistry()
	err := r.Restore(nil)
	if err == nil {
		t.Fatal("expected error for nil snapshot")
	}
}

func TestRestoreOverwritesExisting(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("old", "old-comp")

	snap := &Snapshot{
		Entries: []SnapshotEntry{
			{Name: "new", Version: "1.0", Category: "test"},
		},
	}

	_ = r.Restore(snap)

	if r.Contains("old") {
		t.Fatal("expected 'old' to be removed after restore")
	}
	if !r.Contains("new") {
		t.Fatal("expected 'new' to exist after restore")
	}
}

func TestReplace(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterWithMeta("svc", "1.0", "old", "old-comp", map[string]string{"a": "b"})

	err := r.Replace("svc", "2.0", "new", "new-comp", map[string]string{"x": "y"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry, _ := r.GetEntry("svc")
	if entry.Version != "2.0" {
		t.Fatalf("expected version 2.0, got %s", entry.Version)
	}
	if entry.Category != "new" {
		t.Fatalf("expected category new, got %s", entry.Category)
	}
	if entry.Component != "new-comp" {
		t.Fatalf("expected new-comp, got %v", entry.Component)
	}
	if entry.Metadata["x"] != "y" {
		t.Fatalf("expected metadata x=y, got %v", entry.Metadata)
	}
	if entry.Metadata["a"] == "b" {
		t.Fatal("expected old metadata to be replaced")
	}
}

func TestReplaceNotFound(t *testing.T) {
	r := NewRegistry()
	err := r.Replace("missing", "1.0", "", "comp", nil)
	if err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestReplaceClearsMetadata(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterWithMeta("svc", "1.0", "", "comp", map[string]string{"k": "v"})

	_ = r.Replace("svc", "1.0", "", "comp", nil)

	entry, _ := r.GetEntry("svc")
	if entry.Metadata != nil {
		t.Fatalf("expected nil metadata after replace with nil, got %v", entry.Metadata)
	}
}

func TestFindByMetadata(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterWithMeta("a", "", "", "1", map[string]string{"env": "prod", "region": "us"})
	_ = r.RegisterWithMeta("b", "", "", "2", map[string]string{"env": "staging", "region": "eu"})
	_ = r.RegisterWithMeta("c", "", "", "3", map[string]string{"env": "prod", "region": "eu"})

	results := r.FindByMetadata(map[string]string{"env": "prod"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestFindByMetadataExactMatch(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterWithMeta("a", "", "", "1", map[string]string{"env": "prod", "region": "us"})
	_ = r.RegisterWithMeta("b", "", "", "2", map[string]string{"env": "prod", "region": "eu"})

	results := r.FindByMetadata(map[string]string{"env": "prod", "region": "us"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestFindByMetadataNoMatch(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterWithMeta("a", "", "", "1", map[string]string{"env": "prod"})
	results := r.FindByMetadata(map[string]string{"env": "dev"})
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestFindByMetadataNilMetadata(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("a", "1")
	results := r.FindByMetadata(map[string]string{"k": "v"})
	if len(results) != 0 {
		t.Fatalf("expected 0 results for nil metadata, got %d", len(results))
	}
}

func TestFindByMetadataEmptyCriteria(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("a", "1")
	_ = r.Register("b", "2")
	results := r.FindByMetadata(map[string]string{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results with empty criteria, got %d", len(results))
	}
}

func TestForEach(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("a", "1")
	_ = r.Register("b", "2")
	_ = r.Register("c", "3")

	var count int
	r.ForEach(func(name string, entry *Entry) bool {
		count++
		return true
	})

	if count != 3 {
		t.Fatalf("expected callback to be called 3 times, got %d", count)
	}
}

func TestForEachBreak(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("a", "1")
	_ = r.Register("b", "2")
	_ = r.Register("c", "3")

	var count int
	r.ForEach(func(name string, entry *Entry) bool {
		count++
		return count < 2
	})

	if count != 2 {
		t.Fatalf("expected callback to be called 2 times before break, got %d", count)
	}
}

func TestForEachEmpty(t *testing.T) {
	r := NewRegistry()
	var called bool
	r.ForEach(func(name string, entry *Entry) bool {
		called = true
		return true
	})
	if called {
		t.Fatal("expected callback not to be called on empty registry")
	}
}

func TestForEachConcurrentSafety(t *testing.T) {
	r := NewRegistry()
	for i := 0; i < 100; i++ {
		_ = r.Register(string(rune('a'+i%26))+string(rune('0'+i/26)), i)
	}

	var mu sync.Mutex
	seen := make(map[string]bool)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.ForEach(func(name string, entry *Entry) bool {
				mu.Lock()
				seen[name] = true
				mu.Unlock()
				return true
			})
		}()
	}
	wg.Wait()

	if len(seen) != 100 {
		t.Fatalf("expected 100 unique entries seen, got %d", len(seen))
	}
}

func TestUnregisterCleansTags(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("svc", "comp")
	_ = r.AddTags("svc", "web")
	_ = r.Unregister("svc")

	results := r.FindByTag("web")
	if len(results) != 0 {
		t.Fatalf("expected 0 results after unregister, got %d", len(results))
	}
}

func TestSnapshotPreservesTags(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("a", "1")
	_ = r.AddTags("a", "x", "y", "z")

	snap := r.Snapshot()
	for _, se := range snap.Entries {
		if se.Name == "a" {
			if len(se.Tags) != 3 {
				t.Fatalf("expected 3 tags in snapshot, got %d", len(se.Tags))
			}
		}
	}
}

func TestReplacePreservesTags(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("svc", "old")
	_ = r.AddTags("svc", "important")

	_ = r.Replace("svc", "2.0", "", "new", nil)

	tags, _ := r.GetTags("svc")
	if len(tags) != 1 || tags[0] != "important" {
		t.Fatal("expected tags to be preserved after replace")
	}
}
