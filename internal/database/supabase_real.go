//go:build !nosql

package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/supabase"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type cachedResult struct {
	rows []Row
	expiresAt time.Time
}

type RealSupabase struct {
	db             *sql.DB
	config         *Config
	supabaseClient *supabase.Client
	restMode       bool
	queryCache     map[string]cachedResult
	cacheMu        sync.RWMutex
}

const cacheTTL = 30 * time.Second

func NewRealSupabase() *RealSupabase {
	return &RealSupabase{
		queryCache: make(map[string]cachedResult),
	}
}

func (s *RealSupabase) restClient() *supabase.Client {
	if s.supabaseClient != nil {
		return s.supabaseClient
	}
	supaCfg := &supabase.Config{
		ProjectRef:     s.config.SupabaseProjectRef,
		ServiceRoleKey: s.config.SupabaseServiceRoleKey,
		ManagementURL:  s.config.SupabaseManagementURL,
		AccessToken:    s.config.SupabaseAccessToken,
	}
	supaCfg.URL = s.config.SupabaseURL
	if supaCfg.URL == "" {
		supaCfg.URL = "https://" + s.config.SupabaseProjectRef + ".supabase.co"
	}
	if supaCfg.ManagementURL == "" {
		supaCfg.ManagementURL = "https://api.supabase.com"
	}
	s.supabaseClient = supabase.NewClient(supaCfg)
	return s.supabaseClient
}

func (s *RealSupabase) Name() string {
	return "supabase"
}

func (s *RealSupabase) Connect(config *Config) error {
	s.config = config

	if config.SupabaseProjectRef != "" {
		s.restMode = true
		return nil
	}

	sslmode := config.SSLMode
	if sslmode == "" {
		sslmode = "require"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.Database, sslmode)

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
	if s.restMode {
		return nil
	}
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *RealSupabase) Ping() error {
	if s.restMode {
		if s.config == nil || s.config.SupabaseProjectRef == "" {
			return naeoserr.New(naeoserr.ErrDatabase, "not connected")
		}
		return nil
	}
	if s.db == nil {
		return naeoserr.New(naeoserr.ErrDatabase, "not connected")
	}
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.db.PingContext(ctx)
}

func (s *RealSupabase) Exec(query string, args ...any) (Result, error) {
	if s.restMode {
		return s.execREST(query)
	}
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.ExecContext(ctx, query, args...)
}

func (s *RealSupabase) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	if s.restMode {
		return s.execREST(query)
	}
	if s.db == nil {
		return Result{}, naeoserr.New(naeoserr.ErrDatabase, "not connected")
	}
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, err
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return Result{RowsAffected: affected, LastInsertID: lastID}, nil
}

func (s *RealSupabase) cacheKey(query string) string {
	return strings.TrimSpace(query)
}

func (s *RealSupabase) cacheGet(key string) ([]Row, bool) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	entry, ok := s.queryCache[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		delete(s.queryCache, key)
		return nil, false
	}
	return entry.rows, true
}

func (s *RealSupabase) cacheSet(key string, rows []Row) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.queryCache[key] = cachedResult{rows: rows, expiresAt: time.Now().Add(cacheTTL)}
}

func (s *RealSupabase) cacheClear() {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.queryCache = make(map[string]cachedResult)
}

func (s *RealSupabase) isWriteQuery(query string) bool {
	q := strings.TrimSpace(strings.ToUpper(query))
	for _, prefix := range []string{"INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER", "TRUNCATE", "REPLACE"} {
		if strings.HasPrefix(q, prefix) {
			return true
		}
	}
	return false
}

func (s *RealSupabase) execREST(query string) (Result, error) {
	if s.isWriteQuery(query) {
		defer s.cacheClear()
	}

	ctx, cancel := s.defaultContext()
	defer cancel()

	var result Result
	err := WithRetry(ctx, 3, 200*time.Millisecond, func(ctx context.Context) error {
		client := s.restClient()
		res, err := client.ExecuteSQL(query)
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "execute SQL via REST")
		}
		if res.Error != "" {
			return naeoserr.New(naeoserr.ErrDatabase, "SQL error: "+res.Error)
		}
		if len(res.Rows) == 0 {
			result = Result{RowsAffected: 0}
			return nil
		}
		if v, ok := res.Rows[0]["rows_affected"]; ok {
			switch vv := v.(type) {
			case float64:
				result = Result{RowsAffected: int64(vv)}
			case int64:
				result = Result{RowsAffected: vv}
			}
			return nil
		}
		result = Result{RowsAffected: int64(len(res.Rows))}
		return nil
	})
	return result, err
}

func (s *RealSupabase) Query(query string, args ...any) ([]Row, error) {
	if s.restMode {
		return s.queryREST(query)
	}
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.QueryContext(ctx, query, args...)
}

func (s *RealSupabase) QueryContext(ctx context.Context, query string, args ...any) ([]Row, error) {
	if s.restMode {
		return s.queryREST(query)
	}
	if s.db == nil {
		return nil, naeoserr.New(naeoserr.ErrDatabase, "not connected")
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

func (s *RealSupabase) queryREST(query string) ([]Row, error) {
	if s.isWriteQuery(query) {
		defer s.cacheClear()
	}

	key := s.cacheKey(query)
	if cached, ok := s.cacheGet(key); ok {
		return cached, nil
	}

	ctx, cancel := s.defaultContext()
	defer cancel()

	var rows []Row
	err := WithRetry(ctx, 3, 200*time.Millisecond, func(ctx context.Context) error {
		client := s.restClient()
		result, err := client.ExecuteSQL(query)
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "query via REST")
		}
		if result.Error != "" {
			return naeoserr.New(naeoserr.ErrDatabase, "SQL error: "+result.Error)
		}
		rows = make([]Row, len(result.Rows))
		for i, r := range result.Rows {
			rows[i] = Row(r)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if !s.isWriteQuery(query) {
		s.cacheSet(key, rows)
	}
	return rows, nil
}

func (s *RealSupabase) QueryRow(query string, args ...any) (Row, error) {
	if s.restMode {
		rows, err := s.queryREST(query)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return Row{}, nil
		}
		return rows[0], nil
	}
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.QueryRowContext(ctx, query, args...)
}

func (s *RealSupabase) QueryRowContext(ctx context.Context, query string, args ...any) (Row, error) {
	if s.restMode {
		rows, err := s.queryREST(query)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return Row{}, nil
		}
		return rows[0], nil
	}
	if s.db == nil {
		return nil, naeoserr.New(naeoserr.ErrDatabase, "not connected")
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
	if s.restMode {
		return nil, naeoserr.New(naeoserr.ErrDatabase, "transactions not supported in Supabase REST API mode; use direct PostgreSQL connection instead")
	}
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.BeginTx(ctx)
}

func (s *RealSupabase) BeginTx(ctx context.Context) (Transaction, error) {
	if s.restMode {
		return nil, naeoserr.New(naeoserr.ErrDatabase, "transactions not supported in Supabase REST API mode; use direct PostgreSQL connection instead")
	}
	if s.db == nil {
		return nil, naeoserr.New(naeoserr.ErrDatabase, "not connected")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &RealSupabaseTx{tx: tx}, nil
}

func (s *RealSupabase) Migrate(migrations []Migration) error {
	if s.restMode {
		return s.migrateREST(migrations)
	}
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.MigrateContext(ctx, migrations)
}

func (s *RealSupabase) MigrateContext(ctx context.Context, migrations []Migration) error {
	if s.restMode {
		return s.migrateREST(migrations)
	}
	if s.db == nil {
		return naeoserr.New(naeoserr.ErrDatabase, "not connected")
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
		return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "create migrations table")
	}

	for _, m := range migrations {
		var count int
		err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM _migrations WHERE version = $1", m.Version).Scan(&count)
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "check migration %d", m.Version)
		}
		if count > 0 {
			continue
		}

		tx, err := s.db.BeginTx(ctx, nil)
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

func (s *RealSupabase) migrateREST(migrations []Migration) error {
	client := s.restClient()

	_, err := client.ExecuteSQL("CREATE TABLE IF NOT EXISTS _migrations (version INTEGER PRIMARY KEY, name TEXT, down_sql TEXT, applied_at TIMESTAMPTZ DEFAULT NOW())")
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "create migrations table via REST")
	}

	for _, m := range migrations {
		result, err := client.ExecuteSQL(fmt.Sprintf("SELECT COUNT(*) as cnt FROM _migrations WHERE version = %d", m.Version))
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "check migration %d", m.Version)
		}
		if result.Error != "" {
			return naeoserr.New(naeoserr.ErrDatabase, "check migration error: "+result.Error)
		}
		if len(result.Rows) > 0 {
			if cnt, ok := result.Rows[0]["cnt"]; ok {
				var alreadyApplied bool
				switch v := cnt.(type) {
				case float64:
					alreadyApplied = int64(v) > 0
				case int64:
					alreadyApplied = v > 0
				case int:
					alreadyApplied = v > 0
				}
				if alreadyApplied {
					continue
				}
			}
		}

		upResult, err := client.ExecuteSQL(m.Up)
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "apply migration %d via REST", m.Version)
		}
		if upResult.Error != "" {
			return naeoserr.New(naeoserr.ErrDatabase, fmt.Sprintf("apply migration %d (%s) error: %s", m.Version, m.Name, upResult.Error))
		}

		escapedDown := strings.ReplaceAll(m.Down, "'", "''")
		_, err = client.ExecuteSQL(fmt.Sprintf("INSERT INTO _migrations (version, name, down_sql) VALUES (%d, '%s', '%s')", m.Version, m.Name, escapedDown))
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "record migration %d via REST", m.Version)
		}
	}

	return nil
}

func (s *RealSupabase) Rollback(version int) error {
	if s.restMode {
		return s.rollbackREST(version)
	}
	ctx, cancel := s.defaultContext()
	defer cancel()
	return s.RollbackContext(ctx, version)
}

func (s *RealSupabase) RollbackContext(ctx context.Context, version int) error {
	if s.restMode {
		return s.rollbackREST(version)
	}
	if s.db == nil {
		return naeoserr.New(naeoserr.ErrDatabase, "not connected")
	}

	var migrations []Migration
	rows, err := s.db.QueryContext(ctx, "SELECT version, name, down_sql FROM _migrations WHERE version > $1 ORDER BY version DESC", version)
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
		tx, err := s.db.BeginTx(ctx, nil)
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

func (s *RealSupabase) rollbackREST(version int) error {
	client := s.restClient()

	result, err := client.ExecuteSQL(fmt.Sprintf("SELECT version, name, down_sql FROM _migrations WHERE version > %d ORDER BY version DESC", version))
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "query migrations via REST")
	}
	if result.Error != "" && !strings.Contains(result.Error, "does not exist") {
		return naeoserr.New(naeoserr.ErrDatabase, "query migrations error: "+result.Error)
	}

	for _, row := range result.Rows {
		var m Migration
		if v, ok := row["version"]; ok {
			switch vv := v.(type) {
			case float64:
				m.Version = int(vv)
			case int64:
				m.Version = int(vv)
			}
		}
		if n, ok := row["name"]; ok {
			m.Name = fmt.Sprintf("%v", n)
		}
		if d, ok := row["down_sql"]; ok {
			m.Down = fmt.Sprintf("%v", d)
		}

		if m.Down != "" {
			downResult, err := client.ExecuteSQL(m.Down)
			if err != nil {
				return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "execute down migration %d", m.Version)
			}
			if downResult.Error != "" {
				return naeoserr.New(naeoserr.ErrDatabase, fmt.Sprintf("down migration %d error: %s", m.Version, downResult.Error))
			}
		}

		_, err = client.ExecuteSQL(fmt.Sprintf("DELETE FROM _migrations WHERE version = %d", m.Version))
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrDatabase, "remove migration record %d via REST", m.Version)
		}
	}

	return nil
}

func (s *RealSupabase) HealthCheck() error {
	if s.restMode {
		if s.config == nil || s.config.SupabaseProjectRef == "" {
			return naeoserr.New(naeoserr.ErrDatabase, "not connected")
		}
		client := s.restClient()
		_, err := client.ExecuteSQL("SELECT 1")
		return err
	}
	if s.db == nil {
		return naeoserr.New(naeoserr.ErrDatabase, "not connected")
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
