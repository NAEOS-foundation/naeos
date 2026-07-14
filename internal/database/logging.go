package database

import (
	"context"
	"log/slog"
	"time"
)

type loggingDatabase struct {
	inner  Database
	logger *slog.Logger
}

func NewLoggingDatabase(inner Database, logger *slog.Logger) Database {
	if logger == nil {
		logger = slog.Default()
	}
	return &loggingDatabase{inner: inner, logger: logger}
}

func (l *loggingDatabase) Name() string {
	return l.inner.Name()
}

func (l *loggingDatabase) Connect(config *Config) error {
	start := time.Now()
	err := l.inner.Connect(config)
	l.logger.Info("connect",
		"driver", l.inner.Name(),
		"host", config.Host,
		"port", config.Port,
		"database", config.Database,
		"duration", time.Since(start),
		"error", err,
	)
	return err
}

func (l *loggingDatabase) Close() error {
	start := time.Now()
	err := l.inner.Close()
	l.logger.Info("close", "driver", l.inner.Name(), "duration", time.Since(start), "error", err)
	return err
}

func (l *loggingDatabase) Ping() error {
	start := time.Now()
	err := l.inner.Ping()
	l.logger.Debug("ping", "driver", l.inner.Name(), "duration", time.Since(start), "error", err)
	return err
}

func (l *loggingDatabase) Exec(query string, args ...any) (Result, error) {
	return l.ExecContext(context.Background(), query, args...)
}

func (l *loggingDatabase) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	start := time.Now()
	result, err := l.inner.ExecContext(ctx, query, args...)
	l.logQuery("exec", query, args, start, err)
	return result, err
}

func (l *loggingDatabase) Query(query string, args ...any) ([]Row, error) {
	return l.QueryContext(context.Background(), query, args...)
}

func (l *loggingDatabase) QueryContext(ctx context.Context, query string, args ...any) ([]Row, error) {
	start := time.Now()
	rows, err := l.inner.QueryContext(ctx, query, args...)
	l.logQuery("query", query, args, start, err)
	return rows, err
}

func (l *loggingDatabase) QueryRow(query string, args ...any) (Row, error) {
	return l.QueryRowContext(context.Background(), query, args...)
}

func (l *loggingDatabase) QueryRowContext(ctx context.Context, query string, args ...any) (Row, error) {
	start := time.Now()
	row, err := l.inner.QueryRowContext(ctx, query, args...)
	l.logQuery("query_row", query, args, start, err)
	return row, err
}

func (l *loggingDatabase) Begin() (Transaction, error) {
	return l.BeginTx(context.Background())
}

func (l *loggingDatabase) BeginTx(ctx context.Context) (Transaction, error) {
	start := time.Now()
	tx, err := l.inner.BeginTx(ctx)
	l.logger.Debug("begin_tx", "driver", l.inner.Name(), "duration", time.Since(start), "error", err)
	return tx, err
}

func (l *loggingDatabase) Migrate(migrations []Migration) error {
	return l.MigrateContext(context.Background(), migrations)
}

func (l *loggingDatabase) MigrateContext(ctx context.Context, migrations []Migration) error {
	start := time.Now()
	err := l.inner.MigrateContext(ctx, migrations)
	l.logger.Info("migrate",
		"driver", l.inner.Name(),
		"count", len(migrations),
		"duration", time.Since(start),
		"error", err,
	)
	return err
}

func (l *loggingDatabase) Rollback(version int) error {
	return l.RollbackContext(context.Background(), version)
}

func (l *loggingDatabase) RollbackContext(ctx context.Context, version int) error {
	start := time.Now()
	err := l.inner.RollbackContext(ctx, version)
	l.logger.Info("rollback",
		"driver", l.inner.Name(),
		"target_version", version,
		"duration", time.Since(start),
		"error", err,
	)
	return err
}

func (l *loggingDatabase) HealthCheck() error {
	start := time.Now()
	err := l.inner.HealthCheck()
	l.logger.Debug("health_check", "driver", l.inner.Name(), "duration", time.Since(start), "error", err)
	return err
}

func (l *loggingDatabase) logQuery(op, query string, args []any, start time.Time, err error) {
	duration := time.Since(start)
	level := slog.LevelDebug
	if duration > time.Second {
		level = slog.LevelWarn
	}

	truncated := query
	if len(truncated) > 200 {
		truncated = truncated[:200] + "..."
	}

	l.logger.Log(context.Background(), level, op,
		"driver", l.inner.Name(),
		"query", truncated,
		"args_count", len(args),
		"duration", duration,
		"error", err,
	)
}
