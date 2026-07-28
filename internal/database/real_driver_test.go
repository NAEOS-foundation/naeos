//go:build !nosql

package database

import "testing"

func TestRealPostgreSQLCloseNilDB(t *testing.T) {
	t.Parallel()
	db := NewRealPostgreSQL()
	if err := db.Close(); err != nil {
		t.Fatalf("Close nil db: %v", err)
	}
}

func TestRealMySQLCloseNilDB(t *testing.T) {
	t.Parallel()
	db := NewRealMySQL()
	if err := db.Close(); err != nil {
		t.Fatalf("Close nil db: %v", err)
	}
}

func TestRealSQLiteCloseNilDB(t *testing.T) {
	t.Parallel()
	db := NewRealSQLite()
	if err := db.Close(); err != nil {
		t.Fatalf("Close nil db: %v", err)
	}
}
