package securityext

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Secret Manager

type Secret struct {
	Name      string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SecretManager struct {
	secrets map[string]*Secret
	key     []byte
	mu      sync.RWMutex
}

func NewSecretManager(encryptionKey string) *SecretManager {
	hash := sha256.Sum256([]byte(encryptionKey))
	return &SecretManager{
		secrets: make(map[string]*Secret),
		key:     hash[:],
	}
}

func (sm *SecretManager) Set(name, value string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	encrypted, err := sm.encrypt(value)
	if err != nil {
		return err
	}

	now := time.Now()
	if existing, ok := sm.secrets[name]; ok {
		existing.Value = encrypted
		existing.UpdatedAt = now
	} else {
		sm.secrets[name] = &Secret{
			Name:      name,
			Value:     encrypted,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	return nil
}

func (sm *SecretManager) Get(name string) (string, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	secret, ok := sm.secrets[name]
	if !ok {
		return "", false
	}

	decrypted, err := sm.decrypt(secret.Value)
	if err != nil {
		return "", false
	}
	return decrypted, true
}

func (sm *SecretManager) Delete(name string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.secrets[name]; ok {
		delete(sm.secrets, name)
		return true
	}
	return false
}

func (sm *SecretManager) List() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	names := make([]string, 0, len(sm.secrets))
	for name := range sm.secrets {
		names = append(names, name)
	}
	return names
}

func (sm *SecretManager) Exists(name string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, ok := sm.secrets[name]
	return ok
}

// FileSecretManager wraps SecretManager with file-backed persistence.

type FileSecretManager struct {
	sm       *SecretManager
	filePath string
	mu       sync.Mutex
}

func NewFileSecretManager(key string) (*FileSecretManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".config", "naeos")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	fp := filepath.Join(dir, "secrets.enc")
	fsm := &FileSecretManager{
		sm:       NewSecretManager(key),
		filePath: fp,
	}
	fsm.load()
	return fsm, nil
}

type fileEntry struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (fsm *FileSecretManager) load() {
	data, err := os.ReadFile(fsm.filePath)
	if err != nil {
		return
	}
	var entries []fileEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return
	}
	fsm.sm.mu.Lock()
	defer fsm.sm.mu.Unlock()
	for _, e := range entries {
		createdAt, _ := time.Parse(time.RFC3339Nano, e.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339Nano, e.UpdatedAt)
		fsm.sm.secrets[e.Name] = &Secret{
			Name:      e.Name,
			Value:     e.Value,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
	}
}

func (fsm *FileSecretManager) Save() error {
	fsm.sm.mu.RLock()
	defer fsm.sm.mu.RUnlock()
	entries := make([]fileEntry, 0, len(fsm.sm.secrets))
	for _, s := range fsm.sm.secrets {
		entries = append(entries, fileEntry{
			Name:      s.Name,
			Value:     s.Value,
			CreatedAt: s.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt: s.UpdatedAt.Format(time.RFC3339Nano),
		})
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("marshal secrets: %w", err)
	}
	if err := os.WriteFile(fsm.filePath, data, 0o600); err != nil {
		return fmt.Errorf("write secrets file: %w", err)
	}
	return nil
}

func (fsm *FileSecretManager) Set(name, value string) error {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()
	if err := fsm.sm.Set(name, value); err != nil {
		return err
	}
	return fsm.Save()
}

func (fsm *FileSecretManager) Get(name string) (string, bool) {
	return fsm.sm.Get(name)
}

func (fsm *FileSecretManager) Delete(name string) bool {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()
	ok := fsm.sm.Delete(name)
	if ok {
		_ = fsm.Save()
	}
	return ok
}

func (fsm *FileSecretManager) List() []string {
	return fsm.sm.List()
}

func (fsm *FileSecretManager) Exists(name string) bool {
	return fsm.sm.Exists(name)
}

func (sm *SecretManager) encrypt(value string) (string, error) {
	block, err := aes.NewCipher(sm.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(value), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (sm *SecretManager) decrypt(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(sm.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// Input Sanitizer

type Sanitizer struct {
	patterns map[string]*regexp.Regexp
	mu       sync.RWMutex
}

func NewSanitizer() *Sanitizer {
	s := &Sanitizer{
		patterns: make(map[string]*regexp.Regexp),
	}

	s.patterns["html"] = regexp.MustCompile(`<[^>]*>`)
	s.patterns["sql"] = regexp.MustCompile(`['";\\]`)
	s.patterns["xss"] = regexp.MustCompile(`<script[^>]*>.*?</script>`)
	s.patterns["path"] = regexp.MustCompile(`\.\./`)
	s.patterns["email"] = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return s
}

func (s *Sanitizer) SanitizeHTML(input string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.patterns["html"].ReplaceAllString(input, "")
}

func (s *Sanitizer) SanitizeSQL(input string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.patterns["sql"].ReplaceAllString(input, "")
}

func (s *Sanitizer) SanitizeXSS(input string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.patterns["xss"].ReplaceAllString(input, "")
}

func (s *Sanitizer) SanitizePath(input string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.patterns["path"].ReplaceAllString(input, "")
}

func (s *Sanitizer) ValidateEmail(email string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.patterns["email"].MatchString(email)
}

func (s *Sanitizer) SanitizeAll(input string) string {
	result := input
	result = s.SanitizeHTML(result)
	result = s.SanitizeXSS(result)
	result = s.SanitizePath(result)
	return result
}

// ValidateFilePath checks that path stays within allowedBase.
// Returns the cleaned absolute path if valid, or an error if traversal is detected.
func ValidateFilePath(path, allowedBase string) (string, error) {
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	absBase, err := filepath.Abs(allowedBase)
	if err != nil {
		return "", fmt.Errorf("resolve base: %w", err)
	}
	if !strings.HasPrefix(absPath, absBase+string(os.PathSeparator)) && absPath != absBase {
		return "", fmt.Errorf("path traversal detected: %s", path)
	}
	return absPath, nil
}

// ValidatePluginName checks that a name is safe to use as a filesystem directory name.
// It rejects empty names, names with path separators or relative components.
func ValidatePluginName(name string) error {
	if name == "" {
		return fmt.Errorf("name must not be empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("name must not contain path separators")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("name must not contain relative path components")
	}
	clean := filepath.Clean(name)
	if clean != name {
		return fmt.Errorf("name must be a simple name without path components")
	}
	return nil
}

// Hash

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash), nil
}

func VerifyPassword(password, hash string) bool {
	decoded, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword(decoded, []byte(password)) == nil
}

// Token Generator

func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Validator

type Validator struct {
	rules map[string]func(string) error
	mu    sync.RWMutex
}

func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string]func(string) error),
	}
}

func (v *Validator) AddRule(name string, rule func(string) error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.rules[name] = rule
}

func (v *Validator) Validate(name, value string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	rule, ok := v.rules[name]
	if !ok {
		return fmt.Errorf("rule not found: %s", name)
	}
	return rule(value)
}

func (v *Validator) ValidateAll(values map[string]string) []error {
	var errors []error

	for name, value := range values {
		if err := v.Validate(name, value); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// Default Validator Rules

func RequiredRule(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func MinLengthRule(min int) func(string) error {
	return func(value string) error {
		if len(value) < min {
			return fmt.Errorf("value must be at least %d characters", min)
		}
		return nil
	}
}

func MaxLengthRule(max int) func(string) error {
	return func(value string) error {
		if len(value) > max {
			return fmt.Errorf("value must be at most %d characters", max)
		}
		return nil
	}
}

// EncryptConfig encrypts raw config bytes using AES-256-GCM with the given passphrase.
// Returns base64-encoded ciphertext.
func EncryptConfig(plaintext []byte, passphrase string) (string, error) {
	key := sha256.Sum256([]byte(passphrase))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptConfig decrypts base64-encoded ciphertext using AES-256-GCM with the given passphrase.
func DecryptConfig(encrypted string, passphrase string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}

	key := sha256.Sum256([]byte(passphrase))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func PatternRule(pattern string) func(string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return func(string) error {
			return fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}
	}
	return func(value string) error {
		if !re.MatchString(value) {
			return fmt.Errorf("value does not match pattern")
		}
		return nil
	}
}
