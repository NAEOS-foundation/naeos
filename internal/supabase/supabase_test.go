// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package supabase

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func onCI() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

func supabaseEnvConfig() *Config {
	url := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")
	if anonKey == "" {
		anonKey = os.Getenv("SUPABASE_PUBLISHABEL_KEY")
	}
	if anonKey == "" {
		anonKey = os.Getenv("SUPABASE_PUBLISHABLE_KEY")
	}
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if serviceKey == "" {
		serviceKey = os.Getenv("SUPABASE_SECRET_KEY")
	}
	jwksURL := os.Getenv("SUPABASE_JWKS_URL")
	if url == "" || anonKey == "" {
		return nil
	}
	return &Config{
		URL:            url,
		AnonKey:        anonKey,
		ServiceRoleKey: serviceKey,
		JWKSURL:        jwksURL,
	}
}

func TestIntegrationListBuckets(t *testing.T) {
	cfg := supabaseEnvConfig()
	if cfg == nil {
		t.Skip("SUPABASE_URL and SUPABASE_ANON_KEY not set")
	}

	client := NewClient(cfg)

	buckets, err := client.ListBuckets()
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}

	t.Logf("Found %d buckets", len(buckets))
}

func TestIntegrationExecuteSQL(t *testing.T) {
	cfg := supabaseEnvConfig()
	if cfg == nil || cfg.ServiceRoleKey == "" {
		t.Skip("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY not set")
	}
	accessToken := os.Getenv("SUPABASE_ACCESS_TOKEN")
	if accessToken == "" {
		t.Skip("SUPABASE_ACCESS_TOKEN not set; Management API tests skipped")
	}
	cfg.AccessToken = accessToken

	client := NewClient(cfg)

	result, err := client.ExecuteSQL("SELECT 1 as num")
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			if onCI() {
				t.Fatalf("ExecuteSQL: SUPABASE_ACCESS_TOKEN rejected by the Management API on CI: %v", err)
			}
			t.Skipf("SUPABASE_ACCESS_TOKEN rejected by the Management API (expired/invalid): %v", err)
		}
		t.Fatalf("ExecuteSQL: %v", err)
	}
	if len(result.Rows) == 0 {
		t.Fatal("expected at least 1 row")
	}
	t.Logf("SQL: num = %v", result.Rows[0]["num"])
}

func TestIntegrationSignUpSignInFlow(t *testing.T) {
	cfg := supabaseEnvConfig()
	if cfg == nil {
		t.Skip("SUPABASE_URL and SUPABASE_ANON_KEY not set")
	}

	client := NewClient(cfg)

	email := "test-" + randString(8) + "@naeos-test.com"
	password := "Test1234!@#$"

	result, err := client.SignUp(SignUpParams{Email: email, Password: password})
	if err != nil {
		if strings.Contains(err.Error(), "email_address_invalid") || strings.Contains(err.Error(), "rate_limit") {
			if onCI() {
				t.Fatalf("SignUp: %v", err)
			}
			t.Skipf("Supabase auth policy rejected test email (likely restricted domains or rate limit): %v", err)
		}
		t.Fatalf("SignUp: %v", err)
	}
	t.Logf("Signed up: %s (%s)", result.Email, result.ID)

	session, err := client.SignInWithEmail(email, password)
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	t.Logf("Signed in: %s", session.User.Email)

	user, err := client.GetUser()
	if err != nil {
		t.Fatalf("GetUser after signin: %v", err)
	}
	if user.Email != email {
		t.Errorf("expected email %s, got %s", email, user.Email)
	}

	if err := client.SignOut(); err != nil {
		t.Fatalf("SignOut: %v", err)
	}
	t.Log("Signed out")
}

func TestIntegrationStorageUploadDownload(t *testing.T) {
	cfg := supabaseEnvConfig()
	if cfg == nil {
		t.Skip("SUPABASE_URL and SUPABASE_ANON_KEY not set")
	}
	if cfg.ServiceRoleKey == "" {
		t.Skip("SUPABASE_SERVICE_ROLE_KEY not set; storage lifecycle tests skipped")
	}

	privileged := *cfg
	// Storage lifecycle (create/delete bucket, upload/download/delete object)
	// requires service-role privileges, so the service role key is used as the
	// Storage API apikey here. Do not reuse SUPABASE_ACCESS_TOKEN for this:
	// it is a Management API credential, a distinct token type. When the
	// apikey carries the service-role JWT, the Storage gateway authenticates
	// through it and no Authorization header is required.
	privileged.AnonKey = cfg.ServiceRoleKey
	client := NewClient(&privileged)
	tmpDir := t.TempDir()

	bucketName := "test-" + randString(6)

	bucket, err := client.CreateBucket(bucketName, false)
	if err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	t.Logf("Created bucket: %s", bucket.Name)
	t.Cleanup(func() { client.DeleteBucket(bucketName) })

	srcContent := "hello supabase storage"
	srcFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte(srcContent), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := client.UploadFile(bucketName, srcFile, "uploads/test.txt"); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	t.Log("Uploaded test.txt")

	destFile := filepath.Join(tmpDir, "downloaded.txt")
	if err := client.DownloadFile(bucketName, "uploads/test.txt", destFile); err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}

	data, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != srcContent {
		t.Errorf("content mismatch: got %q, want %q", string(data), srcContent)
	}
	t.Logf("Download verified: %s", string(data))

	if err := client.DeleteFile(bucketName, "uploads/test.txt"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}
	t.Log("Deleted test.txt")
}

func TestIntegrationAdminCreateUser(t *testing.T) {
	cfg := supabaseEnvConfig()
	if cfg == nil || cfg.ServiceRoleKey == "" {
		t.Skip("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY not set")
	}

	client := NewClient(cfg)

	email := "admin-test-" + randString(8) + "@naeos-test.com"
	password := "Admin4567!@#$"

	user, err := client.AdminCreateUser(email, password, nil)
	if err != nil {
		t.Fatalf("AdminCreateUser: %v", err)
	}
	t.Logf("Admin created: %s (%s)", user.Email, user.ID)
	t.Cleanup(func() { client.AdminDeleteUser(user.ID) })

	users, err := client.AdminListUsers()
	if err != nil {
		t.Fatalf("AdminListUsers: %v", err)
	}

	found := false
	for _, u := range users {
		if u.Email == email {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("user %s not found in admin list", email)
	}
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[(i*7+13)%len(letters)]
	}
	return string(b)
}
