// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

//go:build !nosql

package database

func New(driver string) Database {
	switch driver {
	case "postgresql", "postgres":
		return NewRealPostgreSQL()
	case "mysql", "mariadb":
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
