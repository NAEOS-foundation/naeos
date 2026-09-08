package sbom

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// NewSerialNumber returns a random RFC 4122 v4 UUID suitable for use as a
// CycloneDX BOM serial number. The format is urn:uuid:<uuid>.
func NewSerialNumber() string {
	var buf [16]byte
	_, _ = rand.Read(buf[:])
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("urn:uuid:%08x-%04x-%04x-%04x-%012x",
		buf[:4], buf[4:6], buf[6:8], buf[8:10], buf[10:])
}

// Timestamp returns the current UTC time in RFC 3339 format suitable
// for CycloneDX metadata timestamps.
func Timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// RandomHex returns n random bytes as a hex string.
func RandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
