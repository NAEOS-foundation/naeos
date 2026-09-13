// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	sm := NewStateManagerWithDir(dir)

	record := &DeploymentRecord{
		Project:     "myapp",
		Provider:    AWS,
		Environment: "prod",
		Region:      "us-east-1",
		Resources: []DeployedResource{
			{Name: "uploads", Type: "aws_s3_bucket", ID: "arn:aws:s3:::myapp-prod-uploads"},
			{Name: "api", Type: "aws_ecs_service", ID: "arn:aws:ecs:::myapp-prod-api"},
		},
		TerraformDir: "/tmp/tf-myapp",
		Status:       "deployed",
		Timestamp:    time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	if err := sm.Save(record); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := sm.Load("myapp", AWS)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Project != record.Project {
		t.Errorf("project: got %s, want %s", loaded.Project, record.Project)
	}
	if loaded.Provider != record.Provider {
		t.Errorf("provider: got %s, want %s", loaded.Provider, record.Provider)
	}
	if loaded.Environment != record.Environment {
		t.Errorf("environment: got %s, want %s", loaded.Environment, record.Environment)
	}
	if loaded.Region != record.Region {
		t.Errorf("region: got %s, want %s", loaded.Region, record.Region)
	}
	if loaded.TerraformDir != record.TerraformDir {
		t.Errorf("terraform_dir: got %s, want %s", loaded.TerraformDir, record.TerraformDir)
	}
	if loaded.Status != record.Status {
		t.Errorf("status: got %s, want %s", loaded.Status, record.Status)
	}
	if len(loaded.Resources) != 2 {
		t.Fatalf("resources: got %d, want 2", len(loaded.Resources))
	}
	if loaded.Resources[0].Name != "uploads" {
		t.Errorf("resource[0].name: got %s, want uploads", loaded.Resources[0].Name)
	}
}

func TestSaveMultipleProviders(t *testing.T) {
	dir := t.TempDir()
	sm := NewStateManagerWithDir(dir)

	for _, provider := range []CloudProvider{AWS, GCP, Azure} {
		record := &DeploymentRecord{
			Project:  "multi",
			Provider: provider,
			Status:   "deployed",
		}
		if err := sm.Save(record); err != nil {
			t.Fatalf("Save %s failed: %v", provider, err)
		}
	}

	for _, provider := range []CloudProvider{AWS, GCP, Azure} {
		loaded, err := sm.Load("multi", provider)
		if err != nil {
			t.Fatalf("Load %s failed: %v", provider, err)
		}
		if loaded.Provider != provider {
			t.Errorf("provider: got %s, want %s", loaded.Provider, provider)
		}
	}
}

func TestListDeployments(t *testing.T) {
	dir := t.TempDir()
	sm := NewStateManagerWithDir(dir)

	records := []*DeploymentRecord{
		{Project: "alpha", Provider: AWS, Status: "deployed", Timestamp: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)},
		{Project: "beta", Provider: GCP, Status: "deployed", Timestamp: time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)},
		{Project: "gamma", Provider: Azure, Status: "deployed", Timestamp: time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)},
	}

	for _, r := range records {
		if err := sm.Save(r); err != nil {
			t.Fatalf("Save %s failed: %v", r.Project, err)
		}
	}

	list, err := sm.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 3 {
		t.Fatalf("expected 3 deployments, got %d", len(list))
	}

	if list[0].Project != "beta" {
		t.Errorf("expected beta first (newest), got %s", list[0].Project)
	}
	if list[1].Project != "alpha" {
		t.Errorf("expected alpha second, got %s", list[1].Project)
	}
	if list[2].Project != "gamma" {
		t.Errorf("expected gamma third (oldest), got %s", list[2].Project)
	}
}

func TestListEmpty(t *testing.T) {
	dir := t.TempDir()
	sm := NewStateManagerWithDir(dir)

	list, err := sm.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}

func TestListNonexistentDir(t *testing.T) {
	sm := NewStateManagerWithDir("/nonexistent/path/naeos-test")
	list, err := sm.List()
	if err != nil {
		t.Fatalf("List on nonexistent dir should return empty, got error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}

func TestDeleteDeployment(t *testing.T) {
	dir := t.TempDir()
	sm := NewStateManagerWithDir(dir)

	record := &DeploymentRecord{
		Project:  "deleteme",
		Provider: AWS,
		Status:   "deployed",
	}
	if err := sm.Save(record); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := sm.Delete("deleteme", AWS); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := sm.Load("deleteme", AWS)
	if err == nil {
		t.Error("expected error loading deleted deployment")
	}
}

func TestDeleteNonexistent(t *testing.T) {
	dir := t.TempDir()
	sm := NewStateManagerWithDir(dir)

	err := sm.Delete("ghost", AWS)
	if err == nil {
		t.Error("expected error deleting nonexistent deployment")
	}
}

func TestLoadNonexistentReturnsError(t *testing.T) {
	dir := t.TempDir()
	sm := NewStateManagerWithDir(dir)

	_, err := sm.Load("nothere", AWS)
	if err == nil {
		t.Error("expected error loading nonexistent deployment")
	}
}

func TestSaveMissingProject(t *testing.T) {
	sm := NewStateManagerWithDir(t.TempDir())
	err := sm.Save(&DeploymentRecord{Provider: AWS})
	if err == nil {
		t.Error("expected error for missing project")
	}
}

func TestSaveMissingProvider(t *testing.T) {
	sm := NewStateManagerWithDir(t.TempDir())
	err := sm.Save(&DeploymentRecord{Project: "test"})
	if err == nil {
		t.Error("expected error for missing provider")
	}
}

func TestDeleteCleansUpEmptyParent(t *testing.T) {
	dir := t.TempDir()
	sm := NewStateManagerWithDir(dir)

	record := &DeploymentRecord{
		Project:  "orphan",
		Provider: GCP,
		Status:   "deployed",
	}
	if err := sm.Save(record); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	projectDir := filepath.Join(dir, "orphan")
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Fatal("project dir should exist before delete")
	}

	if err := sm.Delete("orphan", GCP); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		t.Error("project dir should be removed after deleting last provider")
	}
}

func TestSaveSetsTimestamp(t *testing.T) {
	sm := NewStateManagerWithDir(t.TempDir())

	record := &DeploymentRecord{
		Project:  "ts-test",
		Provider: AWS,
		Status:   "deployed",
	}

	if err := sm.Save(record); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := sm.Load("ts-test", AWS)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}
