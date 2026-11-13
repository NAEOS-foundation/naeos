// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"errors"
	"time"
)

// Repository errors are intentionally distinct: callers must not interpret an
// infrastructure or integrity failure as an authorization denial.
var (
	ErrRepositoryNotFound    = errors.New("repository record not found")
	ErrRepositoryConflict    = errors.New("repository state conflict")
	ErrRepositoryUnavailable = errors.New("repository unavailable")
	ErrRepositoryIntegrity   = errors.New("repository integrity failure")
)

// LedgerRepository is the context-aware persistence boundary for canonical
// append-only control-plane evidence. Append returns only after commit.
type LedgerRepository interface {
	Append(context.Context, LedgerEvent) (LedgerEvent, int64, error)
	LoadVerified(context.Context) ([]LedgerEvent, error)
	Decision(context.Context, string) (LedgerEvent, error)
	HasExecution(context.Context, string) (bool, error)
}

// PolicyRepository stores immutable policy versions and authoritative active
// version state. Implementations must enforce monotonic versions atomically.
type PolicyRepository interface {
	PutVersion(context.Context, *Policy) error
	Active(context.Context, string) (*Policy, error)
	Version(context.Context, string, int) (*Policy, error)
	// CheckActive is an authoritative point-in-time check. It does not promise
	// a lock across external side effects; execution must use durable intent and
	// reconciliation rather than holding a database transaction over callbacks.
	CheckActive(context.Context, string, int) error
}

// GrantRepository persists grant lifecycle and authoritative revocation state.
type GrantRepository interface {
	Create(context.Context, Grant) error
	Get(context.Context, string) (Grant, error)
	GetByAgent(context.Context, string) (Grant, error)
	Revoke(context.Context, string, string, time.Time) error
}

// ApprovalRepository provides durable approval lifecycle operations. Transition
// must be compare-and-set on expected status, so concurrent consume has one
// winner. Bindings in Approval are validated by the service layer.
type ApprovalRepository interface {
	Create(context.Context, Approval) error
	Get(context.Context, string) (Approval, error)
	Transition(context.Context, string, string, string, time.Time) (Approval, error)
}
