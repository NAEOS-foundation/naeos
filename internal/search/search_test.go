// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package search

import (
	"testing"
	"time"
)

func TestInMemory(t *testing.T) {
	e := NewInMemory()

	if e.Name() != "inmemory" {
		t.Errorf("expected name 'inmemory', got %s", e.Name())
	}

	doc := &Document{
		ID:      "doc1",
		Index:   "articles",
		Title:   "Test Article",
		Content: "This is a test article",
		Tags:    []string{"test", "article"},
	}

	err := e.Index(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if e.Count() != 1 {
		t.Errorf("expected count 1, got %d", e.Count())
	}
}

func TestBulkIndex(t *testing.T) {
	e := NewInMemory()

	docs := []*Document{
		{ID: "doc1", Title: "Article 1", Content: "Content 1"},
		{ID: "doc2", Title: "Article 2", Content: "Content 2"},
		{ID: "doc3", Title: "Article 3", Content: "Content 3"},
	}

	err := e.BulkIndex(docs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if e.Count() != 3 {
		t.Errorf("expected count 3, got %d", e.Count())
	}
}

func TestSearch(t *testing.T) {
	e := NewInMemory()

	e.Index(&Document{ID: "1", Title: "Go Programming", Content: "Learn Go"})
	e.Index(&Document{ID: "2", Title: "Python Programming", Content: "Learn Python"})
	e.Index(&Document{ID: "3", Title: "Go Tips", Content: "Go best practices"})

	result, err := e.Search(&Query{Text: "Go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected 2 results, got %d", result.Total)
	}

	if len(result.Hits) != 2 {
		t.Errorf("expected 2 hits, got %d", len(result.Hits))
	}
}

func TestSearchByIndex(t *testing.T) {
	e := NewInMemory()

	e.Index(&Document{ID: "1", Index: "articles", Title: "Article 1"})
	e.Index(&Document{ID: "2", Index: "posts", Title: "Post 1"})

	result, _ := e.Search(&Query{Index: "articles"})
	if result.Total != 1 {
		t.Errorf("expected 1 result, got %d", result.Total)
	}
}

func TestSearchByTags(t *testing.T) {
	e := NewInMemory()

	e.Index(&Document{ID: "1", Tags: []string{"go", "programming"}})
	e.Index(&Document{ID: "2", Tags: []string{"python", "programming"}})
	e.Index(&Document{ID: "3", Tags: []string{"go", "tips"}})

	result, _ := e.Search(&Query{Tags: []string{"go"}})
	if result.Total != 2 {
		t.Errorf("expected 2 results, got %d", result.Total)
	}
}

func TestSearchLimitOffset(t *testing.T) {
	e := NewInMemory()

	for i := 0; i < 10; i++ {
		e.Index(&Document{ID: string(rune('0' + i)), Title: "Article"})
	}

	result, _ := e.Search(&Query{Limit: 3, Offset: 2})
	if len(result.Hits) != 3 {
		t.Errorf("expected 3 hits, got %d", len(result.Hits))
	}
}

func TestDelete(t *testing.T) {
	e := NewInMemory()

	e.Index(&Document{ID: "doc1", Title: "Test"})
	e.Delete("doc1")

	if e.Count() != 0 {
		t.Errorf("expected count 0, got %d", e.Count())
	}
}

func TestDeleteByQuery(t *testing.T) {
	e := NewInMemory()

	e.Index(&Document{ID: "1", Index: "articles", Title: "Article 1"})
	e.Index(&Document{ID: "2", Index: "posts", Title: "Post 1"})

	deleted, _ := e.DeleteByQuery(&Query{Index: "articles"})
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	if e.Count() != 1 {
		t.Errorf("expected count 1, got %d", e.Count())
	}
}

func TestUpdate(t *testing.T) {
	e := NewInMemory()

	e.Index(&Document{ID: "doc1", Title: "Original"})

	updated := &Document{ID: "doc1", Title: "Updated"}
	err := e.Update("doc1", updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc, _ := e.GetByID("doc1")
	if doc.Title != "Updated" {
		t.Errorf("expected 'Updated', got %s", doc.Title)
	}
}

func TestUpdateNotFound(t *testing.T) {
	e := NewInMemory()

	err := e.Update("nonexistent", &Document{Title: "Test"})
	if err == nil {
		t.Error("expected error for nonexistent document")
	}
}

func TestGetByID(t *testing.T) {
	e := NewInMemory()

	e.Index(&Document{ID: "doc1", Title: "Test"})

	doc, err := e.GetByID("doc1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Title != "Test" {
		t.Errorf("expected 'Test', got %s", doc.Title)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	e := NewInMemory()

	_, err := e.GetByID("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent document")
	}
}

func TestSearchResult(t *testing.T) {
	e := NewInMemory()

	e.Index(&Document{ID: "1", Title: "Test Article", Content: "Test content"})

	start := time.Now()
	result, _ := e.Search(&Query{Text: "Test"})
	took := time.Since(start)

	if result.Total != 1 {
		t.Errorf("expected 1 result, got %d", result.Total)
	}

	if result.Query != "Test" {
		t.Errorf("expected query 'Test', got %s", result.Query)
	}

	if took < 0 {
		t.Error("expected positive duration")
	}
}

func TestManager(t *testing.T) {
	m := NewManager()

	mem := NewInMemory()
	m.Register("inmemory", mem)

	got, ok := m.Get("inmemory")
	if !ok {
		t.Fatal("expected engine to be found")
	}
	if got.Name() != "inmemory" {
		t.Errorf("expected 'inmemory', got %s", got.Name())
	}

	names := m.List()
	if len(names) != 1 {
		t.Errorf("expected 1 engine, got %d", len(names))
	}

	m.Remove("inmemory")
	_, ok = m.Get("inmemory")
	if ok {
		t.Error("expected engine to be removed")
	}
}

func TestManagerCloseAll(t *testing.T) {
	m := NewManager()

	mem := NewInMemory()
	m.Register("inmemory", mem)

	err := m.CloseAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchScore(t *testing.T) {
	e := NewInMemory()

	e.Index(&Document{ID: "1", Title: "Go Programming", Content: "Learn Go"})
	e.Index(&Document{ID: "2", Title: "Python", Content: "Learn Python"})

	result, _ := e.Search(&Query{Text: "Go"})

	if len(result.Hits) > 0 && result.Hits[0].Score <= 0 {
		t.Error("expected positive score")
	}
}
