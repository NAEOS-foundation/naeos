package interop

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gowebpki/jcs"
)

const candidateDigestV1Prefix = "NAEOS-CANDIDATE-DIGEST-V1\n"

type candidateDigestVectorFile struct {
	Profile     string                  `json:"profile"`
	Canonicalization string             `json:"canonicalization"`
	Hash        string                  `json:"hash"`
	DomainPrefix string                 `json:"domain_prefix_utf8"`
	Vectors     []candidateDigestVector `json:"vectors"`
}

type candidateDigestVector struct {
	ID                string          `json:"id"`
	Candidate         json.RawMessage `json:"candidate"`
	CanonicalJSONUTF8 string          `json:"canonical_json_utf8"`
	ExpectedDigestHex string          `json:"expected_digest_hex"`
}

func TestCandidateDigestV1GoldenVectors(t *testing.T) {
	path := filepath.Join("handoff-candidate-digest-v1-vectors.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden vectors: %v", err)
	}

	var file candidateDigestVectorFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("parse golden vectors: %v", err)
	}
	if file.Profile != "naeos-candidate-jcs-sha256-v1" {
		t.Fatalf("unexpected profile: %q", file.Profile)
	}
	if file.Canonicalization != "RFC 8785 JCS" || file.Hash != "SHA-256" {
		t.Fatalf("unexpected canonicalization/hash: %q / %q", file.Canonicalization, file.Hash)
	}
	if file.DomainPrefix != candidateDigestV1Prefix {
		t.Fatalf("domain prefix mismatch: got %q", file.DomainPrefix)
	}
	if len(file.Vectors) == 0 {
		t.Fatal("no golden vectors defined")
	}

	for _, vector := range file.Vectors {
		t.Run(vector.ID, func(t *testing.T) {
			canonical, err := jcs.Transform(vector.Candidate)
			if err != nil {
				t.Fatalf("JCS transform: %v", err)
			}
			if string(canonical) != vector.CanonicalJSONUTF8 {
				t.Fatalf("canonical bytes mismatch\n got: %s\nwant: %s", canonical, vector.CanonicalJSONUTF8)
			}

			h := sha256.New()
			_, _ = h.Write([]byte(candidateDigestV1Prefix))
			_, _ = h.Write(canonical)
			got := hex.EncodeToString(h.Sum(nil))
			if got != vector.ExpectedDigestHex {
				t.Fatalf("digest mismatch: got %s, want %s", got, vector.ExpectedDigestHex)
			}
			if len(got) != 64 || strings.ToLower(got) != got {
				t.Fatalf("digest is not 64 lowercase hex chars: %q", got)
			}
		})
	}
}

func TestCandidateDigestV1DomainSeparation(t *testing.T) {
	canonical := []byte(`{"input":{"path":"README.md"},"tool":"read"}`)
	withPrefix := sha256.Sum256(append([]byte(candidateDigestV1Prefix), canonical...))
	withoutPrefix := sha256.Sum256(canonical)
	if fmt.Sprintf("%x", withPrefix[:]) == fmt.Sprintf("%x", withoutPrefix[:]) {
		t.Fatal("domain separation prefix did not affect digest")
	}
}
