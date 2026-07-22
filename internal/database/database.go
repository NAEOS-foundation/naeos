package database

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Database Adapter Interface

type Database interface {
	Name() string
	Connect(config *Config) error
	Close() error
	Ping() error
	Exec(query string, args ...any) (Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (Result, error)
	Query(query string, args ...any) ([]Row, error)
	QueryContext(ctx context.Context, query string, args ...any) ([]Row, error)
	QueryRow(query string, args ...any) (Row, error)
	QueryRowContext(ctx context.Context, query string, args ...any) (Row, error)
	Begin() (Transaction, error)
	BeginTx(ctx context.Context) (Transaction, error)
	Migrate(migrations []Migration) error
	MigrateContext(ctx context.Context, migrations []Migration) error
	Rollback(version int) error
	RollbackContext(ctx context.Context, version int) error
	HealthCheck() error
}

type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	Timeout         time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 {
		return fmt.Errorf("port must be positive, got %d", c.Port)
	}
	if c.User == "" {
		return fmt.Errorf("user is required")
	}
	if c.Database == "" {
		return fmt.Errorf("database is required")
	}
	if c.SSLMode != "" {
		switch c.SSLMode {
		case "disable", "require", "verify-ca", "verify-full":
		default:
			return fmt.Errorf("invalid sslmode %q: must be disable, require, verify-ca, or verify-full", c.SSLMode)
		}
	}
	if c.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative")
	}
	if c.MaxOpenConns < 0 {
		return fmt.Errorf("max_open_conns must be non-negative")
	}
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns must be non-negative")
	}
	if c.ConnMaxLifetime < 0 {
		return fmt.Errorf("conn_max_lifetime must be non-negative")
	}
	if c.ConnMaxIdleTime < 0 {
		return fmt.Errorf("conn_max_idle_time must be non-negative")
	}
	return nil
}

type Result struct {
	RowsAffected int64
	LastInsertID int64
}

type Row map[string]any

type Transaction interface {
	Exec(query string, args ...any) (Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (Result, error)
	Query(query string, args ...any) ([]Row, error)
	QueryContext(ctx context.Context, query string, args ...any) ([]Row, error)
	Commit() error
	Rollback() error
}

type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// BaseDatabase provides shared mock logic for database stubs.

type BaseDatabase struct {
	mu           sync.RWMutex
	config       *Config
	connected    bool
	tables       map[string][]Row
	migrations   []Migration
	lastVersion  int
	txInProgress bool
}

func (b *BaseDatabase) connect(config *Config) {
	b.config = config
	b.connected = true
}

func (b *BaseDatabase) close() {
	b.connected = false
}

func (b *BaseDatabase) notConnectedError() error {
	return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
}

func (b *BaseDatabase) ping() error {
	if !b.connected {
		return b.notConnectedError()
	}
	return nil
}

func (b *BaseDatabase) exec(_ string, _ ...any) (Result, error) {
	if !b.connected {
		return Result{}, b.notConnectedError()
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return Result{RowsAffected: 1}, nil
}

func (b *BaseDatabase) query(_ string, _ ...any) ([]Row, error) {
	if !b.connected {
		return nil, b.notConnectedError()
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return []Row{}, nil
}

func (b *BaseDatabase) queryRow(_ string, _ ...any) (Row, error) {
	if !b.connected {
		return nil, b.notConnectedError()
	}
	return Row{}, nil
}

func (b *BaseDatabase) begin() (Transaction, error) {
	if !b.connected {
		return nil, b.notConnectedError()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.txInProgress = true
	return &BaseTransaction{db: b}, nil
}

func (b *BaseDatabase) migrate(migrations []Migration) error {
	if !b.connected {
		return b.notConnectedError()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, m := range migrations {
		b.migrations = append(b.migrations, m)
		if m.Version > b.lastVersion {
			b.lastVersion = m.Version
		}
	}
	return nil
}

func (b *BaseDatabase) rollback(version int) error {
	if !b.connected {
		return b.notConnectedError()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for i := len(b.migrations) - 1; i >= 0; i-- {
		if b.migrations[i].Version > version {
			b.migrations = b.migrations[:i]
		}
	}
	if version < b.lastVersion {
		b.lastVersion = version
	}
	return nil
}

func (b *BaseDatabase) migrationVersion() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.lastVersion
}

func (b *BaseDatabase) healthCheck() error {
	return b.ping()
}

// BaseTransaction provides shared mock logic for transaction stubs.

type BaseTransaction struct {
	db        *BaseDatabase
	committed bool
}

func (t *BaseTransaction) Exec(query string, args ...any) (Result, error) {
	return Result{RowsAffected: 1}, nil
}

func (t *BaseTransaction) ExecContext(_ context.Context, query string, args ...any) (Result, error) {
	return Result{RowsAffected: 1}, nil
}

func (t *BaseTransaction) Query(query string, args ...any) ([]Row, error) {
	return []Row{}, nil
}

func (t *BaseTransaction) QueryContext(_ context.Context, query string, args ...any) ([]Row, error) {
	return []Row{}, nil
}

func (t *BaseTransaction) Commit() error {
	t.committed = true
	t.db.mu.Lock()
	t.db.txInProgress = false
	t.db.mu.Unlock()
	return nil
}

func (t *BaseTransaction) Rollback() error {
	t.db.mu.Lock()
	t.db.txInProgress = false
	t.db.mu.Unlock()
	return nil
}

// PostgreSQL Adapter

type PostgreSQL struct {
	BaseDatabase
}

func NewPostgreSQL() *PostgreSQL {
	return &PostgreSQL{
		BaseDatabase: BaseDatabase{
			tables: make(map[string][]Row),
		},
	}
}

func (p *PostgreSQL) Name() string { return "postgresql" }

func (p *PostgreSQL) Connect(config *Config) error { p.connect(config); return nil }
func (p *PostgreSQL) Close() error                 { p.close(); return nil }
func (p *PostgreSQL) Ping() error                  { return p.ping() }
func (p *PostgreSQL) HealthCheck() error           { return p.healthCheck() }

func (p *PostgreSQL) Exec(query string, args ...any) (Result, error) {
	return p.exec(query, args...)
}
func (p *PostgreSQL) ExecContext(_ context.Context, query string, args ...any) (Result, error) {
	return p.exec(query, args...)
}

func (p *PostgreSQL) Query(query string, args ...any) ([]Row, error) {
	return p.query(query, args...)
}
func (p *PostgreSQL) QueryContext(_ context.Context, query string, args ...any) ([]Row, error) {
	return p.query(query, args...)
}

func (p *PostgreSQL) QueryRow(query string, args ...any) (Row, error) {
	return p.queryRow(query, args...)
}
func (p *PostgreSQL) QueryRowContext(_ context.Context, query string, args ...any) (Row, error) {
	return p.queryRow(query, args...)
}

func (p *PostgreSQL) Begin() (Transaction, error)                    { return p.begin() }
func (p *PostgreSQL) BeginTx(_ context.Context) (Transaction, error) { return p.begin() }

func (p *PostgreSQL) Migrate(migrations []Migration) error {
	return p.migrate(migrations)
}
func (p *PostgreSQL) MigrateContext(_ context.Context, migrations []Migration) error {
	return p.migrate(migrations)
}

func (p *PostgreSQL) Rollback(version int) error { return p.rollback(version) }
func (p *PostgreSQL) RollbackContext(_ context.Context, version int) error {
	return p.rollback(version)
}

func (p *PostgreSQL) MigrationVersion() int { return p.migrationVersion() }

// MySQL Adapter

type MySQL struct {
	BaseDatabase
}

func NewMySQL() *MySQL {
	return &MySQL{
		BaseDatabase: BaseDatabase{
			tables: make(map[string][]Row),
		},
	}
}

func (m *MySQL) Name() string { return "mysql" }

func (m *MySQL) Connect(config *Config) error { m.connect(config); return nil }
func (m *MySQL) Close() error                 { m.close(); return nil }
func (m *MySQL) Ping() error                  { return m.ping() }
func (m *MySQL) HealthCheck() error           { return m.healthCheck() }

func (m *MySQL) Exec(query string, args ...any) (Result, error) {
	return m.exec(query, args...)
}
func (m *MySQL) ExecContext(_ context.Context, query string, args ...any) (Result, error) {
	return m.exec(query, args...)
}

func (m *MySQL) Query(query string, args ...any) ([]Row, error) {
	return m.query(query, args...)
}
func (m *MySQL) QueryContext(_ context.Context, query string, args ...any) ([]Row, error) {
	return m.query(query, args...)
}

func (m *MySQL) QueryRow(query string, args ...any) (Row, error) {
	return m.queryRow(query, args...)
}
func (m *MySQL) QueryRowContext(_ context.Context, query string, args ...any) (Row, error) {
	return m.queryRow(query, args...)
}

func (m *MySQL) Begin() (Transaction, error)                    { return m.begin() }
func (m *MySQL) BeginTx(_ context.Context) (Transaction, error) { return m.begin() }

func (m *MySQL) Migrate(migrations []Migration) error {
	return m.migrate(migrations)
}
func (m *MySQL) MigrateContext(_ context.Context, migrations []Migration) error {
	return m.migrate(migrations)
}

func (m *MySQL) Rollback(version int) error { return m.rollback(version) }
func (m *MySQL) RollbackContext(_ context.Context, version int) error {
	return m.rollback(version)
}

func (m *MySQL) MigrationVersion() int { return m.migrationVersion() }

// SQLite Adapter

type SQLite struct {
	BaseDatabase
}

func NewSQLite() *SQLite {
	return &SQLite{
		BaseDatabase: BaseDatabase{
			tables: make(map[string][]Row),
		},
	}
}

func (s *SQLite) Name() string { return "sqlite" }

func (s *SQLite) Connect(config *Config) error { s.connect(config); return nil }
func (s *SQLite) Close() error                 { s.close(); return nil }
func (s *SQLite) Ping() error                  { return s.ping() }
func (s *SQLite) HealthCheck() error           { return s.healthCheck() }

func (s *SQLite) Exec(query string, args ...any) (Result, error) {
	return s.exec(query, args...)
}
func (s *SQLite) ExecContext(_ context.Context, query string, args ...any) (Result, error) {
	return s.exec(query, args...)
}

func (s *SQLite) Query(query string, args ...any) ([]Row, error) {
	return s.query(query, args...)
}
func (s *SQLite) QueryContext(_ context.Context, query string, args ...any) ([]Row, error) {
	return s.query(query, args...)
}

func (s *SQLite) QueryRow(query string, args ...any) (Row, error) {
	return s.queryRow(query, args...)
}
func (s *SQLite) QueryRowContext(_ context.Context, query string, args ...any) (Row, error) {
	return s.queryRow(query, args...)
}

func (s *SQLite) Begin() (Transaction, error)                    { return s.begin() }
func (s *SQLite) BeginTx(_ context.Context) (Transaction, error) { return s.begin() }

func (s *SQLite) Migrate(migrations []Migration) error {
	return s.migrate(migrations)
}
func (s *SQLite) MigrateContext(_ context.Context, migrations []Migration) error {
	return s.migrate(migrations)
}

func (s *SQLite) Rollback(version int) error { return s.rollback(version) }
func (s *SQLite) RollbackContext(_ context.Context, version int) error {
	return s.rollback(version)
}

func (s *SQLite) MigrationVersion() int { return s.migrationVersion() }

// Database Manager

type Manager struct {
	databases map[string]Database
	mu        sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		databases: make(map[string]Database),
	}
}

func (m *Manager) Register(name string, db Database) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.databases[name] = db
}

func (m *Manager) Get(name string) (Database, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	db, ok := m.databases[name]
	return db, ok
}

func (m *Manager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.databases, name)
}

func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.databases))
	for name := range m.databases {
		names = append(names, name)
	}
	return names
}

func (m *Manager) ConnectAll(configs map[string]*Config) error {
	for name, config := range configs {
		db, ok := m.Get(name)
		if !ok {
			continue
		}
		if err := db.Connect(config); err != nil {
			slog.Error("database connect failed", "name", name, "error", err)
			return fmt.Errorf("failed to connect to %s: %w", name, err)
		}
	}
	return nil
}

func (m *Manager) CloseAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, db := range m.databases {
		if err := db.Close(); err != nil {
			slog.Error("database close failed", "name", name, "error", err)
			return fmt.Errorf("failed to close %s: %w", name, err)
		}
	}
	return nil
}

// Connection Pool

type Pool struct {
	maxOpen     int
	maxIdle     int
	maxLifetime time.Duration
	conns       chan Database
}

func NewPool(maxOpen, maxIdle int, maxLifetime time.Duration) *Pool {
	return &Pool{
		maxOpen:     maxOpen,
		maxIdle:     maxIdle,
		maxLifetime: maxLifetime,
		conns:       make(chan Database, maxOpen),
	}
}

func (p *Pool) Get() Database {
	select {
	case conn := <-p.conns:
		return conn
	default:
		return nil
	}
}

func (p *Pool) Put(conn Database) {
	select {
	case p.conns <- conn:
	default:
		conn.Close()
	}
}

func (p *Pool) Size() int {
	return len(p.conns)
}
