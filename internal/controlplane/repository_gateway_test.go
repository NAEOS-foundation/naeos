// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type repositoryGatewayTestLedger struct {
	mu         sync.Mutex
	events     []LedgerEvent
	failAppend bool
}

func (l *repositoryGatewayTestLedger) Append(ctx context.Context, event LedgerEvent) (LedgerEvent, int64, error) {
	if err := ctx.Err(); err != nil {
		return LedgerEvent{}, 0, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.failAppend {
		return LedgerEvent{}, 0, ErrRepositoryUnavailable
	}
	if event.EventType == "EXECUTION_STARTED" && event.DecisionID != "" {
		for _, existing := range l.events {
			if existing.DecisionID == event.DecisionID &&
				(existing.EventType == "EXECUTION_STARTED" || existing.EventType == "EXECUTION_ALLOWED" || existing.EventType == "EXECUTION_UNKNOWN") {
				return LedgerEvent{}, 0, ErrRepositoryConflict
			}
		}
	}
	event = cloneLedgerEvent(event)
	event.ID = controlPlaneID("EVT")
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if len(l.events) > 0 {
		event.PreviousHash = l.events[len(l.events)-1].EventHash
	}
	event.EventHash = hashLedgerEvent(event)
	l.events = append(l.events, event)
	return event, int64(len(l.events)), nil
}

func (l *repositoryGatewayTestLedger) LoadVerified(ctx context.Context) ([]LedgerEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]LedgerEvent, len(l.events))
	for i := range l.events {
		out[i] = cloneLedgerEvent(l.events[i])
	}
	return out, nil
}

func (l *repositoryGatewayTestLedger) Decision(ctx context.Context, id string) (LedgerEvent, error) {
	if err := ctx.Err(); err != nil {
		return LedgerEvent{}, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := len(l.events) - 1; i >= 0; i-- {
		e := l.events[i]
		if e.EventType == "AUTHORIZATION_DECISION" && e.DecisionID == id {
			return cloneLedgerEvent(e), nil
		}
	}
	return LedgerEvent{}, ErrRepositoryNotFound
}

func (l *repositoryGatewayTestLedger) HasExecution(ctx context.Context, id string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, e := range l.events {
		if e.DecisionID == id && (e.EventType == "EXECUTION_STARTED" || e.EventType == "EXECUTION_ALLOWED" || e.EventType == "EXECUTION_UNKNOWN") {
			return true, nil
		}
	}
	return false, nil
}

func (l *repositoryGatewayTestLedger) countType(eventType string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	count := 0
	for _, e := range l.events {
		if e.EventType == eventType {
			count++
		}
	}
	return count
}

func repositoryGatewayFixture(t *testing.T) (*RepositoryDecisionGateway, AuthorizeRequest) {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UTC()
	policy := &Policy{ID: "repo-gateway-policy", Version: 1, Status: "active", AllowedCapabilities: []Capability{"artifact.read"}}
	policyStore := NewPolicyStore()
	if err := policyStore.Set(policy); err != nil {
		t.Fatal(err)
	}
	policies := NewMemoryPolicyRepository(policyStore)
	grants := NewMemoryGrantRepository()
	grant := Grant{GrantID: "repo-gateway-grant", AgentID: "agent-gateway", PolicyID: policy.ID, PolicyVersion: 1, Capabilities: []Capability{"artifact.read"}, CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour), Status: "active"}
	if err := grants.Create(ctx, grant); err != nil {
		t.Fatal(err)
	}
	ledger := &repositoryGatewayTestLedger{}
	gateway := NewRepositoryDecisionGateway(policies, grants, nil, ledger)
	req := AuthorizeRequest{RequestID: "request-gateway", AgentID: "agent-gateway", Action: Action{AgentID: "agent-gateway", Capability: "artifact.read", ArtifactHash: "sha256:artifact"}, Timestamp: now}
	return gateway, req
}

func TestRepositoryDecisionGatewayAuthorizesCanonicalRepositoryState(t *testing.T) {
	gateway, req := repositoryGatewayFixture(t)
	// Deliberately malformed caller-provided grant/policy content must not become authority.
	req.Grant = &Grant{GrantID: "forged", AgentID: "attacker", Status: "active", Capabilities: []Capability{"admin"}}
	req.Policy = &Policy{ID: "forged-policy", Version: 99, Status: "active", AllowedCapabilities: []Capability{"admin"}}
	result, err := gateway.Authorize(context.Background(), req)
	if err != nil {
		t.Fatalf("Authorize error: %v", err)
	}
	if result.Status != DecisionAllow || result.DecisionID == "" {
		t.Fatalf("Authorize = %+v; want committed ALLOW", result)
	}
}

func TestRepositoryDecisionGatewayFailsClosedWhenEvidenceAppendFails(t *testing.T) {
	gateway, req := repositoryGatewayFixture(t)
	gateway.Ledger.(*repositoryGatewayTestLedger).failAppend = true
	result, err := gateway.Authorize(context.Background(), req)
	if !errors.Is(err, ErrRepositoryUnavailable) {
		t.Fatalf("error = %v; want repository unavailable", err)
	}
	if result.Status != DecisionIndeterminate || result.DecisionID != "" {
		t.Fatalf("result = %+v; want INDETERMINATE without decision ID", result)
	}
}

func TestRepositoryDecisionGatewayExecutionIntentPrecedesSideEffect(t *testing.T) {
	gateway, req := repositoryGatewayFixture(t)
	result, err := gateway.Authorize(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	called := false
	executed, receipt, err := gateway.Execute(context.Background(), req, result, func(context.Context) error {
		called = true
		if gateway.Ledger.(*repositoryGatewayTestLedger).countType("EXECUTION_STARTED") != 1 {
			t.Fatal("side effect invoked before durable intent")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if !called || executed.Status != DecisionAllow || receipt.EventType != "EXECUTION_ALLOWED" {
		t.Fatalf("Execute = result %+v receipt %+v called=%v", executed, receipt, called)
	}
	if got := gateway.Ledger.(*repositoryGatewayTestLedger).countType("EXECUTION_ALLOWED"); got != 1 {
		t.Fatalf("success receipts = %d; want 1", got)
	}
}

func TestRepositoryDecisionGatewayAmbiguousSideEffectIsUnknownAndNotRetryable(t *testing.T) {
	gateway, req := repositoryGatewayFixture(t)
	result, err := gateway.Authorize(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	executed, _, sideEffectErr := gateway.Execute(context.Background(), req, result, func(context.Context) error {
		return errors.New("remote timeout after request submission")
	})
	if sideEffectErr == nil || executed.Status != DecisionUnknown {
		t.Fatalf("Execute = %+v, err=%v; want UNKNOWN", executed, sideEffectErr)
	}
	if got := gateway.Ledger.(*repositoryGatewayTestLedger).countType("EXECUTION_UNKNOWN"); got != 1 {
		t.Fatalf("unknown evidence count = %d; want 1", got)
	}
	_, _, retryErr := gateway.Execute(context.Background(), req, result, func(context.Context) error { t.Fatal("must not retry side effect"); return nil })
	if !errors.Is(retryErr, ErrRepositoryConflict) {
		t.Fatalf("retry error = %v; want conflict", retryErr)
	}
}

type unavailableGrantRepository struct {
	GrantRepository
}

func (r unavailableGrantRepository) GetByAgent(context.Context, string) (Grant, error) {
	return Grant{}, ErrRepositoryUnavailable
}

type unavailableLedgerExecutionLookup struct {
	LedgerRepository
}

func (r unavailableLedgerExecutionLookup) HasExecution(context.Context, string) (bool, error) {
	return false, ErrRepositoryUnavailable
}

type unavailablePolicyCheckRepository struct {
	PolicyRepository
}

func (r unavailablePolicyCheckRepository) CheckActive(context.Context, string, int) error {
	return ErrRepositoryUnavailable
}

func TestRepositoryDecisionGatewayGrantStoreOutageIsIndeterminate(t *testing.T) {
	gateway, req := repositoryGatewayFixture(t)
	gateway.Grants = unavailableGrantRepository{GrantRepository: gateway.Grants}
	result, err := gateway.Authorize(context.Background(), req)
	if !errors.Is(err, ErrRepositoryUnavailable) || result.Status != DecisionIndeterminate {
		t.Fatalf("Authorize = %+v, err=%v; want INDETERMINATE on grant-store outage", result, err)
	}
}

func TestRepositoryDecisionGatewayExecutionLookupOutageDoesNotDenyOrInvoke(t *testing.T) {
	gateway, req := repositoryGatewayFixture(t)
	result, err := gateway.Authorize(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	gateway.Ledger = unavailableLedgerExecutionLookup{LedgerRepository: gateway.Ledger}
	called := false
	executed, _, err := gateway.Execute(context.Background(), req, result, func(context.Context) error {
		called = true
		return nil
	})
	if !errors.Is(err, ErrRepositoryUnavailable) || executed.Status != DecisionIndeterminate || called {
		t.Fatalf("Execute = %+v, err=%v, side_effect_called=%t; want INDETERMINATE without side effect", executed, err, called)
	}
}

func TestRepositoryDecisionGatewayPolicyCheckOutageDoesNotDenyOrInvoke(t *testing.T) {
	gateway, req := repositoryGatewayFixture(t)
	result, err := gateway.Authorize(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	gateway.Policies = unavailablePolicyCheckRepository{PolicyRepository: gateway.Policies}
	called := false
	executed, _, err := gateway.Execute(context.Background(), req, result, func(context.Context) error {
		called = true
		return nil
	})
	if !errors.Is(err, ErrRepositoryUnavailable) || executed.Status != DecisionIndeterminate || called {
		t.Fatalf("Execute = %+v, err=%v, side_effect_called=%t; want INDETERMINATE without side effect", executed, err, called)
	}
}

func TestRepositoryDecisionGatewayConcurrentExecuteHasSingleWinner(t *testing.T) {
	gateway, req := repositoryGatewayFixture(t)
	result, err := gateway.Authorize(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	sideEffects, conflicts, unexpected := 0, 0, []error{}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, executeErr := gateway.Execute(context.Background(), req, result, func(context.Context) error {
				mu.Lock()
				sideEffects++
				mu.Unlock()
				return nil
			})
			mu.Lock()
			defer mu.Unlock()
			if errors.Is(executeErr, ErrRepositoryConflict) {
				conflicts++
			} else if executeErr != nil {
				unexpected = append(unexpected, executeErr)
			}
		}()
	}
	wg.Wait()
	if sideEffects != 1 || conflicts != 1 || len(unexpected) != 0 {
		t.Fatalf("concurrent Execute: side_effects=%d conflicts=%d unexpected=%v; want one winner and one conflict", sideEffects, conflicts, unexpected)
	}
}
