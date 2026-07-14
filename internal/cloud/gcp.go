package cloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/version"
)

// GCPAdapter implements CloudAdapter for Google Cloud Platform.
type GCPAdapter struct {
	Runner CommandRunner
}

func (a *GCPAdapter) Name() string {
	return "GCP"
}

func (a *GCPAdapter) Provider() CloudProvider {
	return GCP
}

func (a *GCPAdapter) Validate(config *DeployConfig) error {
	if config.Project == "" {
		return fmt.Errorf("GCP project is required")
	}
	if config.Region == "" {
		return fmt.Errorf("GCP region is required")
	}
	return nil
}

func (a *GCPAdapter) Plan(config *DeployConfig) (*PlanResult, error) {
	resources := []Resource{}
	for _, res := range config.Resources {
		switch res.Type {
		case ResourceStorage:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "google_storage_bucket",
				Spec: map[string]any{
					"name":     fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"location": config.Region,
				},
			})
		case ResourceCompute:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "google_cloud_run_service",
				Spec: map[string]any{
					"name":     res.Name,
					"location": config.Region,
				},
			})
		case ResourceDatabase:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "google_sql_database_instance",
				Spec: map[string]any{
					"name":       fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"region":     config.Region,
					"db_version": "POSTGRES_15",
				},
			})
		case ResourceCache:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "google_redis_instance",
				Spec: map[string]any{
					"name":   fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"region": config.Region,
				},
			})
		case ResourceQueue:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "google_pubsub_topic",
				Spec: map[string]any{
					"name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
				},
			})
	case ResourceCDN:
		bucketName := fmt.Sprintf("%s-%s-%s-cdn", config.Project, config.Environment, res.Name)
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "google_compute_backend_bucket",
			Spec: map[string]any{
				"name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
			},
		})
		resources = append(resources, Resource{
			Name: res.Name + "-cdn-bucket",
			Type: "google_storage_bucket",
			Spec: map[string]any{
				"name":     bucketName,
				"location": config.Region,
			},
		})
	case ResourceServerless:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "google_cloudfunctions2_function",
			Spec: map[string]any{
				"name":     fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
				"location": config.Region,
			},
		})
	case ResourceMonitoring:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "google_monitoring_alert_policy",
			Spec: map[string]any{
				"display_name": fmt.Sprintf("%s %s %s", config.Project, config.Environment, res.Name),
			},
		})
	case ResourceSecrets:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "google_secret_manager_secret",
			Spec: map[string]any{
				"secret_id": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
			},
		})
	case ResourceDNS:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "google_dns_managed_zone",
			Spec: map[string]any{
				"name":        res.Name,
				"dns_name":    fmt.Sprintf("%s.", res.Name),
			},
		})
	case ResourceNetworking:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "google_compute_network",
			Spec: map[string]any{
				"name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
			},
		})
	}
}

	estimator := NewCostEstimator()
	cost := estimator.EstimateCost(string(config.Provider), resources)

	return &PlanResult{
		Resources:    resources,
		CostEstimate: cost,
	}, nil
}

func (a *GCPAdapter) Deploy(config *DeployConfig) (*DeployResult, error) {
	if err := a.Validate(config); err != nil {
		return nil, err
	}

	planResult, err := a.Plan(config)
	if err != nil {
		return nil, err
	}

	tf, err := a.ExportTerraform(config)
	if err != nil {
		return nil, err
	}

	pool := GetDefaultPool()
	tr, pooled := pool.Get(config.Project, GCP)
	if pooled {
		if err := tr.writeHCL(tf); err != nil {
			return nil, err
		}
		if err := tr.Apply(); err != nil {
			return nil, err
		}
	} else {
		workDir, err := TempWorkDir("naeos-gcp")
		if err != nil {
			return nil, err
		}
		tr = NewTerraformRunner(workDir)
		if a.Runner != nil {
			tr.Runner = a.Runner
		}
		if err := tr.Deploy(tf); err != nil {
			return nil, err
		}
		pool.Put(config.Project, GCP, tr, true)
	}

	deployed := []DeployedResource{}
	for _, res := range planResult.Resources {
		deployed = append(deployed, DeployedResource{
			Name: res.Name,
			Type: res.Type,
			ID:   fmt.Sprintf("projects/%s/%s/%s", config.Project, res.Type, res.Name),
		})
	}

	result := &DeployResult{
		Provider:  GCP,
		Resources: deployed,
		Terraform: tf,
		Status:    "deployed",
		Timestamp: time.Now(),
	}

	sm := NewStateManager()
	sm.Save(&DeploymentRecord{
		Project:      config.Project,
		Provider:     GCP,
		Environment:  config.Environment,
		Region:       config.Region,
		Resources:    deployed,
		TerraformDir: tr.WorkDir,
		Timestamp:    result.Timestamp,
		Status:       "deployed",
	})

	return result, nil
}

func (a *GCPAdapter) Destroy(config *DeployConfig) error {
	pool := GetDefaultPool()
	if tr, pooled := pool.Get(config.Project, GCP); pooled {
		if err := tr.ApplyDestroy(); err == nil {
			pool.Remove(config.Project, GCP)
			sm := NewStateManager()
			sm.Delete(config.Project, GCP)
			return nil
		}
	}

	sm := NewStateManager()
	record, err := sm.Load(config.Project, GCP)
	if err == nil && record.TerraformDir != "" {
		tr := NewTerraformRunner(record.TerraformDir)
		if a.Runner != nil {
			tr.Runner = a.Runner
		}
		if derr := tr.DestroyAll(); derr == nil {
			sm.Delete(config.Project, GCP)
			return nil
		}
	}

	planResult, err := a.Plan(config)
	if err != nil {
		return err
	}
	if len(planResult.Resources) == 0 {
		return fmt.Errorf("no resources to destroy")
	}

	tf, err := a.ExportTerraform(config)
	if err != nil {
		return err
	}

	workDir, werr := TempWorkDir("naeos-gcp-destroy")
	if werr != nil {
		return werr
	}

	tr := NewTerraformRunner(workDir)
	if a.Runner != nil {
		tr.Runner = a.Runner
	}
	if err := tr.writeHCL(tf); err != nil {
		return err
	}
	if err := tr.DestroyAll(); err != nil {
		return err
	}

	sm.Delete(config.Project, GCP)
	return nil
}

func (a *GCPAdapter) ExportTerraform(config *DeployConfig) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf(`terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = "%s"
  region  = "%s"
}

`, config.Project, config.Region))

	// Local variables
	sb.WriteString(fmt.Sprintf(`locals {
  project     = "%s"
  environment = "%s"
  common_labels = {
    environment = "%s"
    project     = "%s"
    managed_by  = "%s"
  }
}

`, config.Project, config.Environment, config.Environment, config.Project, version.ProductName))

	for _, res := range config.Resources {
		switch res.Type {
		case ResourceStorage:
			bucketName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "google_storage_bucket" "%s" {
  name     = "%s"
  location = "%s"

  uniform_bucket_level_access = true
  versioning {
    enabled = true
  }

  labels = local.common_labels
}

resource "google_storage_bucket_iam_member" "%s_public" {
  bucket = google_storage_bucket.%s.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

`, res.Name, bucketName, config.Region,
				res.Name, res.Name))

		case ResourceCompute:
			sb.WriteString(fmt.Sprintf(`resource "google_cloud_run_service" "%s" {
  name     = "%s"
  location = "%s"

  template {
    metadata {
      labels = local.common_labels
    }

    spec {
      containers {
        image = "gcr.io/%s/%s:latest"
        ports {
          container_port = 8080
        }
        resources {
          limits = {
            cpu    = "1000m"
            memory = "512Mi"
          }
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  lifecycle {
    ignore_changes = [
      template[0].metadata[0].annotations,
    ]
  }
}

resource "google_cloud_run_service_iam_member" "%s_invoker" {
  service = google_cloud_run_service.%s.name
  role    = "roles/run.invoker"
  member  = "allUsers"
}

`, res.Name, res.Name, config.Region,
			config.Project, res.Name,
			res.Name, res.Name))

		case ResourceDatabase:
			instanceName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			dbName := strings.ReplaceAll(res.Name, "-", "_")
			sb.WriteString(fmt.Sprintf(`resource "random_password" "%s_db" {
  length           = 32
  special          = true
  override_special = "!#$%%^&*()-_=+[]{}<>:?"
}

resource "google_sql_database_instance" "%s" {
  name             = "%s"
  database_version = "POSTGRES_15"
  region           = "%s"

  settings {
    tier              = "db-f1-micro"
    availability_type = "ZONAL"

    disk_size    = 10
    disk_type    = "PD_SSD"

    backup_configuration {
      enabled = true
    }

    ip_configuration {
      ipv4_enabled = true
    }

    database_flags {
      name  = "max_connections"
      value = "100"
    }
  }

  deletion_protection = false

  labels = local.common_labels
}

resource "google_sql_database" "%s" {
  name     = "%s"
  instance = google_sql_database_instance.%s.name
}

resource "google_sql_user" "%s" {
  name     = "app"
  instance = google_sql_database_instance.%s.name
  password = random_password.%s_db.result
}

`, res.Name,
			res.Name, instanceName, config.Region,
			res.Name, dbName, res.Name,
			res.Name, res.Name, res.Name))

		case ResourceCache:
			instanceName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "google_redis_instance" "%s" {
  name           = "%s"
  tier           = "BASIC"
  memory_size_gb = 1

  region = "%s"

  redis_version = "REDIS_7_0"
  display_name  = "%s"

  labels = local.common_labels
}

`, res.Name, instanceName, config.Region,
			fmt.Sprintf("%s %s %s", config.Project, config.Environment, res.Name)))

		case ResourceQueue:
			topicName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "google_pubsub_topic" "%s" {
  name = "%s"

  labels = local.common_labels

  message_retention_duration = "86400s"
}

resource "google_pubsub_subscription" "%s_sub" {
  name  = "%s-subscription"
  topic = google_pubsub_topic.%s.name

  ack_deadline_seconds = 20

  expiration_policy {
    ttl = ""
  }

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  labels = local.common_labels
}

`, res.Name, topicName,
			res.Name, res.Name, res.Name))

		case ResourceCDN:
			bucketName := fmt.Sprintf("%s-%s-%s-cdn", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "google_compute_backend_bucket" "%s" {
  name        = "%s"
  bucket_name = google_storage_bucket.%s_cdn.name
  enable_cdn  = true

  cdn_policy {
    cache_mode                   = "CACHE_ALL_STATIC"
    default_ttl                  = 3600
    max_ttl                      = 86400
    client_ttl                   = 3600
    negative_caching             = true
    signed_url_cache_max_age_sec = 7200
  }
}

resource "google_storage_bucket" "%s_cdn" {
  name     = "%s"
  location = "%s"

  uniform_bucket_level_access = true

  labels = local.common_labels
}

resource "google_compute_url_map" "%s" {
  name            = "%s-url-map"
  default_service = google_compute_backend_bucket.%s.self_link
}

resource "google_compute_target_http_proxy" "%s" {
  name    = "%s-http-proxy"
  url_map = google_compute_url_map.%s.self_link
}

resource "google_compute_global_forwarding_rule" "%s" {
  name       = "%s-forwarding"
  target     = google_compute_target_http_proxy.%s.self_link
  port_range = "80"
}

`, res.Name, res.Name, res.Name,
			res.Name, bucketName, config.Region,
			res.Name, res.Name, res.Name,
			res.Name, res.Name, res.Name,
			res.Name, res.Name, res.Name))

		case ResourceServerless:
			funcName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "google_cloudfunctions2_function" "%s" {
  name     = "%s"
  location = "%s"

  build_config {
    runtime     = "python312"
    entry_point = "handler"
    source {
      storage_source {
        bucket = "%s"
        object = "%s.zip"
      }
    }
  }

  service_config {
    max_instance_count = 100
    available_memory   = "256M"
    timeout_seconds    = 60

    environment_variables = {
      ENVIRONMENT = "%s"
    }
  }

  labels = local.common_labels
}

`, res.Name, funcName, config.Region,
			config.Project, res.Name,
			config.Environment))

		case ResourceMonitoring:
			displayName := fmt.Sprintf("%s %s %s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "google_monitoring_alert_policy" "%s" {
  display_name = "%s"
  combiner     = "OR"

  conditions {
    display_name = "%s condition"
    condition_threshold {
      filter     = "resource.type = \"gce_instance\""
      duration   = "300s"
      comparison = "COMPARISON_GT"
      threshold_value = 80

      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  notification_channels = []

  labels = local.common_labels
}

`, res.Name, displayName, displayName))

		case ResourceSecrets:
			secretID := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "google_secret_manager_secret" "%s" {
  secret_id = "%s"

  replication {
    auto {}
  }

  labels = local.common_labels
}

`, res.Name, secretID))

		case ResourceDNS:
			sb.WriteString(fmt.Sprintf(`resource "google_dns_managed_zone" "%s" {
  name        = "%s"
  dns_name    = "%s."
  description = "%s DNS zone"

  labels = local.common_labels
}

`, res.Name, res.Name, res.Name,
			fmt.Sprintf("%s-%s", config.Project, res.Name)))

		case ResourceNetworking:
			networkName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "google_compute_network" "%s" {
  name                    = "%s"
  auto_create_subnetworks = true
  routing_mode            = "REGIONAL"

  labels = local.common_labels
}

`, res.Name, networkName))

		}
	}

	return sb.String(), nil
}
