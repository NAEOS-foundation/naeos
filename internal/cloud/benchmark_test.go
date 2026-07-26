package cloud

import (
	"testing"
)

func BenchmarkAWSPlan(b *testing.B) {
	adapter := &AWSAdapter{}
	config := &DeployConfig{
		Provider:    AWS,
		Region:      "us-east-1",
		Project:     "benchmark",
		Environment: "test",
		Resources: []Resource{
			{Name: "storage-1", Type: ResourceStorage},
			{Name: "compute-1", Type: ResourceCompute},
			{Name: "database-1", Type: ResourceDatabase},
			{Name: "cache-1", Type: ResourceCache},
			{Name: "queue-1", Type: ResourceQueue},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := adapter.Plan(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGCPPlan(b *testing.B) {
	adapter := &GCPAdapter{}
	config := &DeployConfig{
		Provider:    GCP,
		Region:      "us-central1",
		Project:     "benchmark",
		Environment: "test",
		Resources: []Resource{
			{Name: "storage-1", Type: ResourceStorage},
			{Name: "compute-1", Type: ResourceCompute},
			{Name: "database-1", Type: ResourceDatabase},
			{Name: "cdn-1", Type: ResourceCDN},
			{Name: "monitoring-1", Type: ResourceMonitoring},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := adapter.Plan(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAzurePlan(b *testing.B) {
	adapter := &AzureAdapter{}
	config := &DeployConfig{
		Provider:    Azure,
		Region:      "eastus",
		Project:     "benchmark",
		Environment: "test",
		Resources: []Resource{
			{Name: "storage-1", Type: ResourceStorage},
			{Name: "compute-1", Type: ResourceCompute},
			{Name: "database-1", Type: ResourceDatabase},
			{Name: "queue-1", Type: ResourceQueue},
			{Name: "dns-1", Type: ResourceDNS},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := adapter.Plan(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExportTerraform(b *testing.B) {
	adapters := []CloudAdapter{&AWSAdapter{}, &GCPAdapter{}, &AzureAdapter{}}
	config := &DeployConfig{
		Region:      "us-east-1",
		Project:     "benchmark",
		Environment: "test",
		Resources: []Resource{
			{Name: "storage-1", Type: ResourceStorage},
			{Name: "compute-1", Type: ResourceCompute},
		},
	}

	for _, a := range adapters {
		b.Run(a.Name(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := a.ExportTerraform(config)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
