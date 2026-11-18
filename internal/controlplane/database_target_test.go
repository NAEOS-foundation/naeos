// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import "testing"

func TestValidateControlPlaneDatabaseTarget(t *testing.T) {
	const productionURL = "postgres://naeos:secret@db.oruykkehukoiipldzcvc.supabase.co:5432/postgres?sslmode=require"
	const stagingURL = "postgres://naeos:secret@db.kqmhpsftijskesjdovjm.supabase.co:5432/postgres?sslmode=verify-full"
	const productionPoolerURL = "postgres://postgres.oruykkehukoiipldzcvc:secret@aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres?sslmode=require"
	const stagingPoolerURL = "postgres://postgres.kqmhpsftijskesjdovjm:secret@aws-0-ap-northeast-1.pooler.supabase.com:5432/postgres?sslmode=require"
	const productionRuntimePoolerURL = "postgres://naeos_control_plane_runtime_login.oruykkehukoiipldzcvc:secret@aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres?sslmode=require"
	const stagingRuntimePoolerURL = "postgres://naeos_control_plane_runtime_login.kqmhpsftijskesjdovjm:secret@aws-0-ap-northeast-1.pooler.supabase.com:5432/postgres?sslmode=require"

	tests := []struct {
		name        string
		raw         string
		environment string
		wantErr     bool
	}{
		{name: "production direct target", raw: productionURL, environment: "production"},
		{name: "staging direct target", raw: stagingURL, environment: "staging"},
		{name: "preview maps to staging", raw: stagingURL, environment: "preview"},
		{name: "production session pooler", raw: productionPoolerURL, environment: "production"},
		{name: "staging session pooler", raw: stagingPoolerURL, environment: "staging"},
		{name: "production custom runtime role pooler", raw: productionRuntimePoolerURL, environment: "production"},
		{name: "staging custom runtime role pooler", raw: stagingRuntimePoolerURL, environment: "staging"},
		{name: "production cannot use staging", raw: stagingURL, environment: "production", wantErr: true},
		{name: "staging cannot use production", raw: productionURL, environment: "staging", wantErr: true},
		{name: "unknown environment", raw: productionURL, environment: "development", wantErr: true},
		{name: "missing environment", raw: productionURL, environment: "", wantErr: true},
		{name: "wrong pooler project username", raw: "postgres://postgres.kqmhpsftijskesjdovjm:secret@aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres?sslmode=require", environment: "production", wantErr: true},
		{name: "custom pooler role with wrong project suffix", raw: "postgres://naeos_control_plane_runtime_login.kqmhpsftijskesjdovjm:secret@aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres?sslmode=require", environment: "production", wantErr: true},
		{name: "invalid custom pooler role name", raw: "postgres://naeos-control-plane-runtime.oruykkehukoiipldzcvc:secret@aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres?sslmode=require", environment: "production", wantErr: true},
		{name: "pooler username suffix spoof", raw: "postgres://evil.oruykkehukoiipldzcvc:secret@aws-0-ap-southeast-1.pooler.supabase.com:5432/postgres?sslmode=require", environment: "production", wantErr: true},
		{name: "untrusted host", raw: "postgres://naeos:secret@db.example.com:5432/postgres?sslmode=require", environment: "production", wantErr: true},
		{name: "tls disabled", raw: "postgres://naeos:secret@db.oruykkehukoiipldzcvc.supabase.co:5432/postgres?sslmode=disable", environment: "production", wantErr: true},
		{name: "empty url", raw: "", environment: "production", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateControlPlaneDatabaseTarget(tt.raw, tt.environment)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateControlPlaneDatabaseTarget() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
