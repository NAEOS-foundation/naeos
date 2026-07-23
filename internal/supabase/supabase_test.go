package supabase

import (
	"os"
	"path/filepath"
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
		t.Skip("SUPABASE_URL and SUPABASE_ANON_KEY not set")
	}

	client := NewClient(cfg)

	user, err := client.GetUser()
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	t.Logf("Authenticated as: %s (%s)", user.Email, user.ID)
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

	client := NewClient(cfg)

	result, err := client.ExecuteSQL("SELECT 1 as num")
	if err != nil {
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

	client := NewClient(cfg)
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
		b[i] = letters[int(i*7+13)%len(letters)]
	}
	return string(b)
}

