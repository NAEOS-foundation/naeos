package interop

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/gowebpki/jcs"
)

const candidateDigestPrefix = "NAEOS-CANDIDATE-DIGEST-V1\n"

type digestVectorFixture struct {
	Profile          string `json:"profile"`
	DomainPrefixUTF8 string `json:"domain_prefix_utf8"`
	Vectors          []struct {
		ID                string          `json:"id"`
		Candidate         json.RawMessage `json:"candidate"`
		CanonicalJSONUTF8 string          `json:"canonical_json_utf8"`
		ExpectedDigestHex string          `json:"expected_digest_hex"`
	} `json:"vectors"`
}

// candidateDigestV1 fails closed: it returns no digest unless the exact
// candidate envelope can be parsed and canonicalized using RFC 8785 JCS.
func candidateDigestV1(raw []byte) ([]byte, string, error) {
	canonical, err := jcs.Transform(raw)
	if err != nil {
		return nil, "", fmt.Errorf("canonicalize candidate: %w", err)
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &envelope); err != nil {
		return nil, "", fmt.Errorf("candidate must be a JSON object: %w", err)
	}
	if len(envelope) != 2 {
		return nil, "", fmt.Errorf("candidate envelope must contain exactly tool and input")
	}
	toolRaw, ok := envelope["tool"]
	if !ok {
		return nil, "", fmt.Errorf("candidate tool is required")
	}
	inputRaw, ok := envelope["input"]
	if !ok || len(inputRaw) == 0 {
		return nil, "", fmt.Errorf("candidate input is required")
	}
	var tool string
	if err := json.Unmarshal(toolRaw, &tool); err != nil || tool == "" {
		return nil, "", fmt.Errorf("candidate tool must be a non-empty string")
	}
	preimage := append([]byte(candidateDigestPrefix), canonical...)
	sum := sha256.Sum256(preimage)
	return canonical, hex.EncodeToString(sum[:]), nil
}

func TestCandidateDigestV1GoldenVectors(t *testing.T) {
	raw, err := os.ReadFile("handoff-candidate-digest-v1-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture digestVectorFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode vectors: %v", err)
	}
	if fixture.Profile != "naeos-candidate-jcs-sha256-v1" {
		t.Fatalf("unexpected profile: %q", fixture.Profile)
	}
	if fixture.DomainPrefixUTF8 != candidateDigestPrefix {
		t.Fatalf("unexpected prefix: %q", fixture.DomainPrefixUTF8)
	}
	if len(fixture.Vectors) == 0 {
		t.Fatal("no golden vectors")
	}
	for _, vector := range fixture.Vectors {
		t.Run(vector.ID, func(t *testing.T) {
			canonical, digest, err := candidateDigestV1(vector.Candidate)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(canonical, []byte(vector.CanonicalJSONUTF8)) {
				t.Fatalf("canonical bytes mismatch: got %s want %s", canonical, vector.CanonicalJSONUTF8)
			}
			if digest != vector.ExpectedDigestHex {
				t.Fatalf("digest mismatch: got %s want %s", digest, vector.ExpectedDigestHex)
			}
		})
	}
}

func TestCandidateDigestV1KeyOrderInvariantArrayOrderSensitive(t *testing.T) {
	a := []byte(`{"tool":"update","input":{"a":1,"b":2,"labels":["x","y"]}}`)
	b := []byte(`{"input":{"labels":["x","y"],"b":2,"a":1},"tool":"update"}`)
	c := []byte(`{"input":{"labels":["y","x"],"b":2,"a":1},"tool":"update"}`)
	_, da, err := candidateDigestV1(a)
	if err != nil {
		t.Fatal(err)
	}
	_, db, err := candidateDigestV1(b)
	if err != nil {
		t.Fatal(err)
	}
	_, dc, err := candidateDigestV1(c)
	if err != nil {
		t.Fatal(err)
	}
	if da != db {
		t.Fatalf("object key order changed digest: %s != %s", da, db)
	}
	if da == dc {
		t.Fatal("array element order must affect digest")
	}
}

func TestCandidateDigestV1RejectsInvalidCandidates(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"not-object", `[]`},
		{"missing-tool", `{"input":{}}`},
		{"empty-tool", `{"tool":"","input":{}}`},
		{"non-string-tool", `{"tool":42,"input":{}}`},
		{"missing-input", `{"tool":"read"}`},
		{"extra-member", `{"tool":"read","input":{},"subject":"x"}`},
		{"duplicate-top-level-key", `{"tool":"read","tool":"write","input":{}}`},
		{"duplicate-nested-key", `{"tool":"read","input":{"x":1,"x":2}}`},
		{"non-json-number", `{"tool":"read","input":{"n":NaN}}`},
		{"out-of-range-number", `{"tool":"read","input":{"n":1e9999}}`},
		{"invalid-surrogate", `{"tool":"read","input":{"s":"\ud800"}}`},
		{"trailing-data", `{"tool":"read","input":{}} true`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := candidateDigestV1([]byte(tc.raw)); err == nil {
				t.Fatalf("candidate unexpectedly accepted: %s", tc.raw)
			}
		})
	}
}
