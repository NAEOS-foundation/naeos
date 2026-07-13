package cloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/version"
)

type GCPAdapter struct{}

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

func (a *GCPAdapter) Plan(config *DeployConfig) ([]Resource, error) {
	resources := []Resource{}
	for _, res := range config.Resources {
		switch res.Type {
		case ResourceStorage:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "google_storage_bucket",
				Spec: map[string]interface{}{
					"name":     fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"location": config.Region,
				},
			})
		case ResourceCompute:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "google_cloud_run_service",
				Spec: map[string]interface{}{
					"name":     res.Name,
					"location": config.Region,
				},
			})
		case ResourceDatabase:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "google_sql_database_instance",
				Spec: map[string]interface{}{
					"name":       fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"region":     config.Region,
					"db_version": "POSTGRES_15",
				},
			})
		case ResourceCache:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "google_redis_instance",
				Spec: map[string]interface{}{
					"name":   fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"region": config.Region,
				},
			})
		case ResourceQueue:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "google_pubsub_topic",
				Spec: map[string]interface{}{
					"name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
				},
			})
		case ResourceCDN:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "google_compute_backend_bucket",
				Spec: map[string]interface{}{
					"name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
				},
			})
		}
	}
	return resources, nil
}

func (a *GCPAdapter) Deploy(config *DeployConfig) (*DeployResult, error) {
	plan, err := a.Plan(config)
	if err != nil {
		return nil, err
	}

	deployed := []DeployedResource{}
	for _, res := range plan {
		deployed = append(deployed, DeployedResource{
			Name: res.Name,
			Type: res.Type,
			ID:   fmt.Sprintf("projects/%s/%s/%s", config.Project, res.Type, res.Name),
		})
	}

	tf, _ := a.ExportTerraform(config)

	return &DeployResult{
		Provider:  GCP,
		Resources: deployed,
		Terraform: tf,
		Status:    "deployed",
		Timestamp: time.Now(),
	}, nil
}

func (a *GCPAdapter) Destroy(config *DeployConfig) error {
	plan, err := a.Plan(config)
	if err != nil {
		return err
	}
	if len(plan) == 0 {
		return fmt.Errorf("no resources to destroy")
	}
	var resources []string
	for _, r := range plan {
		resources = append(resources, r.Type+"."+r.Name)
	}
	return fmt.Errorf("terraform destroy planned: %d GCP resources would be destroyed: %s", len(resources), strings.Join(resources, ", "))
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
			sb.WriteString(fmt.Sprintf(`resource "google_sql_database_instance" "%s" {
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
  password = ""
}

`, res.Name, instanceName, config.Region,
			res.Name, dbName, res.Name,
			res.Name, res.Name))

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
		}
	}

	return sb.String(), nil
}
