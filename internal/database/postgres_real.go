//go:build !nosql

package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type RealPostgreSQL struct {
	db     *sql.DB
	config *Config
}

func NewRealPostgreSQL() *RealPostgreSQL {
	return &RealPostgreSQL{}
}

func (p *RealPostgreSQL) Name() string {
	return "postgresql"
}

func (p *RealPostgreSQL) Connect(config *Config) error {
	p.config = config
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.Database, config.SSLMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	if config.Timeout > 0 {
		db.SetConnMaxLifetime(config.Timeout)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("ping database: %w", err)
	}

	p.db = db
	return nil
}

func (p *RealPostgreSQL) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *RealPostgreSQL) Ping() error {
	if p.db == nil {
		return fmt.Errorf("not connected")
	}
	return p.db.Ping()
}

func (p *RealPostgreSQL) Exec(query string, args ...interface{}) (Result, error) {
	if p.db == nil {
		return Result{}, fmt.Errorf("not connected")
	}
	res, err := p.db.Exec(query, args...)
	if err != nil {
		return Result{}, err
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return Result{RowsAffected: affected, LastInsertID: lastID}, nil
}

func (p *RealPostgreSQL) Query(query string, args ...interface{}) ([]Row, error) {
	if p.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	rows, err := p.db.Query(query, args...)
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
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
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

func (p *RealPostgreSQL) QueryRow(query string, args ...interface{}) (Row, error) {
	if p.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	row := p.db.QueryRow(query, args...)
	return Row{"_row": row}, nil
}

func (p *RealPostgreSQL) Begin() (Transaction, error) {
	if p.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}
	return &RealPostgreSQLTx{tx: tx}, nil
}

func (p *RealPostgreSQL) Migrate(migrations []Migration) error {
	if p.db == nil {
		return fmt.Errorf("not connected")
	}

	ctx := context.Background()

	_, err := p.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS _migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	for _, m := range migrations {
		var count int
		err := p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM _migrations WHERE version = $1", m.Version).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", m.Version, err)
		}
		if count > 0 {
			continue
		}

		tx, err := p.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.Version, err)
		}

		if _, err := tx.ExecContext(ctx, m.Up); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %d: %w", m.Version, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO _migrations (version, name) VALUES ($1, $2)", m.Version, m.Name); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.Version, err)
		}
	}

	return nil
}

func (p *RealPostgreSQL) Rollback(version int) error {
	if p.db == nil {
		return fmt.Errorf("not connected")
	}

	ctx := context.Background()

	var migrations []Migration
	rows, err := p.db.QueryContext(ctx, "SELECT version, name FROM _migrations WHERE version > $1 ORDER BY version DESC", version)
	if err != nil {
		return fmt.Errorf("query migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m Migration
		if err := rows.Scan(&m.Version, &m.Name); err != nil {
			return err
		}
		migrations = append(migrations, m)
	}

	for _, m := range migrations {
		tx, err := p.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin rollback %d: %w", m.Version, err)
		}

		if _, err := tx.ExecContext(ctx, "DELETE FROM _migrations WHERE version = $1", m.Version); err != nil {
			tx.Rollback()
			return fmt.Errorf("remove migration record %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit rollback %d: %w", m.Version, err)
		}
	}

	return nil
}

type RealPostgreSQLTx struct {
	tx *sql.Tx
}

func (t *RealPostgreSQLTx) Exec(query string, args ...interface{}) (Result, error) {
	res, err := t.tx.Exec(query, args...)
	if err != nil {
		return Result{}, err
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return Result{RowsAffected: affected, LastInsertID: lastID}, nil
}

func (t *RealPostgreSQLTx) Query(query string, args ...interface{}) ([]Row, error) {
	rows, err := t.tx.Query(query, args...)
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
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
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

func (t *RealPostgreSQLTx) Commit() error {
	return t.tx.Commit()
}

func (t *RealPostgreSQLTx) Rollback() error {
	return t.tx.Rollback()
}
