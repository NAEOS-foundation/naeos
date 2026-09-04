package signing

import (
	"crypto/ed25519"
	"os"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	t.Parallel()
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	if len(kp.PublicKey) != ed25519.PublicKeySize {
		t.Errorf("unexpected public key size: %d", len(kp.PublicKey))
	}
	if len(kp.PrivateKey) != ed25519.PrivateKeySize {
		t.Errorf("unexpected private key size: %d", len(kp.PrivateKey))
	}
}

func TestKeyPairBase64Roundtrip(t *testing.T) {
	t.Parallel()
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	pubB64 := kp.PublicKeyBase64()
	privB64 := kp.PrivateKeyBase64()

	pub, err := ParsePublicKeyBase64(pubB64)
	if err != nil {
		t.Fatalf("ParsePublicKeyBase64: %v", err)
	}
	if len(pub) != ed25519.PublicKeySize {
		t.Error("public key roundtrip failed")
	}

	priv, err := ParsePrivateKeyBase64(privB64)
	if err != nil {
		t.Fatalf("ParsePrivateKeyBase64: %v", err)
	}
	if len(priv) != ed25519.PrivateKeySize {
		t.Error("private key roundtrip failed")
	}
}

func TestParsePublicKeyInvalid(t *testing.T) {
	t.Parallel()
	_, err := ParsePublicKeyBase64("not-base64!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestParsePublicKeyWrongLength(t *testing.T) {
	t.Parallel()
	_, err := ParsePublicKeyBase64("dGVzdA==") // "test"
	if err == nil {
		t.Error("expected error for wrong key length")
	}
}

func TestParsePrivateKeyInvalid(t *testing.T) {
	t.Parallel()
	_, err := ParsePrivateKeyBase64("not-base64!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestParsePrivateKeyWrongLength(t *testing.T) {
	t.Parallel()
	_, err := ParsePrivateKeyBase64("dGVzdA==")
	if err == nil {
		t.Error("expected error for wrong key length")
	}
}

func TestHash(t *testing.T) {
	t.Parallel()
	h := Hash([]byte("hello"))
	if len(h) != 64 {
		t.Errorf("expected 64-char hex hash, got %d", len(h))
	}
	// Deterministic.
	h2 := Hash([]byte("hello"))
	if h != h2 {
		t.Error("expected deterministic hash")
	}
}

func TestHashFile(t *testing.T) {
	t.Parallel()
	path := t.TempDir() + "/test.txt"
	if err := writeFile(path, []byte("content")); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	hash, size, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}
	if len(hash) != 64 {
		t.Errorf("expected 64-char hex hash, got %d", len(hash))
	}
	if size != 7 {
		t.Errorf("expected size 7, got %d", size)
	}
}

func TestHashFileNotExist(t *testing.T) {
	t.Parallel()
	_, _, err := HashFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestSignAndVerify(t *testing.T) {
	t.Parallel()
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	content := []byte("test artifact content")
	bundle, err := Sign("artifact.bin", content, kp, WithSigner("test-user"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if bundle.Algorithm != Ed25519 {
		t.Errorf("expected Ed25519 algorithm, got %s", bundle.Algorithm)
	}
	if bundle.Artifact.Hash != Hash(content) {
		t.Error("hash mismatch")
	}
	if bundle.Signer != "test-user" {
		t.Errorf("expected signer test-user, got %s", bundle.Signer)
	}

	result, err := Verify(bundle)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid signature, got: %s", result.Message)
	}
}

func TestVerifyTamperedHash(t *testing.T) {
	t.Parallel()
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	bundle, err := Sign("a", []byte("x"), kp)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	bundle.Artifact.Hash = "tampered"
	result, err := Verify(bundle)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid for tampered hash")
	}
}

func TestVerifyTamperedSignature(t *testing.T) {
	t.Parallel()
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	bundle, err := Sign("a", []byte("x"), kp)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	bundle.Signature = "tampered"
	result, err := Verify(bundle)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid for tampered signature")
	}
}

func TestVerifyWrongKey(t *testing.T) {
	t.Parallel()
	kp1, _ := GenerateKeyPair()
	kp2, _ := GenerateKeyPair()
	bundle, err := Sign("a", []byte("x"), kp1)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	bundle.PublicKey = kp2.PublicKeyBase64()
	result, err := Verify(bundle)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid for wrong key")
	}
}

func TestVerifyNilBundle(t *testing.T) {
	t.Parallel()
	_, err := Verify(nil)
	if err == nil {
		t.Error("expected error for nil bundle")
	}
}

func TestVerifyUnsupportedVersion(t *testing.T) {
	t.Parallel()
	kp, _ := GenerateKeyPair()
	bundle, _ := Sign("a", []byte("x"), kp)
	bundle.SpecVersion = "99.0"
	result, err := Verify(bundle)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid for unsupported version")
	}
}

func TestVerifyInvalidSignatureEncoding(t *testing.T) {
	t.Parallel()
	kp, _ := GenerateKeyPair()
	bundle, _ := Sign("a", []byte("x"), kp)
	bundle.Signature = string([]byte{0xff, 0xfe, 0xfd, 0xfc, 0xfb})
	_, err := Verify(bundle)
	if err == nil {
		t.Error("expected error for invalid signature encoding")
	}
}

func TestSignOptions(t *testing.T) {
	t.Parallel()
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	bundle, err := Sign("a", []byte("x"), kp,
		WithSigner("alice"),
		WithPath("/a.bin"),
		WithSize(100),
	)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if bundle.Signer != "alice" {
		t.Errorf("expected signer alice, got %s", bundle.Signer)
	}
	if bundle.Artifact.Path != "/a.bin" {
		t.Errorf("expected path /a.bin, got %s", bundle.Artifact.Path)
	}
	if bundle.Artifact.Size != 100 {
		t.Errorf("expected size 100, got %d", bundle.Artifact.Size)
	}
}

func TestBundleJSONRoundtrip(t *testing.T) {
	t.Parallel()
	kp, _ := GenerateKeyPair()
	bundle, err := Sign("roundtrip", []byte("data"), kp, WithSigner("tester"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	data, err := encodeJSON(bundle)
	if err != nil {
		t.Fatalf("encodeJSON: %v", err)
	}

	var loaded Bundle
	if err := decodeJSON(data, &loaded); err != nil {
		t.Fatalf("decodeJSON: %v", err)
	}

	result, err := Verify(&loaded)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !result.Valid {
		t.Error("expected valid after roundtrip")
	}
}

func TestVerifyFile(t *testing.T) {
	t.Parallel()
	kp, _ := GenerateKeyPair()
	bundle, _ := Sign("file-test", []byte("content"), kp)
	path := t.TempDir() + "/sig.json"
	if err := Write(bundle, path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	result, err := VerifyFile(path)
	if err != nil {
		t.Fatalf("VerifyFile: %v", err)
	}
	if !result.Valid {
		t.Error("expected valid from file")
	}
}

func TestVerifyFileNotExist(t *testing.T) {
	t.Parallel()
	_, err := VerifyFile("/nonexistent/sig.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadBundle(t *testing.T) {
	t.Parallel()
	kp, _ := GenerateKeyPair()
	bundle, _ := Sign("load-test", []byte("x"), kp)
	path := t.TempDir() + "/bundle.json"
	Write(bundle, path)
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Artifact.Name != "load-test" {
		t.Errorf("expected name load-test, got %s", loaded.Artifact.Name)
	}
}

func TestLoadBundleNotExist(t *testing.T) {
	t.Parallel()
	_, err := Load("/nonexistent/bundle.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestConcurrentSignVerify(t *testing.T) {
	kp, _ := GenerateKeyPair()
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			content := []byte("concurrent-content")
			bundle, err := Sign("c", content, kp)
			if err != nil {
				t.Errorf("Sign: %v", err)
			}
			result, err := Verify(bundle)
			if err != nil {
				t.Errorf("Verify: %v", err)
			}
			if !result.Valid {
				t.Error("expected valid in concurrent test")
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}
