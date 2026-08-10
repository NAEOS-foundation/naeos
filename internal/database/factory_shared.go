package database

import (
	"log/slog"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

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
