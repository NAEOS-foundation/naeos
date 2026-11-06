// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package multitenant

import (
	"testing"
)

func TestCreateAndGetTenant(t *testing.T) {
	w := New()
	tenant, err := w.CreateTenant("tenant-1", "Acme Corp", "enterprise")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenant.ID != "tenant-1" {
		t.Errorf("expected ID tenant-1, got %s", tenant.ID)
	}
	if !tenant.Active {
		t.Error("expected tenant to be active")
	}

	got, err := w.GetTenant("tenant-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Acme Corp" {
		t.Errorf("expected name Acme Corp, got %s", got.Name)
	}
}

func TestCreateDuplicateTenant(t *testing.T) {
	w := New()
	w.CreateTenant("t1", "Test", "free")
	_, err := w.CreateTenant("t1", "Test2", "free")
	if err == nil {
		t.Fatal("expected error for duplicate tenant")
	}
}

func TestGetNonexistentTenant(t *testing.T) {
	w := New()
	_, err := w.GetTenant("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent tenant")
	}
}

func TestDeactivateAndActivate(t *testing.T) {
	w := New()
	w.CreateTenant("t1", "Test", "free")

	if err := w.DeactivateTenant("t1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := w.GetTenant("t1")
	if err == nil {
		t.Fatal("expected error for deactivated tenant")
	}

	if err := w.ActivateTenant("t1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = w.GetTenant("t1")
	if err != nil {
		t.Fatalf("unexpected error after reactivation: %v", err)
	}
}

func TestListTenants(t *testing.T) {
	w := New()
	w.CreateTenant("t1", "A", "free")
	w.CreateTenant("t2", "B", "enterprise")

	tenants := w.ListTenants()
	if len(tenants) != 2 {
		t.Errorf("expected 2 tenants, got %d", len(tenants))
	}
}

func TestTenantCount(t *testing.T) {
	w := New()
	if c := w.TenantCount(); c != 0 {
		t.Errorf("expected 0, got %d", c)
	}
	w.CreateTenant("t1", "A", "free")
	if c := w.TenantCount(); c != 1 {
		t.Errorf("expected 1, got %d", c)
	}
}

func TestCreateTenantValidations(t *testing.T) {
	w := New()
	if _, err := w.CreateTenant("", "name", "free"); err == nil {
		t.Error("expected error for empty ID")
	}
	if _, err := w.CreateTenant("id", "", "free"); err == nil {
		t.Error("expected error for empty name")
	}
}


func TestTenantMembershipAndSingleTenantResolution(t *testing.T) {
	w := New()
	if _, err := w.CreateTenantForUser("tenant-a", "Tenant A", "free", "user-1"); err != nil {
		t.Fatal(err)
	}
	if !w.IsMember("tenant-a", "user-1") {
		t.Fatal("creator should be the initial member")
	}
	if id, ok := w.SingleTenantForUser("user-1"); !ok || id != "tenant-a" {
		t.Fatalf("expected single membership tenant-a, got %q %v", id, ok)
	}
	if w.IsMember("tenant-a", "user-2") {
		t.Fatal("unassigned user must not be a tenant member")
	}
	if err := w.AddMember("tenant-a", "user-2"); err != nil {
		t.Fatal(err)
	}
	if !w.IsMember("tenant-a", "user-2") {
		t.Fatal("added user should be a member")
	}
	if _, err := w.CreateTenantForUser("tenant-b", "Tenant B", "free", "user-1"); err != nil {
		t.Fatal(err)
	}
	if id, ok := w.SingleTenantForUser("user-1"); ok || id != "" {
		t.Fatalf("multiple memberships must require explicit tenant selection, got %q %v", id, ok)
	}
}
