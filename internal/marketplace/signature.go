package marketplace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type SignatureVerifier interface {
	Verify(data []byte, signature []byte, publicKey []byte) bool
}

type SHA256Verifier struct{}

func (v *SHA256Verifier) Verify(data []byte, signature []byte, publicKey []byte) bool {
	h := sha256.Sum256(data)
	expected := hex.EncodeToString(signature)
	return hex.EncodeToString(h[:]) == expected
}

func VerifyPlugin(pluginPath, expectedChecksum string) error {
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrPlugin, "read plugin file")
	}

	h := sha256.Sum256(data)
	actual := hex.EncodeToString(h[:])

	if actual != expectedChecksum {
		return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("checksum mismatch: expected %s, got %s", expectedChecksum, actual))
	}

	return nil
}

func GenerateChecksum(pluginPath string) (string, error) {
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		return "", naeoserr.Wrapf(err, naeoserr.ErrPlugin, "read plugin file")
	}

	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}
