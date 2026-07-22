//go:build !nosql

package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type RealMySQL struct {
	db     *sql.DB
	config *Config
}

func NewRealMySQL() *RealMySQL {
	return &RealMySQL{}
}

func (m *RealMySQL) Name() string {
	return "mysql"
}

func (m *RealMySQL) Connect(config *Config) error {
	m.config = config
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
		config.User, config.Password, config.Host, config.Port, config.Database)

	db, err := sql.Open("mysql", dsn)
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

	m.db = db
	return nil
}

func (m *RealMySQL) defaultContext() (context.Context, context.CancelFunc) {
	if m.config != nil && m.config.Timeout > 0 {
		return context.WithTimeout(context.Background(), m.config.Timeout)
	}
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func (m *RealMySQL) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

func (m *RealMySQL) Ping() error {
	if m.db == nil {
		return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	ctx, cancel := m.defaultContext()
	defer cancel()
	return m.db.PingContext(ctx)
}

func (m *RealMySQL) Exec(query string, args ...any) (Result, error) {
	ctx, cancel := m.defaultContext()
	defer cancel()
	return m.ExecContext(ctx, query, args...)
}

func (m *RealMySQL) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	if m.db == nil {
		return Result{}, fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	res, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, err
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return Result{RowsAffected: affected, LastInsertID: lastID}, nil
}

func (m *RealMySQL) Query(query string, args ...any) ([]Row, error) {
	ctx, cancel := m.defaultContext()
	defer cancel()
	return m.QueryContext(ctx, query, args...)
}

func (m *RealMySQL) QueryContext(ctx context.Context, query string, args ...any) ([]Row, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	rows, err := m.db.QueryContext(ctx, query, args...)
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

func (m *RealMySQL) QueryRow(query string, args ...any) (Row, error) {
	ctx, cancel := m.defaultContext()
	defer cancel()
	return m.QueryRowContext(ctx, query, args...)
}

func (m *RealMySQL) QueryRowContext(ctx context.Context, query string, args ...any) (Row, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	rows, err := m.db.QueryContext(ctx, query, args...)
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

func (m *RealMySQL) Begin() (Transaction, error) {
	ctx, cancel := m.defaultContext()
	defer cancel()
	return m.BeginTx(ctx)
}

func (m *RealMySQL) BeginTx(ctx context.Context) (Transaction, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &RealMySQLTx{tx: tx}, nil
}

func (m *RealMySQL) Migrate(migrations []Migration) error {
	ctx, cancel := m.defaultContext()
	defer cancel()
	return m.MigrateContext(ctx, migrations)
}

func (m *RealMySQL) MigrateContext(ctx context.Context, migrations []Migration) error {
	if m.db == nil {
		return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}

	_, err := m.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS _migrations (
			version INT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			down_sql TEXT,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	for _, migration := range migrations {
		var count int
		err := m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM _migrations WHERE version = ?", migration.Version).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", migration.Version, err)
		}
		if count > 0 {
			continue
		}

		tx, err := m.db.BeginTx(ctx, nil)
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

func (m *RealMySQL) Rollback(version int) error {
	ctx, cancel := m.defaultContext()
	defer cancel()
	return m.RollbackContext(ctx, version)
}

func (m *RealMySQL) RollbackContext(ctx context.Context, version int) error {
	if m.db == nil {
		return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}

	var migrations []Migration
	rows, err := m.db.QueryContext(ctx, "SELECT version, name, down_sql FROM _migrations WHERE version > ? ORDER BY version DESC", version)
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
		tx, err := m.db.BeginTx(ctx, nil)
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

func (m *RealMySQL) HealthCheck() error {
	if m.db == nil {
		return fmt.Errorf("database not connected; call Connect() with a valid config before performing operations")
	}
	ctx, cancel := m.defaultContext()
	defer cancel()
	return m.db.PingContext(ctx)
}

type RealMySQLTx struct {
	tx *sql.Tx
}

func (t *RealMySQLTx) Exec(query string, args ...any) (Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return t.ExecContext(ctx, query, args...)
}

func (t *RealMySQLTx) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	res, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, err
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return Result{RowsAffected: affected, LastInsertID: lastID}, nil
}

func (t *RealMySQLTx) Query(query string, args ...any) ([]Row, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return t.QueryContext(ctx, query, args...)
}

func (t *RealMySQLTx) QueryContext(ctx context.Context, query string, args ...any) ([]Row, error) {
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

func (t *RealMySQLTx) Commit() error {
	return t.tx.Commit()
}

func (t *RealMySQLTx) Rollback() error {
	return t.tx.Rollback()
}
