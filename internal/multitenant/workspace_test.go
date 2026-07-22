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
