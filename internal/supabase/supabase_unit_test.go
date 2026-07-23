package supabase

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newTestServer(fn func(w http.ResponseWriter, r *http.Request)) (*Client, *httptest.Server) {
	srv := httptest.NewServer(http.HandlerFunc(fn))
	c := NewClient(&Config{
		URL:            srv.URL,
		AnonKey:        "test-anon-key",
		ServiceRoleKey: "test-svc-role-key",
		ProjectRef:     "test-proj",
	})
	return c, srv
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	SetConfigDir(dir)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestSignUp(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/auth/v1/signup" {
			t.Errorf("path = %q, want /auth/v1/signup", r.URL.Path)
		}
		if r.Header.Get("apikey") != "test-anon-key" {
			t.Errorf("apikey = %q", r.Header.Get("apikey"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SignUpResponse{
			ID: "u1", Email: "a@b.com", Aud: "authenticated",
			Role: "authenticated", CreatedAt: "2024-01-01T00:00:00Z",
		})
	})
	defer srv.Close()

	result, err := c.SignUp(SignUpParams{Email: "a@b.com", Password: "pw"})
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}
	if result.Email != "a@b.com" {
		t.Errorf("Email = %q, want %q", result.Email, "a@b.com")
	}
	if result.ID != "u1" {
		t.Errorf("ID = %q, want %q", result.ID, "u1")
	}
}

func TestSignInWithEmail(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/auth/v1/token" || r.URL.RawQuery != "grant_type=password" {
			t.Errorf("unexpected URL: %s", r.URL.String())
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Session{
			AccessToken:  "at-123",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			RefreshToken: "rt-456",
			User: User{
				ID: "u1", Email: "a@b.com", Role: "authenticated",
				CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z",
			},
		})
	})
	defer srv.Close()

	session, err := c.SignInWithEmail("a@b.com", "pw")
	if err != nil {
		t.Fatalf("SignInWithEmail: %v", err)
	}
	if session.AccessToken != "at-123" {
		t.Errorf("AccessToken = %q", session.AccessToken)
	}
	if c.AuthToken() != "at-123" {
		t.Errorf("client AuthToken = %q, want at-123", c.AuthToken())
	}
}

func TestSignOut(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/auth/v1/logout" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	c.SetAuthToken("tok")
	if err := c.SignOut(); err != nil {
		t.Fatalf("SignOut: %v", err)
	}
	if tok := c.AuthToken(); tok != "" {
		t.Errorf("AuthToken after signout = %q, want empty", tok)
	}
}

func TestGetUser(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/auth/v1/user" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer at-123" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(User{
			ID: "u1", Email: "a@b.com", Role: "authenticated",
			CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z",
		})
	})
	defer srv.Close()

	c.SetAuthToken("at-123")
	user, err := c.GetUser()
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if user.Email != "a@b.com" {
		t.Errorf("Email = %q", user.Email)
	}
}

func TestRefreshToken(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "grant_type=refresh_token" {
			t.Errorf("query = %q", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Session{
			AccessToken: "at-new", RefreshToken: "rt-new",
			TokenType: "bearer", ExpiresIn: 3600,
			User: User{ID: "u1", Email: "a@b.com", Role: "authenticated",
				CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"},
		})
	})
	defer srv.Close()

	session, err := c.RefreshToken("rt-old")
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}
	if session.AccessToken != "at-new" {
		t.Errorf("AccessToken = %q", session.AccessToken)
	}
	if c.AuthToken() != "at-new" {
		t.Errorf("client AuthToken = %q", c.AuthToken())
	}
}

func TestSignInWithOAuth(t *testing.T) {
	c := NewClient(&Config{URL: "https://example.supabase.co"})
	url, err := c.SignInWithOAuth("github")
	if err != nil {
		t.Fatalf("SignInWithOAuth: %v", err)
	}
	if url != "https://example.supabase.co/auth/v1/authorize?provider=github&redirect_to=http%3A%2F%2Flocalhost%3A9999%2Fauth%2Fcallback" {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestAdminListUsers(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/auth/v1/admin/users" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("apikey") != "test-svc-role-key" {
			t.Errorf("apikey = %q", r.Header.Get("apikey"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(AdminUserResponse{
			Users: []User{
				{ID: "u1", Email: "a@b.com", Role: "authenticated",
					CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z"},
			},
		})
	})
	defer srv.Close()

	users, err := c.AdminListUsers()
	if err != nil {
		t.Fatalf("AdminListUsers: %v", err)
	}
	if len(users) != 1 || users[0].Email != "a@b.com" {
		t.Errorf("unexpected users: %+v", users)
	}
}

func TestAdminCreateUser(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/auth/v1/admin/users" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(User{
			ID: "u-new", Email: "new@b.com", Role: "authenticated",
			CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z",
		})
	})
	defer srv.Close()

	user, err := c.AdminCreateUser("new@b.com", "pw", nil)
	if err != nil {
		t.Fatalf("AdminCreateUser: %v", err)
	}
	if user.Email != "new@b.com" || user.ID != "u-new" {
		t.Errorf("unexpected user: %+v", user)
	}
}

func TestAdminDeleteUser(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/auth/v1/admin/users/u-del" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	if err := c.AdminDeleteUser("u-del"); err != nil {
		t.Fatalf("AdminDeleteUser: %v", err)
	}
}

func TestListBuckets(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/storage/v1/bucket" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Bucket{
			{ID: "b1", Name: "Bucket 1", Public: true, CreatedAt: "now", UpdatedAt: "now"},
		})
	})
	defer srv.Close()

	buckets, err := c.ListBuckets()
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}
	if len(buckets) != 1 || buckets[0].Name != "Bucket 1" {
		t.Errorf("unexpected buckets: %+v", buckets)
	}
}

func TestCreateBucket(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/storage/v1/bucket" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Bucket{ID: "b-new", Name: "new-bucket", Public: false, CreatedAt: "now", UpdatedAt: "now"})
	})
	defer srv.Close()

	b, err := c.CreateBucket("new-bucket", false)
	if err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	if b.Name != "new-bucket" || b.Public {
		t.Errorf("unexpected bucket: %+v", b)
	}
}

func TestGetBucket(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/storage/v1/bucket/b1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Bucket{ID: "b1", Name: "Bucket 1", Public: true, CreatedAt: "now", UpdatedAt: "now"})
	})
	defer srv.Close()

	b, err := c.GetBucket("b1")
	if err != nil {
		t.Fatalf("GetBucket: %v", err)
	}
	if b.ID != "b1" {
		t.Errorf("ID = %q", b.ID)
	}
}

func TestDeleteBucket(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/storage/v1/bucket/b-del" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	if err := c.DeleteBucket("b-del"); err != nil {
		t.Fatalf("DeleteBucket: %v", err)
	}
}

func TestListFiles(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/storage/v1/object/list/mybucket" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]FileObject{
			{Name: "file.txt", BucketID: "mybucket", ID: "f1", CreatedAt: "now", UpdatedAt: "now", LastAccessedAt: "now"},
		})
	})
	defer srv.Close()

	files, err := c.ListFiles("mybucket", "")
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if len(files) != 1 || files[0].Name != "file.txt" {
		t.Errorf("unexpected files: %+v", files)
	}
}

func TestDeleteFile(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/storage/v1/object/mybucket" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	if err := c.DeleteFile("mybucket", "path/to/file.txt"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}
}

func TestExecuteSQL(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/projects/test-proj/database/query" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("apikey") != "test-svc-role-key" {
			t.Errorf("apikey = %q", r.Header.Get("apikey"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(QueryResult{
			Rows: []map[string]any{{"num": float64(1)}},
		})
	})
	defer srv.Close()

	result, err := c.ExecuteSQL("SELECT 1")
	if err != nil {
		t.Fatalf("ExecuteSQL: %v", err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}
	if result.Rows[0]["num"] != float64(1) {
		t.Errorf("num = %v", result.Rows[0]["num"])
	}
}

func TestGetRoles(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/projects/test-proj/database/roles" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Role{{Name: "anon"}, {Name: "authenticated"}})
	})
	defer srv.Close()

	roles, err := c.GetRoles()
	if err != nil {
		t.Fatalf("GetRoles: %v", err)
	}
	if len(roles) != 2 || roles[0].Name != "anon" {
		t.Errorf("unexpected roles: %+v", roles)
	}
}

func TestListFunctions(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/projects/test-proj/functions" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]EdgeFunction{
			{ID: "fn1", Name: "hello", Slug: "hello", Status: "active", Version: 1,
				CreatedAt: "now", UpdatedAt: "now", Entrypoint: "index.ts", ImportMap: false, VerifyJWT: true},
		})
	})
	defer srv.Close()

	fns, err := c.ListFunctions()
	if err != nil {
		t.Fatalf("ListFunctions: %v", err)
	}
	if len(fns) != 1 || fns[0].Slug != "hello" {
		t.Errorf("unexpected functions: %+v", fns)
	}
}

func TestDeployFunction(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/projects/test-proj/functions" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(EdgeFunction{
			ID: "fn-new", Name: "my-fn", Slug: "my-fn", Status: "active", Version: 1,
			CreatedAt: "now", UpdatedAt: "now", Entrypoint: "index.ts",
		})
	})
	defer srv.Close()

	fn, err := c.DeployFunction("my-fn", "my-fn", "index.ts", "export default () => {}", true, false)
	if err != nil {
		t.Fatalf("DeployFunction: %v", err)
	}
	if fn.Slug != "my-fn" {
		t.Errorf("Slug = %q", fn.Slug)
	}
}

func TestDeleteFunction(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/api/v1/projects/test-proj/functions/my-fn" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	if err := c.DeleteFunction("my-fn"); err != nil {
		t.Fatalf("DeleteFunction: %v", err)
	}
}

func TestInvokeFunction(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/functions/v1/my-fn" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("apikey") != "test-anon-key" {
			t.Errorf("apikey = %q", r.Header.Get("apikey"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"ok"}`))
	})
	defer srv.Close()

	data, err := c.InvokeFunction("my-fn", map[string]any{"input": "test"})
	if err != nil {
		t.Fatalf("InvokeFunction: %v", err)
	}
	if string(data) != `{"result":"ok"}` {
		t.Errorf("body = %q", string(data))
	}
}

func TestInvokeFunctionWithAuth(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer my-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`ok`))
	})
	defer srv.Close()

	c.SetAuthToken("my-token")
	_, err := c.InvokeFunction("my-fn", nil)
	if err != nil {
		t.Fatalf("InvokeFunction: %v", err)
	}
}

func TestGetFunctionURL(t *testing.T) {
	c := NewClient(&Config{URL: "https://example.supabase.co"})
	u := c.GetFunctionURL("hello")
	if u != "https://example.supabase.co/functions/v1/hello" {
		t.Errorf("url = %q", u)
	}
}

func TestDefaultFunctionsDir(t *testing.T) {
	dir := DefaultFunctionsDir()
	if dir == "" {
		t.Error("DefaultFunctionsDir() returned empty")
	}
}

func TestAPIError(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"msg":"bad request"}`))
	})
	defer srv.Close()

	_, err := c.ListBuckets()
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
}

func TestMalformedJSON(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{bad json`))
	})
	defer srv.Close()

	_, err := c.ListBuckets()
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestSignOutError(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"msg":"invalid token"}`))
	})
	defer srv.Close()

	if err := c.SignOut(); err == nil {
		t.Fatal("expected error for 401 SignOut")
	}
}

func TestDeleteBucketError(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	if err := c.DeleteBucket("nonexistent"); err == nil {
		t.Fatal("expected error for 404 DeleteBucket")
	}
}

func TestAdminDeleteUserError(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	if err := c.AdminDeleteUser("nonexistent"); err == nil {
		t.Fatal("expected error for 404 AdminDeleteUser")
	}
}

func TestDeleteFunctionError(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	if err := c.DeleteFunction("nonexistent"); err == nil {
		t.Fatal("expected error for 404 DeleteFunction")
	}
}

func TestSignUpError(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"msg":"user already exists"}`))
	})
	defer srv.Close()

	_, err := c.SignUp(SignUpParams{Email: "exists@b.com", Password: "pw"})
	if err == nil {
		t.Fatal("expected error for 409 SignUp")
	}
}

func TestSignInError(t *testing.T) {
	c, srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"msg":"invalid credentials"}`))
	})
	defer srv.Close()

	_, err := c.SignInWithEmail("bad@b.com", "wrong")
	if err == nil {
		t.Fatal("expected error for 400 SignIn")
	}
}
