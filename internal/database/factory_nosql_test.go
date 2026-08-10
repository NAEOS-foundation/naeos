//go:build nosql

package database

import (
	"testing"
)

func TestNewNoSQL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		driver string
		want   Database
	}{
		{driver: "mock-postgresql", want: NewPostgreSQL()},
		{driver: "mock-mysql", want: NewMySQL()},
		{driver: "mock-sqlite", want: NewSQLite()},
		{driver: "mock-supabase", want: NewSupabase()},
		{driver: "unknown", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			got := New(tt.driver)
			if tt.want == nil && got != nil {
				t.Errorf("New(%q) = %v, want nil", tt.driver, got)
			}
		})
	}
}

func TestNewFromConfigNoSQL(t *testing.T) {
	t.Parallel()

	t.Run("valid driver", func(t *testing.T) {
		db, err := NewFromConfig("mock-sqlite", &Config{
			Host: "localhost",
			Port: 5432,
			User: "test",
		})
		if err != nil {
			t.Fatalf("NewFromConfig() error = %v", err)
		}
		if db == nil {
			t.Fatal("NewFromConfig() returned nil")
		}
	})

	t.Run("unsupported driver", func(t *testing.T) {
		_, err := NewFromConfig("unknown", &Config{
			Host: "localhost",
			Port: 5432,
			User: "test",
		})
		if err == nil {
			t.Fatal("expected error for unsupported driver")
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		_, err := NewFromConfig("mock-sqlite", &Config{})
		if err == nil {
			t.Fatal("expected error for empty config")
		}
	})
}
