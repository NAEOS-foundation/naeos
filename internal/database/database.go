package database

import (
	"fmt"
	"sync"
	"time"
)

// Database Adapter Interface

type Database interface {
	Name() string
	Connect(config *Config) error
	Close() error
	Ping() error
	Exec(query string, args ...interface{}) (Result, error)
	Query(query string, args ...interface{}) ([]Row, error)
	QueryRow(query string, args ...interface{}) (Row, error)
	Begin() (Transaction, error)
	Migrate(migrations []Migration) error
	Rollback(version int) error
}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	Timeout  time.Duration
}

type Result struct {
	RowsAffected int64
	LastInsertID int64
}

type Row map[string]interface{}

type Transaction interface {
	Exec(query string, args ...interface{}) (Result, error)
	Query(query string, args ...interface{}) ([]Row, error)
	Commit() error
	Rollback() error
}

type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// PostgreSQL Adapter

type PostgreSQL struct {
	config       *Config
	connected    bool
	mu           sync.RWMutex
	tables       map[string][]Row
	migrations   []Migration
	lastVersion  int
	txInProgress bool
}

func NewPostgreSQL() *PostgreSQL {
	return &PostgreSQL{
		tables: make(map[string][]Row),
	}
}

func (p *PostgreSQL) Name() string {
	return "postgresql"
}

func (p *PostgreSQL) Connect(config *Config) error {
	p.config = config
	p.connected = true
	return nil
}

func (p *PostgreSQL) Close() error {
	p.connected = false
	return nil
}

func (p *PostgreSQL) Ping() error {
	if !p.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (p *PostgreSQL) Exec(query string, args ...interface{}) (Result, error) {
	if !p.connected {
		return Result{}, fmt.Errorf("not connected")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return Result{RowsAffected: 1}, nil
}

func (p *PostgreSQL) Query(query string, args ...interface{}) ([]Row, error) {
	if !p.connected {
		return nil, fmt.Errorf("not connected")
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return []Row{}, nil
}

func (p *PostgreSQL) QueryRow(query string, args ...interface{}) (Row, error) {
	if !p.connected {
		return nil, fmt.Errorf("not connected")
	}
	return Row{}, nil
}

func (p *PostgreSQL) Begin() (Transaction, error) {
	if !p.connected {
		return nil, fmt.Errorf("not connected")
	}
	p.mu.Lock()
	p.txInProgress = true
	p.mu.Unlock()
	return &PostgreSQLTx{db: p}, nil
}

func (p *PostgreSQL) Migrate(migrations []Migration) error {
	if !p.connected {
		return fmt.Errorf("not connected")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, m := range migrations {
		p.migrations = append(p.migrations, m)
		if m.Version > p.lastVersion {
			p.lastVersion = m.Version
		}
	}
	return nil
}

func (p *PostgreSQL) Rollback(version int) error {
	if !p.connected {
		return fmt.Errorf("not connected")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := len(p.migrations) - 1; i >= 0; i-- {
		if p.migrations[i].Version > version {
			p.migrations = p.migrations[:i]
		}
	}
	if version < p.lastVersion {
		p.lastVersion = version
	}
	return nil
}

func (p *PostgreSQL) MigrationVersion() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastVersion
}

type PostgreSQLTx struct {
	db     *PostgreSQL
	committed bool
}

func (t *PostgreSQLTx) Exec(query string, args ...interface{}) (Result, error) {
	return Result{RowsAffected: 1}, nil
}

func (t *PostgreSQLTx) Query(query string, args ...interface{}) ([]Row, error) {
	return []Row{}, nil
}

func (t *PostgreSQLTx) Commit() error {
	t.committed = true
	t.db.mu.Lock()
	t.db.txInProgress = false
	t.db.mu.Unlock()
	return nil
}

func (t *PostgreSQLTx) Rollback() error {
	t.db.mu.Lock()
	t.db.txInProgress = false
	t.db.mu.Unlock()
	return nil
}

// MySQL Adapter

type MySQL struct {
	config       *Config
	connected    bool
	mu           sync.RWMutex
	tables       map[string][]Row
	migrations   []Migration
	lastVersion  int
	txInProgress bool
}

func NewMySQL() *MySQL {
	return &MySQL{
		tables: make(map[string][]Row),
	}
}

func (m *MySQL) Name() string {
	return "mysql"
}

func (m *MySQL) Connect(config *Config) error {
	m.config = config
	m.connected = true
	return nil
}

func (m *MySQL) Close() error {
	m.connected = false
	return nil
}

func (m *MySQL) Ping() error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (m *MySQL) Exec(query string, args ...interface{}) (Result, error) {
	if !m.connected {
		return Result{}, fmt.Errorf("not connected")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return Result{RowsAffected: 1}, nil
}

func (m *MySQL) Query(query string, args ...interface{}) ([]Row, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return []Row{}, nil
}

func (m *MySQL) QueryRow(query string, args ...interface{}) (Row, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	return Row{}, nil
}

func (m *MySQL) Begin() (Transaction, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	m.mu.Lock()
	m.txInProgress = true
	m.mu.Unlock()
	return &MySQLTx{db: m}, nil
}

func (m *MySQL) Migrate(migrations []Migration) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, migration := range migrations {
		m.migrations = append(m.migrations, migration)
		if migration.Version > m.lastVersion {
			m.lastVersion = migration.Version
		}
	}
	return nil
}

func (m *MySQL) Rollback(version int) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.migrations) - 1; i >= 0; i-- {
		if m.migrations[i].Version > version {
			m.migrations = m.migrations[:i]
		}
	}
	if version < m.lastVersion {
		m.lastVersion = version
	}
	return nil
}

func (m *MySQL) MigrationVersion() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastVersion
}

type MySQLTx struct {
	db     *MySQL
	committed bool
}

func (t *MySQLTx) Exec(query string, args ...interface{}) (Result, error) {
	return Result{RowsAffected: 1}, nil
}

func (t *MySQLTx) Query(query string, args ...interface{}) ([]Row, error) {
	return []Row{}, nil
}

func (t *MySQLTx) Commit() error {
	t.committed = true
	t.db.mu.Lock()
	t.db.txInProgress = false
	t.db.mu.Unlock()
	return nil
}

func (t *MySQLTx) Rollback() error {
	t.db.mu.Lock()
	t.db.txInProgress = false
	t.db.mu.Unlock()
	return nil
}

// SQLite Adapter

type SQLite struct {
	config       *Config
	connected    bool
	mu           sync.RWMutex
	tables       map[string][]Row
	migrations   []Migration
	lastVersion  int
	txInProgress bool
}

func NewSQLite() *SQLite {
	return &SQLite{
		tables: make(map[string][]Row),
	}
}

func (s *SQLite) Name() string {
	return "sqlite"
}

func (s *SQLite) Connect(config *Config) error {
	s.config = config
	s.connected = true
	return nil
}

func (s *SQLite) Close() error {
	s.connected = false
	return nil
}

func (s *SQLite) Ping() error {
	if !s.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (s *SQLite) Exec(query string, args ...interface{}) (Result, error) {
	if !s.connected {
		return Result{}, fmt.Errorf("not connected")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return Result{RowsAffected: 1}, nil
}

func (s *SQLite) Query(query string, args ...interface{}) ([]Row, error) {
	if !s.connected {
		return nil, fmt.Errorf("not connected")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return []Row{}, nil
}

func (s *SQLite) QueryRow(query string, args ...interface{}) (Row, error) {
	if !s.connected {
		return nil, fmt.Errorf("not connected")
	}
	return Row{}, nil
}

func (s *SQLite) Begin() (Transaction, error) {
	if !s.connected {
		return nil, fmt.Errorf("not connected")
	}
	s.mu.Lock()
	s.txInProgress = true
	s.mu.Unlock()
	return &SQLiteTx{db: s}, nil
}

func (s *SQLite) Migrate(migrations []Migration) error {
	if !s.connected {
		return fmt.Errorf("not connected")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, migration := range migrations {
		s.migrations = append(s.migrations, migration)
		if migration.Version > s.lastVersion {
			s.lastVersion = migration.Version
		}
	}
	return nil
}

func (s *SQLite) Rollback(version int) error {
	if !s.connected {
		return fmt.Errorf("not connected")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.migrations) - 1; i >= 0; i-- {
		if s.migrations[i].Version > version {
			s.migrations = s.migrations[:i]
		}
	}
	if version < s.lastVersion {
		s.lastVersion = version
	}
	return nil
}

func (s *SQLite) MigrationVersion() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastVersion
}

type SQLiteTx struct {
	db     *SQLite
	committed bool
}

func (t *SQLiteTx) Exec(query string, args ...interface{}) (Result, error) {
	return Result{RowsAffected: 1}, nil
}

func (t *SQLiteTx) Query(query string, args ...interface{}) ([]Row, error) {
	return []Row{}, nil
}

func (t *SQLiteTx) Commit() error {
	t.committed = true
	t.db.mu.Lock()
	t.db.txInProgress = false
	t.db.mu.Unlock()
	return nil
}

func (t *SQLiteTx) Rollback() error {
	t.db.mu.Lock()
	t.db.txInProgress = false
	t.db.mu.Unlock()
	return nil
}

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
			return fmt.Errorf("failed to close %s: %w", name, err)
		}
	}
	return nil
}

// Connection Pool

type Pool struct {
	maxOpen    int
	maxIdle    int
	maxLifetime time.Duration
	conns      chan Database
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
