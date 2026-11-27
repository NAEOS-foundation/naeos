// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
 "context"
 "errors"
 "net/url"
 "strings"
 "time"
 "github.com/jackc/pgx/v5/pgxpool"
)

type PostgresEvidenceVerifierReplayGuard struct { pool *pgxpool.Pool }

func NewPostgresEvidenceVerifierReplayGuard(ctx context.Context, databaseURL string) (*PostgresEvidenceVerifierReplayGuard, error) {
 raw := strings.TrimSpace(databaseURL)
 u, err := url.Parse(raw)
 if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" { return nil, errors.New("evidence verifier database URL must be PostgreSQL") }
 sslmode := strings.ToLower(u.Query().Get("sslmode"))
 if sslmode != "require" && sslmode != "verify-ca" && sslmode != "verify-full" { return nil, errors.New("evidence verifier database URL must require TLS") }
 cfg, err := pgxpool.ParseConfig(raw); if err != nil { return nil, errors.New("invalid evidence verifier database URL") }
 cfg.MaxConns=4; cfg.MinConns=0; cfg.MaxConnLifetime=30*time.Minute; cfg.MaxConnIdleTime=5*time.Minute; cfg.ConnConfig.ConnectTimeout=5*time.Second
 pool, err := pgxpool.NewWithConfig(ctx,cfg); if err != nil { return nil, errors.New("unable to configure verifier PostgreSQL pool") }
 pingCtx,cancel:=context.WithTimeout(ctx,5*time.Second); defer cancel()
 if err:=pool.Ping(pingCtx); err!=nil { pool.Close(); return nil,errors.New("verifier PostgreSQL connectivity failed") }
 var ready bool
 const check = `SELECT to_regclass('naeos_evidence_verifier.replay_claims') IS NOT NULL,
 has_schema_privilege(current_user,'naeos_evidence_verifier','USAGE'),
 has_table_privilege(current_user,'naeos_evidence_verifier.replay_claims','INSERT'),
 NOT has_table_privilege(current_user,'naeos_evidence_verifier.replay_claims','SELECT'),
 NOT has_table_privilege(current_user,'naeos_evidence_verifier.replay_claims','UPDATE'),
 NOT has_table_privilege(current_user,'naeos_evidence_verifier.replay_claims','DELETE')`
 var tableExists, schemaUsage, insertAllowed, selectDenied, updateDenied, deleteDenied bool
 if err:=pool.QueryRow(pingCtx,check).Scan(&tableExists,&schemaUsage,&insertAllowed,&selectDenied,&updateDenied,&deleteDenied);err!=nil || !tableExists || !schemaUsage || !insertAllowed || !selectDenied || !updateDenied || !deleteDenied { pool.Close(); return nil,errors.New("verifier database schema or least-privilege role validation failed") }
 return &PostgresEvidenceVerifierReplayGuard{pool:pool},nil
}
func (g *PostgresEvidenceVerifierReplayGuard) Claim(ctx context.Context, scope, requestID string) (bool,error) {
 if g==nil || g.pool==nil || strings.TrimSpace(scope)=="" || len(requestID)!=32 { return false,errors.New("durable replay guard is not configured or claim is invalid") }
 ctx,cancel:=context.WithTimeout(ctx,3*time.Second); defer cancel()
 const query = `INSERT INTO naeos_evidence_verifier.replay_claims (verifier_scope, request_id)
 VALUES ($1,$2) ON CONFLICT (verifier_scope,request_id) DO NOTHING`
 tag,err:=g.pool.Exec(ctx,query,scope,strings.ToLower(requestID)); if err!=nil { return false,errors.New("durable replay claim failed") }
 return tag.RowsAffected()==1,nil
}
func (g *PostgresEvidenceVerifierReplayGuard) ClaimReadiness(ctx context.Context) (bool,error) {
 if g==nil || g.pool==nil { return false,errors.New("replay guard unavailable") }
 ctx,cancel:=context.WithTimeout(ctx,2*time.Second); defer cancel()
 if err:=g.pool.Ping(ctx); err!=nil { return false,errors.New("replay guard readiness failed") }
 return true,nil
}
func (g *PostgresEvidenceVerifierReplayGuard) Close() { if g!=nil && g.pool!=nil { g.pool.Close() } }
