package auth

import (
	"testing"
	"time"
)

func TestNewRBAC(t *testing.T) {
	r := NewRBAC()
	if r == nil {
		t.Fatal("expected RBAC to be created")
	}
}

func TestAddRole(t *testing.T) {
	r := NewRBAC()
	role := &Role{Name: "admin", Permissions: []string{"users.*"}}
	r.AddRole(role)

	got, ok := r.GetRole("admin")
	if !ok {
		t.Fatal("expected role to be found")
	}
	if got.Name != "admin" {
		t.Errorf("expected name 'admin', got %s", got.Name)
	}
}

func TestRemoveRole(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{Name: "admin"})
	r.RemoveRole("admin")

	_, ok := r.GetRole("admin")
	if ok {
		t.Error("expected role to be removed")
	}
}

func TestListRoles(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{Name: "admin"})
	r.AddRole(&Role{Name: "user"})

	roles := r.ListRoles()
	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}
}

func TestHasPermission(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{Name: "admin", Permissions: []string{"users"}})
	r.AddPermission(&Permission{Resource: "users", Actions: []string{"read", "write", "delete"}})

	user := &User{Roles: []string{"admin"}}

	if !r.HasPermission(user, "users", "read") {
		t.Error("expected permission to be granted")
	}

	if !r.HasPermission(user, "users", "delete") {
		t.Error("expected permission to be granted")
	}

	if r.HasPermission(user, "posts", "read") {
		t.Error("expected permission to be denied")
	}
}

func TestHasPermissionWildcard(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{Name: "superadmin", Permissions: []string{"*"}})
	r.AddPermission(&Permission{Resource: "*", Actions: []string{"*"}})

	user := &User{Roles: []string{"superadmin"}}

	if !r.HasPermission(user, "anything", "do") {
		t.Error("expected wildcard permission")
	}
}

func TestAssignRole(t *testing.T) {
	r := NewRBAC()
	user := &User{Roles: []string{}}
	r.AssignRole(user, "admin")

	if len(user.Roles) != 1 {
		t.Errorf("expected 1 role, got %d", len(user.Roles))
	}
}

func TestRemoveRoleFromUser(t *testing.T) {
	r := NewRBAC()
	user := &User{Roles: []string{"admin", "user"}}
	r.RemoveRoleFromUser(user, "admin")

	if len(user.Roles) != 1 {
		t.Errorf("expected 1 role, got %d", len(user.Roles))
	}
	if user.Roles[0] != "user" {
		t.Errorf("expected 'user', got %s", user.Roles[0])
	}
}

func TestGoogleOAuth2(t *testing.T) {
	g := NewGoogleOAuth2("client-id", "client-secret", "http://localhost:8080/callback")

	if g.Name() != "google" {
		t.Errorf("expected name 'google', got %s", g.Name())
	}

	url := g.GetAuthorizationURL("state123")
	if url == "" {
		t.Error("expected non-empty URL")
	}

	token, err := g.ExchangeCode("code123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken == "" {
		t.Error("expected access token")
	}

	user, err := g.GetUser(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email == "" {
		t.Error("expected user email")
	}
}

func TestGitHubOAuth2(t *testing.T) {
	g := NewGitHubOAuth2("client-id", "client-secret", "http://localhost:8080/callback")

	if g.Name() != "github" {
		t.Errorf("expected name 'github', got %s", g.Name())
	}

	url := g.GetAuthorizationURL("state123")
	if url == "" {
		t.Error("expected non-empty URL")
	}

	token, err := g.ExchangeCode("code123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken == "" {
		t.Error("expected access token")
	}
}

func TestAPIKeyManager(t *testing.T) {
	m := NewAPIKeyManager()

	key, err := m.Generate("user-1", "my-key", []string{"read", "write"}, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if key == "" {
		t.Error("expected non-empty key")
	}

	apiKey, ok := m.Validate(key)
	if !ok {
		t.Fatal("expected key to be valid")
	}
	if apiKey.UserID != "user-1" {
		t.Errorf("expected user 'user-1', got %s", apiKey.UserID)
	}
}

func TestAPIKeyExpired(t *testing.T) {
	m := NewAPIKeyManager()

	key, _ := m.Generate("user-1", "my-key", nil, time.Now().Add(-time.Hour))

	_, ok := m.Validate(key)
	if ok {
		t.Error("expected key to be expired")
	}
}

func TestAPIKeyRevoke(t *testing.T) {
	m := NewAPIKeyManager()

	key, _ := m.Generate("user-1", "my-key", nil, time.Now().Add(time.Hour))

	revoked := m.Revoke(key)
	if !revoked {
		t.Error("expected key to be revoked")
	}

	_, ok := m.Validate(key)
	if ok {
		t.Error("expected key to be invalid after revoke")
	}
}

func TestAPIKeyListByUser(t *testing.T) {
	m := NewAPIKeyManager()

	m.Generate("user-1", "key1", nil, time.Now().Add(time.Hour))
	m.Generate("user-1", "key2", nil, time.Now().Add(time.Hour))
	m.Generate("user-2", "key3", nil, time.Now().Add(time.Hour))

	keys := m.ListByUser("user-1")
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestSessionManager(t *testing.T) {
	m := NewSessionManager()

	id, err := m.Create("user-1", map[string]any{"role": "admin"}, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	session, ok := m.Get(id)
	if !ok {
		t.Fatal("expected session to be found")
	}
	if session.UserID != "user-1" {
		t.Errorf("expected user 'user-1', got %s", session.UserID)
	}
}

func TestSessionExpired(t *testing.T) {
	m := NewSessionManager()

	id, _ := m.Create("user-1", nil, time.Now().Add(-time.Hour))

	_, ok := m.Get(id)
	if ok {
		t.Error("expected session to be expired")
	}
}

func TestSessionDelete(t *testing.T) {
	m := NewSessionManager()

	id, _ := m.Create("user-1", nil, time.Now().Add(time.Hour))

	deleted := m.Delete(id)
	if !deleted {
		t.Error("expected session to be deleted")
	}

	_, ok := m.Get(id)
	if ok {
		t.Error("expected session to be deleted")
	}
}

func TestSessionCleanup(t *testing.T) {
	m := NewSessionManager()

	m.Create("user-1", nil, time.Now().Add(-time.Hour))
	m.Create("user-2", nil, time.Now().Add(time.Hour))

	removed := m.Cleanup()
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}
}

func TestAuthManager(t *testing.T) {
	m := NewManager()

	// Create user
	user := &User{ID: "user-1", Email: "test@example.com", Name: "Test User"}
	m.CreateUser(user)

	// Generate API key
	key, _ := m.APIKeys().Generate("user-1", "my-key", nil, time.Now().Add(time.Hour))

	// Authenticate
	authUser, ok := m.AuthenticateAPIKey(key)
	if !ok {
		t.Fatal("expected authentication to succeed")
	}
	if authUser.ID != "user-1" {
		t.Errorf("expected user 'user-1', got %s", authUser.ID)
	}
}

func TestAuthManagerInvalidKey(t *testing.T) {
	m := NewManager()

	_, ok := m.AuthenticateAPIKey("invalid-key")
	if ok {
		t.Error("expected authentication to fail")
	}
}

func TestRegisterOAuth2(t *testing.T) {
	m := NewManager()

	google := NewGoogleOAuth2("id", "secret", "http://localhost/callback")
	m.RegisterOAuth2(google)

	provider, ok := m.GetOAuth2("google")
	if !ok {
		t.Fatal("expected OAuth2 provider to be found")
	}
	if provider.Name() != "google" {
		t.Errorf("expected 'google', got %s", provider.Name())
	}
}
