//go:build !nosql

package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type RealSupabase struct {
	db     *sql.DB
	config *Config
}

func NewRealSupabase() *RealSupabase {
	return &RealSupabase{}
}

func (s *RealSupabase) Name() string {
	return "supabase"
}

func (s *RealSupabase) Connect(config *Config) error {
	s.config = config

	sslmode := config.SSLMode
	if sslmode == "" {
		sslmode = "require"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.Database, sslmode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	if config.Timeout > 0 {
		db.SetConnMaxLifetime(config.Timeout)
	}

	maxOpen := 25
	if config.MaxOpenConns > 0 {
		maxOpen = config.MaxOpenConns
	}
	db.SetMaxOpenConns(maxOpen)

	maxIdle := 5
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

	s.db = db
	return nil
}

func (s *RealSupabase) defaultContext() (context.Context, context.CancelFunc) {
	if s.config != nil && s.config.Timeout > 0 {
		return context.WithTimeout(context.Background(), s.config.Timeout)
	}
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func (s *RealSupabase) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *RealSupabase) Ping() error {
	if s.db == nil {
		return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.db.PingContext(ctx)
}

func (s *RealSupabase) Exec(query string, args ...any) (Result, error) {
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.ExecContext(ctx, query, args...)
}

func (s *RealSupabase) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
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

func (s *RealSupabase) Query(query string, args ...any) ([]Row, error) {
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.QueryContext(ctx, query, args...)
}

func (s *RealSupabase) QueryContext(ctx context.Context, query string, args ...any) ([]Row, error) {
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

func (s *RealSupabase) QueryRow(query string, args ...any) (Row, error) {
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.QueryRowContext(ctx, query, args...)
}

func (s *RealSupabase) QueryRowContext(ctx context.Context, query string, args ...any) (Row, error) {
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

func (s *RealSupabase) Begin() (Transaction, error) {
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.BeginTx(ctx)
}

func (s *RealSupabase) BeginTx(ctx context.Context) (Transaction, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &RealSupabaseTx{tx: tx}, nil
}

func (s *RealSupabase) Migrate(migrations []Migration) error {
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.MigrateContext(ctx, migrations)
}

func (s *RealSupabase) MigrateContext(ctx context.Context, migrations []Migration) error {
	if s.db == nil {
		return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}

	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS _migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			down_sql TEXT,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	for _, m := range migrations {
		var count int
		err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM _migrations WHERE version = $1", m.Version).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", m.Version, err)
		}
		if count > 0 {
			continue
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.Version, err)
		}

		if _, err := tx.ExecContext(ctx, m.Up); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %d: %w", m.Version, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO _migrations (version, name, down_sql) VALUES ($1, $2, $3)", m.Version, m.Name, m.Down); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.Version, err)
		}
	}

	return nil
}

func (s *RealSupabase) Rollback(version int) error {
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.RollbackContext(ctx, version)
}

func (s *RealSupabase) RollbackContext(ctx context.Context, version int) error {
	if s.db == nil {
		return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}

	var migrations []Migration
	rows, err := s.db.QueryContext(ctx, "SELECT version, name, down_sql FROM _migrations WHERE version > $1 ORDER BY version DESC", version)
	if err != nil {
		return fmt.Errorf("query migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m Migration
		if err := rows.Scan(&m.Version, &m.Name, &m.Down); err != nil {
			return err
		}
		migrations = append(migrations, m)
	}

	for _, m := range migrations {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin rollback %d: %w", m.Version, err)
		}

		if m.Down != "" {
			if _, err := tx.ExecContext(ctx, m.Down); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("execute down migration %d (%s): %w", m.Version, m.Name, err)
			}
		}

		if _, err := tx.ExecContext(ctx, "DELETE FROM _migrations WHERE version = $1", m.Version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("remove migration record %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit rollback %d: %w", m.Version, err)
		}
	}

	return nil
}

func (s *RealSupabase) HealthCheck() error {
	if s.db == nil {
		return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.db.PingContext(ctx)
}

type RealSupabaseTx struct {
	tx *sql.Tx
}

func (t *RealSupabaseTx) Exec(query string, args ...any) (Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return t.ExecContext(ctx, query, args...)
}

func (t *RealSupabaseTx) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	res, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, err
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return Result{RowsAffected: affected, LastInsertID: lastID}, nil
}

func (t *RealSupabaseTx) Query(query string, args ...any) ([]Row, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return t.QueryContext(ctx, query, args...)
}

func (t *RealSupabaseTx) QueryContext(ctx context.Context, query string, args ...any) ([]Row, error) {
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

func (t *RealSupabaseTx) Commit() error {
	return t.tx.Commit()
}

func (t *RealSupabaseTx) Rollback() error {
	return t.tx.Rollback()
}
