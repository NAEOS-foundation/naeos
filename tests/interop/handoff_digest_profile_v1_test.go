package interop

import (
	"bytes"
	"testing"

	"github.com/gowebpki/jcs"
)

func TestCandidateDigestV1CanonicalizesRFC8785Properties(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"sorts-recursive-object-keys", `{"z":1,"a":{"y":2,"b":3}}`, `{"a":{"b":3,"y":2},"z":1}`},
		{"preserves-array-order", `{"v":[3,1,2]}`, `{"v":[3,1,2]}`},
		{"normalizes-number-rendering", `{"n":4.50,"m":2e-3}`, `{"m":0.002,"n":4.5}`},
		{"does-not-unicode-normalize", `{"v":"café"}`, `{"v":"café"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := jcs.Transform([]byte(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, []byte(tc.want)) {
				t.Fatalf("canonical mismatch: got %s want %s", got, tc.want)
			}
		})
	}
}

func TestCandidateDigestV1RejectsDuplicateKeysAtEveryDepth(t *testing.T) {
	for _, raw := range []string{
		`{"a":1,"a":2}`,
		`{"outer":{"x":1,"x":2}}`,
		`{"outer":{"x":1},"outer":{"x":2}}`,
	} {
		if _, err := jcs.Transform([]byte(raw)); err == nil {
			t.Fatalf("duplicate keys accepted: %s", raw)
		}
	}
}
