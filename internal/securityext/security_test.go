package securityext

import (
	"fmt"
	"strings"
	"testing"
)

func TestSecretManager(t *testing.T) {
	sm := NewSecretManager("test-key-123")

	err := sm.Set("db-password", "secret123")
	if err != nil {
		t.Fatalf("failed to set secret: %v", err)
	}

	val, ok := sm.Get("db-password")
	if !ok {
		t.Fatal("expected secret to exist")
	}
	if val != "secret123" {
		t.Errorf("expected 'secret123', got %s", val)
	}
}

func TestSecretManagerUpdate(t *testing.T) {
	sm := NewSecretManager("test-key-123")

	sm.Set("key", "val1")
	sm.Set("key", "val2")

	val, _ := sm.Get("key")
	if val != "val2" {
		t.Errorf("expected 'val2', got %s", val)
	}
}

func TestSecretManagerDelete(t *testing.T) {
	sm := NewSecretManager("test-key-123")

	sm.Set("key", "val")
	if !sm.Delete("key") {
		t.Error("expected true")
	}
	if sm.Delete("key") {
		t.Error("expected false")
	}
}

func TestSecretManagerList(t *testing.T) {
	sm := NewSecretManager("test-key-123")

	sm.Set("a", "1")
	sm.Set("b", "2")

	names := sm.List()
	if len(names) != 2 {
		t.Errorf("expected 2, got %d", len(names))
	}
}

func TestSecretManagerExists(t *testing.T) {
	sm := NewSecretManager("test-key-123")

	sm.Set("key", "val")
	if !sm.Exists("key") {
		t.Error("expected true")
	}
	if sm.Exists("missing") {
		t.Error("expected false")
	}
}

func TestSanitizer(t *testing.T) {
	s := NewSanitizer()

	html := s.SanitizeHTML("<b>bold</b>")
	if strings.Contains(html, "<b>") {
		t.Error("expected HTML tags removed")
	}

	xss := s.SanitizeXSS("<script>alert('xss')</script>")
	if strings.Contains(xss, "<script>") {
		t.Error("expected script tags removed")
	}

	path := s.SanitizePath("../../etc/passwd")
	if strings.Contains(path, "../") {
		t.Error("expected path traversal removed")
	}

	if !s.ValidateEmail("test@example.com") {
		t.Error("expected valid email")
	}
	if s.ValidateEmail("invalid") {
		t.Error("expected invalid email")
	}
}

func TestSanitizerSanitizeAll(t *testing.T) {
	s := NewSanitizer()

	result := s.SanitizeAll("<script>alert('xss')</script>")
	if strings.Contains(result, "<script>") {
		t.Error("expected script removed")
	}
}

func TestHashPassword(t *testing.T) {
	hash := HashPassword("password123")
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if !VerifyPassword("password123", hash) {
		t.Error("expected password to verify")
	}
	if VerifyPassword("wrong", hash) {
		t.Error("expected wrong password to fail")
	}
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(32)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(token) == 0 {
		t.Error("expected non-empty token")
	}
}

func TestValidator(t *testing.T) {
	v := NewValidator()

	v.AddRule("name", RequiredRule)
	v.AddRule("email", func(val string) error {
		if !strings.Contains(val, "@") {
			return fmt.Errorf("invalid email: must contain @")
		}
		if !strings.Contains(val, ".") {
			return fmt.Errorf("invalid email: must contain domain")
		}
		return nil
	})

	if err := v.Validate("name", "John"); err != nil {
		t.Error("expected no error")
	}
	if err := v.Validate("name", ""); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestValidatorMinMaxLength(t *testing.T) {
	v := NewValidator()

	v.AddRule("short", MinLengthRule(3))
	v.AddRule("long", MaxLengthRule(5))

	if err := v.Validate("short", "ab"); err == nil {
		t.Error("expected error")
	}
	if err := v.Validate("short", "abc"); err != nil {
		t.Error("expected no error")
	}
	if err := v.Validate("long", "abcdef"); err == nil {
		t.Error("expected error")
	}
	if err := v.Validate("long", "abc"); err != nil {
		t.Error("expected no error")
	}
}

func TestValidatorValidateAll(t *testing.T) {
	v := NewValidator()

	v.AddRule("name", RequiredRule)
	v.AddRule("email", RequiredRule)

	errors := v.ValidateAll(map[string]string{"name": ""})
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
}
