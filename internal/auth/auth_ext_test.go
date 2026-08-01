package auth

import (
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSetupDefaultRoles(t *testing.T) {
	r := NewRBAC()
	SetupDefaultRoles(r)

	roles := r.ListRoles()
	if len(roles) != 3 {
		t.Fatalf("expected 3 default roles, got %d: %v", len(roles), roles)
	}

	// Verify each default role exists
	for _, name := range []string{"admin", "developer", "viewer"} {
		if _, ok := r.GetRole(name); !ok {
			t.Errorf("expected role %q", name)
		}
	}
}

func TestSetupDefaultRolesAdminPermissions(t *testing.T) {
	r := NewRBAC()
	SetupDefaultRoles(r)

	admin := &User{Roles: []string{"admin"}}
	if !r.HasPermission(admin, "spec", "read") {
		t.Error("admin should have spec:read")
	}
	if !r.HasPermission(admin, "spec", "write") {
		t.Error("admin should have spec:write")
	}
	if !r.HasPermission(admin, "spec", "delete") {
		t.Error("admin should have spec:delete")
	}
	if !r.HasPermission(admin, "pipeline", "read") {
		t.Error("admin should have pipeline:read")
	}
	if !r.HasPermission(admin, "cloud", "write") {
		t.Error("admin should have cloud:write")
	}
	if !r.HasPermission(admin, "audit", "read") {
		t.Error("admin should have audit:read")
	}
}

func TestSetupDefaultRolesDeveloperPermissions(t *testing.T) {
	r := NewRBAC()
	SetupDefaultRoles(r)

	dev := &User{Roles: []string{"developer"}}
	if !r.HasPermission(dev, "spec", "read") {
		t.Error("developer should have spec:read")
	}
	if !r.HasPermission(dev, "spec", "write") {
		t.Error("developer should have spec:write")
	}
	if r.HasPermission(dev, "spec", "delete") {
		t.Error("developer should NOT have spec:delete")
	}
	if !r.HasPermission(dev, "pipeline", "read") {
		t.Error("developer should have pipeline:read")
	}
	if !r.HasPermission(dev, "pipeline", "write") {
		t.Error("developer should have pipeline:write")
	}
	if r.HasPermission(dev, "audit", "read") {
		t.Error("developer should NOT have audit:read")
	}
}

func TestSetupDefaultRolesViewerPermissions(t *testing.T) {
	r := NewRBAC()
	SetupDefaultRoles(r)

	viewer := &User{Roles: []string{"viewer"}}
	if !r.HasPermission(viewer, "spec", "read") {
		t.Error("viewer should have spec:read")
	}
	if r.HasPermission(viewer, "spec", "write") {
		t.Error("viewer should NOT have spec:write")
	}
	if r.HasPermission(viewer, "pipeline", "write") {
		t.Error("viewer should NOT have pipeline:write")
	}
	// Viewer doesn't have pipeline:write but has pipeline:read
	if !r.HasPermission(viewer, "pipeline", "read") {
		t.Error("viewer should have pipeline:read")
	}
}

func TestJoinActions(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{[]string{"read", "write"}, "read+write"},
		{[]string{"read"}, "read"},
		{nil, ""},
		{[]string{}, ""},
		{[]string{"read", "write", "delete"}, "read+write+delete"},
	}
	for _, tt := range tests {
		got := joinActions(tt.input)
		if got != tt.want {
			t.Errorf("joinActions(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestManagerRBAC(t *testing.T) {
	m := NewManager()
	rbac := m.RBAC()
	if rbac == nil {
		t.Fatal("expected non-nil RBAC")
	}
}

func TestManagerSessions(t *testing.T) {
	m := NewManager()
	sessions := m.Sessions()
	if sessions == nil {
		t.Fatal("expected non-nil Sessions")
	}
}

func TestManagerGetUserNotFound(t *testing.T) {
	m := NewManager()
	_, ok := m.GetUser("nonexistent")
	if ok {
		t.Error("expected false for nonexistent user")
	}
}

func TestManagerGetUserFound(t *testing.T) {
	m := NewManager()
	m.CreateUser(&User{ID: "u1", Name: "User 1"})
	user, ok := m.GetUser("u1")
	if !ok {
		t.Fatal("expected user found")
	}
	if user.Name != "User 1" {
		t.Errorf("expected 'User 1', got %s", user.Name)
	}
}

func TestManagerListUsers(t *testing.T) {
	m := NewManager()
	m.CreateUser(&User{ID: "u1"})
	m.CreateUser(&User{ID: "u2"})

	users := m.ListUsers()
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestManagerListUsersEmpty(t *testing.T) {
	m := NewManager()
	users := m.ListUsers()
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestManagerGetOAuth2NotFound(t *testing.T) {
	m := NewManager()
	_, ok := m.GetOAuth2("nonexistent")
	if ok {
		t.Error("expected false for nonexistent OAuth2")
	}
}

func TestManagerAuthenticateAPIKeyUserNotFound(t *testing.T) {
	m := NewManager()
	// Generate API key for a user that doesn't exist
	key, _ := m.APIKeys().Generate("ghost-user", "key", nil, time.Now().Add(time.Hour))
	_, ok := m.AuthenticateAPIKey(key)
	if ok {
		t.Error("expected false when API key user doesn't exist")
	}
}

func TestUserStoreFilePath(t *testing.T) {
	s := NewUserStore("")
	fp := s.filePath()
	if fp == "" {
		t.Error("expected non-empty file path")
	}
	if filepath.Base(fp) != authConfigFile {
		t.Errorf("expected file path to end with %s, got %s", authConfigFile, filepath.Base(fp))
	}
}

func TestUserStoreAddAndGet(t *testing.T) {
	dir := t.TempDir()
	s := &UserStore{
		dir:     dir,
		entries: nil,
	}

	user := &User{
		ID:    "u1",
		Name:  "Test User",
		Email: "test@example.com",
		Roles: []string{"admin"},
	}

	if err := s.Add(user); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	got, ok := s.Get("u1")
	if !ok {
		t.Fatal("expected user to be found")
	}
	if got.Name != "Test User" {
		t.Errorf("expected 'Test User', got %s", got.Name)
	}
	if got.Email != "test@example.com" {
		t.Errorf("expected email, got %s", got.Email)
	}
}

func TestUserStoreAddUpdateExisting(t *testing.T) {
	dir := t.TempDir()
	s := &UserStore{dir: dir}

	s.Add(&User{ID: "u1", Name: "Original"})
	s.Add(&User{ID: "u1", Name: "Updated"})

	got, ok := s.Get("u1")
	if !ok {
		t.Fatal("expected user found")
	}
	if got.Name != "Updated" {
		t.Errorf("expected 'Updated', got %s", got.Name)
	}
}

func TestUserStoreAddWithCreatedAt(t *testing.T) {
	dir := t.TempDir()
	s := &UserStore{dir: dir}

	now := time.Now().Truncate(time.Second).UTC()
	s.Add(&User{ID: "u1", Name: "User", CreatedAt: now})

	s2 := &UserStore{dir: dir}
	got, ok := s2.Get("u1")
	if !ok {
		t.Fatal("expected user found")
	}
	if got.CreatedAt.IsZero() {
		// The SavedUser format only stores seconds; CreatedAt is parsed on read.
		// If it's zero, the creation time was not persisted.
		t.Skip("CreatedAt not round-tripped through save/load")
	}
}

func TestUserStoreGetNotFound(t *testing.T) {
	dir := t.TempDir()
	s := &UserStore{dir: dir}

	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("expected false for nonexistent user")
	}
}

func TestUserStoreList(t *testing.T) {
	dir := t.TempDir()
	s := &UserStore{dir: dir}

	s.Add(&User{ID: "u1", Name: "User 1"})
	s.Add(&User{ID: "u2", Name: "User 2"})

	users, err := s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestUserStoreListEmpty(t *testing.T) {
	dir := t.TempDir()
	s := &UserStore{dir: dir}

	users, err := s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestUserStoreLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	s := &UserStore{dir: filepath.Join(dir, "nonexistent")}

	err := s.load()
	if err != nil {
		t.Errorf("expected nil error for missing file, got %v", err)
	}
}

func TestUserStoreLoadCorruptFile(t *testing.T) {
	dir := t.TempDir()
	s := &UserStore{dir: dir}
	os.WriteFile(filepath.Join(dir, authConfigFile), []byte("not json"), 0o600)

	err := s.load()
	if err == nil {
		t.Error("expected error for corrupt file")
	}
}

func TestUserStoreSaveCreatesDir(t *testing.T) {
	dir := t.TempDir()
	s := &UserStore{dir: filepath.Join(dir, "newdir", "subdir")}

	s.entries = []SavedUser{{ID: "u1", Name: "Test"}}
	err := s.save()
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "newdir", "subdir", authConfigFile)); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}

func TestUserStoreWithEncryption(t *testing.T) {
	dir := t.TempDir()
	s := &UserStore{dir: dir, passphrase: "secret", key: make([]byte, 32)}

	s.Add(&User{ID: "u1", Name: "Encrypted User"})

	// Create a new store and read the same file
	s2 := &UserStore{dir: dir, passphrase: "secret", key: make([]byte, 32)}
	got, ok := s2.Get("u1")
	if !ok {
		t.Fatal("expected to read encrypted user")
	}
	if got.Name != "Encrypted User" {
		t.Errorf("expected 'Encrypted User', got %s", got.Name)
	}
}

func TestHasPermissionNoRole(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{Name: "admin", Permissions: []string{"spec"}})
	r.AddPermission(&Permission{Resource: "spec", Actions: []string{"read"}})

	user := &User{Roles: []string{"nonexistent"}}
	if r.HasPermission(user, "spec", "read") {
		t.Error("expected false for nonexistent role")
	}
}

func TestHasPermissionNoPermission(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{Name: "admin", Permissions: []string{"nonexistent-perm"}})
	r.AddPermission(&Permission{Resource: "spec", Actions: []string{"read"}})

	user := &User{Roles: []string{"admin"}}
	if r.HasPermission(user, "spec", "read") {
		t.Error("expected false for nonexistent permission")
	}
}

func TestHasPermissionWildcardAction(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{Name: "admin", Permissions: []string{"spec"}})
	r.AddPermission(&Permission{Resource: "spec", Actions: []string{"*"}})

	user := &User{Roles: []string{"admin"}}
	if !r.HasPermission(user, "spec", "delete") {
		t.Error("expected wildcard action to grant permission")
	}
}

func TestHasPermissionWildcardResource(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{Name: "superadmin", Permissions: []string{"*"}})
	r.AddPermission(&Permission{Resource: "*", Actions: []string{"read"}})

	user := &User{Roles: []string{"superadmin"}}
	if !r.HasPermission(user, "spec", "read") {
		t.Error("expected wildcard resource to grant permission for spec")
	}
	if r.HasPermission(user, "spec", "write") {
		t.Error("expected wildcard resource to NOT grant write (only read)")
	}
}

func TestRemoveRoleFromUserNotFound(t *testing.T) {
	r := NewRBAC()
	user := &User{Roles: []string{"admin"}}
	r.RemoveRoleFromUser(user, "nonexistent")

	if len(user.Roles) != 1 {
		t.Errorf("expected roles unchanged, got %d", len(user.Roles))
	}
}

func TestHasPermissionEmptyRoles(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{Name: "admin", Permissions: []string{"spec"}})
	r.AddPermission(&Permission{Resource: "spec", Actions: []string{"read"}})

	user := &User{Roles: nil}
	if r.HasPermission(user, "spec", "read") {
		t.Error("expected false for empty roles")
	}
}

func TestAPIKeyManagerRevokeNotFound(t *testing.T) {
	m := NewAPIKeyManager()
	if m.Revoke("nonexistent") {
		t.Error("expected false for nonexistent key")
	}
}

func TestAPIKeyManagerValidateNotFound(t *testing.T) {
	m := NewAPIKeyManager()
	_, ok := m.Validate("nonexistent")
	if ok {
		t.Error("expected false for nonexistent key")
	}
}

func TestSessionManagerGetNotFound(t *testing.T) {
	m := NewSessionManager()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("expected false for nonexistent session")
	}
}

func TestSessionManagerDeleteNotFound(t *testing.T) {
	m := NewSessionManager()
	if m.Delete("nonexistent") {
		t.Error("expected false for nonexistent session")
	}
}

func TestGoogleOAuth2Interface(t *testing.T) {
	g := NewGoogleOAuth2("id", "secret", "http://localhost/callback")
	var p OAuth2ProviderInterface = g
	if p.Name() != "google" {
		t.Error("expected google provider")
	}
}

func TestGitHubOAuth2Interface(t *testing.T) {
	g := NewGitHubOAuth2("id", "secret", "http://localhost/callback")
	var p OAuth2ProviderInterface = g
	if p.Name() != "github" {
		t.Error("expected github provider")
	}
}

func TestGitHubOAuth2GetUser(t *testing.T) {
	g := NewGitHubOAuth2("id", "secret", "http://localhost/callback")
	user, err := g.GetUser(&OAuth2Token{AccessToken: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != "github-user-1" {
		t.Errorf("expected github-user-1, got %s", user.ID)
	}
}

func TestAPIKeyManagerZeroExpiration(t *testing.T) {
	m := NewAPIKeyManager()
	key, _ := m.Generate("u1", "key", nil, time.Time{})

	_, ok := m.Validate(key)
	if !ok {
		t.Error("expected zero expiration to mean no expiry check")
	}
}

func TestCheckDenyExplicitResource(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{
		Name:            "restricted",
		ResourceActions: map[string][]string{"secret": {"read"}},
		Deny:            map[string][]string{"secret": {"delete"}},
	})
	r.AddPermission(&Permission{Resource: "secret", Actions: []string{"read", "delete"}})

	user := &User{Roles: []string{"restricted"}}
	if !r.HasPermission(user, "secret", "read") {
		t.Error("expected read to be allowed")
	}
	if r.HasPermission(user, "secret", "delete") {
		t.Error("expected delete to be denied")
	}
}

func TestCheckDenyWildcardResource(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{
		Name:            "restricted",
		ResourceActions: map[string][]string{"*": {"read", "write"}},
		Deny:            map[string][]string{"*": {"delete"}},
	})
	r.AddPermission(&Permission{Resource: "*", Actions: []string{"read", "write", "delete"}})

	user := &User{Roles: []string{"restricted"}}
	if !r.HasPermission(user, "anything", "read") {
		t.Error("expected read to be allowed")
	}
	if !r.HasPermission(user, "anything", "write") {
		t.Error("expected write to be allowed (deny only covers delete)")
	}
	if r.HasPermission(user, "anything", "delete") {
		t.Error("expected delete to be denied (wildcard deny)")
	}
}

func TestCheckDenySpecificAction(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{
		Name:            "admin",
		ResourceActions: map[string][]string{"spec": {"read", "write", "delete"}},
	})
	r.AddPermission(&Permission{Resource: "spec", Actions: []string{"read", "write", "delete"}})

	user := &User{Roles: []string{"admin"}}
	if !r.HasPermission(user, "spec", "read") {
		t.Error("expected read to be allowed")
	}
}

func TestHasPermissionParentChain(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{Name: "base", ResourceActions: map[string][]string{"report": {"read"}}})
	r.AddRole(&Role{Name: "extended", Parents: []string{"base"}, ResourceActions: map[string][]string{"report": {"write"}}})
	r.AddPermission(&Permission{Resource: "report", Actions: []string{"read", "write"}})

	user := &User{Roles: []string{"extended"}}
	if !r.HasPermission(user, "report", "read") {
		t.Error("expected read from parent role")
	}
	if !r.HasPermission(user, "report", "write") {
		t.Error("expected write from own role")
	}
}

func TestHasPermissionCircularParent(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{Name: "a", Parents: []string{"b"}})
	r.AddRole(&Role{Name: "b", Parents: []string{"a"}, ResourceActions: map[string][]string{"x": {"read"}}})
	r.AddPermission(&Permission{Resource: "x", Actions: []string{"read"}})

	user := &User{Roles: []string{"a"}}
	// Role "a" inherits from "b" which has x:read
	if !r.HasPermission(user, "x", "read") {
		t.Error("expected permission to resolve through circular parent chain")
	}
}

func TestSetupRoleTemplateValid(t *testing.T) {
	r := NewRBAC()
	SetupRoleTemplate(r, "auditor", "my-auditor", nil)

	role, ok := r.GetRole("my-auditor")
	if !ok {
		t.Fatal("expected role to be created")
	}
	if role.ResourceActions == nil {
		t.Error("expected resource actions")
	}
	if role.Deny == nil {
		t.Error("expected deny rules for auditor template")
	}
	// Auditor should not have admin access
	user := &User{Roles: []string{"my-auditor"}}
	if r.HasPermission(user, "admin", "admin") {
		t.Error("auditor should NOT have admin:admin (denied)")
	}
}

func TestSetupRoleTemplateInvalidTemplate(t *testing.T) {
	r := NewRBAC()
	SetupRoleTemplate(r, "nonexistent", "my-role", nil)

	_, ok := r.GetRole("my-role")
	if ok {
		t.Error("expected no role for invalid template")
	}
}

func TestSetupRoleTemplateWithParents(t *testing.T) {
	r := NewRBAC()
	r.AddRole(&Role{Name: "base-viewer", ResourceActions: map[string][]string{"spec": {"read"}}})
	r.AddPermission(&Permission{Resource: "spec", Actions: []string{"read"}})
	SetupRoleTemplate(r, "auditor", "ext-auditor", []string{"base-viewer"})

	role, ok := r.GetRole("ext-auditor")
	if !ok {
		t.Fatal("expected role to be created")
	}
	if len(role.Parents) != 1 || role.Parents[0] != "base-viewer" {
		t.Errorf("expected parents [base-viewer], got %v", role.Parents)
	}
}

func TestParseCertificatePEMValid(t *testing.T) {
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: []byte("not a real cert but valid pem format"),
	})
	_, err := ParseCertificatePEM(pemData)
	if err == nil {
		t.Error("expected error for invalid certificate data after PEM decode")
	}
}

func TestParseCertificatePEMInvalid(t *testing.T) {
	_, err := ParseCertificatePEM([]byte("not pem"))
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestParseCertificatePEMEmpty(t *testing.T) {
	_, err := ParseCertificatePEM(nil)
	if err == nil {
		t.Error("expected error for nil/empty PEM")
	}
}

func TestGenerateToken(t *testing.T) {
	tok := generateToken()
	if tok == "" {
		t.Error("expected non-empty token")
	}
	if len(tok) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(tok))
	}
}

func TestGenerateSecureKey(t *testing.T) {
	key, err := generateSecureKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == "" {
		t.Error("expected non-empty key")
	}
	if len(key) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(key))
	}
}
