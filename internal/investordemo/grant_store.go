package investordemo

import (
	"fmt"
	"sync"
	"time"
)

// GrantStore manages capability grants for agents.
type GrantStore struct {
	mu      sync.RWMutex
	grants  map[string]*CapabilityGrant // grant_id -> grant
	byAgent map[string]string           // agent_id -> grant_id (one grant per agent for simplicity)
}

// NewGrantStore creates a new grant store.
func NewGrantStore() *GrantStore {
	return &GrantStore{
		grants:  make(map[string]*CapabilityGrant),
		byAgent: make(map[string]string),
	}
}

// StoreGrant stores a new grant.
func (gs *GrantStore) StoreGrant(grant *CapabilityGrant) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if grant.GrantID == "" {
		return fmt.Errorf("grant ID is required")
	}

	if grant.AgentID == "" {
		return fmt.Errorf("agent ID is required")
	}

	gs.grants[grant.GrantID] = grant
	gs.byAgent[grant.AgentID] = grant.GrantID

	return nil
}

// GetGrant retrieves a grant by ID.
func (gs *GrantStore) GetGrant(grantID string) (*CapabilityGrant, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	grant, ok := gs.grants[grantID]
	if !ok {
		return nil, fmt.Errorf("grant %s not found", grantID)
	}

	return grant, nil
}

// GetGrantByAgent retrieves the grant for a specific agent.
func (gs *GrantStore) GetGrantByAgent(agentID string) (*CapabilityGrant, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	grantID, ok := gs.byAgent[agentID]
	if !ok {
		return nil, fmt.Errorf("no grant found for agent %s", agentID)
	}

	grant, ok := gs.grants[grantID]
	if !ok {
		return nil, fmt.Errorf("grant %s not found", grantID)
	}

	return grant, nil
}

// RevokeGrant revokes a grant.
func (gs *GrantStore) RevokeGrant(grantID string, reason string, revokedAt *time.Time) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	grant, ok := gs.grants[grantID]
	if !ok {
		return fmt.Errorf("grant %s not found", grantID)
	}

	grant.Revoked = true
	grant.RevokeReason = reason
	grant.RevokedAt = revokedAt
	grant.Status = "revoked"

	return nil
}

// ListGrants returns all grants (for auditing).
func (gs *GrantStore) ListGrants() []*CapabilityGrant {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	var grants []*CapabilityGrant
	for _, g := range gs.grants {
		grants = append(grants, g)
	}
	return grants
}
