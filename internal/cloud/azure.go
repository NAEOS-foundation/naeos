package cloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/version"
)

// AzureAdapter implements CloudAdapter for Microsoft Azure.
type AzureAdapter struct {
	Runner CommandRunner
}

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

func (a *AzureAdapter) Plan(config *DeployConfig) (*PlanResult, error) {
	resources := []Resource{}
	for _, res := range config.Resources {
		switch res.Type {
		case ResourceStorage:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "azurerm_storage_account",
				Spec: map[string]any{
					"name":     fmt.Sprintf("%s%s%s", config.Project, config.Environment, res.Name),
					"location": config.Region,
				},
			})
		case ResourceCompute:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "azurerm_container_group",
				Spec: map[string]any{
					"name":     res.Name,
					"location": config.Region,
				},
			})
		case ResourceDatabase:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "azurerm_postgresql_flexible_server",
				Spec: map[string]any{
					"name":     fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"location": config.Region,
				},
			})
		case ResourceCache:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "azurerm_redis_cache",
				Spec: map[string]any{
					"name":     fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"location": config.Region,
				},
			})
		case ResourceQueue:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "azurerm_servicebus_queue",
				Spec: map[string]any{
					"name":     res.Name,
					"location": config.Region,
				},
			})
	case ResourceCDN:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "azurerm_cdn_frontdoor_profile",
			Spec: map[string]any{
				"name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
			},
		})
	case ResourceServerless:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "azurerm_linux_function_app",
			Spec: map[string]any{
				"name":     fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
				"location": config.Region,
			},
		})
	case ResourceMonitoring:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "azurerm_monitor_action_group",
			Spec: map[string]any{
				"name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
			},
		})
	case ResourceSecrets:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "azurerm_key_vault",
			Spec: map[string]any{
				"name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
			},
		})
	case ResourceDNS:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "azurerm_dns_zone",
			Spec: map[string]any{
				"name": res.Name,
			},
		})
	case ResourceNetworking:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "azurerm_virtual_network",
			Spec: map[string]any{
				"name":     fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
				"location": config.Region,
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

func (a *AzureAdapter) Deploy(config *DeployConfig) (*DeployResult, error) {
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
	tr, pooled := pool.Get(config.Project, Azure)
	if pooled {
		if err := tr.writeHCL(tf); err != nil {
			return nil, err
		}
		if err := tr.Apply(); err != nil {
			return nil, err
		}
	} else {
		workDir, err := TempWorkDir("naeos-azure")
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
		pool.Put(config.Project, Azure, tr, true)
	}

	deployed := []DeployedResource{}
	for _, res := range planResult.Resources {
		deployed = append(deployed, DeployedResource{
			Name: res.Name,
			Type: res.Type,
			ID:   fmt.Sprintf("/subscriptions/.../resourceGroups/%s/providers/%s/%s", config.Project, res.Type, res.Name),
		})
	}

	result := &DeployResult{
		Provider:  Azure,
		Resources: deployed,
		Terraform: tf,
		Status:    "deployed",
		Timestamp: time.Now(),
	}

	sm := NewStateManager()
	sm.Save(&DeploymentRecord{
		Project:      config.Project,
		Provider:     Azure,
		Environment:  config.Environment,
		Region:       config.Region,
		Resources:    deployed,
		TerraformDir: tr.WorkDir,
		Timestamp:    result.Timestamp,
		Status:       "deployed",
	})

	return result, nil
}

func (a *AzureAdapter) Destroy(config *DeployConfig) error {
	pool := GetDefaultPool()
	if tr, pooled := pool.Get(config.Project, Azure); pooled {
		if err := tr.ApplyDestroy(); err == nil {
			pool.Remove(config.Project, Azure)
			sm := NewStateManager()
			sm.Delete(config.Project, Azure)
			return nil
		}
	}

	sm := NewStateManager()
	record, err := sm.Load(config.Project, Azure)
	if err == nil && record.TerraformDir != "" {
		tr := NewTerraformRunner(record.TerraformDir)
		if a.Runner != nil {
			tr.Runner = a.Runner
		}
		if derr := tr.DestroyAll(); derr == nil {
			sm.Delete(config.Project, Azure)
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

	workDir, werr := TempWorkDir("naeos-azure-destroy")
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

	sm.Delete(config.Project, Azure)
	return nil
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
			sb.WriteString(fmt.Sprintf(`resource "random_password" "%s_db" {
  length           = 32
  special          = true
  override_special = "!#$%%^&*()-_=+[]{}<>:?"
}

resource "azurerm_postgresql_flexible_server" "%s" {
  name                = "%s"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku_name            = "B_Standard_B1ms"
  version             = "15"
  storage_mb          = 32768

  backup_retention_days        = 7
  geo_redundant_backup_enabled = false

  admin_login    = "psqladmin"
  admin_password = random_password.%s_db.result

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
  name             = "allow-azure-services"
  server_id        = azurerm_postgresql_flexible_server.%s.id
  start_ip_address = "0.0.0.0"
  end_ip_address   = "0.0.0.0"
}

`, res.Name,
			res.Name, serverName,
			res.Name,
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

		case ResourceServerless:
			funcAppName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "azurerm_linux_function_app" "%s" {
  name                       = "%s"
  resource_group_name        = azurerm_resource_group.main.name
  location                   = azurerm_resource_group.main.location
  storage_account_name       = azurerm_storage_account.%s.name
  storage_account_access_key = azurerm_storage_account.%s.primary_access_key
  service_plan_id            = azurerm_service_plan.%s.id

  site_config {}

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

resource "azurerm_service_plan" "%s" {
  name                = "%s-plan"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  os_type             = "Linux"
  sku_name            = "Y1"

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

`, res.Name, funcAppName,
			res.Name, res.Name,
			res.Name,
			config.Environment, config.Project,
			res.Name, res.Name,
			config.Environment, config.Project))

		case ResourceMonitoring:
			sb.WriteString(fmt.Sprintf(`resource "azurerm_monitor_action_group" "%s" {
  name                = "%s"
  resource_group_name = azurerm_resource_group.main.name
  short_name          = "%s"

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

resource "azurerm_monitor_metric_alert" "%s" {
  name                = "%s-alert"
  resource_group_name = azurerm_resource_group.main.name
  scopes              = []
  description         = "%s %s %s alert"
  severity            = 3
  frequency           = "PT1M"
  window_size         = "PT5M"

  criteria {
    metric_namespace = "Microsoft.Compute/virtualMachines"
    metric_name      = "Percentage CPU"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = 80
  }

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

`, res.Name, res.Name, res.Name,
			config.Environment, config.Project,
			res.Name, res.Name,
			config.Project, config.Environment, res.Name,
			config.Environment, config.Project))

		case ResourceSecrets:
			vaultName := fmt.Sprintf("%s%s%s", config.Project, config.Environment, res.Name)
			vaultName = strings.ToLower(strings.ReplaceAll(vaultName, "-", ""))
			if len(vaultName) > 24 {
				vaultName = vaultName[:24]
			}
			sb.WriteString(fmt.Sprintf(`resource "azurerm_key_vault" "%s" {
  name                = "%s"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"

  enabled_for_deployment          = true
  enabled_for_template_deployment = true

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

data "azurerm_client_config" "current" {}

`, res.Name, vaultName,
			config.Environment, config.Project))

		case ResourceDNS:
			sb.WriteString(fmt.Sprintf(`resource "azurerm_dns_zone" "%s" {
  name                = "%s"
  resource_group_name = azurerm_resource_group.main.name

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

`, res.Name, res.Name,
			config.Environment, config.Project))

		case ResourceNetworking:
			vnetName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "azurerm_virtual_network" "%s" {
  name                = "%s"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  address_space       = ["10.0.0.0/16"]

  tags = {
    environment = "%s"
    project     = "%s"
  }
}

resource "azurerm_subnet" "%s" {
  name                 = "%s-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.%s.name
  address_prefixes     = ["10.0.1.0/24"]
}

`, res.Name, vnetName,
			config.Environment, config.Project,
			res.Name, res.Name, res.Name))

		}
	}

	return sb.String(), nil
}
