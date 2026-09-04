package signing

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// SpecVersion is the signing bundle format version.
const SpecVersion = "1.0"

// Algorithm identifies the signing algorithm used.
type Algorithm string

const (
	Ed25519 Algorithm = "Ed25519"
)

// Bundle is a self-contained signing document that binds an artifact hash
// to a digital signature. The bundle is designed for offline verification
// and is compatible with Cosign-style verification workflows.
type Bundle struct {
	SpecVersion string    `json:"specVersion"`
	Algorithm   Algorithm `json:"algorithm"`
	Artifact    Artifact  `json:"artifact"`
	Signature   string    `json:"signature"`
	PublicKey    string    `json:"publicKey"`
	SignedAt    string    `json:"signedAt"`
	Signer      string    `json:"signer,omitempty"`
}

// Artifact describes the signed artifact and its integrity hash.
type Artifact struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	Size int64  `json:"size,omitempty"`
	Path string `json:"path,omitempty"`
}

// KeyPair holds an Ed25519 key pair for signing and verification.
type KeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// GenerateKeyPair creates a new Ed25519 key pair.
func GenerateKeyPair() (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "generate key pair")
	}
	return &KeyPair{PublicKey: pub, PrivateKey: priv}, nil
}

// PublicKeyBase64 returns the public key encoded as base64.
func (kp *KeyPair) PublicKeyBase64() string {
	return base64.StdEncoding.EncodeToString(kp.PublicKey)
}

// PrivateKeyBase64 returns the private key encoded as base64.
func (kp *KeyPair) PrivateKeyBase64() string {
	return base64.StdEncoding.EncodeToString(kp.PrivateKey)
}

// ParsePublicKeyBase64 decodes a base64-encoded public key.
func ParsePublicKeyBase64(b64 string) (ed25519.PublicKey, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrValidation, "decode public key")
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, naeoserr.New(naeoserr.ErrValidation, "invalid public key length")
	}
	return ed25519.PublicKey(raw), nil
}

// ParsePrivateKeyBase64 decodes a base64-encoded private key.
func ParsePrivateKeyBase64(b64 string) (ed25519.PrivateKey, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrValidation, "decode private key")
	}
	if len(raw) != ed25519.PrivateKeySize {
		return nil, naeoserr.New(naeoserr.ErrValidation, "invalid private key length")
	}
	return ed25519.PrivateKey(raw), nil
}

// Hash computes the SHA-256 hex digest of data.
func Hash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// HashFile computes the SHA-256 hex digest of a file's contents.
func HashFile(path string) (string, int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", 0, naeoserr.Wrapf(err, naeoserr.ErrInternal, "read %s", path)
	}
	return Hash(data), int64(len(data)), nil
}

// Sign creates a signing bundle for the given artifact.
func Sign(name string, content []byte, kp *KeyPair, opts ...SignOption) (*Bundle, error) {
	bundle := &Bundle{
		SpecVersion: SpecVersion,
		Algorithm:   Ed25519,
		Artifact: Artifact{
			Name: name,
			Hash: Hash(content),
			Size: int64(len(content)),
		},
		SignedAt: time.Now().UTC().Format(time.RFC3339),
		PublicKey: kp.PublicKeyBase64(),
	}

	for _, opt := range opts {
		opt(bundle)
	}

	digest := sha256.Sum256([]byte(bundle.Artifact.Hash))
	bundle.Signature = base64.StdEncoding.EncodeToString(
		ed25519.Sign(kp.PrivateKey, digest[:]),
	)

	return bundle, nil
}

// SignOption configures a signing bundle.
type SignOption func(*Bundle)

// WithSigner sets the signer identity.
func WithSigner(name string) SignOption {
	return func(b *Bundle) { b.Signer = name }
}

// WithPath sets the artifact file path.
func WithPath(path string) SignOption {
	return func(b *Bundle) { b.Artifact.Path = path }
}

// WithSize sets the artifact size.
func WithSize(size int64) SignOption {
	return func(b *Bundle) { b.Artifact.Size = size }
}

// Verify checks that a signing bundle's signature is valid for the
// contained artifact hash and public key.
func Verify(bundle *Bundle) (*VerifyResult, error) {
	if bundle == nil {
		return nil, naeoserr.New(naeoserr.ErrValidation, "nil bundle")
	}

	if bundle.SpecVersion != SpecVersion {
		return &VerifyResult{
			Valid:   false,
			Message: fmt.Sprintf("unsupported spec version %s", bundle.SpecVersion),
		}, nil
	}

	pubKey, err := ParsePublicKeyBase64(bundle.PublicKey)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrValidation, "parse public key from bundle")
	}

	sig, err := base64.StdEncoding.DecodeString(bundle.Signature)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrValidation, "decode signature")
	}

	digest := sha256.Sum256([]byte(bundle.Artifact.Hash))
	valid := ed25519.Verify(pubKey, digest[:], sig)

	result := &VerifyResult{
		Valid:          valid,
		Algorithm:      bundle.Algorithm,
		ArtifactHash:   bundle.Artifact.Hash,
		ArtifactName:   bundle.Artifact.Name,
		PublicKey:      bundle.PublicKey,
		SignatureValid: valid,
	}

	if !valid {
		result.Message = "signature verification failed"
	} else {
		result.Message = "signature verified"
	}

	return result, nil
}

// VerifyResult is the outcome of signature verification.
type VerifyResult struct {
	Valid          bool      `json:"valid"`
	Message        string    `json:"message"`
	Algorithm      Algorithm `json:"algorithm"`
	ArtifactHash   string    `json:"artifactHash"`
	ArtifactName   string    `json:"artifactName"`
	PublicKey      string    `json:"publicKey"`
	SignatureValid bool      `json:"signatureValid"`
	Timestamp      time.Time `json:"timestamp"`
}

// VerifyFile loads a bundle from a JSON file and verifies it.
func VerifyFile(path string) (*VerifyResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "read bundle %s", path)
	}
	bundle := new(Bundle)
	if err := json.Unmarshal(data, bundle); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "parse bundle")
	}
	return Verify(bundle)
}

// Write persists a signing bundle to disk as indented JSON.
func Write(bundle *Bundle, path string) error {
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "marshal bundle")
	}
	if err := os.MkdirAll(".", 0o755); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "create dir")
	}
	return os.WriteFile(path, data, 0o600)
}

// Load reads a signing bundle from a JSON file.
func Load(path string) (*Bundle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "read %s", path)
	}
	bundle := new(Bundle)
	if err := json.Unmarshal(data, bundle); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "parse bundle")
	}
	return bundle, nil
}
