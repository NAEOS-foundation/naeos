package interop

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCandidateDigestV1GoldenVectors(t *testing.T) {
	data, err := os.ReadFile("handoff-candidate-digest-v1-vectors.json")
	if err != nil { t.Fatal(err) }
	var fixture struct {
		Profile string `json:"profile"`
		DomainPrefix string `json:"domain_prefix_utf8"`
		Vectors []struct {
			ID string `json:"id"`
			Candidate json.RawMessage `json:"candidate"`
			Canonical string `json:"canonical_json_utf8"`
			Expected string `json:"expected_digest_hex"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil { t.Fatal(err) }
	if fixture.Profile != CandidateDigestProfileV1() { t.Fatalf("profile mismatch: %s", fixture.Profile) }
	if fixture.DomainPrefix != candidateDigestDomainV1 { t.Fatalf("domain prefix mismatch: %q", fixture.DomainPrefix) }
	if len(fixture.Vectors) == 0 { t.Fatal("no golden vectors") }
	for _, vector := range fixture.Vectors {
		t.Run(vector.ID, func(t *testing.T) {
			got, canonical, err := CandidateDigestJSON(vector.Candidate)
			if err != nil { t.Fatal(err) }
			if string(canonical) != vector.Canonical {
				t.Fatalf("canonical bytes mismatch\n got: %s\nwant: %s", canonical, vector.Canonical)
			}
			if got != vector.Expected {
				t.Fatalf("digest mismatch: got %s want %s", got, vector.Expected)
			}
		})
	}
}

func TestCandidateDigestV1RejectsMalformedCandidates(t *testing.T) {
	tests := []struct{name, raw, wantErr string}{
		{"empty tool", `{"tool":"","input":{}}`, "tool must be a non-empty string"},
		{"missing input", `{"tool":"read"}`, "exactly tool and input"},
		{"unknown envelope field", `{"tool":"read","input":{},"extra":true}`, "exactly tool and input"},
		{"duplicate top-level key", `{"tool":"read","tool":"write","input":{}}`, "duplicate object key"},
		{"duplicate nested key", `{"tool":"read","input":{"x":1,"x":2}}`, "duplicate object key"},
		{"trailing JSON", `{"tool":"read","input":{}} {}`, "trailing JSON"},
		{"invalid JSON", `{"tool":"read","input":`, "invalid candidate JSON"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := CandidateDigestJSON([]byte(tc.raw))
			if err == nil || !containsText(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func containsText(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub { return true }
	}
	return false
}
