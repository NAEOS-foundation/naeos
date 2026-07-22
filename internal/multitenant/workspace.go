package multitenant

import (
	"fmt"
	"sync"
	"time"
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
}

func New() *Workspace {
	return &Workspace{
		tenants: make(map[string]*Tenant),
	}
}

func (w *Workspace) CreateTenant(id, name, plan string) (*Tenant, error) {
	if id == "" {
		return nil, fmt.Errorf("tenant ID must not be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("tenant name must not be empty")
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
		return nil, fmt.Errorf("tenant %q already exists", id)
	}

	w.tenants[id] = t
	return t, nil
}

func (w *Workspace) GetTenant(id string) (*Tenant, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	t, ok := w.tenants[id]
	if !ok {
		return nil, fmt.Errorf("tenant %q not found", id)
	}
	if !t.Active {
		return nil, fmt.Errorf("tenant %q is deactivated", id)
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
		return fmt.Errorf("tenant %q not found", id)
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
		return fmt.Errorf("tenant %q not found", id)
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
