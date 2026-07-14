package database

import (
	"context"
	"testing"
	"time"
)

func TestPostgreSQL(t *testing.T) {
	db := NewPostgreSQL()

	if db.Name() != "postgresql" {
		t.Errorf("expected name 'postgresql', got %s", db.Name())
	}

	err := db.Connect(&Config{Host: "localhost", Port: 5432})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := db.Exec("INSERT INTO users (name) VALUES ($1)", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", result.RowsAffected)
	}

	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}

	row, err := db.QueryRow("SELECT * FROM users WHERE id = $1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row == nil {
		t.Error("expected row")
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tx.Exec("INSERT INTO users (name) VALUES ($1)", "tx-user")
	tx.Commit()

	err = db.Migrate([]Migration{{Version: 1, Name: "create_users", Up: "CREATE TABLE users (id SERIAL PRIMARY KEY)"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMySQL(t *testing.T) {
	db := NewMySQL()

	if db.Name() != "mysql" {
		t.Errorf("expected name 'mysql', got %s", db.Name())
	}

	err := db.Connect(&Config{Host: "localhost", Port: 3306})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := db.Exec("INSERT INTO users (name) VALUES (?)", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", result.RowsAffected)
	}

	err = db.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLite(t *testing.T) {
	db := NewSQLite()

	if db.Name() != "sqlite" {
		t.Errorf("expected name 'sqlite', got %s", db.Name())
	}

	err := db.Connect(&Config{Database: ":memory:"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", result.RowsAffected)
	}

	err = db.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransaction(t *testing.T) {
	db := NewPostgreSQL()
	db.Connect(&Config{})

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = tx.Exec("INSERT INTO users (name) VALUES ($1)", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows, err := tx.Query("SELECT * FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows == nil {
		t.Error("expected rows")
	}

	err = tx.Rollback()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManager(t *testing.T) {
	m := NewManager()

	pg := NewPostgreSQL()
	mysql := NewMySQL()
	sqlite := NewSQLite()

	m.Register("pg", pg)
	m.Register("mysql", mysql)
	m.Register("sqlite", sqlite)

	got, ok := m.Get("pg")
	if !ok {
		t.Fatal("expected database to be found")
	}
	if got.Name() != "postgresql" {
		t.Errorf("expected 'postgresql', got %s", got.Name())
	}

	names := m.List()
	if len(names) != 3 {
		t.Errorf("expected 3 databases, got %d", len(names))
	}

	m.Remove("pg")
	_, ok = m.Get("pg")
	if ok {
		t.Error("expected database to be removed")
	}
}

func TestManagerConnectAll(t *testing.T) {
	m := NewManager()

	pg := NewPostgreSQL()
	m.Register("pg", pg)

	configs := map[string]*Config{
		"pg": {Host: "localhost", Port: 5432},
	}

	err := m.ConnectAll(configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManagerCloseAll(t *testing.T) {
	m := NewManager()

	pg := NewPostgreSQL()
	pg.Connect(&Config{})
	m.Register("pg", pg)

	err := m.CloseAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPool(t *testing.T) {
	pool := NewPool(10, 5, time.Hour)

	pg := NewPostgreSQL()
	pg.Connect(&Config{})

	pool.Put(pg)

	if pool.Size() != 1 {
		t.Errorf("expected size 1, got %d", pool.Size())
	}

	got := pool.Get()
	if got == nil {
		t.Error("expected connection")
	}
}

func TestHealthCheck(t *testing.T) {
	db := NewPostgreSQL()
	if err := db.HealthCheck(); err == nil {
		t.Error("expected error when not connected")
	}
	db.Connect(&Config{})
	if err := db.HealthCheck(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestContextMethods(t *testing.T) {
	db := NewPostgreSQL()
	db.Connect(&Config{})
	ctx := context.Background()

	result, err := db.ExecContext(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", result.RowsAffected)
	}

	rows, err := db.QueryContext(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows == nil {
		t.Error("expected rows")
	}

	row, err := db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = $1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row == nil {
		t.Error("expected row")
	}

	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tx.ExecContext(ctx, "INSERT INTO users (name) VALUES ($1)", "ctx-user")
	tx.Commit()

	err = db.MigrateContext(ctx, []Migration{{Version: 1, Name: "create_users", Up: "CREATE TABLE users (id SERIAL PRIMARY KEY)"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = db.RollbackContext(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{"valid", Config{Host: "localhost", Port: 5432, User: "user", Database: "db"}, false},
		{"missing host", Config{Port: 5432, User: "user", Database: "db"}, true},
		{"missing port", Config{Host: "localhost", User: "user", Database: "db"}, true},
		{"invalid port", Config{Host: "localhost", Port: -1, User: "user", Database: "db"}, true},
		{"missing user", Config{Host: "localhost", Port: 5432, Database: "db"}, true},
		{"missing database", Config{Host: "localhost", Port: 5432, User: "user"}, true},
		{"invalid sslmode", Config{Host: "localhost", Port: 5432, User: "user", Database: "db", SSLMode: "invalid"}, true},
		{"valid sslmode", Config{Host: "localhost", Port: 5432, User: "user", Database: "db", SSLMode: "disable"}, false},
		{"negative timeout", Config{Host: "localhost", Port: 5432, User: "user", Database: "db", Timeout: -1}, true},
		{"negative max open", Config{Host: "localhost", Port: 5432, User: "user", Database: "db", MaxOpenConns: -1}, true},
		{"negative max idle", Config{Host: "localhost", Port: 5432, User: "user", Database: "db", MaxIdleConns: -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFactory(t *testing.T) {
	tests := []struct {
		driver string
		want   string
	}{
		{"postgresql", "postgresql"},
		{"postgres", "postgresql"},
		{"mysql", "mysql"},
		{"sqlite", "sqlite"},
		{"mock-postgresql", "postgresql"},
		{"mock-mysql", "mysql"},
		{"mock-sqlite", "sqlite"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			db := New(tt.driver)
			if tt.want == "" {
				if db != nil {
					t.Errorf("expected nil for unknown driver")
				}
			} else {
				if db == nil {
					t.Fatalf("expected non-nil for driver %s", tt.driver)
				}
				if db.Name() != tt.want {
					t.Errorf("expected name %q, got %q", tt.want, db.Name())
				}
			}
		})
	}
}

func TestNotConnected(t *testing.T) {
	db := NewPostgreSQL()

	_, err := db.Exec("SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.Query("SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.QueryRow("SELECT 1")
	if err == nil {
		t.Error("expected error when not connected")
	}

	_, err = db.Begin()
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = db.Migrate(nil)
	if err == nil {
		t.Error("expected error when not connected")
	}

	err = db.Ping()
	if err == nil {
		t.Error("expected error when not connected")
	}
}
