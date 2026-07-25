//go:build !nosql

package database

import (
	"database/sql"
	"time"
)

const (
	DefaultMaxOpenConns    = 5
	DefaultMaxIdleConns    = 2
	DefaultConnMaxLifetime = 30 * time.Minute
	DefaultConnMaxIdleTime = 5 * time.Minute
)

func applyPoolConfig(db *sql.DB, cfg *Config) {
	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = DefaultMaxOpenConns
	}
	db.SetMaxOpenConns(maxOpen)

	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = DefaultMaxIdleConns
	}
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}
	db.SetMaxIdleConns(maxIdle)

	connMaxLifetime := cfg.ConnMaxLifetime
	if connMaxLifetime <= 0 {
		connMaxLifetime = DefaultConnMaxLifetime
	}
	db.SetConnMaxLifetime(connMaxLifetime)

	connMaxIdleTime := cfg.ConnMaxIdleTime
	if connMaxIdleTime <= 0 {
		connMaxIdleTime = DefaultConnMaxIdleTime
	}
	db.SetConnMaxIdleTime(connMaxIdleTime)
}
