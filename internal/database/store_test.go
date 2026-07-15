package database

import (
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) *ConnectionStore {
	t.Helper()
	dir := t.TempDir()
	return &ConnectionStore{dir: dir}
}

func TestConnectionStore_AddAndList(t *testing.T) {
	store := tempStore(t)

	err := store.Add("testdb", "sqlite", &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Database: "test.db",
	})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	conns, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	if conns[0].Name != "testdb" {
		t.Errorf("expected name testdb, got %s", conns[0].Name)
	}
	if conns[0].Driver != "sqlite" {
		t.Errorf("expected driver sqlite, got %s", conns[0].Driver)
	}
}

func TestConnectionStore_AddDuplicate(t *testing.T) {
	store := tempStore(t)

	err := store.Add("testdb", "sqlite", &Config{Host: "localhost", Port: 5432})
	if err != nil {
		t.Fatalf("first Add failed: %v", err)
	}

	err = store.Add("testdb", "mysql", &Config{Host: "localhost", Port: 3306})
	if err == nil {
		t.Fatal("expected error for duplicate, got nil")
	}
}

func TestConnectionStore_Remove(t *testing.T) {
	store := tempStore(t)

	store.Add("db1", "sqlite", &Config{Host: "localhost", Port: 5432})
	store.Add("db2", "mysql", &Config{Host: "localhost", Port: 3306})

	err := store.Remove("db1")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	conns, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection after remove, got %d", len(conns))
	}
	if conns[0].Name != "db2" {
		t.Errorf("expected remaining connection db2, got %s", conns[0].Name)
	}
}

func TestConnectionStore_RemoveNotFound(t *testing.T) {
	store := tempStore(t)

	err := store.Remove("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent, got nil")
	}
}

func TestConnectionStore_Get(t *testing.T) {
	store := tempStore(t)

	cfg := &Config{Host: "localhost", Port: 5432, User: "admin", Database: "mydb"}
	store.Add("mydb", "postgresql", cfg)

	got, err := store.Get("mydb")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != "mydb" {
		t.Errorf("expected name mydb, got %s", got.Name)
	}
	if got.Driver != "postgresql" {
		t.Errorf("expected driver postgresql, got %s", got.Driver)
	}
	if got.Config.Database != "mydb" {
		t.Errorf("expected database mydb, got %s", got.Config.Database)
	}
}

func TestConnectionStore_GetNotFound(t *testing.T) {
	store := tempStore(t)

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent, got nil")
	}
}

func TestConnectionStore_Persistence(t *testing.T) {
	dir := t.TempDir()

	store1 := &ConnectionStore{dir: dir}
	store1.Add("persistent", "sqlite", &Config{Host: "localhost", Port: 5432, Database: "test.db"})

	store2 := &ConnectionStore{dir: dir}
	conns, err := store2.List()
	if err != nil {
		t.Fatalf("List on new store failed: %v", err)
	}
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection from new store, got %d", len(conns))
	}
	if conns[0].Name != "persistent" {
		t.Errorf("expected name persistent, got %s", conns[0].Name)
	}
}

func TestConnectionStore_FileCreated(t *testing.T) {
	dir := t.TempDir()
	store := &ConnectionStore{dir: dir}

	store.Add("test", "sqlite", &Config{Host: "localhost", Port: 5432})

	path := filepath.Join(dir, connectionsFile)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("connections.json should exist after Add")
	}
}
