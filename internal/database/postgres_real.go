//go:build !nosql

package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
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
		return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "open database")
	}

	applyPoolConfig(db, config)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "ping database")
	}

	p.db = db
	return nil
}

func (p *RealPostgreSQL) defaultContext() (context.Context, context.CancelFunc) {
	if p.config != nil && p.config.Timeout > 0 {
		return context.WithTimeout(context.Background(), p.config.Timeout)
	}
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func (p *RealPostgreSQL) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *RealPostgreSQL) Ping() error {
	if p.db == nil {
		return naeoserr.New(naeoserr.ErrDatabase, "not connected")
	}
	ctx, cancel := p.defaultContext()
	defer cancel()
	return p.db.PingContext(ctx)
}

func (p *RealPostgreSQL) Exec(query string, args ...any) (Result, error) {
	ctx, cancel := p.defaultContext()
	defer cancel()
	return p.ExecContext(ctx, query, args...)
}

func (p *RealPostgreSQL) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	if p.db == nil {
		return Result{}, naeoserr.New(naeoserr.ErrDatabase, "not connected")
	}
	res, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, err
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return Result{RowsAffected: affected, LastInsertID: lastID}, nil
}

func (p *RealPostgreSQL) Query(query string, args ...any) ([]Row, error) {
	ctx, cancel := p.defaultContext()
	defer cancel()
	return p.QueryContext(ctx, query, args...)
}

func (p *RealPostgreSQL) QueryContext(ctx context.Context, query string, args ...any) ([]Row, error) {
	if p.db == nil {
		return nil, naeoserr.New(naeoserr.ErrDatabase, "not connected")
	}
	rows, err := p.db.QueryContext(ctx, query, args...)
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

func (p *RealPostgreSQL) QueryRow(query string, args ...any) (Row, error) {
	ctx, cancel := p.defaultContext()
	defer cancel()
	return p.QueryRowContext(ctx, query, args...)
}

func (p *RealPostgreSQL) QueryRowContext(ctx context.Context, query string, args ...any) (Row, error) {
	if p.db == nil {
		return nil, naeoserr.New(naeoserr.ErrDatabase, "not connected")
	}
	rows, err := p.db.QueryContext(ctx, query, args...)
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

func (p *RealPostgreSQL) Begin() (Transaction, error) {
	ctx, cancel := p.defaultContext()
	defer cancel()
	return p.BeginTx(ctx)
}

func (p *RealPostgreSQL) BeginTx(ctx context.Context) (Transaction, error) {
	if p.db == nil {
		return nil, naeoserr.New(naeoserr.ErrDatabase, "not connected")
	}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &RealPostgreSQLTx{tx: tx}, nil
}

func (p *RealPostgreSQL) Migrate(migrations []Migration) error {
	ctx, cancel := p.defaultContext()
	defer cancel()
	return p.MigrateContext(ctx, migrations)
}

func (p *RealPostgreSQL) MigrateContext(ctx context.Context, migrations []Migration) error {
	if p.db == nil {
		return naeoserr.New(naeoserr.ErrDatabase, "not connected")
	}

	_, err := p.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS _migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			down_sql TEXT,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "create migrations table")
	}

	for _, m := range migrations {
		var count int
		err := p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM _migrations WHERE version = $1", m.Version).Scan(&count)
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "check migration %d", m.Version)
		}
		if count > 0 {
			continue
		}

		tx, err := p.db.BeginTx(ctx, nil)
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "begin migration %d", m.Version)
		}

		if _, err := tx.ExecContext(ctx, m.Up); err != nil {
			_ = tx.Rollback()
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "apply migration %d", m.Version)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO _migrations (version, name, down_sql) VALUES ($1, $2, $3)", m.Version, m.Name, m.Down); err != nil {
			_ = tx.Rollback()
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "record migration %d", m.Version)
		}

		if err := tx.Commit(); err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "commit migration %d", m.Version)
		}
	}

	return nil
}

func (p *RealPostgreSQL) Rollback(version int) error {
	ctx, cancel := p.defaultContext()
	defer cancel()
	return p.RollbackContext(ctx, version)
}

func (p *RealPostgreSQL) RollbackContext(ctx context.Context, version int) error {
	if p.db == nil {
		return naeoserr.New(naeoserr.ErrDatabase, "not connected")
	}

	var migrations []Migration
	rows, err := p.db.QueryContext(ctx, "SELECT version, name, down_sql FROM _migrations WHERE version > $1 ORDER BY version DESC", version)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "query migrations")
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
		tx, err := p.db.BeginTx(ctx, nil)
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "begin rollback %d", m.Version)
		}

		if m.Down != "" {
			if _, err := tx.ExecContext(ctx, m.Down); err != nil {
				_ = tx.Rollback()
				return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "execute down migration %d (%s)", m.Version, m.Name)
			}
		}

		if _, err := tx.ExecContext(ctx, "DELETE FROM _migrations WHERE version = $1", m.Version); err != nil {
			_ = tx.Rollback()
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "remove migration record %d", m.Version)
		}

		if err := tx.Commit(); err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "commit rollback %d", m.Version)
		}
	}

	return nil
}

func (p *RealPostgreSQL) HealthCheck() error {
	if p.db == nil {
		return naeoserr.New(naeoserr.ErrDatabase, "not connected")
	}
	ctx, cancel := p.defaultContext()
	defer cancel()
	return p.db.PingContext(ctx)
}

type RealPostgreSQLTx struct {
	tx *sql.Tx
}

func (t *RealPostgreSQLTx) Exec(query string, args ...any) (Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return t.ExecContext(ctx, query, args...)
}

func (t *RealPostgreSQLTx) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	res, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, err
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return Result{RowsAffected: affected, LastInsertID: lastID}, nil
}

func (t *RealPostgreSQLTx) Query(query string, args ...any) ([]Row, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return t.QueryContext(ctx, query, args...)
}

func (t *RealPostgreSQLTx) QueryContext(ctx context.Context, query string, args ...any) ([]Row, error) {
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

func (t *RealPostgreSQLTx) Commit() error {
	return t.tx.Commit()
}

func (t *RealPostgreSQLTx) Rollback() error {
	return t.tx.Rollback()
}
