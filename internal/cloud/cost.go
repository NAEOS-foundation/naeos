package cloud

import (
	"fmt"
	"sort"
	"strings"
)

// CostEstimate holds the total and per-resource monthly cost estimate.
type CostEstimate struct {
	TotalMonthlyUSD float64            `json:"total_monthly_usd"`
	Breakdown       map[string]float64 `json:"breakdown"`
	Currency        string             `json:"currency"`
}

// ResourceCost associates a resource type with its monthly cost.
type ResourceCost struct {
	ResourceType string  `json:"resource_type"`
	MonthlyUSD   float64 `json:"monthly_usd"`
}

// CostEstimator calculates cost estimates for cloud resources.
type CostEstimator struct {
	pricing map[string]map[string]float64
}

// NewCostEstimator creates a cost estimator with built-in pricing data.
func NewCostEstimator() *CostEstimator {
	ce := &CostEstimator{
		pricing: make(map[string]map[string]float64),
	}
	ce.loadPricing()
	return ce
}

func (ce *CostEstimator) loadPricing() {
	ce.pricing["aws"] = map[string]float64{
		"aws_s3_bucket":                  1.15,
		"aws_ecs_service":                8.47,
		"aws_rds_instance":               12.41,
		"aws_elasticache_cluster":        5.72,
		"aws_sqs_queue":                  0.40,
		"aws_cloudfront_distribution":    8.50,
		"aws_lambda_function":            2.08,
		"aws_cloudwatch_metric_alarm":    0.10,
		"aws_secretsmanager_secret":      0.40,
		"aws_route53_zone":               0.50,
		"aws_vpc":                        0.00,
	}

	ce.pricing["gcp"] = map[string]float64{
		"google_storage_bucket":            1.00,
		"google_cloud_run_service":         7.35,
		"google_sql_database_instance":     7.67,
		"google_redis_instance":            5.87,
		"google_pubsub_topic":              0.40,
		"google_compute_backend_bucket":    6.00,
		"google_cloudfunctions2_function":  1.80,
		"google_monitoring_alert_policy":   0.10,
		"google_secret_manager_secret":     0.06,
		"google_dns_managed_zone":          0.20,
		"google_compute_network":           0.00,
	}

	ce.pricing["azure"] = map[string]float64{
		"azurerm_storage_account":                    1.15,
		"azurerm_container_group":                    9.12,
		"azurerm_postgresql_flexible_server":         12.41,
		"azurerm_redis_cache":                        5.30,
		"azurerm_servicebus_queue":                   0.85,
		"azurerm_cdn_frontdoor_profile":              8.50,
		"azurerm_linux_function_app":                 1.90,
		"azurerm_monitor_action_group":               0.10,
		"azurerm_key_vault":                          0.03,
		"azurerm_dns_zone":                           0.50,
		"azurerm_virtual_network":                    0.00,
	}
}

// EstimateCost returns the total monthly cost for the given resources.
func (ce *CostEstimator) EstimateCost(provider string, resources []Resource) CostEstimate {
	providerPricing, ok := ce.pricing[provider]
	if !ok {
		return CostEstimate{
			TotalMonthlyUSD: 0,
			Breakdown:       make(map[string]float64),
			Currency:        "USD",
		}
	}

	breakdown := make(map[string]float64)
	var total float64

	for _, res := range resources {
		cost, exists := providerPricing[res.Type]
		if !exists {
			continue
		}
		key := fmt.Sprintf("%s (%s)", res.Name, res.Type)
		breakdown[key] = cost
		total += cost
	}

	return CostEstimate{
		TotalMonthlyUSD: total,
		Breakdown:       breakdown,
		Currency:        "USD",
	}
}

// EstimateCostByType returns per-resource-type cost breakdown.
func (ce *CostEstimator) EstimateCostByType(provider string, resources []Resource) []ResourceCost {
	providerPricing, ok := ce.pricing[provider]
	if !ok {
		return nil
	}

	seen := make(map[string]float64)
	for _, res := range resources {
		if cost, exists := providerPricing[res.Type]; exists {
			seen[res.Type] = cost
		}
	}

	var costs []ResourceCost
	for rt, cost := range seen {
		costs = append(costs, ResourceCost{ResourceType: rt, MonthlyUSD: cost})
	}
	sort.Slice(costs, func(i, j int) bool {
		return costs[i].ResourceType < costs[j].ResourceType
	})
	return costs
}

// FormatCost returns a human-readable cost breakdown string.
func (e CostEstimate) FormatCost() string {
	if len(e.Breakdown) == 0 {
		return fmt.Sprintf("Estimated cost: $0.00 %s/month", e.Currency)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Estimated cost: $%.2f %s/month\n", e.TotalMonthlyUSD, e.Currency))
	sb.WriteString("Breakdown:\n")

	keys := make([]string, 0, len(e.Breakdown))
	for k := range e.Breakdown {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("  %s: $%.2f/month\n", k, e.Breakdown[k]))
	}

	return sb.String()
}
