//go:build !nosql

package database

import (
	"log/slog"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
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
	case "supabase":
		return NewRealSupabase()
	case "mock-supabase":
		return NewSupabase()
	default:
		return nil
	}
}

func NewFromConfig(driver string, config *Config) (Database, error) {
	if err := config.Validate(); err != nil {
		slog.Error("invalid database config", "driver", driver, "error", err)
		return nil, naeoserr.Wrapf(err, naeoserr.ErrDatabase, "invalid config")
	}
	db := New(driver)
	if db == nil {
		slog.Error("unsupported database driver", "driver", driver)
		return nil, naeoserr.New(naeoserr.ErrDatabase, "unsupported driver: "+driver)
	}
	if err := db.Connect(config); err != nil {
		slog.Error("database connect failed", "driver", driver, "error", err)
		return nil, naeoserr.Wrapf(err, naeoserr.ErrDatabase, "connect")
	}
	return db, nil
}
