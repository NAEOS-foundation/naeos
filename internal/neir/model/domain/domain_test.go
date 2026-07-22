package domain

import "testing"

func TestDomain_ZeroValue(t *testing.T) {
	var d Domain
	if d.Name != "" {
		t.Error("expected empty Name")
	}
	if d.BoundedContexts != nil {
		t.Error("expected nil BoundedContexts")
	}
}

func TestDomain_Full(t *testing.T) {
	d := Domain{
		Name:        "billing",
		Description: "Billing domain",
		BoundedContexts: []BoundedContext{
			{Name: "invoicing", Description: "Invoice management", Modules: []string{"invoices"}},
		},
		Aggregates: []Aggregate{
			{Name: "Invoice", RootEntity: "Invoice", Entities: []string{"Invoice", "LineItem"}},
		},
		Entities: []Entity{
			{Name: "Invoice", Attributes: map[string]string{"amount": "decimal"}},
		},
		ValueObjects: []ValueObject{
			{Name: "Money", Attributes: map[string]string{"currency": "string"}},
		},
		Attributes: map[string]string{"key": "val"},
	}
	if d.Name != "billing" {
		t.Errorf("expected billing, got %s", d.Name)
	}
	if len(d.BoundedContexts) != 1 {
		t.Errorf("expected 1 bounded context, got %d", len(d.BoundedContexts))
	}
	if d.BoundedContexts[0].Name != "invoicing" {
		t.Errorf("expected invoicing, got %s", d.BoundedContexts[0].Name)
	}
	if len(d.Aggregates) != 1 {
		t.Errorf("expected 1 aggregate, got %d", len(d.Aggregates))
	}
	if d.Aggregates[0].RootEntity != "Invoice" {
		t.Errorf("expected Invoice, got %s", d.Aggregates[0].RootEntity)
	}
	if len(d.Entities) != 1 {
		t.Errorf("expected 1 entity, got %d", len(d.Entities))
	}
	if d.Entities[0].Attributes["amount"] != "decimal" {
		t.Errorf("expected decimal, got %s", d.Entities[0].Attributes["amount"])
	}
	if len(d.ValueObjects) != 1 {
		t.Errorf("expected 1 value object, got %d", len(d.ValueObjects))
	}
}

func TestBoundedContext_ZeroValue(t *testing.T) {
	var bc BoundedContext
	if bc.Name != "" {
		t.Error("expected empty Name")
	}
}
