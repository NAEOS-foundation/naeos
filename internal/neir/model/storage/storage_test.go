package storage

import "testing"

func TestStorageTypeConstants(t *testing.T) {
	tests := []struct {
		st   StorageType
		want string
	}{
		{TypeSQL, "sql"},
		{TypeNoSQL, "nosql"},
		{TypeFile, "file"},
		{TypeCache, "cache"},
		{TypeQueue, "queue"},
		{TypeBlob, "blob"},
	}
	for _, tt := range tests {
		if string(tt.st) != tt.want {
			t.Errorf("StorageType(%s) = %s, want %s", tt.want, string(tt.st), tt.want)
		}
	}
}

func TestStorage_ZeroValue(t *testing.T) {
	var s Storage
	if s.Name != "" {
		t.Error("expected empty Name")
	}
	if s.Type != "" {
		t.Error("expected empty Type")
	}
	if s.Collections != nil {
		t.Error("expected nil Collections")
	}
}

func TestStorage_Full(t *testing.T) {
	s := Storage{
		Name:       "users-db",
		Type:       TypeSQL,
		Provider:   "postgres",
		Connection: "postgres://localhost:5432",
		Collections: []Collection{
			{Name: "users", Schema: map[string]string{"id": "uuid", "name": "text"}},
		},
		Attributes: map[string]string{"key": "val"},
	}
	if s.Name != "users-db" {
		t.Errorf("expected users-db, got %s", s.Name)
	}
	if s.Type != TypeSQL {
		t.Errorf("expected sql, got %s", s.Type)
	}
	if s.Provider != "postgres" {
		t.Errorf("expected postgres, got %s", s.Provider)
	}
	if len(s.Collections) != 1 {
		t.Errorf("expected 1 collection, got %d", len(s.Collections))
	}
	if s.Collections[0].Schema["id"] != "uuid" {
		t.Errorf("expected uuid, got %s", s.Collections[0].Schema["id"])
	}
}
