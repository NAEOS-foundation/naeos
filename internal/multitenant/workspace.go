// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package multitenant

import (
	"fmt"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// Tenant represents an isolated workspace with its own data scope.
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Plan      string    `json:"plan"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Workspace struct {
	mu      sync.RWMutex
	tenants map[string]*Tenant
	members map[string]map[string]struct{}
}

func New() *Workspace {
	return &Workspace{
		tenants: make(map[string]*Tenant),
		members: make(map[string]map[string]struct{}),
	}
}

func (w *Workspace) CreateTenant(id, name, plan string) (*Tenant, error) {
	return w.createTenant(id, name, plan, "")
}

// CreateTenantForUser creates a tenant and grants initial membership to its creator.
func (w *Workspace) CreateTenantForUser(id, name, plan, userID string) (*Tenant, error) {
	if userID == "" {
		return nil, naeoserr.New(naeoserr.ErrValidation, "tenant creator user ID must not be empty")
	}
	return w.createTenant(id, name, plan, userID)
}

func (w *Workspace) createTenant(id, name, plan, creatorUserID string) (*Tenant, error) {
	if id == "" {
		return nil, naeoserr.New(naeoserr.ErrValidation, "tenant ID must not be empty")
	}
	if name == "" {
		return nil, naeoserr.New(naeoserr.ErrValidation, "tenant name must not be empty")
	}

	now := time.Now()
	t := &Tenant{
		ID:        id,
		Name:      name,
		Plan:      plan,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.tenants[id]; exists {
		return nil, naeoserr.New(naeoserr.ErrConflict, fmt.Sprintf("tenant %q already exists", id))
	}

	w.tenants[id] = t
	w.members[id] = make(map[string]struct{})
	if creatorUserID != "" {
		w.members[id][creatorUserID] = struct{}{}
	}
	return t, nil
}

func (w *Workspace) GetTenant(id string) (*Tenant, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	t, ok := w.tenants[id]
	if !ok {
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("tenant %q not found", id))
	}
	if !t.Active {
		return nil, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("tenant %q is deactivated", id))
	}
	return t, nil
}

func (w *Workspace) ListTenants() []*Tenant {
	w.mu.RLock()
	defer w.mu.RUnlock()

	out := make([]*Tenant, 0, len(w.tenants))
	for _, t := range w.tenants {
		out = append(out, t)
	}
	return out
}

func (w *Workspace) DeactivateTenant(id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	t, ok := w.tenants[id]
	if !ok {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("tenant %q not found", id))
	}

	t.Active = false
	t.UpdatedAt = time.Now()
	return nil
}

func (w *Workspace) ActivateTenant(id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	t, ok := w.tenants[id]
	if !ok {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("tenant %q not found", id))
	}

	t.Active = true
	t.UpdatedAt = time.Now()
	return nil
}

func (w *Workspace) TenantCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.tenants)
}

// AddMember grants a user membership in an active tenant.
func (w *Workspace) AddMember(tenantID, userID string) error {
	if tenantID == "" || userID == "" {
		return naeoserr.New(naeoserr.ErrValidation, "tenant ID and user ID are required")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	tenant, ok := w.tenants[tenantID]
	if !ok {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("tenant %q not found", tenantID))
	}
	if !tenant.Active {
		return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("tenant %q is deactivated", tenantID))
	}
	if w.members[tenantID] == nil {
		w.members[tenantID] = make(map[string]struct{})
	}
	w.members[tenantID][userID] = struct{}{}
	return nil
}

// IsMember reports whether userID belongs to an active tenant.
func (w *Workspace) IsMember(tenantID, userID string) bool {
	if tenantID == "" || userID == "" {
		return false
	}
	w.mu.RLock()
	defer w.mu.RUnlock()
	tenant, ok := w.tenants[tenantID]
	if !ok || !tenant.Active {
		return false
	}
	_, ok = w.members[tenantID][userID]
	return ok
}

// SingleTenantForUser returns a tenant only when the user has exactly one active membership.
func (w *Workspace) SingleTenantForUser(userID string) (string, bool) {
	if userID == "" {
		return "", false
	}
	w.mu.RLock()
	defer w.mu.RUnlock()
	var found string
	for tenantID, members := range w.members {
		tenant, ok := w.tenants[tenantID]
		if !ok || !tenant.Active {
			continue
		}
		if _, member := members[userID]; member {
			if found != "" {
				return "", false
			}
			found = tenantID
		}
	}
	return found, found != ""
}
