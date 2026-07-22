package search

import (
	"testing"
)

func TestPersistentNewAndIndex(t *testing.T) {
	p, err := NewPersistent(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	doc := &Document{ID: "p1", Title: "Persistent Doc"}
	if err := p.Index(doc); err != nil {
		t.Fatal(err)
	}
	got, err := p.GetByID("p1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Persistent Doc" {
		t.Errorf("expected 'Persistent Doc', got %v", got.Title)
	}
}

func TestPersistentDelete(t *testing.T) {
	p, _ := NewPersistent(t.TempDir())
	p.Index(&Document{ID: "d1"})
	if err := p.Delete("d1"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.GetByID("d1"); err == nil {
		t.Error("expected error after delete")
	}
}

func TestBulkIndexEmpty(t *testing.T) {
	e := NewInMemory()
	if err := e.BulkIndex(nil); err != nil {
		t.Error("expected nil for empty bulk index")
	}
}
