package securityext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileSecretManagerSetGet(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	fsm, err := NewFileSecretManager("test-key")
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := fsm.Set("api-key", "sk-123"); err != nil {
		t.Fatalf("set: %v", err)
	}
	val, ok := fsm.Get("api-key")
	if !ok {
		t.Fatal("expected secret to exist")
	}
	if val != "sk-123" {
		t.Errorf("expected sk-123, got %s", val)
	}
}

func TestFileSecretManagerDelete(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	fsm, _ := NewFileSecretManager("test-key")
	fsm.Set("key", "val")
	if !fsm.Delete("key") {
		t.Error("expected true")
	}
	if fsm.Exists("key") {
		t.Error("expected key to be deleted")
	}
}

func TestFileSecretManagerList(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	fsm, _ := NewFileSecretManager("test-key")
	fsm.Set("a", "1")
	fsm.Set("b", "2")
	names := fsm.List()
	if len(names) != 2 {
		t.Errorf("expected 2, got %d", len(names))
	}
}

func TestFileSecretManagerPersistence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	fsm1, _ := NewFileSecretManager("test-key")
	fsm1.Set("persist-key", "persist-val")

	fsm2, err := NewFileSecretManager("test-key")
	if err != nil {
		t.Fatalf("second new: %v", err)
	}
	val, ok := fsm2.Get("persist-key")
	if !ok {
		t.Fatal("expected persisted key to exist")
	}
	if val != "persist-val" {
		t.Errorf("expected persist-val, got %s", val)
	}
}

func TestSecretManagerGetNonExistent(t *testing.T) {
	sm := NewSecretManager("key")
	_, ok := sm.Get("missing")
	if ok {
		t.Error("expected false")
	}
}

func TestValidateFilePathValid(t *testing.T) {
	base := t.TempDir()
	subDir := filepath.Join(base, "sub")
	os.MkdirAll(subDir, 0o755)

	absPath, err := ValidateFilePath(subDir, base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(absPath, "/sub") && !strings.HasSuffix(absPath, "\\sub") {
		t.Errorf("unexpected path: %s", absPath)
	}
}

func TestValidateFilePathTraversal(t *testing.T) {
	base := t.TempDir()
	_, err := ValidateFilePath("/etc/passwd", base)
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestValidateFilePathInvalidBase(t *testing.T) {
	_, err := ValidateFilePath("clean", "/nonexistent_base_xyz")
	if err == nil {
		t.Error("expected error for invalid base")
	}
}

func TestValidatePluginNameValid(t *testing.T) {
	if err := ValidatePluginName("my-plugin"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidatePluginNameEmpty(t *testing.T) {
	if err := ValidatePluginName(""); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestValidatePluginNameWithSlash(t *testing.T) {
	if err := ValidatePluginName("a/b"); err == nil {
		t.Error("expected error for path separator")
	}
}

func TestValidatePluginNameWithDotDot(t *testing.T) {
	if err := ValidatePluginName(".."); err == nil {
		t.Error("expected error for relative path")
	}
}

func TestValidatePluginNameCleanMismatch(t *testing.T) {
	if err := ValidatePluginName("./foo"); err == nil {
		t.Error("expected error for unclean name")
	}
}

func TestEncryptDecryptConfig(t *testing.T) {
	plaintext := []byte("sensitive config data")
	passphrase := "my-passphrase"

	encrypted, err := EncryptConfig(plaintext, passphrase)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if encrypted == "" {
		t.Fatal("expected non-empty encrypted output")
	}

	decrypted, err := DecryptConfig(encrypted, passphrase)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("expected %s, got %s", plaintext, decrypted)
	}
}

func TestDecryptConfigWrongPassphrase(t *testing.T) {
	plaintext := []byte("data")
	encrypted, _ := EncryptConfig(plaintext, "correct")
	_, err := DecryptConfig(encrypted, "wrong")
	if err == nil {
		t.Error("expected error for wrong passphrase")
	}
}

func TestDecryptConfigInvalidBase64(t *testing.T) {
	_, err := DecryptConfig("not-base64!!!", "key")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestDecryptConfigShortCiphertext(t *testing.T) {
	_, err := DecryptConfig("aGVsbG8=", "key")
	if err == nil {
		t.Error("expected error for short ciphertext")
	}
}

func TestPatternRuleValid(t *testing.T) {
	rule := PatternRule(`^\d{3}-\d{2}-\d{4}$`)
	if err := rule("123-45-6789"); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestPatternRuleInvalid(t *testing.T) {
	rule := PatternRule(`^\d{3}-\d{2}-\d{4}$`)
	if err := rule("abc"); err == nil {
		t.Error("expected error for non-matching value")
	}
}

func TestPatternRuleInvalidPattern(t *testing.T) {
	rule := PatternRule(`[invalid`)
	if err := rule("test"); err == nil {
		t.Error("expected error for invalid pattern")
	}
}

func TestValidatorRuleNotFound(t *testing.T) {
	v := NewValidator()
	err := v.Validate("nonexistent-rule", "value")
	if err == nil {
		t.Error("expected error")
	}
}

func TestValidatorValidateAllNoErrors(t *testing.T) {
	v := NewValidator()
	v.AddRule("name", func(s string) error { return nil })
	errors := v.ValidateAll(map[string]string{"name": "test"})
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}
}

func TestRequiredRuleEmpty(t *testing.T) {
	if err := RequiredRule(""); err == nil {
		t.Error("expected error for empty")
	}
}

func TestRequiredRuleNonEmpty(t *testing.T) {
	if err := RequiredRule("value"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMinLengthRule(t *testing.T) {
	rule := MinLengthRule(3)
	if err := rule("ab"); err == nil {
		t.Error("expected error for short value")
	}
	if err := rule("abc"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMaxLengthRule(t *testing.T) {
	rule := MaxLengthRule(5)
	if err := rule("abcdef"); err == nil {
		t.Error("expected error for long value")
	}
	if err := rule("abc"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewSecretManagerDifferentKeys(t *testing.T) {
	sm1 := NewSecretManager("key1")
	sm2 := NewSecretManager("key2")

	sm1.Set("secret", "value")
	val, ok := sm2.Get("secret")
	if ok {
		t.Error("should not be able to decrypt with different key")
	}
	if val != "" {
		t.Errorf("expected empty, got %s", val)
	}
}

func TestSanitizerSQL(t *testing.T) {
	s := NewSanitizer()
	result := s.SanitizeSQL("Robert'); DROP TABLE Students;--")
	if strings.Contains(result, "'") {
		t.Error("expected SQL injection chars removed")
	}
}

func TestSanitizerPathNoOp(t *testing.T) {
	s := NewSanitizer()
	result := s.SanitizePath("safe/path/file.txt")
	if result != "safe/path/file.txt" {
		t.Errorf("expected unchanged, got %s", result)
	}
}
