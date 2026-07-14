package cloud

import (
	"fmt"
	"time"
)

// CloudProvider identifies a supported cloud infrastructure provider.
type CloudProvider string

const (
	// AWS is the Amazon Web Services provider.
	AWS   CloudProvider = "aws"
	// GCP is the Google Cloud Platform provider.
	GCP   CloudProvider = "gcp"
	// Azure is the Microsoft Azure provider.
	Azure CloudProvider = "azure"
)

// ResourceTypes maps abstract resource types to supported kinds.
const (
	ResourceStorage     = "storage"
	ResourceCompute     = "compute"
	ResourceDatabase    = "database"
	ResourceCache       = "cache"
	ResourceQueue       = "queue"
	ResourceCDN         = "cdn"
	ResourceServerless  = "serverless"
	ResourceMonitoring  = "monitoring"
	ResourceSecrets     = "secrets"
	ResourceDNS         = "dns"
	ResourceNetworking  = "networking"
)

// SupportedResourceTypes lists all abstract resource types across providers.
var SupportedResourceTypes = []string{
	ResourceStorage,
	ResourceCompute,
	ResourceDatabase,
	ResourceCache,
	ResourceQueue,
	ResourceCDN,
	ResourceServerless,
	ResourceMonitoring,
	ResourceSecrets,
	ResourceDNS,
	ResourceNetworking,
}

// DeployConfig holds the parameters for a cloud deployment operation.
type DeployConfig struct {
	Provider    CloudProvider
	Region      string
	Project     string
	Environment string
	Resources   []Resource
}

// Resource describes a single cloud resource to provision.
type Resource struct {
	Name string
	Type string
	Spec map[string]any
}

// DeployResult contains the outcome of a cloud deployment.
type DeployResult struct {
	Provider   CloudProvider
	Resources  []DeployedResource
	Terraform  string
	Status     string
	Timestamp  time.Time
}

// DeployedResource represents a cloud resource that has been provisioned.
type DeployedResource struct {
	Name string
	Type string
	ID   string
	ARN  string
}

// PlanResult contains resources and cost estimates for a planned deployment.
type PlanResult struct {
	Resources     []Resource   `json:"resources"`
	CostEstimate  CostEstimate `json:"cost_estimate"`
}

// CloudAdapter is the interface implemented by cloud provider adapters.
type CloudAdapter interface {
	Name() string
	Provider() CloudProvider
	Validate(config *DeployConfig) error
	Plan(config *DeployConfig) (*PlanResult, error)
	Deploy(config *DeployConfig) (*DeployResult, error)
	Destroy(config *DeployConfig) error
	ExportTerraform(config *DeployConfig) (string, error)
}

var adapterCache = map[CloudProvider]CloudAdapter{
	AWS:   &AWSAdapter{},
	GCP:   &GCPAdapter{},
	Azure: &AzureAdapter{},
}

// GetAdapter returns the CloudAdapter for the given provider.
func GetAdapter(provider CloudProvider) (CloudAdapter, error) {
	if adapter, ok := adapterCache[provider]; ok {
		return adapter, nil
	}
	return nil, fmt.Errorf("unsupported provider: %s", provider)
}
