// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import "testing"

func TestValidateProductionDatabaseURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "empty", url: "", want: false},
		{name: "missing tls", url: "postgres://user:secret@localhost:5432/db", want: false},
		{name: "disable tls", url: "postgres://user:secret@localhost:5432/db?sslmode=disable", want: false},
		{name: "wrong scheme", url: "https://example.com/db?sslmode=require", want: false},
		{name: "malformed", url: "postgres://%zz", want: false},
		{name: "tls require", url: "postgres://user:secret@localhost:5432/db?sslmode=require", want: true},
		{name: "verify ca", url: "postgres://user:secret@localhost:5432/db?sslmode=verify-ca", want: true},
		{name: "verify full", url: "postgresql://user:secret@localhost:5432/db?sslmode=verify-full", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProductionDatabaseURL(tt.url)
			if (err == nil) != tt.want {
				t.Fatalf("validateProductionDatabaseURL() error = %v, want success %v", err, tt.want)
			}
		})
	}
}
