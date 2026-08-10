package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/supabase"
)

func setupSupabaseMock(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *supabase.Config) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	dir := t.TempDir()
	supabase.SetConfigDir(dir)
	t.Cleanup(func() { supabase.SetConfigDir(".naeos/supabase") })

	cfg := &supabase.Config{
		ProjectRef:     "mock-proj",
		URL:            server.URL,
		AnonKey:        "anon-key",
		ServiceRoleKey: "svc-key",
		ManagementURL:  server.URL,
	}
	if err := supabase.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	return server, cfg
}

func TestSupabaseAuthSignupSuccess(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/auth/v1/signup" {
			http.Error(w, "unexpected", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(supabase.SignUpResponse{
			ID: "user-123", Email: "new@example.com", Role: "authenticated",
		})
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "auth", "signup",
		"--email", "new@example.com", "--password", "secret")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}
	if !strings.Contains(output, "user-123") || !strings.Contains(output, "new@example.com") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseAuthSignupAPIError(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "email taken", http.StatusBadRequest)
	})

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "auth", "signup",
		"--email", "dup@example.com", "--password", "secret")
	if err == nil {
		t.Fatal("expected error when API rejects signup")
	}
	if !strings.Contains(err.Error(), "signup failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSupabaseAuthSigninSuccess(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/v1/token" {
			http.Error(w, "unexpected", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(supabase.Session{
			AccessToken: strings.Repeat("t", 30),
			User:        supabase.User{ID: "u1", Email: "me@example.com"},
		})
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "auth", "signin",
		"--email", "me@example.com", "--password", "secret")
	if err != nil {
		t.Fatalf("signin: %v", err)
	}
	if !strings.Contains(output, "Signed in as: me@example.com") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseAuthSigninAPIError(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
	})

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "auth", "signin",
		"--email", "me@example.com", "--password", "wrong")
	if err == nil {
		t.Fatal("expected error for invalid credentials")
	}
	if !strings.Contains(err.Error(), "signin failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSupabaseAuthSignoutSuccess(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/v1/logout" {
			http.Error(w, "unexpected", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "auth", "signout")
	if err != nil {
		t.Fatalf("signout: %v", err)
	}
	if !strings.Contains(output, "Signed out.") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseAuthUserConfirmed(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(supabase.User{
			ID: "u1", Email: "me@example.com", Role: "authenticated",
			EmailConfirmedAt: "2026-01-01T00:00:00Z",
		})
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "auth", "user")
	if err != nil {
		t.Fatalf("auth user: %v", err)
	}
	if !strings.Contains(output, "Email confirmed: 2026-01-01T00:00:00Z") {
		t.Errorf("expected confirmed output, got %q", output)
	}
}

func TestSupabaseAuthUserUnconfirmed(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(supabase.User{ID: "u1", Email: "me@example.com", Role: "authenticated"})
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "auth", "user")
	if err != nil {
		t.Fatalf("auth user: %v", err)
	}
	if strings.Contains(output, "Email confirmed") {
		t.Errorf("did not expect confirmation line, got %q", output)
	}
}

func TestSupabaseAuthUserError(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "auth", "user")
	if err == nil {
		t.Fatal("expected error when API rejects user fetch")
	}
	if !strings.Contains(err.Error(), "get user failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSupabaseAdminListUsersEmpty(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"users": []supabase.User{}})
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "auth", "admin", "list")
	if err != nil {
		t.Fatalf("admin list: %v", err)
	}
	if !strings.Contains(output, "No users found.") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseAdminListUsersWithData(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"users": []supabase.User{
				{ID: "u1", Email: "a@example.com", CreatedAt: "2026-01-01"},
				{ID: "u2", Email: "b@example.com", CreatedAt: "2026-02-01"},
			},
		})
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "auth", "admin", "list")
	if err != nil {
		t.Fatalf("admin list: %v", err)
	}
	if !strings.Contains(output, "a@example.com") || !strings.Contains(output, "b@example.com") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseAdminCreateUser(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/auth/v1/admin/users" {
			http.Error(w, "unexpected", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(supabase.User{ID: "u9", Email: "admin-created@example.com"})
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "auth", "admin", "create",
		"--email", "admin-created@example.com", "--password", "secret")
	if err != nil {
		t.Fatalf("admin create: %v", err)
	}
	if !strings.Contains(output, "Created user: admin-created@example.com (u9)") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseAdminDeleteUser(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "unexpected", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "auth", "admin", "delete", "--user-id", "u5")
	if err != nil {
		t.Fatalf("admin delete: %v", err)
	}
	if !strings.Contains(output, "Deleted user: u5") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseStorageListBucketsEmpty(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "storage", "list-buckets")
	if err != nil {
		t.Fatalf("list-buckets: %v", err)
	}
	if !strings.Contains(output, "No buckets found.") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseStorageListBucketsWithData(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]supabase.Bucket{
			{ID: "b1", Name: "public-bucket", Public: true, CreatedAt: "2026-01-01"},
			{ID: "b2", Name: "private-bucket", Public: false, CreatedAt: "2026-01-02"},
		})
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "storage", "list-buckets")
	if err != nil {
		t.Fatalf("list-buckets: %v", err)
	}
	if !strings.Contains(output, "public-bucket") || !strings.Contains(output, "yes") ||
		!strings.Contains(output, "private-bucket") || !strings.Contains(output, "no") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseStorageCreateBucket(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(supabase.Bucket{ID: "b3", Name: "new-bucket", Public: true})
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "storage", "create-bucket",
		"--name", "new-bucket", "--public")
	if err != nil {
		t.Fatalf("create-bucket: %v", err)
	}
	if !strings.Contains(output, "Bucket created: new-bucket") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseStorageListFilesEmpty(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storage/v1/object/list/media" {
			http.Error(w, "unexpected", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "storage", "list-files", "--bucket", "media")
	if err != nil {
		t.Fatalf("list-files: %v", err)
	}
	if !strings.Contains(output, "No files in bucket 'media'.") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseStorageListFilesWithData(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"name":"photo.png","updated_at":"2026-03-01T00:00:00Z","metadata":{"size":1234}}]`))
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "storage", "list-files",
		"--bucket", "media", "--prefix", "img/")
	if err != nil {
		t.Fatalf("list-files: %v", err)
	}
	if !strings.Contains(output, "photo.png") || !strings.Contains(output, "1234") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseStorageUploadSuccess(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	var uploaded bool
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/storage/v1/object/media/remote/file.txt" && r.Method == http.MethodPost {
			uploaded = true
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "unexpected", http.StatusNotFound)
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "storage", "upload",
		"--bucket", "media", "--source", src, "--dest", "remote/file.txt")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if !uploaded {
		t.Error("upload request not received by mock server")
	}
	if !strings.Contains(output, "Uploaded") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseStorageUploadMissingFile(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unexpected", http.StatusNotFound)
	})

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "storage", "upload",
		"--bucket", "media", "--source", "/nonexistent/file.txt", "--dest", "x.txt")
	if err == nil {
		t.Fatal("expected error for missing local file")
	}
	if !strings.Contains(err.Error(), "upload failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSupabaseStorageDownloadSuccess(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out", "downloaded.txt")

	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storage/v1/object/media/remote/file.txt" {
			http.Error(w, "unexpected", http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte("downloaded content"))
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "storage", "download",
		"--bucket", "media", "--source", "remote/file.txt", "--dest", dest)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if !strings.Contains(output, "Downloaded media/remote/file.txt") {
		t.Errorf("unexpected output: %q", output)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("downloaded file missing: %v", err)
	}
	if string(data) != "downloaded content" {
		t.Errorf("unexpected file content: %q", string(data))
	}
}

func TestSupabaseStorageDeleteFile(t *testing.T) {
	var deleted bool
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/storage/v1/object/media" && r.Method == http.MethodDelete {
			deleted = true
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "unexpected", http.StatusNotFound)
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "storage", "delete",
		"--bucket", "media", "--path", "old/file.txt")
	if err != nil {
		t.Fatalf("storage delete: %v", err)
	}
	if !deleted {
		t.Error("delete request not received by mock server")
	}
	if !strings.Contains(output, "Deleted media/old/file.txt") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseSQLRows(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects/mock-proj/database/query" {
			http.Error(w, "unexpected", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"name":"alice","age":30},{"name":"bob","age":25}]`))
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "sql", "SELECT * FROM users")
	if err != nil {
		t.Fatalf("sql: %v", err)
	}
	if !strings.Contains(output, "name") || !strings.Contains(output, "alice") || !strings.Contains(output, "bob") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseSQLNoRows(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "sql", "DELETE FROM old_rows")
	if err != nil {
		t.Fatalf("sql: %v", err)
	}
	if !strings.Contains(output, "Query executed successfully (0 rows).") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseSQLAPIError(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad query", http.StatusInternalServerError)
	})

	root := NewRootCommand()
	_, err := executeCommand(root, "supabase", "sql", "BROKEN")
	if err == nil {
		t.Fatal("expected error when SQL API fails")
	}
	if !strings.Contains(err.Error(), "SQL execution failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSupabaseStatusConnected(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(supabase.User{ID: "u1", Email: "me@example.com"})
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "status")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(output, "Status: CONNECTED") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseStatusAuthCheckError(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "status")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(output, "Status: CONNECTED (auth check:") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestSupabaseStatusJWKSURL(t *testing.T) {
	setupSupabaseMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(supabase.User{ID: "u1", Email: "me@example.com"})
	})

	dir := t.TempDir()
	supabase.SetConfigDir(dir)
	t.Cleanup(func() { supabase.SetConfigDir(".naeos/supabase") })

	cfg := &supabase.Config{
		ProjectRef:     "mock-proj",
		URL:            "https://mock-proj.supabase.co",
		AnonKey:        "anon-key",
		ServiceRoleKey: "svc-key",
		JWKSURL:        "https://mock-proj.supabase.co/auth/v1/.well-known/jwks.json",
	}
	if err := supabase.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "supabase", "status")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(output, "JWKS URL:") {
		t.Errorf("expected JWKS URL line, got %q", output)
	}
}
