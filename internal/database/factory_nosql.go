//go:build nosql

package database

import naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"

func New(driver string) Database {
	switch driver {
	case "mock-postgresql":
		return NewPostgreSQL()
	case "mock-mysql":
		return NewMySQL()
	case "mock-sqlite":
		return NewSQLite()
	case "mock-supabase":
		return NewSupabase()
	default:
		return nil
	}
}

func NewFromConfig(driver string, config *Config) (Database, error) {
	if err := config.Validate(); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrDatabase, "invalid config")
	}
	db := New(driver)
	if db == nil {
		return nil, naeoserr.New(naeoserr.ErrDatabase, "unsupported driver: "+driver)
	}
	if err := db.Connect(config); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrDatabase, "connect")
	}
	return db, nil
}
