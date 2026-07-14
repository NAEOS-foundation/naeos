package cloud

import (
	"strings"
	"testing"
)

func TestGetAdapterAWS(t *testing.T) {
	t.Parallel()
	adapter, err := GetAdapter(AWS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if adapter.Name() != "AWS" {
		t.Errorf("expected name 'AWS', got %s", adapter.Name())
	}
	if adapter.Provider() != AWS {
		t.Errorf("expected provider AWS, got %s", adapter.Provider())
	}
}

func TestGetAdapterGCP(t *testing.T) {
	t.Parallel()
	adapter, err := GetAdapter(GCP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if adapter.Name() != "GCP" {
		t.Errorf("expected name 'GCP', got %s", adapter.Name())
	}
	if adapter.Provider() != GCP {
		t.Errorf("expected provider GCP, got %s", adapter.Provider())
	}
}

func TestGetAdapterAzure(t *testing.T) {
	t.Parallel()
	adapter, err := GetAdapter(Azure)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if adapter.Name() != "Azure" {
		t.Errorf("expected name 'Azure', got %s", adapter.Name())
	}
	if adapter.Provider() != Azure {
		t.Errorf("expected provider Azure, got %s", adapter.Provider())
	}
}

func TestGetAdapterInvalid(t *testing.T) {
	t.Parallel()
	_, err := GetAdapter("invalid")
	if err == nil {
		t.Error("expected error for invalid provider")
	}
}

func TestSupportedResourceTypes(t *testing.T) {
	t.Parallel()
	expected := []string{"storage", "compute", "database", "cache", "queue", "cdn", "serverless", "monitoring", "secrets", "dns", "networking"}
	if len(SupportedResourceTypes) != len(expected) {
		t.Fatalf("expected %d resource types, got %d", len(expected), len(SupportedResourceTypes))
	}
	for i, v := range expected {
		if SupportedResourceTypes[i] != v {
			t.Errorf("expected %s at index %d, got %s", v, i, SupportedResourceTypes[i])
		}
	}
}

// --- AWS Tests ---

func TestAWSValidate(t *testing.T) {
	t.Parallel()
	adapter := &AWSAdapter{}

	validConfig := &DeployConfig{
		Provider: AWS,
		Region:   "us-east-1",
		Project:  "test",
	}
	if err := adapter.Validate(validConfig); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	invalidConfig := &DeployConfig{
		Provider: AWS,
		Region:   "invalid-region",
		Project:  "test",
	}
	if err := adapter.Validate(invalidConfig); err == nil {
		t.Error("expected error for invalid region")
	}

	emptyRegion := &DeployConfig{
		Provider: AWS,
		Project:  "test",
	}
	if err := adapter.Validate(emptyRegion); err == nil {
		t.Error("expected error for empty region")
	}
}

func TestAWSPlanAllResourceTypes(t *testing.T) {
	t.Parallel()
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider:    AWS,
		Region:      "us-east-1",
		Project:     "myapp",
		Environment: "prod",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
			{Name: "db", Type: ResourceDatabase},
			{Name: "cache", Type: ResourceCache},
			{Name: "queue", Type: ResourceQueue},
			{Name: "cdn", Type: ResourceCDN},
		},
	}

	planResult, err := adapter.Plan(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(planResult.Resources) != 6 {
		t.Errorf("expected 6 resources, got %d", len(planResult.Resources))
	}

	expectedTypes := []string{
		"aws_s3_bucket", "aws_ecs_service", "aws_rds_instance",
		"aws_elasticache_cluster", "aws_sqs_queue", "aws_cloudfront_distribution",
	}
	for i, expected := range expectedTypes {
		if planResult.Resources[i].Type != expected {
			t.Errorf("resource %d: expected %s, got %s", i, expected, planResult.Resources[i].Type)
		}
	}

	if planResult.CostEstimate.TotalMonthlyUSD <= 0 {
		t.Errorf("expected positive cost estimate, got %f", planResult.CostEstimate.TotalMonthlyUSD)
	}
}

func TestAWSTerraformExportAllTypes(t *testing.T) {
	t.Parallel()
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider:    AWS,
		Region:      "us-east-1",
		Project:     "myapp",
		Environment: "prod",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
			{Name: "db", Type: ResourceDatabase},
			{Name: "cache", Type: ResourceCache},
			{Name: "queue", Type: ResourceQueue},
			{Name: "cdn", Type: ResourceCDN},
		},
	}

	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tf == "" {
		t.Fatal("expected non-empty terraform output")
	}

	// Verify provider block
	if !strings.Contains(tf, `source  = "hashicorp/aws"`) {
		t.Error("missing AWS provider source")
	}
	if !strings.Contains(tf, `region = "us-east-1"`) {
		t.Error("missing region")
	}

	// Verify all resource types
	expectedResources := []string{
		`resource "aws_s3_bucket" "uploads"`,
		`resource "aws_ecs_cluster"`,
		`resource "aws_rds_instance" "db"`,
		`resource "aws_elasticache_cluster" "cache"`,
		`resource "aws_sqs_queue" "queue"`,
		`resource "aws_cloudfront_distribution" "cdn"`,
	}
	for _, expected := range expectedResources {
		if !strings.Contains(tf, expected) {
			t.Errorf("missing resource block: %s", expected)
		}
	}

	// Verify local variables
	if !strings.Contains(tf, `ManagedBy   = "naeos"`) {
		t.Error("missing ManagedBy tag")
	}
}

func TestAWSDeploy(t *testing.T) {
	t.Parallel()
	adapter := &AWSAdapter{Runner: &mockRunner{stdout: []byte("ok")}}
	config := &DeployConfig{
		Provider:    AWS,
		Region:      "us-east-1",
		Project:     "myapp",
		Environment: "prod",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
		},
	}

	result, err := adapter.Deploy(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Provider != AWS {
		t.Errorf("expected provider AWS, got %s", result.Provider)
	}
	if result.Status != "deployed" {
		t.Errorf("expected status deployed, got %s", result.Status)
	}
	if len(result.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(result.Resources))
	}
	if result.Terraform == "" {
		t.Error("expected non-empty terraform in result")
	}
}

// --- GCP Tests ---

func TestGCPValidate(t *testing.T) {
	t.Parallel()
	adapter := &GCPAdapter{}

	validConfig := &DeployConfig{
		Provider: GCP,
		Project:  "my-project",
		Region:   "us-central1",
	}
	if err := adapter.Validate(validConfig); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	invalidConfig := &DeployConfig{
		Provider: GCP,
		Region:   "us-central1",
	}
	if err := adapter.Validate(invalidConfig); err == nil {
		t.Error("expected error for missing project")
	}

	emptyRegion := &DeployConfig{
		Provider: GCP,
		Project:  "my-project",
	}
	if err := adapter.Validate(emptyRegion); err == nil {
		t.Error("expected error for empty region")
	}
}

func TestGCPPlanAllResourceTypes(t *testing.T) {
	t.Parallel()
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider:    GCP,
		Region:      "us-central1",
		Project:     "myapp",
		Environment: "dev",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
			{Name: "db", Type: ResourceDatabase},
			{Name: "cache", Type: ResourceCache},
			{Name: "queue", Type: ResourceQueue},
			{Name: "cdn", Type: ResourceCDN},
		},
	}

	planResult, err := adapter.Plan(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(planResult.Resources) != 7 {
		t.Errorf("expected 7 resources, got %d", len(planResult.Resources))
	}

	expectedTypes := []string{
		"google_storage_bucket", "google_cloud_run_service", "google_sql_database_instance",
		"google_redis_instance", "google_pubsub_topic", "google_compute_backend_bucket",
		"google_storage_bucket",
	}
	for i, expected := range expectedTypes {
		if planResult.Resources[i].Type != expected {
			t.Errorf("resource %d: expected %s, got %s", i, expected, planResult.Resources[i].Type)
		}
	}
}

func TestGCPTerraformExportAllTypes(t *testing.T) {
	t.Parallel()
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider:    GCP,
		Region:      "us-central1",
		Project:     "myapp",
		Environment: "dev",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
			{Name: "db", Type: ResourceDatabase},
			{Name: "cache", Type: ResourceCache},
			{Name: "queue", Type: ResourceQueue},
			{Name: "cdn", Type: ResourceCDN},
		},
	}

	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tf == "" {
		t.Fatal("expected non-empty terraform output")
	}

	// Verify provider block
	if !strings.Contains(tf, `source  = "hashicorp/google"`) {
		t.Error("missing Google provider source")
	}
	if !strings.Contains(tf, `project = "myapp"`) {
		t.Error("missing project")
	}

	// Verify all resource types
	expectedResources := []string{
		`resource "google_storage_bucket" "uploads"`,
		`resource "google_cloud_run_service" "api"`,
		`resource "google_sql_database_instance" "db"`,
		`resource "google_redis_instance" "cache"`,
		`resource "google_pubsub_topic" "queue"`,
		`resource "google_compute_backend_bucket" "cdn"`,
	}
	for _, expected := range expectedResources {
		if !strings.Contains(tf, expected) {
			t.Errorf("missing resource block: %s", expected)
		}
	}
}

func TestGCPDeploy(t *testing.T) {
	t.Parallel()
	adapter := &GCPAdapter{Runner: &mockRunner{stdout: []byte("ok")}}
	config := &DeployConfig{
		Provider:    GCP,
		Region:      "us-central1",
		Project:     "myapp",
		Environment: "dev",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
		},
	}

	result, err := adapter.Deploy(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Provider != GCP {
		t.Errorf("expected provider GCP, got %s", result.Provider)
	}
	if len(result.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(result.Resources))
	}
	if !strings.Contains(result.Resources[0].ID, "projects/myapp") {
		t.Errorf("unexpected resource ID: %s", result.Resources[0].ID)
	}
}

// --- Azure Tests ---

func TestAzureValidate(t *testing.T) {
	t.Parallel()
	adapter := &AzureAdapter{}

	validConfig := &DeployConfig{
		Provider: Azure,
		Project:  "my-rg",
		Region:   "eastus",
	}
	if err := adapter.Validate(validConfig); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	invalidConfig := &DeployConfig{
		Provider: Azure,
		Region:   "eastus",
	}
	if err := adapter.Validate(invalidConfig); err == nil {
		t.Error("expected error for missing project")
	}

	emptyRegion := &DeployConfig{
		Provider: Azure,
		Project:  "my-rg",
	}
	if err := adapter.Validate(emptyRegion); err == nil {
		t.Error("expected error for empty region")
	}
}

func TestAzurePlanAllResourceTypes(t *testing.T) {
	t.Parallel()
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider:    Azure,
		Region:      "eastus",
		Project:     "myapp",
		Environment: "staging",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
			{Name: "db", Type: ResourceDatabase},
			{Name: "cache", Type: ResourceCache},
			{Name: "queue", Type: ResourceQueue},
			{Name: "cdn", Type: ResourceCDN},
		},
	}

	planResult, err := adapter.Plan(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(planResult.Resources) != 6 {
		t.Errorf("expected 6 resources, got %d", len(planResult.Resources))
	}

	expectedTypes := []string{
		"azurerm_storage_account", "azurerm_container_group", "azurerm_postgresql_flexible_server",
		"azurerm_redis_cache", "azurerm_servicebus_queue", "azurerm_cdn_frontdoor_profile",
	}
	for i, expected := range expectedTypes {
		if planResult.Resources[i].Type != expected {
			t.Errorf("resource %d: expected %s, got %s", i, expected, planResult.Resources[i].Type)
		}
	}
}

func TestAzureTerraformExportAllTypes(t *testing.T) {
	t.Parallel()
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider:    Azure,
		Region:      "eastus",
		Project:     "myapp",
		Environment: "staging",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
			{Name: "db", Type: ResourceDatabase},
			{Name: "cache", Type: ResourceCache},
			{Name: "queue", Type: ResourceQueue},
			{Name: "cdn", Type: ResourceCDN},
		},
	}

	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tf == "" {
		t.Fatal("expected non-empty terraform output")
	}

	// Verify provider block
	if !strings.Contains(tf, `source  = "hashicorp/azurerm"`) {
		t.Error("missing Azure provider source")
	}

	// Verify resource group
	if !strings.Contains(tf, `resource "azurerm_resource_group" "main"`) {
		t.Error("missing resource group")
	}

	// Verify all resource types
	expectedResources := []string{
		`resource "azurerm_storage_account" "uploads"`,
		`resource "azurerm_container_group" "api"`,
		`resource "azurerm_postgresql_flexible_server" "db"`,
		`resource "azurerm_redis_cache" "cache"`,
		`resource "azurerm_servicebus_queue" "queue"`,
		`resource "azurerm_cdn_frontdoor_profile" "cdn"`,
	}
	for _, expected := range expectedResources {
		if !strings.Contains(tf, expected) {
			t.Errorf("missing resource block: %s", expected)
		}
	}
}

func TestAzureDeploy(t *testing.T) {
	t.Parallel()
	adapter := &AzureAdapter{Runner: &mockRunner{stdout: []byte("ok")}}
	config := &DeployConfig{
		Provider:    Azure,
		Region:      "eastus",
		Project:     "myapp",
		Environment: "staging",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
		},
	}

	result, err := adapter.Deploy(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Provider != Azure {
		t.Errorf("expected provider Azure, got %s", result.Provider)
	}
	if len(result.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(result.Resources))
	}
	if !strings.Contains(result.Resources[0].ID, "resourceGroups/myapp") {
		t.Errorf("unexpected resource ID: %s", result.Resources[0].ID)
	}
}

// --- Edge Cases ---

func TestAWSExportEmptyResources(t *testing.T) {
	t.Parallel()
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider:    AWS,
		Region:      "us-east-1",
		Project:     "myapp",
		Environment: "dev",
		Resources:   []Resource{},
	}

	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(tf, `region = "us-east-1"`) {
		t.Error("should still contain provider block")
	}
}

func TestGCPExportEmptyResources(t *testing.T) {
	t.Parallel()
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider:    GCP,
		Region:      "us-central1",
		Project:     "myapp",
		Environment: "dev",
		Resources:   []Resource{},
	}

	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(tf, `project = "myapp"`) {
		t.Error("should still contain provider block")
	}
}

func TestAzureExportEmptyResources(t *testing.T) {
	t.Parallel()
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider:    Azure,
		Region:      "eastus",
		Project:     "myapp",
		Environment: "dev",
		Resources:   []Resource{},
	}

	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(tf, `resource "azurerm_resource_group" "main"`) {
		t.Error("should still contain resource group")
	}
}

func TestAWSUnknownResourceTypeSkipped(t *testing.T) {
	t.Parallel()
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider:    AWS,
		Region:      "us-east-1",
		Project:     "myapp",
		Environment: "dev",
		Resources: []Resource{
			{Name: "unknown", Type: "unknown_type"},
		},
	}

	planResult, err := adapter.Plan(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(planResult.Resources) != 0 {
		t.Errorf("expected 0 resources for unknown type, got %d", len(planResult.Resources))
	}
}

// --- HCL Content Verification Tests ---

func TestAWSHCLContent(t *testing.T) {
	t.Parallel()
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider:    AWS,
		Region:      "us-east-1",
		Project:     "myapp",
		Environment: "prod",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
			{Name: "db", Type: ResourceDatabase},
			{Name: "cache", Type: ResourceCache},
			{Name: "queue", Type: ResourceQueue},
			{Name: "cdn", Type: ResourceCDN},
		},
	}

	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name     string
		resource string
	}{
		{"S3 bucket", "aws_s3_bucket"},
		{"ECS service", "aws_ecs_service"},
		{"RDS instance", "aws_rds_instance"},
		{"ElastiCache cluster", "aws_elasticache_cluster"},
		{"SQS queue", "aws_sqs_queue"},
		{"CloudFront distribution", "aws_cloudfront_distribution"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(tf, tt.resource) {
				t.Errorf("HCL output missing resource type %q", tt.resource)
			}
		})
	}
}

func TestGCPHCLContent(t *testing.T) {
	t.Parallel()
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider:    GCP,
		Region:      "us-central1",
		Project:     "myapp",
		Environment: "dev",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
			{Name: "db", Type: ResourceDatabase},
			{Name: "cache", Type: ResourceCache},
			{Name: "queue", Type: ResourceQueue},
			{Name: "cdn", Type: ResourceCDN},
		},
	}

	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name     string
		resource string
	}{
		{"GCS bucket", "google_storage_bucket"},
		{"Cloud Run service", "google_cloud_run_service"},
		{"Cloud SQL instance", "google_sql_database_instance"},
		{"Memorystore instance", "google_redis_instance"},
		{"Pub/Sub topic", "google_pubsub_topic"},
		{"CDN backend bucket", "google_compute_backend_bucket"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(tf, tt.resource) {
				t.Errorf("HCL output missing resource type %q", tt.resource)
			}
		})
	}
}

func TestAzureHCLContent(t *testing.T) {
	t.Parallel()
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider:    Azure,
		Region:      "eastus",
		Project:     "myapp",
		Environment: "staging",
		Resources: []Resource{
			{Name: "uploads", Type: ResourceStorage},
			{Name: "api", Type: ResourceCompute},
			{Name: "db", Type: ResourceDatabase},
			{Name: "cache", Type: ResourceCache},
			{Name: "queue", Type: ResourceQueue},
			{Name: "cdn", Type: ResourceCDN},
		},
	}

	tf, err := adapter.ExportTerraform(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name     string
		resource string
	}{
		{"Storage account", "azurerm_storage_account"},
		{"Container group", "azurerm_container_group"},
		{"PostgreSQL server", "azurerm_postgresql_flexible_server"},
		{"Redis cache", "azurerm_redis_cache"},
		{"Service Bus namespace", "azurerm_servicebus_namespace"},
		{"Front Door profile", "azurerm_cdn_frontdoor_profile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(tf, tt.resource) {
				t.Errorf("HCL output missing resource type %q", tt.resource)
			}
		})
	}
}
