//go:build !nosql

package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type RealSQLite struct {
	db     *sql.DB
	config *Config
}

func NewRealSQLite() *RealSQLite {
	return &RealSQLite{}
}

func (s *RealSQLite) Name() string {
	return "sqlite"
}

func (s *RealSQLite) Connect(config *Config) error {
	s.config = config
	dsn := config.Database
	if dsn == "" {
		dsn = ":memory:"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	if config.Timeout > 0 {
		db.SetConnMaxLifetime(config.Timeout)
	}

	maxOpen := 1
	if config.MaxOpenConns > 0 {
		maxOpen = config.MaxOpenConns
	}
	db.SetMaxOpenConns(maxOpen)

	maxIdle := 1
	if config.MaxIdleConns > 0 {
		maxIdle = config.MaxIdleConns
	}
	db.SetMaxIdleConns(maxIdle)

	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	}
	if config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("ping database: %w", err)
	}

	if dsn != ":memory:" {
		if _, err := db.ExecContext(ctx, "PRAGMA journal_mode=WAL"); err != nil {
			db.Close()
			return fmt.Errorf("set WAL mode: %w", err)
		}
	}

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return fmt.Errorf("set busy timeout: %w", err)
	}

	s.db = db
	return nil
}

func (s *RealSQLite) defaultContext() (context.Context, context.CancelFunc) {
	if s.config != nil && s.config.Timeout > 0 {
		return context.WithTimeout(context.Background(), s.config.Timeout)
	}
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func (s *RealSQLite) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *RealSQLite) Ping() error {
	if s.db == nil {
		return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.db.PingContext(ctx)
}

func (s *RealSQLite) Exec(query string, args ...any) (Result, error) {
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.ExecContext(ctx, query, args...)
}

func (s *RealSQLite) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	if s.db == nil {
		return Result{}, fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, err
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return Result{RowsAffected: affected, LastInsertID: lastID}, nil
}

func (s *RealSQLite) Query(query string, args ...any) ([]Row, error) {
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.QueryContext(ctx, query, args...)
}

func (s *RealSQLite) QueryContext(ctx context.Context, query string, args ...any) ([]Row, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []Row
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		row := make(Row)
		for i, col := range columns {
			row[col] = values[i]
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (s *RealSQLite) QueryRow(query string, args ...any) (Row, error) {
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.QueryRowContext(ctx, query, args...)
}

func (s *RealSQLite) QueryRowContext(ctx context.Context, query string, args ...any) (Row, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return Row{}, nil
	}

	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	row := make(Row)
	for i, col := range columns {
		row[col] = values[i]
	}
	return row, nil
}

func (s *RealSQLite) Begin() (Transaction, error) {
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.BeginTx(ctx)
}

func (s *RealSQLite) BeginTx(ctx context.Context) (Transaction, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &RealSQLiteTx{tx: tx}, nil
}

func (s *RealSQLite) Migrate(migrations []Migration) error {
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.MigrateContext(ctx, migrations)
}

func (s *RealSQLite) MigrateContext(ctx context.Context, migrations []Migration) error {
	if s.db == nil {
		return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}

	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS _migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			down_sql TEXT,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	for _, migration := range migrations {
		var count int
		err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM _migrations WHERE version = ?", migration.Version).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", migration.Version, err)
		}
		if count > 0 {
			continue
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", migration.Version, err)
		}

		if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %d: %w", migration.Version, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO _migrations (version, name, down_sql) VALUES (?, ?, ?)", migration.Version, migration.Name, migration.Down); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %d: %w", migration.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

func (s *RealSQLite) Rollback(version int) error {
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.RollbackContext(ctx, version)
}

func (s *RealSQLite) RollbackContext(ctx context.Context, version int) error {
	if s.db == nil {
		return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}

	var migrations []Migration
	rows, err := s.db.QueryContext(ctx, "SELECT version, name, down_sql FROM _migrations WHERE version > ? ORDER BY version DESC", version)
	if err != nil {
		return fmt.Errorf("query migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var migration Migration
		if err := rows.Scan(&migration.Version, &migration.Name, &migration.Down); err != nil {
			return err
		}
		migrations = append(migrations, migration)
	}

	for _, migration := range migrations {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin rollback %d: %w", migration.Version, err)
		}

		if migration.Down != "" {
			if _, err := tx.ExecContext(ctx, migration.Down); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("execute down migration %d (%s): %w", migration.Version, migration.Name, err)
			}
		}

		if _, err := tx.ExecContext(ctx, "DELETE FROM _migrations WHERE version = ?", migration.Version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("remove migration record %d: %w", migration.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit rollback %d: %w", migration.Version, err)
		}
	}

	return nil
}

func (s *RealSQLite) HealthCheck() error {
	if s.db == nil {
		return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.db.PingContext(ctx)
}

type RealSQLiteTx struct {
	tx *sql.Tx
}

func (t *RealSQLiteTx) Exec(query string, args ...any) (Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return t.ExecContext(ctx, query, args...)
}

func (t *RealSQLiteTx) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	res, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, err
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return Result{RowsAffected: affected, LastInsertID: lastID}, nil
}

func (t *RealSQLiteTx) Query(query string, args ...any) ([]Row, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return t.QueryContext(ctx, query, args...)
}

func (t *RealSQLiteTx) QueryContext(ctx context.Context, query string, args ...any) ([]Row, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []Row
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		row := make(Row)
		for i, col := range columns {
			row[col] = values[i]
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (t *RealSQLiteTx) Commit() error {
	return t.tx.Commit()
}

func (t *RealSQLiteTx) Rollback() error {
	return t.tx.Rollback()
}
