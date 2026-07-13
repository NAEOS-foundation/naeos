package cloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/version"
)

type AzureAdapter struct{}

func (a *AzureAdapter) Name() string {
	return "Azure"
}

func (a *AzureAdapter) Provider() CloudProvider {
	return Azure
}

func (a *AzureAdapter) Validate(config *DeployConfig) error {
	if config.Project == "" {
		return fmt.Errorf("Azure resource group is required")
	}
	if config.Region == "" {
		return fmt.Errorf("Azure region is required")
	}
	return nil
}

func (a *AzureAdapter) Plan(config *DeployConfig) ([]Resource, error) {
	resources := []Resource{}
	for _, res := range config.Resources {
		switch res.Type {
		case ResourceStorage:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "azurerm_storage_account",
				Spec: map[string]interface{}{
					"name":     fmt.Sprintf("%s%s%s", config.Project, config.Environment, res.Name),
					"location": config.Region,
				},
			})
		case ResourceCompute:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "azurerm_container_group",
				Spec: map[string]interface{}{
					"name":     res.Name,
					"location": config.Region,
				},
			})
		case ResourceDatabase:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "azurerm_postgresql_flexible_server",
				Spec: map[string]interface{}{
					"name":     fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"location": config.Region,
				},
			})
		case ResourceCache:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "azurerm_redis_cache",
				Spec: map[string]interface{}{
					"name":     fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"location": config.Region,
				},
			})
		case ResourceQueue:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "azurerm_servicebus_queue",
				Spec: map[string]interface{}{
					"name":     res.Name,
					"location": config.Region,
				},
			})
		case ResourceCDN:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "azurerm_cdn_frontdoor_profile",
				Spec: map[string]interface{}{
					"name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
				},
			})
		}
	}
	return resources, nil
}

func (a *AzureAdapter) Deploy(config *DeployConfig) (*DeployResult, error) {
	plan, err := a.Plan(config)
	if err != nil {
		return nil, err
	}

	deployed := []DeployedResource{}
	for _, res := range plan {
		deployed = append(deployed, DeployedResource{
			Name: res.Name,
			Type: res.Type,
			ID:   fmt.Sprintf("/subscriptions/.../resourceGroups/%s/providers/%s/%s", config.Project, res.Type, res.Name),
		})
	}

	tf, _ := a.ExportTerraform(config)

	return &DeployResult{
		Provider:  Azure,
		Resources: deployed,
		Terraform: tf,
		Status:    "deployed",
		Timestamp: time.Now(),
	}, nil
}

func (a *AzureAdapter) Destroy(config *DeployConfig) error {
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
	return fmt.Errorf("terraform destroy planned: %d Azure resources would be destroyed: %s", len(resources), strings.Join(resources, ", "))
}

func (a *AzureAdapter) ExportTerraform(config *DeployConfig) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf(`terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "main" {
  name     = "%s"
  location = "%s"

  tags = {
    environment = "%s"
    project     = "%s"
    managed_by  = "%s"
  }
}

`, config.Project, config.Region, config.Environment, config.Project, version.ProductName))

	for _, res := range config.Resources {
		switch res.Type {
		case ResourceStorage:
			storageName := fmt.Sprintf("%s%s%s", config.Project, config.Environment, res.Name)
			// Azure storage names must be lowercase alphanumeric only
			storageName = strings.ToLower(strings.ReplaceAll(storageName, "-", ""))
			if len(storageName) > 24 {
				storageName = storageName[:24]
			}
			sb.WriteString(fmt.Sprintf(`resource "azurerm_storage_account" "%s" {
  name                     = "%s"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

resource "azurerm_storage_container" "%s" {
  name                  = "%s"
  storage_account_name  = azurerm_storage_account.%s.name
  container_access_type = "private"
}

`, res.Name, storageName,
			config.Environment, config.Project,
			res.Name, res.Name, res.Name))

		case ResourceCompute:
			sb.WriteString(fmt.Sprintf(`resource "azurerm_container_group" "%s" {
  name                = "%s"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  ip_address_type     = "Public"
  os_type             = "Linux"

  container {
    name   = "%s"
    image  = "%s/%s:latest"
    cpu    = "1.0"
    memory = "1.5"

    ports {
      port     = 8080
      protocol = "TCP"
    }

    environment_variables = {
      ENV = "%s"
    }
  }

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

`, res.Name, res.Name,
			res.Name, config.Project, res.Name,
			config.Environment,
			config.Environment, config.Project))

		case ResourceDatabase:
			serverName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			// Azure server names: lowercase alphanumeric and hyphens
			serverName = strings.ToLower(serverName)
			dbName := strings.ReplaceAll(res.Name, "-", "_")
			sb.WriteString(fmt.Sprintf(`resource "azurerm_postgresql_flexible_server" "%s" {
  name                = "%s"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku_name            = "B_Standard_B1ms"
  version             = "15"
  storage_mb          = 32768

  backup_retention_days        = 7
  geo_redundant_backup_enabled = false

  admin_login    = "psqladmin"
  admin_password = ""

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

resource "azurerm_postgresql_flexible_server_database" "%s" {
  name      = "%s"
  server_id = azurerm_postgresql_flexible_server.%s.id
  collation = "en_US.utf8"
  charset   = "utf8"
}

resource "azurerm_postgresql_flexible_server_firewall_rule" "%s" {
  name             = "allow-all"
  server_id        = azurerm_postgresql_flexible_server.%s.id
  start_ip_address = "0.0.0.0"
  end_ip_address   = "255.255.255.255"
}

`, res.Name, serverName,
			config.Environment, config.Project,
			res.Name, dbName, res.Name,
			res.Name, res.Name))

		case ResourceCache:
			redisName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			redisName = strings.ReplaceAll(redisName, "-", "")
			if len(redisName) > 64 {
				redisName = redisName[:64]
			}
			sb.WriteString(fmt.Sprintf(`resource "azurerm_redis_cache" "%s" {
  name                = "%s"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  capacity            = 0
  family              = "C"
  sku_name            = "Basic"
  minimum_tls_version = "1.2"

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

`, res.Name, redisName,
			config.Environment, config.Project))

		case ResourceQueue:
			sb.WriteString(fmt.Sprintf(`resource "azurerm_servicebus_namespace" "main" {
  name                = "%s-sb"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "Basic"

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

resource "azurerm_servicebus_queue" "%s" {
  name         = "%s"
  namespace_id = azurerm_servicebus_namespace.main.id

  enable_partitioning = false
  max_size_in_megabytes = 1024

  default_message_ttl = "P14D"
  dead_lettering_on_message_expiration = true
}

`, config.Project,
			config.Environment, config.Project,
			res.Name, res.Name))

		case ResourceCDN:
			sb.WriteString(fmt.Sprintf(`resource "azurerm_cdn_frontdoor_profile" "%s" {
  name                = "%s"
  resource_group_name = azurerm_resource_group.main.name
  sku_name            = "Standard_AzureFrontDoor"

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

resource "azurerm_cdn_frontdoor_endpoint" "%s" {
  name                     = "%s-endpoint"
  cdn_frontdoor_profile_id = azurerm_cdn_frontdoor_profile.%s.id
}

resource "azurerm_cdn_frontdoor_origin_group" "%s" {
  name                     = "%s-origin-group"
  cdn_frontdoor_profile_id = azurerm_cdn_frontdoor_profile.%s.id

  load_balancing {
    sample_size                        = 4
    successful_samples_required        = 3
    additional_latency_in_milliseconds = 50
  }
}

resource "azurerm_cdn_frontdoor_origin" "%s" {
  name                          = "%s-origin"
  origin_group_id               = azurerm_cdn_frontdoor_origin_group.%s.id
  enabled                       = true
  host_name                     = "example.com"
  http_port                     = 80
  https_port                    = 443
  origin_host_header            = "example.com"
  priority                      = 1
  weight                        = 1000
}

resource "azurerm_cdn_frontdoor_route" "%s" {
  name                          = "%s-route"
  cdn_frontdoor_endpoint_id     = azurerm_cdn_frontdoor_endpoint.%s.id
  origin_group_id               = azurerm_cdn_frontdoor_origin_group.%s.id
  origin_path                   = "/"
  patterns_to_match             = ["/*"]
  supported_protocols           = ["Http", "Https"]
  https_redirect_enabled        = true
  forward_to_origin_group       = true
}

`, res.Name, res.Name,
			config.Environment, config.Project,
			res.Name, res.Name, res.Name,
			res.Name, res.Name, res.Name,
			res.Name, res.Name, res.Name,
			res.Name, res.Name, res.Name, res.Name))
		}
	}

	return sb.String(), nil
}
