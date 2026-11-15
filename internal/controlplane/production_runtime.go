// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/NAEOS-foundation/naeos/internal/api"
)

type productionTokenValidator interface {
	Validate(context.Context, string) (*api.JWTClaims, error)
}

// ProductionRuntime owns the dedicated PostgreSQL connections and repository
// instances for the production Control Plane. It does not seed demo state.
type ProductionRuntime struct {
	pool      *pgxpool.Pool
	sqlDB     *sql.DB
	Policies  *PostgresPolicyRepository
	Grants    *PostgresGrantRepository
	Approvals *PostgresApprovalRepository
	Ledger    *PostgresLedger
	identity  productionTokenValidator
	gateway   *RepositoryDecisionGateway
}

// OpenProductionRuntime validates the connection contract, opens bounded pools,
// verifies the least-privilege runtime role and private schema, and validates
// the persisted ledger before returning a production runtime.
func OpenProductionRuntime(ctx context.Context, databaseURL string) (*ProductionRuntime, error) {
	if err := validateProductionDatabaseURL(databaseURL); err != nil {
		return nil, err
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.New("parse Control Plane DATABASE_URL: invalid PostgreSQL URL")
	}
	// The runtime uses two drivers over one DATABASE_URL. Split the total
	// connection budget evenly so the combined maximum remains eight.
	cfg.MaxConns = 4
	cfg.MinConns = 0
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("configure Control Plane PostgreSQL pool: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			pool.Close()
		}
	}()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("control plane PostgreSQL ping failed: %w", err)
	}
	if err := verifyControlPlaneRuntimeRole(ctx, pool); err != nil {
		return nil, err
	}
	if err := verifyControlPlaneSchema(ctx, pool); err != nil {
		return nil, err
	}

	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open Control Plane SQL repositories: %w", err)
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("control plane SQL repository ping failed: %w", err)
	}

	ledger, err := NewPostgresLedger(pool)
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("initialize Control Plane ledger: %w", err)
	}
	// Do not bind the listener over an unverified or tampered ledger. The
	// complete chain and authoritative head are checked from one DB snapshot.
	if _, err := ledger.Load(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("control plane ledger startup verification failed: %w", err)
	}

	cleanup = false
	return &ProductionRuntime{
		pool: pool, sqlDB: sqlDB,
		Policies:  NewPostgresPolicyRepository(sqlDB),
		Grants:    NewPostgresGrantRepository(sqlDB),
		Approvals: NewPostgresApprovalRepository(sqlDB),
		Ledger:    ledger,
	}, nil
}

// Close releases both PostgreSQL pools. Call it during graceful shutdown.
func (r *ProductionRuntime) Close() {
	if r == nil {
		return
	}
	if r.sqlDB != nil {
		_ = r.sqlDB.Close()
	}
	if r.pool != nil {
		r.pool.Close()
	}
}

func validateProductionDatabaseURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return errors.New("control plane DATABASE_URL is required")
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" {
		return errors.New("control plane DATABASE_URL must be a valid PostgreSQL URL")
	}
	sslmode := strings.ToLower(u.Query().Get("sslmode"))
	switch sslmode {
	case "require", "verify-ca", "verify-full":
		return nil
	default:
		return errors.New("control plane DATABASE_URL must require TLS (sslmode=require, verify-ca, or verify-full)")
	}
}

// verifyControlPlaneRuntimeRole rejects admin, BYPASSRLS, and non-login
// connections. The application must authenticate as a dedicated LOGIN role
// that inherits only the NOLOGIN/NOBYPASSRLS Control Plane runtime group.
func verifyControlPlaneRuntimeRole(ctx context.Context, pool *pgxpool.Pool) error {
	const query = `SELECT
		login_role.rolcanlogin,
		login_role.rolsuper,
		login_role.rolcreatedb,
		login_role.rolcreaterole,
		login_role.rolbypassrls,
		login_role.rolinherit,
		runtime_role.rolcanlogin,
		runtime_role.rolsuper,
		runtime_role.rolcreatedb,
		runtime_role.rolcreaterole,
		runtime_role.rolbypassrls,
		pg_catalog.pg_has_role(login_role.oid, runtime_role.oid, 'MEMBER'),
		current_user = session_user,
		NOT EXISTS (
			SELECT 1
			FROM pg_catalog.pg_roles privileged_role
			WHERE privileged_role.rolname IN ('service_role', 'supabase_admin', 'postgres')
			  AND pg_catalog.pg_has_role(login_role.oid, privileged_role.oid, 'MEMBER')
		)
	FROM pg_catalog.pg_roles login_role
	JOIN pg_catalog.pg_roles runtime_role
	  ON runtime_role.rolname = 'naeos_control_plane_runtime'
	WHERE login_role.rolname = session_user`
	var (
		login, loginSuper, loginCreateDB, loginCreateRole, loginBypassRLS, loginInherit bool
		groupLogin, groupSuper, groupCreateDB, groupCreateRole, groupBypassRLS bool
		member, sameRole, noPrivilegedMembership bool
	)
	if pool == nil {
		return errors.New("control plane DATABASE_URL must use the dedicated least-privilege runtime role")
	}
	if err := pool.QueryRow(ctx, query).Scan(
		&login, &loginSuper, &loginCreateDB, &loginCreateRole, &loginBypassRLS, &loginInherit,
		&groupLogin, &groupSuper, &groupCreateDB, &groupCreateRole, &groupBypassRLS,
		&member, &sameRole, &noPrivilegedMembership,
	); err != nil {
		return errors.New("control plane DATABASE_URL runtime role verification failed")
	}
	if !login || loginSuper || loginCreateDB || loginCreateRole || loginBypassRLS || !loginInherit ||
		groupLogin || groupSuper || groupCreateDB || groupCreateRole || groupBypassRLS ||
		!member || !sameRole || !noPrivilegedMembership {
		return errors.New("control plane DATABASE_URL must authenticate as a dedicated LOGIN member of naeos_control_plane_runtime without privileged role membership")
	}
	return nil
}

func verifyControlPlaneSchema(ctx context.Context, pool *pgxpool.Pool) error {
	const required = `SELECT
		to_regclass('naeos_control_plane.policy_versions') IS NOT NULL,
		to_regclass('naeos_control_plane.active_policies') IS NOT NULL,
		to_regclass('naeos_control_plane.grants') IS NOT NULL,
		to_regclass('naeos_control_plane.approvals') IS NOT NULL,
		to_regclass('naeos_control_plane.ledger_head') IS NOT NULL,
		to_regclass('naeos_control_plane.ledger_events') IS NOT NULL,
		to_regprocedure('naeos_control_plane.append_ledger_event(jsonb,text,text)') IS NOT NULL,
		to_regprocedure('naeos_control_plane.read_ledger_head()') IS NOT NULL`
	var policyVersions, activePolicies, grants, approvals, ledgerHead, ledgerEvents, appendFn, headFn bool
	if pool == nil {
		return errors.New("control plane schema compatibility check failed")
	}
	if err := pool.QueryRow(ctx, required).Scan(
		&policyVersions, &activePolicies, &grants, &approvals, &ledgerHead, &ledgerEvents, &appendFn, &headFn,
	); err != nil {
		return errors.New("control plane schema compatibility check failed")
	}
	if !policyVersions || !activePolicies || !grants || !approvals || !ledgerHead || !ledgerEvents || !appendFn || !headFn {
		return errors.New("control plane schema is missing required tables or functions")
	}
	return nil
}

// ProductionHealthHandler exposes only a liveness probe and dependency-backed
// readiness probe. It intentionally does not register demo or decision routes.
func (r *ProductionRuntime) ProductionHealthHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"live"}`))
	})
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, req *http.Request) {
		if r == nil || r.pool == nil || r.sqlDB == nil || r.identity == nil || r.gateway == nil {
			http.Error(w, `{"status":"not_ready"}`, http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()
		if err := r.pool.Ping(ctx); err != nil {
			http.Error(w, `{"status":"not_ready"}`, http.StatusServiceUnavailable)
			return
		}
		if err := r.sqlDB.PingContext(ctx); err != nil {
			http.Error(w, `{"status":"not_ready"}`, http.StatusServiceUnavailable)
			return
		}
		if err := verifyControlPlaneSchema(ctx, r.pool); err != nil {
			http.Error(w, `{"status":"not_ready"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})
	return mux
}
