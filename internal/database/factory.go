//go:build !nosql

package database

import (
	"fmt"
	"log/slog"
)

func New(driver string) Database {
	switch driver {
	case "postgresql", "postgres":
		return NewRealPostgreSQL()
	case "mysql":
		return NewRealMySQL()
	case "sqlite":
		return NewRealSQLite()
	case "mock-postgresql":
		return NewPostgreSQL()
	case "mock-mysql":
		return NewMySQL()
	case "mock-sqlite":
		return NewSQLite()
	default:
		return nil
	}
}

func NewFromConfig(driver string, config *Config) (Database, error) {
	if err := config.Validate(); err != nil {
		slog.Error("invalid database config", "driver", driver, "error", err)
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	db := New(driver)
	if db == nil {
		slog.Error("unsupported database driver", "driver", driver)
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
	if err := db.Connect(config); err != nil {
		slog.Error("database connect failed", "driver", driver, "error", err)
		return nil, fmt.Errorf("connect: %w", err)
	}
	return db, nil
}
