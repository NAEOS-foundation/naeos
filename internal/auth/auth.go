package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// User

type User struct {
	ID        string
	Email     string
	Name      string
	Roles     []string
	APIKeys   []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Role

type Role struct {
	Name        string
	Permissions []string
}

type Permission struct {
	Resource string
	Actions  []string
}

// RBAC

type RBAC struct {
	roles       map[string]*Role
	permissions map[string]*Permission
	mu          sync.RWMutex
}

func NewRBAC() *RBAC {
	return &RBAC{
		roles:       make(map[string]*Role),
		permissions: make(map[string]*Permission),
	}
}

func (r *RBAC) AddRole(role *Role) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.roles[role.Name] = role
}

func (r *RBAC) RemoveRole(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.roles, name)
}

func (r *RBAC) GetRole(name string) (*Role, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	role, ok := r.roles[name]
	return role, ok
}

func (r *RBAC) ListRoles() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.roles))
	for name := range r.roles {
		names = append(names, name)
	}
	return names
}

func (r *RBAC) AddPermission(perm *Permission) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.permissions[perm.Resource] = perm
}

func (r *RBAC) HasPermission(user *User, resource, action string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, roleName := range user.Roles {
		role, ok := r.roles[roleName]
		if !ok {
			continue
		}

		for _, permName := range role.Permissions {
			perm, ok := r.permissions[permName]
			if !ok {
				continue
			}

			if perm.Resource == "*" || perm.Resource == resource {
				for _, a := range perm.Actions {
					if a == "*" || a == action {
						return true
					}
				}
			}
		}
	}
	return false
}

func (r *RBAC) AssignRole(user *User, roleName string) {
	user.Roles = append(user.Roles, roleName)
}

func (r *RBAC) RemoveRoleFromUser(user *User, roleName string) {
	for i, role := range user.Roles {
		if role == roleName {
			user.Roles = append(user.Roles[:i], user.Roles[i+1:]...)
			return
		}
	}
}

// OAuth2

type OAuth2Provider struct {
	Name         string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type OAuth2Token struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	ExpiresAt    time.Time
	Scope        string
}

type OAuth2User struct {
	ID      string
	Email   string
	Name    string
	Picture string
}

type OAuth2ProviderInterface interface {
	Name() string
	GetAuthorizationURL(state string) string
	ExchangeCode(code string) (*OAuth2Token, error)
	GetUser(token *OAuth2Token) (*OAuth2User, error)
}

// Google OAuth2

type GoogleOAuth2 struct {
	Config *OAuth2Provider
}

func NewGoogleOAuth2(clientID, clientSecret, redirectURL string) *GoogleOAuth2 {
	return &GoogleOAuth2{
		Config: &OAuth2Provider{
			Name:         "google",
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
		},
	}
}

func (g *GoogleOAuth2) Name() string {
	return "google"
}

func (g *GoogleOAuth2) GetAuthorizationURL(state string) string {
	return fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		g.Config.ClientID,
		g.Config.RedirectURL,
		"openid email profile",
		state,
	)
}

func (g *GoogleOAuth2) ExchangeCode(code string) (*OAuth2Token, error) {
	// Simulated exchange
	return &OAuth2Token{
		AccessToken:  generateToken(),
		TokenType:    "Bearer",
		RefreshToken: generateToken(),
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil
}

func (g *GoogleOAuth2) GetUser(token *OAuth2Token) (*OAuth2User, error) {
	return &OAuth2User{
		ID:    "google-user-1",
		Email: "user@gmail.com",
		Name:  "Google User",
	}, nil
}

// GitHub OAuth2

type GitHubOAuth2 struct {
	Config *OAuth2Provider
}

func NewGitHubOAuth2(clientID, clientSecret, redirectURL string) *GitHubOAuth2 {
	return &GitHubOAuth2{
		Config: &OAuth2Provider{
			Name:         "github",
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"user:email"},
		},
	}
}

func (g *GitHubOAuth2) Name() string {
	return "github"
}

func (g *GitHubOAuth2) GetAuthorizationURL(state string) string {
	return fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		g.Config.ClientID,
		g.Config.RedirectURL,
		"user:email",
		state,
	)
}

func (g *GitHubOAuth2) ExchangeCode(code string) (*OAuth2Token, error) {
	return &OAuth2Token{
		AccessToken:  generateToken(),
		TokenType:    "Bearer",
		RefreshToken: generateToken(),
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil
}

func (g *GitHubOAuth2) GetUser(token *OAuth2Token) (*OAuth2User, error) {
	return &OAuth2User{
		ID:    "github-user-1",
		Email: "user@github.com",
		Name:  "GitHub User",
	}, nil
}

// API Key Manager

type APIKeyManager struct {
	keys map[string]*APIKey
	mu   sync.RWMutex
}

type APIKey struct {
	Key       string
	UserID    string
	Name      string
	Scopes    []string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewAPIKeyManager() *APIKeyManager {
	return &APIKeyManager{
		keys: make(map[string]*APIKey),
	}
}

func (m *APIKeyManager) Generate(userID, name string, scopes []string, expiresAt time.Time) (string, error) {
	key, err := generateSecureKey()
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.keys[key] = &APIKey{
		Key:       key,
		UserID:    userID,
		Name:      name,
		Scopes:    scopes,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	return key, nil
}

func (m *APIKeyManager) Validate(key string) (*APIKey, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	apiKey, ok := m.keys[key]
	if !ok {
		return nil, false
	}

	if !apiKey.ExpiresAt.IsZero() && time.Now().After(apiKey.ExpiresAt) {
		return nil, false
	}

	return apiKey, true
}

func (m *APIKeyManager) Revoke(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.keys[key]; ok {
		delete(m.keys, key)
		return true
	}
	return false
}

func (m *APIKeyManager) ListByUser(userID string) []*APIKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var keys []*APIKey
	for _, key := range m.keys {
		if key.UserID == userID {
			keys = append(keys, key)
		}
	}
	return keys
}

// Session Manager

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

type Session struct {
	ID        string
	UserID    string
	Data      map[string]any
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (m *SessionManager) Create(userID string, data map[string]any, expiresAt time.Time) (string, error) {
	id, err := generateSecureKey()
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[id] = &Session{
		ID:        id,
		UserID:    userID,
		Data:      data,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	return id, nil
}

func (m *SessionManager) Get(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[id]
	if !ok {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, false
	}

	return session, true
}

func (m *SessionManager) Delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[id]; ok {
		delete(m.sessions, id)
		return true
	}
	return false
}

func (m *SessionManager) Cleanup() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	removed := 0
	for id, session := range m.sessions {
		if time.Now().After(session.ExpiresAt) {
			delete(m.sessions, id)
			removed++
		}
	}
	return removed
}

// Auth Manager

type Manager struct {
	rbac     *RBAC
	apiKeys  *APIKeyManager
	sessions *SessionManager
	oauth2   map[string]OAuth2ProviderInterface
	users    map[string]*User
	mu       sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		rbac:     NewRBAC(),
		apiKeys:  NewAPIKeyManager(),
		sessions: NewSessionManager(),
		oauth2:   make(map[string]OAuth2ProviderInterface),
		users:    make(map[string]*User),
	}
}

func (m *Manager) RBAC() *RBAC {
	return m.rbac
}

func (m *Manager) APIKeys() *APIKeyManager {
	return m.apiKeys
}

func (m *Manager) Sessions() *SessionManager {
	return m.sessions
}

func (m *Manager) RegisterOAuth2(provider OAuth2ProviderInterface) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.oauth2[provider.Name()] = provider
}

func (m *Manager) GetOAuth2(name string) (OAuth2ProviderInterface, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	provider, ok := m.oauth2[name]
	return provider, ok
}

func (m *Manager) CreateUser(user *User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
}

func (m *Manager) GetUser(id string) (*User, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	user, ok := m.users[id]
	return user, ok
}

func (m *Manager) AuthenticateAPIKey(key string) (*User, bool) {
	apiKey, ok := m.apiKeys.Validate(key)
	if !ok {
		return nil, false
	}

	user, ok := m.GetUser(apiKey.UserID)
	if !ok {
		return nil, false
	}

	return user, true
}

func (m *Manager) ListUsers() []*User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	users := make([]*User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	return users
}

// Helpers

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateSecureKey() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]), nil
}
