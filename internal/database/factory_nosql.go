//go:build nosql

package database

import "fmt"

func New(driver string) Database {
	switch driver {
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
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	db := New(driver)
	if db == nil {
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
	if err := db.Connect(config); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	return db, nil
}
