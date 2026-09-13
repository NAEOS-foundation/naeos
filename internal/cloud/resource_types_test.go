// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"strings"
	"testing"
)

func TestSupportedResourceTypesExpanded(t *testing.T) {
	expected := []string{
		"storage", "compute", "database", "cache", "queue", "cdn",
		"serverless", "monitoring", "secrets", "dns", "networking",
	}
	if len(SupportedResourceTypes) != len(expected) {
		t.Fatalf("expected %d resource types, got %d", len(expected), len(SupportedResourceTypes))
	}
	for i, v := range expected {
		if SupportedResourceTypes[i] != v {
			t.Errorf("expected %s at index %d, got %s", v, i, SupportedResourceTypes[i])
		}
	}
}

// --- AWS New Resource Types ---

func TestAWSExportServerless(t *testing.T) {
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider: AWS, Region: "us-east-1", Project: "myapp", Environment: "prod",
		Resources: []Resource{{Name: "api", Type: ResourceServerless}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "aws_lambda_function" "api"`) {
		t.Error("missing aws_lambda_function resource")
	}
	if !strings.Contains(tf, `aws_iam_role`) {
		t.Error("missing IAM role for Lambda")
	}
}

func TestAWSExportMonitoring(t *testing.T) {
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider: AWS, Region: "us-east-1", Project: "myapp", Environment: "prod",
		Resources: []Resource{{Name: "cpu-alert", Type: ResourceMonitoring}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "aws_cloudwatch_metric_alarm" "cpu-alert"`) {
		t.Error("missing aws_cloudwatch_metric_alarm resource")
	}
}

func TestAWSExportSecrets(t *testing.T) {
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider: AWS, Region: "us-east-1", Project: "myapp", Environment: "prod",
		Resources: []Resource{{Name: "db-creds", Type: ResourceSecrets}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "aws_secretsmanager_secret" "db-creds"`) {
		t.Error("missing aws_secretsmanager_secret resource")
	}
}

func TestAWSExportDNS(t *testing.T) {
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider: AWS, Region: "us-east-1", Project: "myapp", Environment: "prod",
		Resources: []Resource{{Name: "example.com", Type: ResourceDNS}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "aws_route53_zone" "example.com"`) {
		t.Error("missing aws_route53_zone resource")
	}
}

func TestAWSExportNetworking(t *testing.T) {
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider: AWS, Region: "us-east-1", Project: "myapp", Environment: "prod",
		Resources: []Resource{{Name: "main", Type: ResourceNetworking}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "aws_vpc" "main"`) {
		t.Error("missing aws_vpc resource")
	}
	if !strings.Contains(tf, `resource "aws_subnet" "main"`) {
		t.Error("missing aws_subnet resource")
	}
}

func TestAWSPlanNewResourceTypes(t *testing.T) {
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider: AWS, Region: "us-east-1", Project: "myapp", Environment: "prod",
		Resources: []Resource{
			{Name: "api", Type: ResourceServerless},
			{Name: "alerts", Type: ResourceMonitoring},
			{Name: "creds", Type: ResourceSecrets},
			{Name: "zone", Type: ResourceDNS},
			{Name: "vpc", Type: ResourceNetworking},
		},
	}
	planResult, err := adapter.Plan(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(planResult.Resources) != 5 {
		t.Fatalf("expected 5 resources, got %d", len(planResult.Resources))
	}
	expectedTypes := []string{
		"aws_lambda_function",
		"aws_cloudwatch_metric_alarm",
		"aws_secretsmanager_secret",
		"aws_route53_zone",
		"aws_vpc",
	}
	for i, expected := range expectedTypes {
		if planResult.Resources[i].Type != expected {
			t.Errorf("resource %d: expected %s, got %s", i, expected, planResult.Resources[i].Type)
		}
	}
}

// --- GCP New Resource Types ---

func TestGCPExportServerless(t *testing.T) {
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider: GCP, Region: "us-central1", Project: "myapp", Environment: "dev",
		Resources: []Resource{{Name: "func", Type: ResourceServerless}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "google_cloudfunctions2_function" "func"`) {
		t.Error("missing google_cloudfunctions2_function resource")
	}
}

func TestGCPExportMonitoring(t *testing.T) {
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider: GCP, Region: "us-central1", Project: "myapp", Environment: "dev",
		Resources: []Resource{{Name: "alert", Type: ResourceMonitoring}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "google_monitoring_alert_policy" "alert"`) {
		t.Error("missing google_monitoring_alert_policy resource")
	}
}

func TestGCPExportSecrets(t *testing.T) {
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider: GCP, Region: "us-central1", Project: "myapp", Environment: "dev",
		Resources: []Resource{{Name: "keys", Type: ResourceSecrets}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "google_secret_manager_secret" "keys"`) {
		t.Error("missing google_secret_manager_secret resource")
	}
}

func TestGCPExportDNS(t *testing.T) {
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider: GCP, Region: "us-central1", Project: "myapp", Environment: "dev",
		Resources: []Resource{{Name: "example.com", Type: ResourceDNS}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "google_dns_managed_zone" "example.com"`) {
		t.Error("missing google_dns_managed_zone resource")
	}
}

func TestGCPExportNetworking(t *testing.T) {
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider: GCP, Region: "us-central1", Project: "myapp", Environment: "dev",
		Resources: []Resource{{Name: "mynetwork", Type: ResourceNetworking}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "google_compute_network" "mynetwork"`) {
		t.Error("missing google_compute_network resource")
	}
}

func TestGCPPlanNewResourceTypes(t *testing.T) {
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider: GCP, Region: "us-central1", Project: "myapp", Environment: "dev",
		Resources: []Resource{
			{Name: "func", Type: ResourceServerless},
			{Name: "alert", Type: ResourceMonitoring},
			{Name: "keys", Type: ResourceSecrets},
			{Name: "zone", Type: ResourceDNS},
			{Name: "vpc", Type: ResourceNetworking},
		},
	}
	planResult, err := adapter.Plan(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(planResult.Resources) != 5 {
		t.Fatalf("expected 5 resources, got %d", len(planResult.Resources))
	}
	expectedTypes := []string{
		"google_cloudfunctions2_function",
		"google_monitoring_alert_policy",
		"google_secret_manager_secret",
		"google_dns_managed_zone",
		"google_compute_network",
	}
	for i, expected := range expectedTypes {
		if planResult.Resources[i].Type != expected {
			t.Errorf("resource %d: expected %s, got %s", i, expected, planResult.Resources[i].Type)
		}
	}
}

// --- Azure New Resource Types ---

func TestAzureExportServerless(t *testing.T) {
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider: Azure, Region: "eastus", Project: "myapp", Environment: "staging",
		Resources: []Resource{{Name: "myfunc", Type: ResourceServerless}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "azurerm_linux_function_app" "myfunc"`) {
		t.Error("missing azurerm_linux_function_app resource")
	}
	if !strings.Contains(tf, `resource "azurerm_service_plan"`) {
		t.Error("missing azurerm_service_plan resource")
	}
}

func TestAzureExportMonitoring(t *testing.T) {
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider: Azure, Region: "eastus", Project: "myapp", Environment: "staging",
		Resources: []Resource{{Name: "alerts", Type: ResourceMonitoring}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "azurerm_monitor_action_group" "alerts"`) {
		t.Error("missing azurerm_monitor_action_group resource")
	}
	if !strings.Contains(tf, `resource "azurerm_monitor_metric_alert" "alerts"`) {
		t.Error("missing azurerm_monitor_metric_alert resource")
	}
}

func TestAzureExportSecrets(t *testing.T) {
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider: Azure, Region: "eastus", Project: "myapp", Environment: "staging",
		Resources: []Resource{{Name: "vault", Type: ResourceSecrets}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "azurerm_key_vault" "vault"`) {
		t.Error("missing azurerm_key_vault resource")
	}
	if !strings.Contains(tf, `data "azurerm_client_config" "current"`) {
		t.Error("missing azurerm_client_config data source")
	}
}

func TestAzureExportDNS(t *testing.T) {
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider: Azure, Region: "eastus", Project: "myapp", Environment: "staging",
		Resources: []Resource{{Name: "example.com", Type: ResourceDNS}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "azurerm_dns_zone" "example.com"`) {
		t.Error("missing azurerm_dns_zone resource")
	}
}

func TestAzureExportNetworking(t *testing.T) {
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider: Azure, Region: "eastus", Project: "myapp", Environment: "staging",
		Resources: []Resource{{Name: "main", Type: ResourceNetworking}},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(tf, `resource "azurerm_virtual_network" "main"`) {
		t.Error("missing azurerm_virtual_network resource")
	}
	if !strings.Contains(tf, `resource "azurerm_subnet" "main"`) {
		t.Error("missing azurerm_subnet resource")
	}
}

func TestAzurePlanNewResourceTypes(t *testing.T) {
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider: Azure, Region: "eastus", Project: "myapp", Environment: "staging",
		Resources: []Resource{
			{Name: "func", Type: ResourceServerless},
			{Name: "alerts", Type: ResourceMonitoring},
			{Name: "vault", Type: ResourceSecrets},
			{Name: "zone", Type: ResourceDNS},
			{Name: "vnet", Type: ResourceNetworking},
		},
	}
	planResult, err := adapter.Plan(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(planResult.Resources) != 5 {
		t.Fatalf("expected 5 resources, got %d", len(planResult.Resources))
	}
	expectedTypes := []string{
		"azurerm_linux_function_app",
		"azurerm_monitor_action_group",
		"azurerm_key_vault",
		"azurerm_dns_zone",
		"azurerm_virtual_network",
	}
	for i, expected := range expectedTypes {
		if planResult.Resources[i].Type != expected {
			t.Errorf("resource %d: expected %s, got %s", i, expected, planResult.Resources[i].Type)
		}
	}
}

// --- Validation tests ---

func TestAWSValidateWithNewResourceTypes(t *testing.T) {
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider: AWS, Region: "us-east-1", Project: "myapp", Environment: "prod",
		Resources: []Resource{
			{Name: "api", Type: ResourceServerless},
			{Name: "alerts", Type: ResourceMonitoring},
			{Name: "creds", Type: ResourceSecrets},
			{Name: "zone", Type: ResourceDNS},
			{Name: "vpc", Type: ResourceNetworking},
		},
	}
	if err := adapter.Validate(config); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestGCPValidateWithNewResourceTypes(t *testing.T) {
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider: GCP, Region: "us-central1", Project: "myapp", Environment: "dev",
		Resources: []Resource{
			{Name: "func", Type: ResourceServerless},
			{Name: "alert", Type: ResourceMonitoring},
			{Name: "keys", Type: ResourceSecrets},
			{Name: "zone", Type: ResourceDNS},
			{Name: "vpc", Type: ResourceNetworking},
		},
	}
	if err := adapter.Validate(config); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestAzureValidateWithNewResourceTypes(t *testing.T) {
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider: Azure, Region: "eastus", Project: "myapp", Environment: "staging",
		Resources: []Resource{
			{Name: "func", Type: ResourceServerless},
			{Name: "alerts", Type: ResourceMonitoring},
			{Name: "vault", Type: ResourceSecrets},
			{Name: "zone", Type: ResourceDNS},
			{Name: "vnet", Type: ResourceNetworking},
		},
	}
	if err := adapter.Validate(config); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

// --- All types together ---

func TestAWSExportAllResourceTypes(t *testing.T) {
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider: AWS, Region: "us-east-1", Project: "myapp", Environment: "prod",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
			{Name: "db", Type: ResourceDatabase},
			{Name: "cache", Type: ResourceCache},
			{Name: "queue", Type: ResourceQueue},
			{Name: "cdn", Type: ResourceCDN},
			{Name: "func", Type: ResourceServerless},
			{Name: "alerts", Type: ResourceMonitoring},
			{Name: "creds", Type: ResourceSecrets},
			{Name: "zone", Type: ResourceDNS},
			{Name: "vpc", Type: ResourceNetworking},
		},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedBlocks := []string{
		`resource "aws_s3_bucket" "uploads"`,
		`resource "aws_ecs_cluster"`,
		`resource "aws_rds_instance" "db"`,
		`resource "aws_elasticache_cluster" "cache"`,
		`resource "aws_sqs_queue" "queue"`,
		`resource "aws_cloudfront_distribution" "cdn"`,
		`resource "aws_lambda_function" "func"`,
		`resource "aws_cloudwatch_metric_alarm" "alerts"`,
		`resource "aws_secretsmanager_secret" "creds"`,
		`resource "aws_route53_zone" "zone"`,
		`resource "aws_vpc" "vpc"`,
	}
	for _, block := range expectedBlocks {
		if !strings.Contains(tf, block) {
			t.Errorf("missing resource block: %s", block)
		}
	}
}

func TestGCPExportAllResourceTypes(t *testing.T) {
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider: GCP, Region: "us-central1", Project: "myapp", Environment: "dev",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
			{Name: "db", Type: ResourceDatabase},
			{Name: "cache", Type: ResourceCache},
			{Name: "queue", Type: ResourceQueue},
			{Name: "cdn", Type: ResourceCDN},
			{Name: "func", Type: ResourceServerless},
			{Name: "alert", Type: ResourceMonitoring},
			{Name: "keys", Type: ResourceSecrets},
			{Name: "zone", Type: ResourceDNS},
			{Name: "net", Type: ResourceNetworking},
		},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedBlocks := []string{
		`resource "google_storage_bucket" "uploads"`,
		`resource "google_cloud_run_service" "api"`,
		`resource "google_sql_database_instance" "db"`,
		`resource "google_redis_instance" "cache"`,
		`resource "google_pubsub_topic" "queue"`,
		`resource "google_compute_backend_bucket" "cdn"`,
		`resource "google_cloudfunctions2_function" "func"`,
		`resource "google_monitoring_alert_policy" "alert"`,
		`resource "google_secret_manager_secret" "keys"`,
		`resource "google_dns_managed_zone" "zone"`,
		`resource "google_compute_network" "net"`,
	}
	for _, block := range expectedBlocks {
		if !strings.Contains(tf, block) {
			t.Errorf("missing resource block: %s", block)
		}
	}
}

func TestAzureExportAllResourceTypes(t *testing.T) {
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider: Azure, Region: "eastus", Project: "myapp", Environment: "staging",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
			{Name: "db", Type: ResourceDatabase},
			{Name: "cache", Type: ResourceCache},
			{Name: "queue", Type: ResourceQueue},
			{Name: "cdn", Type: ResourceCDN},
			{Name: "func", Type: ResourceServerless},
			{Name: "alerts", Type: ResourceMonitoring},
			{Name: "vault", Type: ResourceSecrets},
			{Name: "zone", Type: ResourceDNS},
			{Name: "vnet", Type: ResourceNetworking},
		},
	}
	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedBlocks := []string{
		`resource "azurerm_storage_account" "uploads"`,
		`resource "azurerm_container_group" "api"`,
		`resource "azurerm_postgresql_flexible_server" "db"`,
		`resource "azurerm_redis_cache" "cache"`,
		`resource "azurerm_servicebus_queue" "queue"`,
		`resource "azurerm_cdn_frontdoor_profile" "cdn"`,
		`resource "azurerm_linux_function_app" "func"`,
		`resource "azurerm_monitor_action_group" "alerts"`,
		`resource "azurerm_key_vault" "vault"`,
		`resource "azurerm_dns_zone" "zone"`,
		`resource "azurerm_virtual_network" "vnet"`,
	}
	for _, block := range expectedBlocks {
		if !strings.Contains(tf, block) {
			t.Errorf("missing resource block: %s", block)
		}
	}
}
