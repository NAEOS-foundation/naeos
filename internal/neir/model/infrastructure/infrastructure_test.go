package infrastructure

import "testing"

func TestProviderConstants(t *testing.T) {
	tests := []struct {
		p    Provider
		want string
	}{
		{ProviderAWS, "aws"},
		{ProviderGCP, "gcp"},
		{ProviderAzure, "azure"},
		{ProviderLocal, "local"},
	}
	for _, tt := range tests {
		if string(tt.p) != tt.want {
			t.Errorf("Provider(%s) = %s, want %s", tt.want, string(tt.p), tt.want)
		}
	}
}

func TestInfrastructure_ZeroValue(t *testing.T) {
	var i Infrastructure
	if i.Provider != "" {
		t.Error("expected empty Provider")
	}
	if i.Resources != nil {
		t.Error("expected nil Resources")
	}
}

func TestInfrastructure_Full(t *testing.T) {
	i := Infrastructure{
		Provider:    ProviderAWS,
		Region:      "us-east-1",
		Project:     "my-project",
		Environment: "production",
		Resources: []Resource{
			{Name: "db", Kind: "rds", Type: "postgres", Spec: map[string]string{"version": "16"}},
		},
		Networking: []Network{
			{Name: "vpc", Kind: "vpc", Ports: []int{443, 80}},
		},
		Attributes: map[string]string{"key": "val"},
	}
	if i.Provider != ProviderAWS {
		t.Errorf("expected aws, got %s", i.Provider)
	}
	if i.Region != "us-east-1" {
		t.Errorf("expected us-east-1, got %s", i.Region)
	}
	if len(i.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(i.Resources))
	}
	if i.Resources[0].Spec["version"] != "16" {
		t.Errorf("expected version 16, got %s", i.Resources[0].Spec["version"])
	}
	if len(i.Networking) != 1 {
		t.Errorf("expected 1 network, got %d", len(i.Networking))
	}
	if len(i.Networking[0].Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(i.Networking[0].Ports))
	}
}
