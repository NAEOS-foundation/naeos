package supabase

import (
	"os"
	"testing"
)

func supabaseEnvConfig() *Config {
	url := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if url == "" || anonKey == "" {
		return nil
	}
	return &Config{
		URL:            url,
		AnonKey:        anonKey,
		ServiceRoleKey: serviceKey,
	}
}

func TestIntegrationAuthFlow(t *testing.T) {
	cfg := supabaseEnvConfig()
	if cfg == nil {
		t.Skip("SUPABASE_URL and SUPABASE_ANON_KEY not set; skipping integration test")
	}

	client := NewClient(cfg)

	user, err := client.GetUser()
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	t.Logf("Authenticated as: %s (%s)", user.Email, user.ID)
}

func TestIntegrationListBuckets(t *testing.T) {
	cfg := supabaseEnvConfig()
	if cfg == nil {
		t.Skip("SUPABASE_URL and SUPABASE_ANON_KEY not set; skipping integration test")
	}

	client := NewClient(cfg)

	buckets, err := client.ListBuckets()
	if err != nil {
		t.Fatalf("ListBuckets failed: %v", err)
	}

	t.Logf("Found %d buckets", len(buckets))
	for _, b := range buckets {
		t.Logf("  - %s (public: %v)", b.Name, b.Public)
	}
}

func TestIntegrationExecuteSQL(t *testing.T) {
	cfg := supabaseEnvConfig()
	if cfg == nil || cfg.ServiceRoleKey == "" {
		t.Skip("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY not set; skipping SQL integration test")
	}

	client := NewClient(cfg)

	result, err := client.ExecuteSQL("SELECT 1 as num")
	if err != nil {
		t.Fatalf("ExecuteSQL failed: %v", err)
	}

	if len(result.Rows) == 0 {
		t.Fatal("expected at least 1 row")
	}

	val := result.Rows[0]["num"]
	t.Logf("SQL result: num = %v", val)
}
