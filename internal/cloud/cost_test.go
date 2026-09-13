// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"fmt"
	"strings"
	"testing"
)

func TestCostEstimatorKnownResourceTypePositiveCost(t *testing.T) {
	ce := NewCostEstimator()
	resources := []Resource{
		{Name: "uploads", Type: "aws_s3_bucket"},
	}
	estimate := ce.EstimateCost("aws", resources)

	if estimate.TotalMonthlyUSD <= 0 {
		t.Errorf("expected positive cost for aws_s3_bucket, got %f", estimate.TotalMonthlyUSD)
	}
	if estimate.Currency != "USD" {
		t.Errorf("expected USD currency, got %s", estimate.Currency)
	}
}

func TestCostEstimatorAllResourceTypesPositive(t *testing.T) {
	ce := NewCostEstimator()
	providers := []string{"aws", "gcp", "azure"}
	allTypes := SupportedResourceTypes

	for _, provider := range providers {
		for _, resType := range allTypes {
			resources := []Resource{
				{Name: "test", Type: resType},
			}
			estimate := ce.EstimateCost(provider, resources)
			if estimate.TotalMonthlyUSD < 0 {
				t.Errorf("provider %s, type %s: expected non-negative cost, got %f", provider, resType, estimate.TotalMonthlyUSD)
			}
		}
	}
}

func TestCostEstimatorTotalIsSumOfBreakdown(t *testing.T) {
	ce := NewCostEstimator()
	resources := []Resource{
		{Name: "uploads", Type: "aws_s3_bucket"},
		{Name: "api", Type: "aws_ecs_service"},
		{Name: "db", Type: "aws_rds_instance"},
	}
	estimate := ce.EstimateCost("aws", resources)

	var sum float64
	for _, cost := range estimate.Breakdown {
		sum += cost
	}

	if estimate.TotalMonthlyUSD != sum {
		t.Errorf("total %f does not equal breakdown sum %f", estimate.TotalMonthlyUSD, sum)
	}
}

func TestCostEstimatorTotalMatchesKnownPricing(t *testing.T) {
	ce := NewCostEstimator()
	resources := []Resource{
		{Name: "uploads", Type: "aws_s3_bucket"},
		{Name: "api", Type: "aws_ecs_service"},
	}
	estimate := ce.EstimateCost("aws", resources)

	expected := 1.15 + 8.47
	if diff := estimate.TotalMonthlyUSD - expected; diff < -0.001 || diff > 0.001 {
		t.Errorf("expected total ~%f, got %f", expected, estimate.TotalMonthlyUSD)
	}
}

func TestCostEstimatorUnknownProviderReturnsZero(t *testing.T) {
	ce := NewCostEstimator()
	resources := []Resource{
		{Name: "test", Type: "aws_s3_bucket"},
	}
	estimate := ce.EstimateCost("digitalocean", resources)

	if estimate.TotalMonthlyUSD != 0 {
		t.Errorf("expected 0 for unknown provider, got %f", estimate.TotalMonthlyUSD)
	}
	if len(estimate.Breakdown) != 0 {
		t.Errorf("expected empty breakdown for unknown provider, got %d entries", len(estimate.Breakdown))
	}
}

func TestCostEstimatorUnknownResourceTypeReturnsZero(t *testing.T) {
	ce := NewCostEstimator()
	resources := []Resource{
		{Name: "unknown", Type: "unknown_resource_type"},
	}
	estimate := ce.EstimateCost("aws", resources)

	if estimate.TotalMonthlyUSD != 0 {
		t.Errorf("expected 0 for unknown resource type, got %f", estimate.TotalMonthlyUSD)
	}
	if len(estimate.Breakdown) != 0 {
		t.Errorf("expected empty breakdown for unknown type, got %d entries", len(estimate.Breakdown))
	}
}

func TestCostEstimatorEmptyResources(t *testing.T) {
	ce := NewCostEstimator()
	estimate := ce.EstimateCost("aws", []Resource{})

	if estimate.TotalMonthlyUSD != 0 {
		t.Errorf("expected 0 for empty resources, got %f", estimate.TotalMonthlyUSD)
	}
	if estimate.Currency != "USD" {
		t.Errorf("expected USD, got %s", estimate.Currency)
	}
}

func TestFormatCostWithBreakdown(t *testing.T) {
	ce := NewCostEstimator()
	resources := []Resource{
		{Name: "uploads", Type: "aws_s3_bucket"},
		{Name: "api", Type: "aws_ecs_service"},
	}
	estimate := ce.EstimateCost("aws", resources)
	formatted := estimate.FormatCost()

	if !strings.Contains(formatted, "$") {
		t.Error("formatted output should contain dollar sign")
	}
	if !strings.Contains(formatted, "USD/month") {
		t.Error("formatted output should contain USD/month")
	}
	if !strings.Contains(formatted, "Breakdown:") {
		t.Error("formatted output should contain Breakdown:")
	}
	if !strings.Contains(formatted, "uploads") {
		t.Error("formatted output should contain resource name")
	}
}

func TestFormatCostEmptyBreakdown(t *testing.T) {
	estimate := CostEstimate{
		TotalMonthlyUSD: 0,
		Breakdown:       make(map[string]float64),
		Currency:        "USD",
	}
	formatted := estimate.FormatCost()

	if !strings.Contains(formatted, "$0.00") {
		t.Error("formatted output should show $0.00")
	}
}

func TestEstimateCostByType(t *testing.T) {
	ce := NewCostEstimator()
	resources := []Resource{
		{Name: "a", Type: "aws_s3_bucket"},
		{Name: "b", Type: "aws_s3_bucket"},
		{Name: "c", Type: "aws_ecs_service"},
	}
	costs := ce.EstimateCostByType("aws", resources)

	if len(costs) != 2 {
		t.Errorf("expected 2 distinct types, got %d", len(costs))
	}

	seen := make(map[string]bool)
	for _, c := range costs {
		seen[c.ResourceType] = true
		if c.MonthlyUSD <= 0 {
			t.Errorf("expected positive cost for %s", c.ResourceType)
		}
	}
	if !seen["aws_s3_bucket"] || !seen["aws_ecs_service"] {
		t.Error("expected aws_s3_bucket and aws_ecs_service in results")
	}
}

func TestEstimateCostByTypeUnknownProvider(t *testing.T) {
	ce := NewCostEstimator()
	costs := ce.EstimateCostByType("unknown", []Resource{{Name: "x", Type: "y"}})

	if costs != nil {
		t.Errorf("expected nil for unknown provider, got %v", costs)
	}
}

func TestCostEstimateBreakdownKeys(t *testing.T) {
	ce := NewCostEstimator()
	resources := []Resource{
		{Name: "uploads", Type: "aws_s3_bucket"},
	}
	estimate := ce.EstimateCost("aws", resources)

	expectedKey := "uploads (aws_s3_bucket)"
	if _, ok := estimate.Breakdown[expectedKey]; !ok {
		t.Errorf("expected breakdown key %q, got keys: %v", expectedKey, keysOf(estimate.Breakdown))
	}
}

func keysOf(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestFormatCostDetailedContent(t *testing.T) {
	estimate := CostEstimate{
		TotalMonthlyUSD: 22.02,
		Breakdown: map[string]float64{
			"uploads (aws_s3_bucket)": 1.15,
			"api (aws_ecs_service)":   8.47,
			"db (aws_rds_instance)":   12.41,
		},
		Currency: "USD",
	}
	formatted := estimate.FormatCost()

	expected := fmt.Sprintf("Estimated cost: $%.2f USD/month", 22.02)
	if !strings.Contains(formatted, expected) {
		t.Errorf("expected formatted total %q in output", expected)
	}

	lines := strings.Split(strings.TrimSpace(formatted), "\n")
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines (header + 'Breakdown:' + 3 items), got %d", len(lines))
	}
}
